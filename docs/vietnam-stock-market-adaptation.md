# Vietnam Stock Market Adaptation Guide

## Executive Summary

This document outlines how to adapt the BanExg cryptocurrency exchange SDK for the Vietnam stock market. The adaptation involves creating a new exchange module (`vietnam/`) that implements the `BanExchange` interface while accounting for the fundamental differences between crypto and traditional stock markets.

---

## 1. Vietnam Stock Market Overview

### 1.1 Stock Exchanges

| Exchange | Description | Index |
|----------|-------------|-------|
| **HOSE** | Ho Chi Minh Stock Exchange - Main exchange for large-cap stocks | VN-Index, VN30 |
| **HNX** | Hanoi Stock Exchange - Second-tier exchange | HNX-Index, HNX30 |
| **UPCOM** | Unlisted Public Company Market - OTC market | UPCoM-Index |

### 1.2 Trading Hours (NOT 24/7)

| Session | Time (ICT, UTC+7) | Order Types |
|---------|-------------------|-------------|
| Opening Period | 09:00-09:15 | ATO, LO, Negotiated Deals |
| Morning Session | 09:15-11:30 | LO, MTL, Negotiated Deals |
| Lunch Break | 11:30-13:00 | No trading |
| Afternoon Session | 13:00-14:30 | LO, MTL, Negotiated Deals |
| Closing Period | 14:30-14:45 | ATC, LO, Negotiated Deals |
| Negotiated Deals | 14:45-15:00 | Negotiated only |

### 1.3 Key Differences from Crypto

| Aspect | Crypto (Binance) | Vietnam Stock |
|--------|------------------|---------------|
| **Trading Hours** | 24/7/365 | Mon-Fri 09:00-15:00 (excluding holidays) |
| **Settlement** | Instant | T+1.5 (transitioning to T+1 in Q3/2026) |
| **Order Types** | Market, Limit, Stop | ATO, ATC, LO, MTL, MAK |
| **Leverage** | Up to 125x | Limited margin trading (~11-14% interest) |
| **Short Selling** | Freely available | Heavily restricted |
| **Price Limits** | None | ±7% daily limit |
| **Lot Size** | Flexible | Usually 100 shares per lot |
| **API Model** | Exchange-direct | Broker-intermediated |

---

## 2. Available APIs for Vietnam Market

### 2.1 Primary Recommendation: SSI FastConnect API

**Best choice for production trading integration.**

| Feature | Details |
|---------|---------|
| **Documentation** | https://fc-data.ssi.com.vn/ |
| **GitHub SDK** | https://github.com/SSI-Securities-Corporation/python-fcdata |
| **Coverage** | HOSE, HNX, UPCOM, Derivatives, Bonds |
| **Data Types** | Real-time quotes, OHLCV, Order Book, Company info |
| **Trading** | Full order management (place/amend/cancel) |
| **WebSocket** | Yes - Information Distribution Service (IDS) |

**Authentication:**
```
POST /api/v2/Market/AccessToken
{
  "consumerID": "your_consumer_id",
  "consumerSecret": "your_consumer_secret"
}
→ Returns JWT token (valid 8 hours)
→ Use in header: Authorization: Bearer {Token}
```

**Key Endpoints:**
- `/api/Market/GetSecuritiesList` - List all securities
- `/api/Market/GetSecuritiesDetails` - Security details
- `/api/Market/GetDailyOHLC` - Daily candlestick
- `/api/Market/GetIntradayOHLC` - Intraday data
- `/api/Market/GetDailyStockPrice` - Historical prices

### 2.2 Secondary: iTick API

**Best for real-time market data, especially WebSocket.**

| Feature | Details |
|---------|---------|
| **Documentation** | https://docs.itick.org/en |
| **Coverage** | Vietnam HOSE (VN30), plus global markets |
| **Protocols** | REST, WebSocket, FIX (institutional) |
| **Latency** | As low as 100ms |
| **Free Tier** | Available for testing |

### 2.3 Research Library: vnstock (Python)

