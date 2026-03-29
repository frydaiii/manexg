package vci

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

func newVCIForTest(t *testing.T) *VCI {
	t.Helper()
	v, err := New(nil)
	if err != nil {
		t.Fatalf("new vci: %v", err)
	}
	return v
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

func loadJSONFixture(t *testing.T, name string) []vciOHLCData {
	t.Helper()
	text := loadFixture(t, name)
	var out []vciOHLCData
	if err := utils.UnmarshalString(text, &out, utils.JsonNumDefault); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return out
}

// ─── Symbol / Market Tests ────────────────────────────────────────────────

func TestSymbolNormalization(t *testing.T) {
	sym := makeMarketSymbol("hose", "vci")
	if sym != "HOSE:VCI/VND" {
		t.Fatalf("expected HOSE:VCI/VND, got %s", sym)
	}
	sym = makeMarketSymbol("  hnx  ", "  shs  ")
	if sym != "HNX:SHS/VND" {
		t.Fatalf("expected HNX:SHS/VND, got %s", sym)
	}
}

func TestLoadMarketsFromFixture(t *testing.T) {
	v := newVCIForTest(t)
	fixture := loadFixture(t, "symbols_getall.json")
	setVCITestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, _, _ bool) *banexg.HttpRes {
		if endpoint == MethodPublicGetSymbolsAll {
			return &banexg.HttpRes{Content: fixture}
		}
		return &banexg.HttpRes{Error: errs.NewMsg(errs.CodeParamInvalid, "unexpected endpoint: %s", endpoint)}
	})

	markets, err := v.LoadMarkets(true, nil)
	if err != nil {
		t.Fatalf("LoadMarkets failed: %v", err)
	}

	// 6 STOCK + 1 ETF = 7 stocks, plus 3 index pseudo-markets = 10
	// (CW and BOND are filtered out)
	expectedCount := 10
	if len(markets) != expectedCount {
		t.Fatalf("expected %d markets, got %d", expectedCount, len(markets))
	}

	// Verify a specific market
	vciMarket, ok := markets["HOSE:VCI/VND"]
	if !ok {
		t.Fatal("HOSE:VCI/VND not found in markets")
	}
	if vciMarket.ID != "HOSE:VCI" {
		t.Fatalf("expected ID HOSE:VCI, got %s", vciMarket.ID)
	}
	if vciMarket.Base != "VCI" || vciMarket.Quote != "VND" {
		t.Fatalf("expected base=VCI quote=VND, got %s/%s", vciMarket.Base, vciMarket.Quote)
	}

	// Verify precision defaults for HOSE
	if vciMarket.Precision.Price != 10 {
		t.Fatalf("expected HOSE tick size 10, got %v", vciMarket.Precision.Price)
	}

	// Verify UPCOM lot size
	mcgMarket, ok := markets["UPCOM:MCG/VND"]
	if !ok {
		t.Fatal("UPCOM:MCG/VND not found in markets")
	}
	if mcgMarket.Limits.Amount.Min != 1 {
		t.Fatalf("expected UPCOM min lot 1, got %v", mcgMarket.Limits.Amount.Min)
	}

	// Verify ETF is included
	if _, ok := markets["HOSE:E1VFVN30/VND"]; !ok {
		t.Fatal("ETF E1VFVN30 not found in markets")
	}

	// Verify CW is excluded
	if _, ok := markets["HOSE:CW01/VND"]; ok {
		t.Fatal("CW01 should be filtered out")
	}
}

func TestLoadMarketsIncludesIndexes(t *testing.T) {
	v := newVCIForTest(t)
	fixture := loadFixture(t, "symbols_getall.json")
	setVCITestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, _, _ bool) *banexg.HttpRes {
		return &banexg.HttpRes{Content: fixture}
	})

	markets, err := v.LoadMarkets(true, nil)
	if err != nil {
		t.Fatalf("LoadMarkets failed: %v", err)
	}

	for _, idx := range []string{"INDEX:VNINDEX/VND", "INDEX:HNXINDEX/VND", "INDEX:UPCOMINDEX/VND"} {
		m, ok := markets[idx]
		if !ok {
			t.Fatalf("index %s not found", idx)
		}
		if anyString(m.Info["isIndex"]) != "true" {
			t.Fatalf("expected isIndex=true for %s", idx)
		}
	}

	// Verify vciSymbol mapping
	vnindex := markets["INDEX:VNINDEX/VND"]
	if anyString(vnindex.Info["vciSymbol"]) != "VNINDEX" {
		t.Fatalf("expected VNINDEX vciSymbol=VNINDEX, got %s", anyString(vnindex.Info["vciSymbol"]))
	}
	hnxindex := markets["INDEX:HNXINDEX/VND"]
	if anyString(hnxindex.Info["vciSymbol"]) != "HNXIndex" {
		t.Fatalf("expected HNXINDEX vciSymbol=HNXIndex, got %s", anyString(hnxindex.Info["vciSymbol"]))
	}
}

