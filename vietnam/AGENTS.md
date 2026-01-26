# Vietnam Module Development Guide

## 1. Core File Responsibilities

| File | Core Responsibility | Key Content |
| :--- | :--- | :--- |
| **`entry.go`** | **Entry & Routing** | `New` constructor; **`Apis` mapping table** (defines all REST API routes, Host, Cost); SSI authentication & token refresh. |
| **`data.go`** | **Constants Definition** | `MethodXXX` method name constants; Host type constants; API enum value mappings (order types, order sides, order status, exchange IDs). |
| **`types.go`** | **Data Structures** | `Vietnam` main struct; **SSI API raw JSON response structures** (prefixed with `SSI`). |
| **`common.go`** | **Common Utilities** | Time zone handling (ICT/UTC+7); Symbol normalization; Trading session detection; Price/quantity validation; Order type/side/status mapping. |
| **`biz_*.go`** | **Business Logic** | REST API concrete implementations. e.g. `biz_order.go` (orders), `biz_market.go` (market data). |
| **`ws_*.go`** | **WebSocket** | WS connection management, subscription logic, message routing. |

## 2. Adding New REST API Development Flow (Standard Flow)

Adding a new endpoint requires following this **4-step** process:

### Step 1: Define Method Constant (`data.go`)
Add a unique Method constant in `data.go`.
```go
// Naming convention: Method + [DataAPI/TradingAPI] + Action
const MethodDataAPIGetSecurityDetail = "dataApiGetSecurityDetail"
```

### Step 2: Register API Route (`entry.go`)
Register the endpoint configuration in `entry.go`'s `Apis` map.
```go
MethodDataAPIGetSecurityDetail: {
    Path:   "Securities/Detail/{symbol}", // URL relative path
    Host:   HostDataAPI,                  // References Host constant from data.go
    Method: "GET",                        // HTTP method
    Cost:   1,                            // Weight cost
},
```

### Step 3: Define Raw Data Structure (`types.go`)
Define the **raw JSON structure** returned by the SSI API in `types.go`.
```go
type SSISecurityDetail struct {
    Symbol     string  `json:"symbol"`
    CompanyName string `json:"companyName"`
    Exchange   string  `json:"exchange"`
    // ... other fields
}
```

### Step 4: Implement Business Method (`biz_*.go`)
Implement the method in an appropriate `biz_` file following the unified pattern:

```go
func (e *Vietnam) FetchSecurityDetail(symbol string, params map[string]interface{}) (*banexg.Market, *errs.Error) {
    // 1. Preprocess parameters & market (required)
    args, market, err := e.LoadArgsMarket(symbol, params)
    if err != nil {
        return nil, err
    }
    
    // 2. Prepare request parameters
    args["symbol"] = market.ID // SSI typically uses internal symbol ID
    
    // 3. Make request (auto handles token injection & retry)
    method := MethodDataAPIGetSecurityDetail
    rsp := e.RequestApiRetry(context.Background(), method, args, 1)
    if rsp.Error != nil {
        return nil, rsp.Error
    }
    
    // 4. Parse & convert (convert SSI raw structure to banexg standard structure)
    var raw SSISecurityDetail
    if err := utils.UnmarshalString(rsp.Content, &raw, utils.JsonNumDefault); err != nil {
        return nil, errs.NewMsg(errs.CodeDataLost, "parse error: %v", err)
    }
    
    // Convert to standard banexg.Market structure
    return &banexg.Market{
        ID:     raw.Symbol,
        Symbol: raw.Symbol,
        // ... field mapping
    }, nil
}
```

## 3. Important Development Conventions

### 3.1 Market & Parameter Handling
- **`LoadArgsMarket`**: All business methods must call this first. It:
  - Copies `params` to prevent side effects
  - Parses `symbol` string to `*banexg.Market` object
  - Validates market exists
- **SSI Symbol Format**: SSI API expects "EXCHANGE:CODE" format (e.g., "HOSE:VNM")

### 3.2 Host Constant Selection Guide
- **`HostDataAPI`**: Market data endpoints (securities, OHLC, orderbook, etc.)
- **`HostTradingAPI`**: Trading operations (create/cancel order, balance, positions)
- **`HostWS`**: WebSocket real-time data streams

### 3.3 Authentication & Token Management
- **Bearer Token**: SSI uses Bearer token (not API key/secret signing like Binance)
- **8-hour Expiry**: Tokens expire after 8 hours, auto-refreshed by `ensureValidToken()`
- **Token Injection**: Handled automatically in `RequestApiRetry()` override

### 3.4 Trading Session Awareness
Vietnam stock market has specific trading sessions:
- **Pre-open**: 08:30-09:00 (ATO orders only)
- **Morning**: 09:00-11:30 (continuous matching)
- **Break**: 11:30-13:00 (lunch break, no trading)
- **Afternoon**: 13:00-14:30 (continuous matching)
- **Close**: 14:30-14:45 (ATC orders only)

**Session-aware validation**:
```go
// Check if market is open
if !IsMarketOpen(time.Now(), Vietnam.TradingSchedule) {
    return nil, errs.NewMsg(errs.CodeMarketClosed, "market is closed")
}

// Check if order type is allowed in current session
sessionType := GetCurrentSessionType(time.Now(), Vietnam.TradingSchedule)
if !CanPlaceOrderType(orderType, sessionType) {
    return nil, errs.NewMsg(errs.CodeParamInvalid, "order type %s not allowed in %s session", orderType, sessionType)
}
```

### 3.5 Vietnam-Specific Order Types
Vietnam stock market has unique order types (different from crypto "Market/Limit"):
- **ATO**: At The Open (pre-open auction)
- **ATC**: At The Close (closing auction)
- **LO**: Limit Order (standard limit order)
- **MTL**: Market To Limit (executes at market, unfilled becomes limit)
- **MAK**: Market At Kill (market order, unfilled cancels)

