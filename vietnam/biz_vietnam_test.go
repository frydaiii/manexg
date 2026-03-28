package vietnam

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/utils"
)

func loadJSONFixture(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	path := filepath.Join("testdata", name)
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	obj := map[string]interface{}{}
	if err = utils.Unmarshal(text, &obj, utils.JsonNumDefault); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return obj
}

func TestSymbolNormalizationParsing(t *testing.T) {
	symbol := makeMarketSymbol("hose", "ssi")
	if symbol != "HOSE:SSI/VND" {
		t.Fatalf("unexpected symbol: %s", symbol)
	}
	b, tk := splitRawMarketID("HNX:SHS")
	if b != "HNX" || tk != "SHS" {
		t.Fatalf("unexpected split: %s %s", b, tk)
	}
	b, tk = splitRawMarketID("SSI")
	if b != "" || tk != "" {
		t.Fatalf("invalid raw id should not split")
	}
}

func TestMarketLookupAmbiguousTicker(t *testing.T) {
	v := &Vietnam{Exchange: &banexg.Exchange{ExgInfo: &banexg.ExgInfo{Markets: banexg.MarketMap{}}}}
	hose := newStockMarket("HOSE", "SSI", map[string]interface{}{"symbol": "SSI"})
	hnx := newStockMarket("HNX", "SSI", map[string]interface{}{"symbol": "SSI"})
	v.Markets[hose.Symbol] = hose
	v.Markets[hnx.Symbol] = hnx
	v.marketsByRawID = map[string]*banexg.Market{hose.ID: hose, hnx.ID: hnx}
	v.marketsByTicker = map[string][]*banexg.Market{"SSI": {hose, hnx}}

	if mar, err := v.MapMarket("HOSE:SSI", 0); err != nil || mar.ID != "HOSE:SSI" {
		t.Fatalf("expected direct board lookup, got market=%v err=%v", mar, err)
	}
	if _, err := v.MapMarket("SSI", 0); err == nil {
		t.Fatalf("expected ambiguous ticker error")
	}
}

func TestOHLCVTimestampConversionAndOrdering(t *testing.T) {
	intra := loadJSONFixture(t, "intraday_ohlc.json")
	rows := toMapRows(t, intra["data"])
	out := parseSSIKlines(rows, false)
	if len(out) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(out))
	}
	if out[0].Time >= out[1].Time {
		t.Fatalf("expected ascending order, got %d >= %d", out[0].Time, out[1].Time)
	}

	daily := loadJSONFixture(t, "daily_ohlc.json")
	drows := toMapRows(t, daily["data"])
	dout := parseSSIKlines(drows, true)
	if len(dout) != 2 {
		t.Fatalf("expected 2 daily rows, got %d", len(dout))
	}
	if dout[0].Time >= dout[1].Time {
		t.Fatalf("expected ascending daily order, got %d >= %d", dout[0].Time, dout[1].Time)
	}
}

func TestTokenFixtureJWTExpiry(t *testing.T) {
	obj := loadJSONFixture(t, "token_response.json")
	data := mapStringAny(obj["data"])
	token := anyString(data["accessToken"])
	ms := parseJWTExpMS(token)
	if ms <= 0 {
		t.Fatalf("expected positive jwt expiry, got %d", ms)
	}
}

func toMapRows(t *testing.T, value interface{}) []map[string]interface{} {
	t.Helper()
	arr, ok := value.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", value)
	}
	rows := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map row, got %T", item)
		}
		rows = append(rows, m)
	}
	return rows
}

func newVietnamForTest(t *testing.T) *Vietnam {
	t.Helper()
	v, err := New(map[string]interface{}{
		"ConsumerID":     "cid",
		"ConsumerSecret": "csec",
	})
	if err != nil {
		t.Fatalf("new vietnam: %v", err)
	}
	return v
}