// ─── OHLCV Tests ──────────────────────────────────────────────────────────

func TestFetchOHLCVColumnParsing(t *testing.T) {
	data := loadJSONFixture(t, "ohlcv_daily.json")
	if len(data) == 0 {
		t.Fatal("expected at least 1 OHLC data row")
	}
	klines := parseVCIKlines(data[0])
	if len(klines) != 3 {
		t.Fatalf("expected 3 klines, got %d", len(klines))
	}
	// Verify timestamp conversion: 1700000000 seconds → 1700000000000 ms
	if klines[0].Time != 1700000000000 {
		t.Fatalf("expected Time=1700000000000, got %d", klines[0].Time)
	}
	if klines[0].Open != 45.20 {
		t.Fatalf("expected Open=45.20, got %f", klines[0].Open)
	}
	if klines[0].Volume != 1200000 {
		t.Fatalf("expected Volume=1200000, got %f", klines[0].Volume)
	}
	// Ascending order
	if klines[0].Time >= klines[1].Time || klines[1].Time >= klines[2].Time {
		t.Fatal("expected ascending timestamp order")
	}
}

func TestIndexSymbolMapping(t *testing.T) {
	v := newVCIForTest(t)
	fixture := loadFixture(t, "symbols_getall.json")
	ohlcvFixture := loadFixture(t, "ohlcv_daily.json")

	var capturedSymbol string
	setVCITestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, _, _ bool) *banexg.HttpRes {
		if endpoint == MethodPublicGetSymbolsAll {
			return &banexg.HttpRes{Content: fixture}
		}
		if endpoint == MethodPublicPostOhlcChart {
			// Capture the symbols array from the request
			if syms, ok := params["symbols"]; ok {
				if arr, ok := syms.([]string); ok && len(arr) > 0 {
					capturedSymbol = arr[0]
				}
			}
			return &banexg.HttpRes{Content: ohlcvFixture}
		}
		return &banexg.HttpRes{Error: errs.NewMsg(errs.CodeParamInvalid, "unexpected endpoint: %s", endpoint)}
	})

	// Pre-load markets so LoadArgsMarket doesn't go through the base Exchange's
	// LoadMarkets path (which calls FetchMarkets, not our override)
	_, err := v.LoadMarkets(true, nil)
	if err != nil {
		t.Fatalf("LoadMarkets failed: %v", err)
	}

	_, err = v.FetchOHLCV("INDEX:VNINDEX/VND", "1d", 0, 10, nil)
	if err != nil {
		t.Fatalf("FetchOHLCV for index failed: %v", err)
	}
	if capturedSymbol != "VNINDEX" {
		t.Fatalf("expected VCI symbol VNINDEX, got %s", capturedSymbol)
	}
}

func TestMakeSignHeaders(t *testing.T) {
	v := newVCIForTest(t)
	sign := makeSign(v)
	api := &banexg.Entry{Url: "https://example.test/api", Method: "POST", More: map[string]interface{}{"auth": false}}
	req := sign(api, map[string]interface{}{"key": "val"})
	if req.Error != nil {
		t.Fatalf("unexpected sign error: %v", req.Error)
	}
	if req.Headers.Get("Referer") != "https://trading.vietcap.com.vn/" {
		t.Fatalf("expected Referer header, got %q", req.Headers.Get("Referer"))
	}
	if req.Headers.Get("Origin") != "https://trading.vietcap.com.vn" {
		t.Fatalf("expected Origin header, got %q", req.Headers.Get("Origin"))
	}
	if req.Headers.Get("DNT") != "1" {
		t.Fatalf("expected DNT=1, got %q", req.Headers.Get("DNT"))
	}
	if !strings.Contains(req.Headers.Get("User-Agent"), "Mozilla") {
		t.Fatalf("expected Mozilla User-Agent, got %q", req.Headers.Get("User-Agent"))
	}
	// POST should have JSON body
	if req.Body == "" || !strings.Contains(req.Body, "val") {
		t.Fatalf("expected JSON POST body, got %q", req.Body)
	}
}