**Most comprehensive Python library - useful for reference implementation.**

| Feature | Details |
|---------|---------|
| **GitHub** | https://github.com/thinh-vu/vnstock |
| **PyPI** | `pip install vnstock` |
| **Stars** | 1.1k+ |
| **Documentation** | https://vnstocks.com/docs |

**Supported Data:**
- Stock prices (real-time & historical)
- Financial reports (Balance Sheet, Income, Cash Flow)
- Company fundamentals
- Indices, Derivatives/Futures (VN30F)
- Investment Funds (ETF, Mutual Funds)
- Bonds, Forex, Crypto, Gold

---

## 3. Implementation Architecture

### 3.1 Proposed Directory Structure

```
banexg/
├── vietnam/                    # New Vietnam market module
│   ├── entry.go               # Exchange constructor, API mappings
│   ├── data.go                # Method constants, host definitions
│   ├── types.go               # Vietnam struct, response types
│   ├── biz.go                 # Init, LoadMarkets, common logic
│   ├── biz_market.go          # FetchTicker, FetchOHLCV, FetchOrderBook
│   ├── biz_order.go           # CreateOrder, CancelOrder, FetchOrders
│   ├── biz_account.go         # FetchBalance, FetchPositions
│   ├── biz_financial.go       # Vietnam-specific: financials, company info
│   ├── ws_client.go           # WebSocket client management
│   ├── ws_biz.go              # WebSocket message routing
│   ├── common.go              # Helper functions, time zone handling
│   ├── markets.yml            # Static market data (like china/markets.yml)
│   ├── testdata/              # Test fixtures
│   └── AGENTS.md              # Module development guide
└── bex/
    └── entrys.go              # Add: "vietnam": vietnam.NewExchange
```

### 3.2 Core Types Definition

```go
// vietnam/types.go

package vietnam

import "github.com/banbox/banexg"

type Vietnam struct {
    *banexg.Exchange
    
    // SSI API specific
    ConsumerID     string
    ConsumerSecret string
    AccessToken    string
    TokenExpiry    int64
    
    // Trading session state
    SessionState   string  // "pre-open", "open", "lunch", "close", etc.
}

// Vietnam-specific order types
const (
    OdTypeATO = "ATO"  // At-the-Opening
    OdTypeATC = "ATC"  // At-the-Close  
    OdTypeLO  = "LO"   // Limit Order
    OdTypeMTL = "MTL"  // Market-to-Limit
    OdTypeMAK = "MAK"  // Market to Match and Cancel (HNX only)
)

// Stock exchange identifiers
const (
    ExchangeHOSE  = "HOSE"
    ExchangeHNX   = "HNX"
    ExchangeUPCOM = "UPCOM"
)

// SSI API response structures
type SSISecurityInfo struct {
    Symbol         string  `json:"symbol"`
    StockName      string  `json:"stockName"`
    Exchange       string  `json:"exchange"`
    LotSize        int     `json:"lotSize"`
    TickSize       float64 `json:"tickSize"`
    Ceiling        float64 `json:"ceiling"`
    Floor          float64 `json:"floor"`
    RefPrice       float64 `json:"refPrice"`
}

type SSIOHLCData struct {
    TradingDate string  `json:"tradingDate"`
    Open        float64 `json:"open"`
    High        float64 `json:"high"`
    Low         float64 `json:"low"`
    Close       float64 `json:"close"`
    Volume      float64 `json:"volume"`
    Value       float64 `json:"value"`
}

type SSIOrderBook struct {
    Symbol    string      `json:"symbol"`
    Bids      [][]float64 `json:"bids"`  // [price, volume]
    Asks      [][]float64 `json:"asks"`
    Timestamp int64       `json:"timestamp"`
}

// Financial data (Vietnam-specific feature)
type FinancialReport struct {
    Symbol    string                 `json:"symbol"`
    Period    string                 `json:"period"`
    Year      int                    `json:"year"`
    Quarter   int                    `json:"quarter"`
    Data      map[string]interface{} `json:"data"`
}
```

### 3.3 Entry Point Implementation

