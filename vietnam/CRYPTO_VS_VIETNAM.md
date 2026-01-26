# Vietnam vs Crypto Trading - Key Differences

**For developers familiar with crypto exchange APIs (Binance, Bybit, etc.)**

This guide highlights the key differences when trading Vietnamese stocks vs cryptocurrencies using the BanExg library.

---

## Quick Comparison Table

| Aspect | Crypto (Binance/Bybit) | Vietnam Stock (SSI) |
|--------|------------------------|---------------------|
| **Trading Hours** | 24/7/365 | Mon-Fri 09:00-14:30 ICT (with lunch break) |
| **Weekends** | Open | Closed |
| **Holidays** | Open | Closed (Vietnamese public holidays) |
| **Authentication** | API Key + Secret (HMAC signing) | Bearer Token (consumerID + consumerSecret) |
| **Token Expiry** | None (sign each request) | 8 hours (auto-refreshed) |
| **Order Types** | Market, Limit, Stop, Stop-Limit, etc. | LO, MTL, ATO, ATC, MAK |
| **Symbol Format** | `"BTC/USDT"` or `"BTCUSDT"` | `"HOSE:VNM"` (exchange:code) |
| **Price Limits** | None (typically) | ±7% daily limits |
| **Tick Size** | Arbitrary precision (e.g., 0.01) | Fixed (10/50/100 VND depending on price) |
| **Lot Size** | Any amount (0.001 BTC) | 100 shares minimum (1 lot) |
| **Settlement** | Instant | T+1.5 (cash at T+2) |
| **Leverage** | Yes (1x-125x) | No (cash only) |
| **Short Selling** | Yes (easy) | Limited (requires margin account) |
| **Market Orders** | Execute immediately | Converted to MTL/ATO/ATC based on session |
| **Funding Rates** | Yes (perpetual futures) | No |
| **Position Modes** | One-way / Hedge | N/A (stock ownership) |
| **Time Zone** | UTC | ICT (UTC+7) |
| **WebSocket** | Real-time (sub-ms) | Not yet implemented (REST polling) |
| **API Rate Limits** | Published (e.g., 1200 req/min) | Not publicly documented |

---

## Code Pattern Differences

### 1. Initialization

#### Crypto (Binance)

```go
import "github.com/banbox/banexg/binance"

exg, err := binance.New(map[string]interface{}{
    "apiKey":    "your_api_key",
    "apiSecret": "your_api_secret",
})
```

#### Vietnam (SSI)

```go
import "github.com/banbox/banexg/vietnam"

exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "your_consumer_id",
    "consumerSecret": "your_consumer_secret",
})
```

**Key Difference**: Different credential types, but same pattern.

---

### 2. Symbol Format

#### Crypto

```go
// Spot
ticker, _ := exg.FetchTicker("BTC/USDT", nil)

// Futures
ticker, _ := exg.FetchTicker("BTC/USDT:USDT", nil)
```

#### Vietnam

```go
// Must include exchange prefix
ticker, _ := exg.FetchTicker("HOSE:VNM", nil)
ticker, _ := exg.FetchTicker("HNX:SHB", nil)
ticker, _ := exg.FetchTicker("UPCOM:VNG", nil)

// ❌ This won't work
ticker, _ := exg.FetchTicker("VNM", nil) // Missing exchange
```

**Key Difference**: Vietnam requires `EXCHANGE:CODE` format.

---

### 3. Creating Orders

#### Crypto

```go
// Market order - executes immediately
order, _ := exg.CreateOrder(
    "BTC/USDT",
    banexg.OdTypeMarket,
    banexg.OdSideBuy,
    0.001,    // Amount (any size)
    0,        // No price for market orders
    nil,
)

// Limit order
order, _ := exg.CreateOrder(
    "BTC/USDT",
    banexg.OdTypeLimit,
    banexg.OdSideBuy,
    0.001,       // Amount
    50000.50,    // Price (any precision)
    nil,
)
```

#### Vietnam

```go
// Market order - converts to session-appropriate type
order, _ := exg.CreateOrder(
    "HOSE:VNM",
    banexg.OdTypeMarket,  // → MTL/ATO/ATC based on time
    banexg.OdSideBuy,
    100,                   // Must be multiple of 100
    0,
    map[string]interface{}{
        "accountNo": "0001234567", // REQUIRED
    },
)

// Limit order
order, _ := exg.CreateOrder(
    "HOSE:VNM",
    banexg.OdTypeLimit,
    banexg.OdSideBuy,
    100,                    // Must be multiple of 100
    82500,                  // Must be at tick size (100 VND)
    map[string]interface{}{
        "accountNo": "0001234567", // REQUIRED
    },
)
```

**Key Differences**:
1. Vietnam requires `accountNo` parameter
2. Quantity must be multiples of 100 shares
3. Price must be at tick size (100 VND for most stocks)
4. Market orders are converted based on trading session

---

### 4. Market Hours Check

#### Crypto

```go
// Not needed - always open
ticker, _ := exg.FetchTicker("BTC/USDT", nil)
```

#### Vietnam