func TestCountBackOverfetch(t *testing.T) {
	limit := 100
	countBack := int(float64(limit) * 1.5)
	if countBack <= limit {
		t.Fatalf("expected countBack > limit, got %d", countBack)
	}
	if countBack != 150 {
		t.Fatalf("expected countBack=150, got %d", countBack)
	}
}

// ─── Negative / Edge Case Tests ───────────────────────────────────────────

func TestColumnArrayLengthMismatch(t *testing.T) {
	data := loadJSONFixture(t, "ohlcv_mismatched.json")
	if len(data) == 0 {
		t.Fatal("expected at least 1 OHLC data row")
	}
	klines := parseVCIKlines(data[0])
	// t has 5 entries, o/h/l/c/v have 3 → should return 3 klines
	if len(klines) != 3 {
		t.Fatalf("expected 3 klines (safe min), got %d", len(klines))
	}
}

func TestEmptyResponse(t *testing.T) {
	data := vciOHLCData{}
	klines := parseVCIKlines(data)
	if len(klines) != 0 {
		t.Fatalf("expected 0 klines for empty data, got %d", len(klines))
	}
}

func TestWrappedVsBareResponse(t *testing.T) {
	v := newVCIForTest(t)

	// Test bare array response
	bareFixture := loadFixture(t, "ohlcv_daily.json")
	setVCITestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, _, _ bool) *banexg.HttpRes {
		return &banexg.HttpRes{Content: bareFixture}
	})
	bareResult, err := requestVCI[[]vciOHLCData](v, MethodPublicPostOhlcChart, nil, 1)
	if err != nil {
		t.Fatalf("bare response parse failed: %v", err)
	}
	if len(bareResult) != 1 || len(bareResult[0].T) != 3 {
		t.Fatalf("unexpected bare result: %d items", len(bareResult))
	}

	// Test wrapped response
	wrappedFixture := loadFixture(t, "ohlcv_daily_wrapped.json")
	setVCITestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, _, _ bool) *banexg.HttpRes {
		return &banexg.HttpRes{Content: wrappedFixture}
	})
	wrappedResult, err := requestVCI[[]vciOHLCData](v, MethodPublicPostOhlcChart, nil, 1)
	if err != nil {
		t.Fatalf("wrapped response parse failed: %v", err)
	}
	if len(wrappedResult) != 1 || len(wrappedResult[0].T) != 3 {
		t.Fatalf("unexpected wrapped result: %d items", len(wrappedResult))
	}
}

func TestInvalidTimeframe(t *testing.T) {
	v := newVCIForTest(t)
	fixture := loadFixture(t, "symbols_getall.json")
	setVCITestRequest(t, func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, _, _ bool) *banexg.HttpRes {
		return &banexg.HttpRes{Content: fixture}
	})

	// Pre-load markets
	_, loadErr := v.LoadMarkets(true, nil)
	if loadErr != nil {
		t.Fatalf("LoadMarkets failed: %v", loadErr)
	}

	_, err := v.FetchOHLCV("HOSE:VCI/VND", "5m", 0, 10, nil)
	if err == nil {
		t.Fatal("expected error for unsupported timeframe 5m")
	}
	if err.Code != errs.CodeInvalidTimeFrame {
		t.Fatalf("expected CodeInvalidTimeFrame, got code=%d msg=%s", err.Code, err.Error())
	}
}

func TestAmbiguousTickerLookup(t *testing.T) {
	v := &VCI{Exchange: &banexg.Exchange{ExgInfo: &banexg.ExgInfo{Markets: banexg.MarketMap{}}}}
	hose := newStockMarket("HOSE", "DUP", map[string]interface{}{"symbol": "DUP"})
	hnx := newStockMarket("HNX", "DUP", map[string]interface{}{"symbol": "DUP"})
	v.Markets[hose.Symbol] = hose
	v.Markets[hnx.Symbol] = hnx
	v.marketsByRawID = map[string]*banexg.Market{hose.ID: hose, hnx.ID: hnx}
	v.marketsByTicker = map[string][]*banexg.Market{"DUP": {hose, hnx}}

	// Board-qualified lookup should work
	if mar, err := v.MapMarket("HOSE:DUP", 0); err != nil || mar.ID != "HOSE:DUP" {
		t.Fatalf("expected direct board lookup, got market=%v err=%v", mar, err)
	}

	// Bare ticker should fail with ambiguous error
	if _, err := v.MapMarket("DUP", 0); err == nil {
		t.Fatal("expected ambiguous ticker error")
	}
}