```go
// vietnam/entry.go

package vietnam

import (
    "github.com/banbox/banexg"
    "github.com/banbox/banexg/errs"
)

func New(Options map[string]interface{}) (*Vietnam, *errs.Error) {
    exg := &Vietnam{
        Exchange: &banexg.Exchange{
            ExgInfo: &banexg.ExgInfo{
                ID:        "vietnam",
                Name:      "Vietnam Stock Market",
                Countries: []string{"VN"},
                FullDay:   false,  // Not 24/7
            },
            RateLimit: 100,  // Adjust based on SSI API limits
            Options:   Options,
            Hosts: &banexg.ExgHosts{
                Prod: map[string]string{
                    HostDataAPI:    "https://fc-data.ssi.com.vn",
                    HostTradingAPI: "https://fc-tradeapi.ssi.com.vn",
                    HostWS:         "wss://fc-datahub.ssi.com.vn",
                },
                Www: "https://www.ssi.com.vn",
                Doc: []string{
                    "https://guide.ssi.com.vn/ssi-products",
                    "https://fc-data.ssi.com.vn/",
                },
            },
            Fees: &banexg.ExgFee{
                Main: &banexg.TradeFee{
                    FeeSide:    "quote",
                    TierBased:  true,
                    Percentage: true,
                    Taker:      0.0015,  // ~0.15% typical brokerage fee
                    Maker:      0.0015,
                },
            },
            Apis: map[string]*banexg.Entry{
                MethodGetAccessToken:       {Path: "api/v2/Market/AccessToken", Host: HostDataAPI, Method: "POST", Cost: 1},
                MethodGetSecuritiesList:    {Path: "api/Market/GetSecuritiesList", Host: HostDataAPI, Method: "GET", Cost: 1},
                MethodGetSecuritiesDetails: {Path: "api/Market/GetSecuritiesDetails", Host: HostDataAPI, Method: "GET", Cost: 1},
                MethodGetDailyOHLC:         {Path: "api/Market/GetDailyOHLC", Host: HostDataAPI, Method: "GET", Cost: 1},
                MethodGetIntradayOHLC:      {Path: "api/Market/GetIntradayOHLC", Host: HostDataAPI, Method: "GET", Cost: 1},
                MethodGetDailyStockPrice:   {Path: "api/Market/GetDailyStockPrice", Host: HostDataAPI, Method: "GET", Cost: 1},
                MethodGetIndexComponents:   {Path: "api/Market/GetIndexComponents", Host: HostDataAPI, Method: "GET", Cost: 1},
                // Trading APIs
                MethodPlaceOrder:           {Path: "api/Trading/NewOrder", Host: HostTradingAPI, Method: "POST", Cost: 1},
                MethodCancelOrder:          {Path: "api/Trading/CancelOrder", Host: HostTradingAPI, Method: "POST", Cost: 1},
                MethodModifyOrder:          {Path: "api/Trading/ModifyOrder", Host: HostTradingAPI, Method: "POST", Cost: 1},
                MethodGetOrderHistory:      {Path: "api/Trading/GetOrderHistory", Host: HostTradingAPI, Method: "GET", Cost: 1},
                MethodGetAccountBalance:    {Path: "api/Account/GetAccountBalance", Host: HostTradingAPI, Method: "GET", Cost: 1},
                MethodGetStockPosition:     {Path: "api/Account/GetStockPosition", Host: HostTradingAPI, Method: "GET", Cost: 1},
            },
            Has: map[string]map[string]int{
                "": {
                    banexg.ApiFetchTicker:           banexg.HasOk,
                    banexg.ApiFetchTickers:          banexg.HasOk,
                    banexg.ApiFetchOHLCV:            banexg.HasOk,
                    banexg.ApiFetchOrderBook:        banexg.HasOk,
                    banexg.ApiFetchOrder:            banexg.HasOk,
                    banexg.ApiFetchOrders:           banexg.HasOk,
                    banexg.ApiFetchBalance:          banexg.HasOk,
                    banexg.ApiFetchOpenOrders:       banexg.HasOk,
                    banexg.ApiCreateOrder:           banexg.HasOk,
                    banexg.ApiCancelOrder:           banexg.HasOk,
                    banexg.ApiEditOrder:             banexg.HasOk,
                    // Not applicable for stocks
                    banexg.ApiFetchFundingRate:      banexg.HasFail,
                    banexg.ApiSetLeverage:           banexg.HasFail,
                    banexg.ApiFetchPositions:        banexg.HasFail, // Use stock positions instead
                    // WebSocket
                    banexg.ApiWatchOrderBooks:       banexg.HasOk,
                    banexg.ApiWatchTrades:           banexg.HasOk,
                    banexg.ApiWatchOHLCVs:           banexg.HasOk,
                },
            },
        },
    }
    
    err := exg.Init()
    if err != nil {
        return nil, err
    }
    
    return exg, nil
}

func NewExchange(Options map[string]interface{}) (banexg.BanExchange, *errs.Error) {
    return New(Options)
}
```