```go
import "time"

// Must check market hours
now := time.Now().In(time.FixedZone("ICT", 7*3600))
hour, min := now.Hour(), now.Minute()

isOpen := (hour >= 9 && hour < 11) || 
          (hour == 11 && min <= 30) ||
          (hour >= 13 && hour < 14) ||
          (hour == 14 && min <= 30)

if !isOpen {
    log.Fatal("Market is closed")
}

ticker, _ := exg.FetchTicker("HOSE:VNM", nil)
```

**Key Difference**: Vietnam market closes daily (lunch + after-hours).

---

### 5. Balance & Positions

#### Crypto

```go
// Balance shows crypto assets
balance, _ := exg.FetchBalance(nil)
fmt.Println(balance.Free["BTC"])  // Bitcoin balance
fmt.Println(balance.Free["USDT"]) // USDT balance

// Positions (futures only)
positions, _ := exg.FetchPositions(nil, nil)
for _, pos := range positions {
    fmt.Printf("%s: %.3f contracts\n", pos.Symbol, pos.Contracts)
}
```

#### Vietnam

```go
// Balance shows cash + stock value
balance, _ := exg.FetchBalance(map[string]interface{}{
    "accountNo": "0001234567", // REQUIRED
})
fmt.Println(balance.Free["VND"])              // Cash in VND
fmt.Println(balance.Info["totalStockValue"])  // Stock value

// Positions (stock holdings)
positions, _ := exg.FetchPositions(nil, map[string]interface{}{
    "accountNo": "0001234567", // REQUIRED
})
for _, pos := range positions {
    fmt.Printf("%s: %.0f shares\n", pos.Symbol, pos.Contracts)
}
```

**Key Differences**:
1. Vietnam requires `accountNo` for all account operations
2. Balance currency is VND (not crypto)
3. Positions show stock holdings (not futures contracts)

---

### 6. WebSocket Streaming

#### Crypto

```go
// Real-time ticker stream
tickerChan, _ := exg.WatchTicker("BTC/USDT", nil)

for ticker := range tickerChan {
    fmt.Printf("BTC: $%.2f\n", ticker.Last)
}
```

#### Vietnam

```go
// WebSocket NOT YET IMPLEMENTED
// Use REST API polling instead:

import "time"

ticker := time.NewTicker(5 * time.Second)
for range ticker.C {
    t, err := exg.FetchTicker("HOSE:VNM", nil)
    if err != nil {
        log.Println(err)
        continue
    }
    fmt.Printf("VNM: %.0f VND\n", t.Last)
}
```

**Key Difference**: Vietnam WebSocket not yet available, use polling.

---

## Trading Strategy Adaptations

### 1. Time Management

#### Crypto Strategy

```go
// 24/7 monitoring
for {
    ticker, _ := exg.FetchTicker("BTC/USDT", nil)
    
    if ticker.Last < 50000 {
        exg.CreateOrder("BTC/USDT", "market", "buy", 0.001, 0, nil)
    }
    
    time.Sleep(1 * time.Second)
}
```

#### Vietnam Strategy

```go
// Only trade during market hours
for {
    now := time.Now().In(time.FixedZone("ICT", 7*3600))
    
    // Skip if market closed
    if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
        time.Sleep(1 * time.Hour)
        continue
    }
    
    hour := now.Hour()
    if hour < 9 || hour >= 15 {
        time.Sleep(10 * time.Minute)
        continue
    }
    
    ticker, _ := exg.FetchTicker("HOSE:VNM", nil)
    
    if ticker.Last < 82000 {
        exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82000,
            map[string]interface{}{"accountNo": "0001234567"})
    }
    
    time.Sleep(5 * time.Second)
}
```

### 2. Price Precision

#### Crypto

```go
// Any precision allowed
buyPrice := ticker.Last * 0.99  // e.g., 49875.3245
exg.CreateOrder("BTC/USDT", "limit", "buy", 0.001, buyPrice, nil)
```

#### Vietnam

```go
// Must round to tick size
buyPrice := ticker.Last * 0.99
buyPrice = math.Round(buyPrice/100) * 100  // Round to nearest 100 VND
exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, buyPrice,
    map[string]interface{}{"accountNo": "0001234567"})
```

### 3. Position Sizing

#### Crypto

```go
// Fractional amounts allowed
portfolioValue := 10000.0  // $10k
allocation := 0.05         // 5%
amountUSD := portfolioValue * allocation  // $500

// Buy $500 worth of BTC
amount := amountUSD / ticker.Last  // e.g., 0.01 BTC
exg.CreateOrder("BTC/USDT", "market", "buy", amount, 0, nil)
```

#### Vietnam

```go
// Must use lot sizes (100 shares)
portfolioValue := 100000000.0  // 100M VND
allocation := 0.05              // 5%
amountVND := portfolioValue * allocation  // 5M VND

// Calculate shares (must be multiple of 100)
shares := amountVND / ticker.Last     // e.g., 60.6 shares
shares = math.Floor(shares/100) * 100 // Round down to 0 shares

if shares < 100 {
    log.Println("Amount too small for 1 lot")
    return
}

exg.CreateOrder("HOSE:VNM", "limit", "buy", shares, ticker.Last,
    map[string]interface{}{"accountNo": "0001234567"})
```