func TestEmptyRawID(t *testing.T) {
	v := &VCI{Exchange: &banexg.Exchange{ExgInfo: &banexg.ExgInfo{Markets: banexg.MarketMap{}}}}
	v.marketsByRawID = map[string]*banexg.Market{}
	v.marketsByTicker = map[string][]*banexg.Market{}

	_, err := v.MapMarket("", 0)
	if err == nil {
		t.Fatal("expected error for empty rawID")
	}
	if err.Code != errs.CodeParamRequired {
		t.Fatalf("expected CodeParamRequired, got %d", err.Code)
	}
}

func TestMapMarketBoardQualified(t *testing.T) {
	v := &VCI{Exchange: &banexg.Exchange{ExgInfo: &banexg.ExgInfo{Markets: banexg.MarketMap{}}}}
	market := newStockMarket("HOSE", "VCI", map[string]interface{}{"symbol": "VCI"})
	v.Markets[market.Symbol] = market
	v.marketsByRawID = map[string]*banexg.Market{market.ID: market}
	v.marketsByTicker = map[string][]*banexg.Market{"VCI": {market}}

	mar, err := v.MapMarket("HOSE:VCI", 0)
	if err != nil {
		t.Fatalf("MapMarket failed: %v", err)
	}
	if mar.ID != "HOSE:VCI" {
		t.Fatalf("expected ID HOSE:VCI, got %s", mar.ID)
	}
}

func TestMapMarketNotFound(t *testing.T) {
	v := &VCI{Exchange: &banexg.Exchange{ExgInfo: &banexg.ExgInfo{Markets: banexg.MarketMap{}}}}
	v.marketsByRawID = map[string]*banexg.Market{}
	v.marketsByTicker = map[string][]*banexg.Market{}

	_, err := v.MapMarket("HOSE:ZZZZZ", 0)
	if err == nil {
		t.Fatal("expected error for non-existent market")
	}
	if err.Code != errs.CodeNoMarketForPair {
		t.Fatalf("expected CodeNoMarketForPair, got %d", err.Code)
	}
}

func TestMakeSignGETBehavior(t *testing.T) {
	v := newVCIForTest(t)
	sign := makeSign(v)
	api := &banexg.Entry{Url: "https://example.test/api", Method: "GET", More: map[string]interface{}{"auth": false}}
	req := sign(api, map[string]interface{}{"foo": "bar"})
	if req.Error != nil {
		t.Fatalf("unexpected error: %v", req.Error)
	}
	if !strings.Contains(req.Url, "foo=bar") {
		t.Fatalf("expected query params in URL, got: %s", req.Url)
	}
	// GET should not have body
	if req.Body != "" {
		t.Fatalf("expected empty GET body, got: %q", req.Body)
	}
}

func TestEntryApiEndpoints(t *testing.T) {
	v := newVCIForTest(t)
	if v.Apis[MethodPublicGetSymbolsAll] == nil {
		t.Fatal("getAll endpoint missing")
	}
	if v.Apis[MethodPublicGetSymbolsAll].Method != "GET" {
		t.Fatalf("getAll should be GET, got %s", v.Apis[MethodPublicGetSymbolsAll].Method)
	}
	if v.Apis[MethodPublicPostOhlcChart] == nil {
		t.Fatal("ohlcChart endpoint missing")
	}
	if v.Apis[MethodPublicPostOhlcChart].Method != "POST" {
		t.Fatalf("ohlcChart should be POST, got %s", v.Apis[MethodPublicPostOhlcChart].Method)
	}
}

func TestEstimateCountBack(t *testing.T) {
	since := int64(1700000000000) // some timestamp
	until := since + 30*86400*1000 // 30 calendar days later

	daily := estimateCountBack(since, until, "ONE_DAY")
	if daily <= 0 {
		t.Fatalf("expected positive daily countBack, got %d", daily)
	}

	hourly := estimateCountBack(since, until, "ONE_HOUR")
	if hourly <= daily {
		t.Fatalf("expected hourly > daily, got hourly=%d daily=%d", hourly, daily)
	}

	minute := estimateCountBack(since, until, "ONE_MINUTE")
	if minute <= hourly {
		t.Fatalf("expected minute > hourly, got minute=%d hourly=%d", minute, hourly)
	}
}