func TestEntryApiMethodsAndBoards(t *testing.T) {
	v := newVietnamForTest(t)
	if v.Apis[MethodPublicPostMarketSecurities].Method != "GET" {
		t.Fatalf("securities should be GET")
	}
	if v.Apis[MethodPublicPostMarketSecuritiesInfo].Method != "GET" {
		t.Fatalf("securities details should be GET")
	}
	if v.Apis[MethodPublicGetMarketDailyOhlc] == nil {
		t.Fatalf("daily ohlc endpoint missing")
	}
	if v.Apis[MethodPublicGetMarketIndexList] == nil || v.Apis[MethodPublicGetMarketIndexComponents] == nil || v.Apis[MethodPublicGetMarketDailyIndex] == nil {
		t.Fatalf("index endpoints missing")
	}
	foundDER := false
	for _, b := range marketBoards {
		if b == "DER" {
			foundDER = true
			break
		}
	}
	if !foundDER {
		t.Fatalf("expected DER board in marketBoards")
	}
}

func TestMakeSignGetAndPostBehavior(t *testing.T) {
	v := newVietnamForTest(t)
	sign := makeSign(v)

	getApi := &banexg.Entry{Url: "https://example.test/api", Method: "GET", More: map[string]interface{}{"auth": false}}
	getReq := sign(getApi, map[string]interface{}{"Symbol": "SSI", "PageIndex": 1})
	if getReq.Error != nil {
		t.Fatalf("unexpected get sign error: %v", getReq.Error)
	}
	if !strings.Contains(getReq.Url, "Symbol=SSI") || !strings.Contains(getReq.Url, "PageIndex=1") {
		t.Fatalf("expected query params in url, got: %s", getReq.Url)
	}
	if getReq.Body != "" {
		t.Fatalf("expected empty GET body, got: %q", getReq.Body)
	}
	if getReq.Headers.Get("Content-Type") != "" {
		t.Fatalf("GET should not set content-type")
	}

	postApi := &banexg.Entry{Url: "https://example.test/api", Method: "POST", More: map[string]interface{}{"auth": false}}
	postReq := sign(postApi, map[string]interface{}{"Symbol": "SSI"})
	if postReq.Error != nil {
		t.Fatalf("unexpected post sign error: %v", postReq.Error)
	}
	if postReq.Body == "" || !strings.Contains(postReq.Body, "SSI") {
		t.Fatalf("expected JSON POST body, got: %q", postReq.Body)
	}
	if postReq.Headers.Get("Content-Type") != "application/json" {
		t.Fatalf("POST should set content-type json")
	}
}

func TestGetAccessTokenTopLevelAndCache(t *testing.T) {
	v := newVietnamForTest(t)
	count := 0
	token := "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjQxMDAwMDAwMDB9.signature"
	setVietnamTestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
		if endpoint != MethodPublicPostMarketAccessToken {
			return &banexg.HttpRes{Error: errs.NewMsg(errs.CodeParamInvalid, "unexpected endpoint: %s", endpoint)}
		}
		count++
		return &banexg.HttpRes{Content: "{\"responseCode\":0,\"token\":\"" + token + "\"}"}
	})

	got1, err := v.getAccessToken()
	if err != nil {
		t.Fatalf("first getAccessToken failed: %v", err)
	}
	got2, err := v.getAccessToken()
	if err != nil {
		t.Fatalf("second getAccessToken failed: %v", err)
	}
	if got1 != token || got2 != token {
		t.Fatalf("unexpected token values: %q %q", got1, got2)
	}
	if count != 1 {
		t.Fatalf("expected one token request due cache, got %d", count)
	}
}