---

## Authentication Differences

### Crypto (HMAC Signing)

```
1. Every request signed with API Secret
2. Include timestamp and signature in headers
3. No token storage needed
4. Keys never expire (unless revoked)
```

**Example Header:**
```
X-MBX-APIKEY: your_api_key
signature: sha256(params + timestamp + secret)
timestamp: 1640995200000
```

### Vietnam (Bearer Token)

```
1. First call: Get access token (8-hour validity)
2. Subsequent calls: Include token in Authorization header
3. Library stores token internally
4. Auto-refreshes when expired
```

**Example Header:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**What this means for you:**
- Crypto: Each request is independent
- Vietnam: First call might be slower (token fetch), subsequent calls faster

---

## Error Handling Differences

### Crypto

```go
order, err := exg.CreateOrder(...)
if err != nil {
    // Common crypto errors:
    // - Insufficient balance
    // - Invalid symbol
    // - Rate limit exceeded
    // - Invalid signature
}
```

### Vietnam

```go
order, err := exg.CreateOrder(...)
if err != nil {
    // Additional Vietnam-specific errors:
    // - Market closed
    // - Price limit exceeded (±7%)
    // - Invalid tick size
    // - Invalid lot size
    // - Order type not allowed in current session
    // - Token expired (auto-refreshes)
}
```

---

## Migration Checklist

If you're migrating from crypto to Vietnam stocks:

- [ ] Change credential keys: `apiKey/apiSecret` → `consumerID/consumerSecret`
- [ ] Add `accountNo` to all trading operations
- [ ] Update symbol format: `"BTC/USDT"` → `"HOSE:VNM"`
- [ ] Add market hours check before trading
- [ ] Round prices to tick size (100 VND)
- [ ] Ensure quantities are multiples of 100 shares
- [ ] Change from WebSocket to REST polling (temporary)
- [ ] Update time zone: UTC → ICT (UTC+7)
- [ ] Remove leverage/margin logic (not applicable)
- [ ] Update balance currency: USDT/BTC → VND
- [ ] Add weekend/holiday handling

---

## Side-by-Side Example

### Crypto Trading Bot

```go
package main

import (
    "github.com/banbox/banexg/binance"
    "github.com/banbox/banexg"
)

func main() {
    exg, _ := binance.New(map[string]interface{}{
        "apiKey":    "xxx",
        "apiSecret": "yyy",
    })
    
    // 24/7 monitoring
    for {
        ticker, _ := exg.FetchTicker("BTC/USDT", nil)
        
        if ticker.Last < 50000 {
            exg.CreateOrder(
                "BTC/USDT",
                banexg.OdTypeMarket,
                banexg.OdSideBuy,
                0.001,  // Any amount
                0,
                nil,    // No extra params
            )
        }
    }
}
```

### Vietnam Trading Bot

```go
package main

import (
    "github.com/banbox/banexg/vietnam"
    "github.com/banbox/banexg"
    "time"
)

func main() {
    exg, _ := vietnam.New(map[string]interface{}{
        "consumerID":     "xxx",
        "consumerSecret": "yyy",
    })
    
    accountNo := "0001234567"
    
    // Market hours only
    for {
        now := time.Now().In(time.FixedZone("ICT", 7*3600))
        
        // Check market open
        if !isMarketOpen(now) {
            time.Sleep(10 * time.Minute)
            continue
        }
        
        ticker, _ := exg.FetchTicker("HOSE:VNM", nil)
        
        if ticker.Last < 82000 {
            // Round to tick size
            price := 82000
            
            exg.CreateOrder(
                "HOSE:VNM",
                banexg.OdTypeLimit,
                banexg.OdSideBuy,
                100,  // Multiple of 100
                float64(price),
                map[string]interface{}{
                    "accountNo": accountNo,  // Required
                },
            )
        }
        
        time.Sleep(5 * time.Second)
    }
}

func isMarketOpen(t time.Time) bool {
    if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
        return false
    }
    h, m := t.Hour(), t.Minute()
    return (h == 9 || h == 10 || (h == 11 && m <= 30)) ||
           (h == 13 || (h == 14 && m <= 30))
}
```

---

## Summary

**Main Takeaways:**

1. ✅ **Same library, similar patterns** - core methods work the same way
2. ⚠️ **Add `accountNo`** - required for all trading operations
3. ⏰ **Check market hours** - Vietnam market closes daily
4. 📏 **Respect constraints** - tick size (100 VND) and lot size (100 shares)
5. 🔐 **Different auth** - Bearer token instead of HMAC signing
6. 🇻🇳 **Symbol format** - use `EXCHANGE:CODE` pattern

**For detailed examples**, see:
- [QUICK_START.md](./QUICK_START.md) - Get started in 5 minutes
- [USER_GUIDE.md](./USER_GUIDE.md) - Complete guide with all features

---

**Good luck with your Vietnam stock trading! 🚀🇻🇳**
