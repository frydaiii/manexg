# SSI → VCI Migration Guide

Switching from the SSI (`"vietnam"`) provider to the VCI (`"vci"`) provider in `banexg`.

---

## 1. Exchange Name

```diff
- exg, err := bex.New("vietnam", opts)
+ exg, err := bex.New("vci", nil)
```

VCI requires **no credentials** — pass `nil` or an empty map. Drop `ConsumerID` / `ConsumerSecret`.

---

## 2. Symbol Format — Identical

Both providers use the same format. No changes needed.

```
HOSE:VCI/VND
HNX:SHS/VND
UPCOM:MCG/VND
```

`MapMarket` works the same way:
- Board-qualified: `"HOSE:VCI"` → exact match
- Bare ticker: `"VCI"` → works if unambiguous, errors if ticker exists on multiple boards

---

## 3. Index Symbols

SSI uses separate index endpoints (`IndexList`, `DailyIndex`). VCI integrates indexes as **pseudo-markets** loaded during `LoadMarkets`.

```diff
  // SSI — indexes are a special case, separate from stocks
- // (your code may have special-cased index fetching)

  // VCI — indexes are regular markets, same API for everything
  klines, err := exg.FetchOHLCV("INDEX:VNINDEX/VND", "1d", since, limit, nil)
  klines, err := exg.FetchOHLCV("INDEX:HNXINDEX/VND", "1d", since, limit, nil)
  klines, err := exg.FetchOHLCV("INDEX:UPCOMINDEX/VND", "1d", since, limit, nil)
```

The internal VCI→API symbol mapping is automatic:

| Your Symbol | VCI API Symbol |
|-------------|---------------|
| `INDEX:VNINDEX/VND` | `VNINDEX` |
| `INDEX:HNXINDEX/VND` | `HNXIndex` |
| `INDEX:UPCOMINDEX/VND` | `HNXUpcomIndex` |

---

## 4. Supported Timeframes — Reduced in v1

This is the **biggest breaking change**. VCI v1 only supports native resolutions.

| Timeframe | SSI | VCI v1 | Notes |
|-----------|-----|--------|-------|
| `1m` | ✅ | ✅ | |
| `5m` | ✅ | ❌ | Returns `CodeInvalidTimeFrame` |
| `15m` | ✅ | ❌ | Returns `CodeInvalidTimeFrame` |
| `30m` | ✅ | ❌ | Returns `CodeInvalidTimeFrame` |
| `1h` | ✅ | ✅ | |
| `1d` | ✅ | ✅ | |

**If your code uses 5m/15m/30m**, you must either:
1. Fetch `1m` data and resample client-side
2. Stay on SSI for those timeframes
3. Wait for VCI v2 which will add resampling with VN lunch-break awareness (11:30–13:00)

```go
// Check for the error
klines, err := exg.FetchOHLCV("HOSE:VCI/VND", "5m", since, limit, nil)
if err != nil && err.Code == errs.CodeInvalidTimeFrame {
    // Fallback: fetch 1m and resample yourself
}
```

---

## 5. FetchOHLCV — Behavioral Differences

### Request model

| | SSI | VCI |
|-|-----|-----|
| **Time params** | `FromDate` / `ToDate` (DD/MM/YYYY strings) | `to` (Unix seconds) + `countBack` (int) |
| **Pagination** | `PageIndex` / `PageSize` | `countBack` with 1.5× overfetch |
| **Daily vs Intraday** | Separate endpoints + separate parsers | Single endpoint, different `timeFrame` value |

### Response model

| | SSI | VCI |
|-|-----|-----|
| **Format** | Row-oriented: `[{TradingDate, Open, High, ...}, ...]` | Column-oriented: `{t:[], o:[], h:[], l:[], c:[], v:[]}` |
| **Timestamps** | Date strings `"15/02/2026"` + optional time `"09:05:00"` | Unix seconds in `t[]` array |
| **Envelope** | Always `{"status":"SUCCESS", "dataList":[...]}` | Either bare `[{...}]` or `{"data":[{...}]}` |

**From the caller's perspective, the `FetchOHLCV` signature and return type are identical** — `[]*banexg.Kline` with `Time` in milliseconds. The differences are internal.

---

## 6. Authentication — Removed

| | SSI | VCI |
|-|-----|-----|
| **Auth** | OAuth2 token via `ConsumerID`/`ConsumerSecret` | None |
| **Token refresh** | Automatic JWT refresh with expiry cache | N/A |
| **Rate limit** | 1 req/s (documented, enforced with 429 + retry) | ~2 req/s (guessed, no documented limit) |
| **Headers** | `Authorization: Bearer <token>` | Browser-mimicking headers (Referer, Origin, DNT, User-Agent) |