### 3.4 Data Constants

```go
// vietnam/data.go

package vietnam

// Host constants
const (
    HostDataAPI    = "dataApi"
    HostTradingAPI = "tradingApi"
    HostWS         = "ws"
)

// Method constants
const (
    // Authentication
    MethodGetAccessToken = "getAccessToken"
    
    // Market Data
    MethodGetSecuritiesList    = "getSecuritiesList"
    MethodGetSecuritiesDetails = "getSecuritiesDetails"
    MethodGetDailyOHLC         = "getDailyOHLC"
    MethodGetIntradayOHLC      = "getIntradayOHLC"
    MethodGetDailyStockPrice   = "getDailyStockPrice"
    MethodGetIndexComponents   = "getIndexComponents"
    
    // Trading
    MethodPlaceOrder        = "placeOrder"
    MethodCancelOrder       = "cancelOrder"
    MethodModifyOrder       = "modifyOrder"
    MethodGetOrderHistory   = "getOrderHistory"
    MethodGetAccountBalance = "getAccountBalance"
    MethodGetStockPosition  = "getStockPosition"
)

// Trading session times (ICT = UTC+7)
var TradingSchedule = map[string][2]string{
    "pre_open":  {"09:00", "09:15"},
    "morning":   {"09:15", "11:30"},
    "lunch":     {"11:30", "13:00"},
    "afternoon": {"13:00", "14:30"},
    "closing":   {"14:30", "14:45"},
    "after":     {"14:45", "15:00"},
}

// Price tick sizes by price range
var PriceTickSizes = map[string]float64{
    "HOSE_below_10000":  10,     // 10 VND for prices < 10,000
    "HOSE_10000_50000":  50,     // 50 VND for 10,000 <= price < 50,000
    "HOSE_above_50000":  100,    // 100 VND for price >= 50,000
    "HNX":               100,    // 100 VND for all HNX stocks
}
```

---

## 4. Implementation Roadmap

### Phase 1: Core Infrastructure (Week 1-2)

| Task | Priority | Description |
|------|----------|-------------|
| Create module skeleton | High | entry.go, data.go, types.go basic structure |
| Register in bex | High | Add "vietnam" to bex/entrys.go |
| Implement authentication | High | Token management with 8-hour refresh |
| LoadMarkets | High | Fetch securities list from SSI |
| Time zone handling | High | ICT (UTC+7) support, trading hours validation |

### Phase 2: Market Data (Week 2-3)

| Task | Priority | Description |
|------|----------|-------------|
| FetchTicker | High | Current price, bid/ask, volume |
| FetchTickers | High | Multiple securities at once |
| FetchOHLCV | High | Daily and intraday candlesticks |
| FetchOrderBook | Medium | Bid/ask depth |
| MapMarket | Medium | Symbol normalization (VN30 vs SSI format) |

### Phase 3: Trading (Week 3-4)

