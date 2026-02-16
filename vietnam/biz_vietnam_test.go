package vietnam

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/banbox/banexg"
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