**Order Type Mapping**:
```go
// BanExg -> SSI
func orderTypeToBanExg(ssiType string) string {
    switch ssiType {
    case OrderTypeATO:
        return banexg.OdTypeATO
    case OrderTypeATC:
        return banexg.OdTypeATC
    case OrderTypeLO:
        return banexg.OdTypeLimit
    case OrderTypeMTL:
        return banexg.OdTypeMarket
    case OrderTypeMAK:
        return banexg.OdTypeMarket
    default:
        return ""
    }
}
```

### 3.6 Price & Quantity Validation
Vietnam exchanges have specific rules:
- **Price Tick Size**: Varies by exchange and price range (see `GetPriceTickSize` in `common.go`)
- **Lot Size**: Default 100 shares per lot
- **Price Limits**: ±7% daily price limits (see `CalculatePriceLimits`)

**Validation Helpers**:
```go
// Round price to valid tick size
validPrice := RoundPrice(price, exchange)

// Validate price within limits
if err := ValidatePrice(price, refPrice, exchange); err != nil {
    return err
}

// Validate quantity is multiple of lot size
if err := ValidateQuantity(quantity, lotSize); err != nil {
    return err
}
```

### 3.7 Time Zone Handling
- **All times in ICT (UTC+7)**, not UTC
- **Use** `parseVietnamTime()` and `formatVietnamTime()` helpers
- **Trading schedule** defined in `data.go` (see `VietnamTradingSchedule`)

### 3.8 Error Handling
- Unified use of `*errs.Error`
- Parameter errors use `errs.NewMsg(errs.CodeParamInvalid, "msg")`
- Network/API errors automatically wrapped by `RequestApiRetry`

### 3.9 Symbol Normalization
SSI uses "EXCHANGE:CODE" format:
```go
// Normalize input symbol to SSI format
symbol := NormalizeSymbol("VNM")          // -> "HOSE:VNM" (auto-detects exchange)
symbol := BuildSymbol("HOSE", "VNM")      // -> "HOSE:VNM" (explicit exchange)
```

## 4. Vietnam vs Crypto Exchange Key Differences

| Aspect | Crypto (Binance/Bybit) | Vietnam Stock (SSI) |
|--------|------------------------|---------------------|
| **Trading Hours** | 24/7 | Mon-Fri 09:00-14:45 ICT (with lunch break) |
| **Order Types** | Market/Limit/StopLoss/etc. | ATO/ATC/LO/MTL/MAK |
| **Authentication** | API Key + Secret (HMAC signing) | Bearer Token (8-hour expiry) |
| **Time Zone** | UTC | ICT (UTC+7) |
| **Symbol Format** | "BTCUSDT" | "HOSE:VNM" (exchange:code) |
| **Settlement** | Instant | T+1.5 (transitioning to T+1) |
| **Price Limits** | None (generally) | ±7% daily limits |
| **Lot Size** | Arbitrary precision | 100 shares per lot |
| **Leverage** | Yes (futures) | No (cash only) |
| **Funding Rates** | Yes (perpetual) | No |

## 5. Testing Considerations

### 5.1 Test Data Location
- Place test fixtures in `vietnam/testdata/`
- Mock SSI API responses for offline testing

### 5.2 Credentials
- Store SSI credentials in `vietnam/local.json` (gitignored)
- Example format:
```json
{
  "consumerID": "YOUR_CONSUMER_ID",
  "consumerSecret": "YOUR_CONSUMER_SECRET",
  "sandbox": true
}
```

### 5.3 Test Environment
- SSI provides sandbox environment for testing
- Set `sandbox: true` in options to use test endpoints

### 5.4 Market Hours Testing
- Many tests will fail outside trading hours (09:00-14:45 Mon-Fri ICT)
- Use mocked responses or skip time-sensitive tests outside trading hours

## 6. Reference Implementations

### 6.1 Similar Module
- **`china/`** module - also handles non-24/7 markets with trading sessions
- Use `china/` as reference for session-aware logic

### 6.2 Crypto Reference
- **`binance/`** module - comprehensive REST & WebSocket implementation
- Use for general API implementation patterns (but adapt for session awareness)

## 7. Common Pitfalls

❌ **DON'T**:
- Use UTC timestamps without converting to ICT
- Assume market is always open (24/7 mentality)
- Use crypto order types (Market/Limit) directly without mapping
- Hardcode API endpoints (use Host constants)
- Skip token expiry checks (tokens expire after 8 hours)

✅ **DO**:
- Always check trading hours before placing orders
- Map between BanExg standard types and SSI types
- Use helper functions from `common.go` for validation
- Handle session-specific order type restrictions
- Test with mock data outside trading hours

## 8. Development Checklist

When implementing a new feature:
- [ ] Define Method constant in `data.go`
- [ ] Register API route in `entry.go` Apis map
- [ ] Define SSI response struct in `types.go`
- [ ] Implement business logic in appropriate `biz_*.go`
- [ ] Add order type/session validation if applicable
- [ ] Add price/quantity validation if applicable
- [ ] Convert SSI types to BanExg standard types
- [ ] Handle token refresh if needed
- [ ] Add test cases with mock data
- [ ] Update Has flags in `entry.go` if new capability
- [ ] Document Vietnam-specific behavior differences

## 9. AI Development Principles

- **DRY (Don't Repeat Yourself)**: Extract duplicated code into helper functions in `common.go`
- **Minimal Code**: Only generate code currently needed, avoid over-engineering
- **Vietnam-First Thinking**: Don't blindly copy crypto patterns - adapt for Vietnam stock market realities
- **Session Awareness**: Always consider trading hours/sessions when implementing order-related features