| Task | Priority | Description |
|------|----------|-------------|
| CreateOrder | High | Support ATO, ATC, LO, MTL order types |
| CancelOrder | High | Order cancellation |
| EditOrder | Medium | Order modification (LO only) |
| FetchOrders | High | Order history |
| FetchOpenOrders | High | Current active orders |
| FetchBalance | High | Cash balance |
| FetchPositions | High | Stock holdings |

### Phase 4: WebSocket (Week 4-5)

| Task | Priority | Description |
|------|----------|-------------|
| WS Client setup | Medium | Connection management, auth |
| WatchOrderBooks | Medium | Real-time order book updates |
| WatchTrades | Medium | Real-time trade feed |
| WatchOHLCVs | Low | Real-time candlestick updates |

### Phase 5: Vietnam-Specific Features (Week 5-6)

| Task | Priority | Description |
|------|----------|-------------|
| FetchFinancials | Medium | Balance sheet, income statement |
| FetchCompanyInfo | Medium | Company profile, fundamentals |
| Trading session checks | High | Prevent orders outside trading hours |
| Price limit validation | Medium | Validate orders within ±7% ceiling/floor |

---

## 5. Key Implementation Considerations

### 5.1 Order Type Mapping

```go
// Map BanExg standard order types to Vietnam-specific types
func (e *Vietnam) mapOrderType(banexgType string, session string) string {
    switch banexgType {
    case banexg.OdTypeMarket:
        // Market orders depend on session
        if session == "pre_open" {
            return OdTypeATO
        } else if session == "closing" {
            return OdTypeATC
        }
        return OdTypeMTL  // Market-to-Limit during continuous
    case banexg.OdTypeLimit:
        return OdTypeLO
    default:
        return OdTypeLO
    }
}
```

### 5.2 Trading Hours Validation

```go
// Check if current time is within trading hours
func (e *Vietnam) isMarketOpen() bool {
    now := time.Now().In(locICT)
    
    // Check if weekend
    if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
        return false
    }
    
    // Check trading hours (09:00 - 14:45)
    hour, min := now.Hour(), now.Minute()
    timeVal := hour*100 + min
    
    if timeVal >= 900 && timeVal < 1130 {
        return true  // Morning session
    }
    if timeVal >= 1300 && timeVal < 1445 {
        return true  // Afternoon session
    }
    
    return false
}
```

### 5.3 Token Refresh Logic

```go
func (e *Vietnam) ensureValidToken() *errs.Error {
    if e.AccessToken != "" && time.Now().UnixMilli() < e.TokenExpiry-60000 {
        return nil  // Token still valid (with 1-minute buffer)
    }
    
    // Refresh token
    args := map[string]interface{}{
        "consumerID":     e.ConsumerID,
        "consumerSecret": e.ConsumerSecret,
    }
    
    rsp := e.RequestApi(context.Background(), MethodGetAccessToken, args)
    if rsp.Error != nil {
        return rsp.Error
    }
    
    // Parse and store new token
    // Token validity: 8 hours
    e.TokenExpiry = time.Now().UnixMilli() + 8*60*60*1000
    
    return nil
}
```

### 5.4 Symbol Normalization

```go
// Vietnam symbols follow format: {EXCHANGE}:{SYMBOL}
// Example: HOSE:VNM, HNX:SHB, UPCOM:BSR

func (e *Vietnam) normalizeSymbol(symbol string) (exchange, code string) {
    parts := strings.Split(symbol, ":")
    if len(parts) == 2 {
        return parts[0], parts[1]
    }
    // Default to HOSE if no exchange specified
    return "HOSE", symbol
}
```

---

## 6. Testing Strategy

### 6.1 Unit Tests