Constructor changes:

```diff
- exg, err := bex.New("vietnam", map[string]interface{}{
-     "ConsumerID":     "your-id",
-     "ConsumerSecret": "your-secret",
- })
+ exg, err := bex.New("vci", nil)
```

---

## 7. Market Loading

### Boards

| | SSI | VCI |
|-|-----|-----|
| **Boards** | HOSE, HNX, UPCOM, DER | HOSE, HNX, UPCOM (no derivatives) |
| **Types** | All types loaded | Filtered to STOCK + ETF only |
| **Index markets** | Not in Markets map | Added as pseudo-markets (`INDEX:VNINDEX/VND`, etc.) |

### Precision Defaults

VCI doesn't return tick sizes in its API, so board-based defaults are used:

| Board | Tick Size | Lot Size |
|-------|-----------|----------|
| HOSE | 10 VND | 100 shares |
| HNX | 100 VND | 100 shares |
| UPCOM | 100 VND | 1 share (odd lots allowed) |

SSI may return different precision from its `SecuritiesDetails` endpoint. If your code depends on per-symbol tick sizes, be aware VCI uses board-level defaults.

### Market.Info Fields

```diff
  // Both providers set:
  market.Info["board"]   // "HOSE", "HNX", "UPCOM"
  market.Info["ticker"]  // "VCI", "TCB", etc.
  market.Info["rawId"]   // "HOSE:VCI"

  // SSI-specific (not in VCI):
- market.Info["StockSymbol"]
- market.Info["Code"]
- market.Info["RepeatedInfoList"]

  // VCI-specific:
+ market.Info["organName"]      // Vietnamese company name
+ market.Info["enOrganName"]    // English company name
+ market.Info["type"]           // "STOCK" or "ETF"

  // VCI index markets only:
+ market.Info["isIndex"]        // true
+ market.Info["vciSymbol"]      // API symbol ("HNXIndex", etc.)
```

---

## 8. Rate Limiting

```
SSI:  RateLimit = 1100ms  (documented 1 req/s, enforced server-side)
VCI:  RateLimit =  500ms  (guessed ~2 req/s, no documented limit)
```

VCI is faster but the limit is a guess. If you hit issues in production, increase `RateLimit` via options or expect tuning.

SSI has built-in retry logic for `"quota exceeded"` responses (up to 3 retries with backoff). VCI has no equivalent because it doesn't return structured rate-limit errors.

---

## 9. Quick Migration Checklist

- [ ] Change `bex.New("vietnam", opts)` → `bex.New("vci", nil)`
- [ ] Remove `ConsumerID` / `ConsumerSecret` from config
- [ ] Replace any 5m/15m/30m timeframes with 1m+resample or 1h/1d
- [ ] Update index symbol references to use `INDEX:VNINDEX/VND` format
- [ ] Remove any SSI-specific `market.Info` field access (`StockSymbol`, `Code`)
- [ ] Remove DER (derivatives) market handling if present — VCI doesn't support it
- [ ] Test with `exg.LoadMarkets(true, nil)` and verify expected market count
- [ ] Test OHLCV for a stock, an ETF, and an index symbol
- [ ] Monitor rate limiting in production — VCI's 500ms limit is a guess

---

## 10. Running Both Providers

If you need SSI for some features (5m timeframes, derivatives) and VCI for others (no auth needed, faster):

```go
ssi, _ := bex.New("vietnam", map[string]interface{}{
    "ConsumerID":     "...",
    "ConsumerSecret": "...",
})
vci, _ := bex.New("vci", nil)

// Use VCI for daily/hourly data (no auth hassle)
klines, _ := vci.FetchOHLCV("HOSE:VCI/VND", "1d", since, limit, nil)

// Use SSI for 5m data (VCI doesn't support it yet)
klines5m, _ := ssi.FetchOHLCV("HOSE:VCI/VND", "5m", since, limit, nil)
```

Both use the same symbol format, so results are interchangeable.

---

## 11. What's Coming in VCI v2

| Feature | Status |
|---------|--------|
| OHLCV resampling (5m, 15m, 30m, 1W, 1M) | Planned — needs VN lunch-break handling |
| FetchTicker / FetchTickers (price board) | Planned — `POST /price/symbols/getList` |
| GraphQL endpoints (company profile, financials) | Planned |
| Intraday tick data | Planned — `POST /market-watch/LEData/getAll` |