func TestRequestSSIDataListAndDataFallback(t *testing.T) {
	t.Run("dataList", func(t *testing.T) {
		v := newVietnamForTest(t)
		setVietnamTestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
			return &banexg.HttpRes{Content: "{\"status\":\"SUCCESS\",\"dataList\":[{\"symbol\":\"SSI\"}]}"}
		})
		rows, err := requestSSI[[]map[string]interface{}](v, MethodPublicGetMarketDailyOhlc, map[string]interface{}{}, 1)
		if err != nil {
			t.Fatalf("requestSSI dataList failed: %v", err)
		}
		if len(rows) != 1 || anyString(rows[0]["symbol"]) != "SSI" {
			t.Fatalf("unexpected rows from dataList: %#v", rows)
		}
	})

	t.Run("data fallback", func(t *testing.T) {
		v := newVietnamForTest(t)
		setVietnamTestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
			return &banexg.HttpRes{Content: "{\"status\":\"SUCCESS\",\"data\":[{\"symbol\":\"SSI\"}]}"}
		})
		rows, err := requestSSI[[]map[string]interface{}](v, MethodPublicPostMarketSecurities, map[string]interface{}{}, 1)
		if err != nil {
			t.Fatalf("requestSSI data fallback failed: %v", err)
		}
		if len(rows) != 1 || anyString(rows[0]["symbol"]) != "SSI" {
			t.Fatalf("unexpected rows from data fallback: %#v", rows)
		}
	})
}

func TestFetchSecuritiesDetailRowsPayloadAndFlatten(t *testing.T) {
	v := newVietnamForTest(t)
	setVietnamTestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
		if endpoint != MethodPublicPostMarketSecuritiesInfo {
			return &banexg.HttpRes{Error: errs.NewMsg(errs.CodeParamInvalid, "unexpected endpoint: %s", endpoint)}
		}
		if anyString(params["Market"]) != "HOSE" {
			return &banexg.HttpRes{Error: errs.NewMsg(errs.CodeParamInvalid, "missing Market param")}
		}
		if v, ok := params["lookupRequest.pageSize"]; !ok || v != 1000 {
			return &banexg.HttpRes{Error: errs.NewMsg(errs.CodeParamInvalid, "missing lookupRequest.pageSize")}
		}
		return &banexg.HttpRes{Content: "{\"status\":\"SUCCESS\",\"dataList\":[{\"RePeAtEdInFoLiSt\":[{\"StockSymbol\":\"SSI\"},{\"ticker\":\"VND\"}]}]}"}
	})
	rows, err := v.fetchSecuritiesDetailRows("HOSE")
	if err != nil {
		t.Fatalf("fetchSecuritiesDetailRows failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 flattened rows, got %d", len(rows))
	}
	if parseTicker(rows[0]) != "SSI" || parseTicker(rows[1]) != "VND" {
		t.Fatalf("unexpected flattened ticker rows: %#v", rows)
	}
}

func TestFetchOHLCVPayloadsUseUpdatedEndpointAndAscending(t *testing.T) {
	v := newVietnamForTest(t)
	seen := map[string]map[string]interface{}{}
	setVietnamTestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
		cp := map[string]interface{}{}
		for k, val := range params {
			cp[k] = val
		}
		seen[endpoint] = cp
		if endpoint == MethodPublicGetMarketDailyOhlc {
			return &banexg.HttpRes{Content: "{\"status\":\"SUCCESS\",\"dataList\":[{\"TradingDate\":\"15/02/2026\",\"OpenPrice\":\"40\",\"HighestPrice\":\"41\",\"LowestPrice\":\"39\",\"ClosePrice\":\"40\"}]}"}
		}
		if endpoint == MethodPublicPostMarketIntradayOHLC {
			return &banexg.HttpRes{Content: "{\"status\":\"SUCCESS\",\"dataList\":[{\"TradingDate\":\"15/02/2026\",\"Time\":\"09:00:00\",\"Open\":\"40\",\"High\":\"41\",\"Low\":\"39\",\"Close\":\"40\"}]}"}
		}
		return &banexg.HttpRes{Error: errs.NewMsg(errs.CodeParamInvalid, "unexpected endpoint: %s", endpoint)}
	})

	if _, err := v.fetchDailyOHLCV("SSI", 1700000000000, 1700003600000, 10); err != nil {
		t.Fatalf("fetchDailyOHLCV failed: %v", err)
	}
	if _, err := v.fetchIntradayOHLCV("SSI", "1P", 1700000000000, 1700003600000, 10); err != nil {
		t.Fatalf("fetchIntradayOHLCV failed: %v", err)
	}

	dailyPayload, ok := seen[MethodPublicGetMarketDailyOhlc]
	if !ok {
		t.Fatalf("daily endpoint was not used")
	}
	if anyString(dailyPayload["Symbol"]) != "SSI" || dailyPayload["ascending"] != true {
		t.Fatalf("unexpected daily payload: %#v", dailyPayload)
	}

	intraPayload, ok := seen[MethodPublicPostMarketIntradayOHLC]
	if !ok {
		t.Fatalf("intraday endpoint was not used")
	}
	if anyString(intraPayload["Symbol"]) != "SSI" || intraPayload["ascending"] != true || anyString(intraPayload["resolution"]) != "1P" {
		t.Fatalf("unexpected intraday payload: %#v", intraPayload)
	}
}