```go
// vietnam/biz_test.go

func TestLoadMarkets(t *testing.T) {
    exg, _ := New(map[string]interface{}{
        "consumerID":     os.Getenv("SSI_CONSUMER_ID"),
        "consumerSecret": os.Getenv("SSI_CONSUMER_SECRET"),
    })
    
    markets, err := exg.LoadMarkets(true, nil)
    assert.Nil(t, err)
    assert.NotEmpty(t, markets)
    
    // Check VN30 stocks are present
    _, exists := markets["HOSE:VNM"]
    assert.True(t, exists)
}

func TestTradingHoursValidation(t *testing.T) {
    exg, _ := New(nil)
    
    // Test weekend
    weekend := time.Date(2026, 1, 31, 10, 0, 0, 0, locICT) // Saturday
    assert.False(t, exg.isMarketOpenAt(weekend))
    
    // Test lunch break
    lunch := time.Date(2026, 1, 26, 12, 0, 0, 0, locICT) // Monday noon
    assert.False(t, exg.isMarketOpenAt(lunch))
    
    // Test trading hours
    trading := time.Date(2026, 1, 26, 10, 0, 0, 0, locICT) // Monday 10am
    assert.True(t, exg.isMarketOpenAt(trading))
}
```

### 6.2 Integration Tests

Use SSI sandbox/test environment if available, or mock responses from `testdata/` directory.

---

## 7. Resources & References

### Official Documentation
- **SSI FastConnect API**: https://fc-data.ssi.com.vn/
- **SSI Developer Guide**: https://guide.ssi.com.vn/ssi-products
- **SSI GitHub**: https://github.com/SSI-Securities-Corporation

### Data Providers
- **iTick API**: https://docs.itick.org/en
- **FiinGroup/FiinQuant**: https://fiinpro.com
- **Vietstock DataFeed**: https://dichvu.vietstock.vn

### Reference Implementations
- **vnstock (Python)**: https://github.com/thinh-vu/vnstock
- **vdatafeed**: https://github.com/quant-vn/vdatafeed
- **vbroker**: https://github.com/quant-vn/vbroker

### Vietnam Market Information
- **HOSE Official**: https://www.hsx.vn
- **HNX Official**: https://www.hnx.vn
- **Vietnam Market Tutorial (2026)**: https://blog.itick.org/en/stock-api/2026-vietnam-stock-exchange-api-python-tutorial

---

## 8. Appendix: BanExchange Interface Mapping

| BanExchange Method | Vietnam Implementation | Notes |
|--------------------|----------------------|-------|
| `LoadMarkets` | ✅ Required | Fetch from SSI GetSecuritiesList |
| `FetchTicker` | ✅ Required | Real-time quote |
| `FetchOHLCV` | ✅ Required | Daily + Intraday |
| `FetchOrderBook` | ✅ Required | Bid/Ask depth |
| `CreateOrder` | ✅ Required | Handle ATO/ATC/LO/MTL |
| `CancelOrder` | ✅ Required | Standard |
| `EditOrder` | ✅ Required | LO orders only |
| `FetchBalance` | ✅ Required | Cash balance |
| `FetchPositions` | ⚠️ Adapt | Map to stock holdings |
| `FetchFundingRate` | ❌ N/A | Not applicable to stocks |
| `SetLeverage` | ❌ N/A | Margin managed by broker |
| `WatchOrderBooks` | ✅ Optional | SSI WebSocket |
| `WatchTrades` | ✅ Optional | SSI WebSocket |
| `FetchFinancials` | 🆕 Vietnam-specific | Balance sheet, Income statement |
| `FetchCompanyInfo` | 🆕 Vietnam-specific | Company profile |

---

## 9. Estimated Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Phase 1: Core | 2 weeks | Module skeleton, auth, LoadMarkets |
| Phase 2: Data | 1 week | Ticker, OHLCV, OrderBook |
| Phase 3: Trading | 1.5 weeks | Order management, Balance |
| Phase 4: WebSocket | 1 week | Real-time data feeds |
| Phase 5: Extras | 1 week | Financials, validation, polish |
| Testing & Docs | 0.5 week | Unit tests, integration tests |

**Total: ~7 weeks** for full implementation

---

## 10. Next Steps

1. **Register for SSI FastConnect API** - Get production credentials
2. **Study vnstock source code** - Understand Vietnam market quirks
3. **Create module skeleton** - entry.go, data.go, types.go
4. **Implement authentication** - Token management
5. **Start with LoadMarkets** - Foundation for all other methods