func TestParseTickerAndSanitizePayload(t *testing.T) {
	if parseTicker(map[string]interface{}{"StockSymbol": "ssi"}) != "SSI" {
		t.Fatalf("expected parseTicker from StockSymbol")
	}
	if parseTicker(map[string]interface{}{"Code": "abc"}) != "ABC" {
		t.Fatalf("expected parseTicker from Code")
	}
	safe := sanitizePayload(map[string]interface{}{
		"apiToken":       "abc",
		"consumerSecret": "xyz",
		"password":       "p",
		"Symbol":         "SSI",
	})
	if anyString(safe["apiToken"]) != "***" || anyString(safe["consumerSecret"]) != "***" || anyString(safe["password"]) != "***" {
		t.Fatalf("expected sensitive keys to be masked: %#v", safe)
	}
	if anyString(safe["Symbol"]) != "SSI" {
		t.Fatalf("non-sensitive keys should remain unchanged: %#v", safe)
	}
}

func TestRequestSSIRateLimitRetry(t *testing.T) {
	v := newVietnamForTest(t)
	callCount := 0
	setVietnamTestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
		callCount++
		if callCount <= 2 {
			// First two calls return quota exceeded
			return &banexg.HttpRes{Content: `{"status":"RateLimit","message":"API calls quota exceeded! maximum admitted 1 per 1s."}`}
		}
		// Third call succeeds
		return &banexg.HttpRes{Content: `{"status":"SUCCESS","dataList":[{"symbol":"SSI"}]}`}
	})
	rows, err := requestSSI[[]map[string]interface{}](v, MethodPublicGetMarketDailyOhlc, map[string]interface{}{}, 1)
	if err != nil {
		t.Fatalf("requestSSI should have retried and succeeded: %v", err)
	}
	if len(rows) != 1 || anyString(rows[0]["symbol"]) != "SSI" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
	if callCount != 3 {
		t.Fatalf("expected 3 calls (2 retries + 1 success), got %d", callCount)
	}
}

func TestRequestSSIRateLimitExhausted(t *testing.T) {
	v := newVietnamForTest(t)
	callCount := 0
	setVietnamTestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
		callCount++
		return &banexg.HttpRes{Content: `{"status":"RateLimit","message":"API calls quota exceeded! maximum admitted 1 per 1s."}`}
	})
	_, err := requestSSI[[]map[string]interface{}](v, MethodPublicGetMarketDailyOhlc, map[string]interface{}{}, 1)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// 1 initial + 3 retries = 4 total calls
	if callCount != 4 {
		t.Fatalf("expected 4 calls (1 initial + 3 retries), got %d", callCount)
	}
}
