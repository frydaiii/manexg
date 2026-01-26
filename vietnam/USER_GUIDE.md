# Vietnam Stock Market Integration - Complete User Guide

This guide walks you through everything you need to use the BanExg Vietnam module for trading on Vietnamese stock exchanges (HOSE, HNX, UPCOM) via SSI FastConnect API.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Getting SSI API Credentials](#getting-ssi-api-credentials)
3. [Installation & Setup](#installation--setup)
4. [Quick Start](#quick-start)
5. [Complete Usage Examples](#complete-usage-examples)
6. [Configuration Patterns](#configuration-patterns)
7. [Trading Operations Guide](#trading-operations-guide)
8. [Common Patterns & Best Practices](#common-patterns--best-practices)
9. [Error Handling](#error-handling)
10. [Testing](#testing)
11. [Troubleshooting](#troubleshooting)
12. [FAQ](#faq)

---

## Prerequisites

Before you can use this library, you need:

1. ✅ **SSI Trading Account** - Active trading account at SSI Securities Corporation
2. ✅ **SSI API Service Registration** - FC Trading and/or FC Data service subscription
3. ✅ **API Credentials** - `consumerID` and `consumerSecret` from SSI
4. ✅ **Go 1.21+** - For using this library

---

## Getting SSI API Credentials

### Step 1: Register for SSI API Service

**⚠️ Important**: This **cannot** be done online. You must:

**Option A: Visit SSI Branch**
- Go to any SSI Securities branch with your ID (CCCD/Passport)
- Request to register for **FC Trading** and **FC Data** API services
- Fill out the registration form

**Option B: Via Account Executive**
- Contact your SSI account executive
- Request API service registration
- Submit required documents (can be done via post)

**Requirements:**
- Must have an active SSI trading account
- Service costs: Contact SSI for pricing
- Service validity: **1 year** (must renew annually)

### Step 2: Receive Approval Email

After approval (usually 1-3 business days):
- You'll receive an email from SSI with a registration link
- Click the link to access SSI iBoard portal

### Step 3: Create API Keys on iBoard

1. Log into [SSI iBoard API Service Management](https://iboard.ssi.com.vn/support/api-service/management)
2. Navigate to **"Dịch vụ API"** (API Service) section
3. Click **"Create new connection key"**
4. Enter OTP verification code
5. **⚠️ CRITICAL**: The system will display three credentials **ONLY ONCE**:

```
ConsumerID:     XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
ConsumerSecret: YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY
PrivateKey:     -----BEGIN RSA PRIVATE KEY-----
                ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ
                -----END RSA PRIVATE KEY-----
```

**IMMEDIATELY save these credentials securely!** You cannot retrieve them later.

### Step 4: Store Credentials Safely

**For Development (local testing):**

Create `vietnam/local.json` in your project:

```json
{
  "consumerID": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "consumerSecret": "YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY",
  "sandbox": false
}
```

**For Production:**

Use environment variables or a secrets manager:

```bash
export SSI_CONSUMER_ID="XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
export SSI_CONSUMER_SECRET="YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY"
```

---

## Installation & Setup

### 1. Install the Library

```bash
go get github.com/banbox/banexg
```

### 2. Import in Your Code

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/banbox/banexg"
    "github.com/banbox/banexg/vietnam"
)
```

### 3. Initialize the Exchange

```go
func main() {
    // Method 1: Direct credentials
    exg, err := vietnam.New(map[string]interface{}{
        "consumerID":     "YOUR_CONSUMER_ID",
        "consumerSecret": "YOUR_CONSUMER_SECRET",
    })
    if err != nil {
        log.Fatal("Failed to initialize:", err.Message)
    }
    
    fmt.Println("✅ Connected to Vietnam Stock Market")
}
```

---

## Quick Start

### Complete Example: Your First Trade

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/banbox/banexg"
    "github.com/banbox/banexg/vietnam"
)

func main() {
    // 1. Initialize exchange
    exg, err := vietnam.New(map[string]interface{}{
        "consumerID":     os.Getenv("SSI_CONSUMER_ID"),
        "consumerSecret": os.Getenv("SSI_CONSUMER_SECRET"),
    })
    if err != nil {
        log.Fatal("Init failed:", err.Message)
    }
    
    // 2. Load markets (triggers authentication)
    markets, err := exg.LoadMarkets(false, nil)
    if err != nil {
        log.Fatal("LoadMarkets failed:", err.Message)
    }
    fmt.Printf("✅ Loaded %d markets\n", len(markets))
    
    // 3. Fetch current price
    ticker, err := exg.FetchTicker("HOSE:VNM", nil)
    if err != nil {
        log.Fatal("FetchTicker failed:", err.Message)
    }
    fmt.Printf("📊 VNM Price: %.0f VND (Change: %.2f%%)\n", 
        ticker.Last, ticker.Percentage)
    
    // 4. Get account balance
    balance, err := exg.FetchBalance(map[string]interface{}{
        "accountNo": "0001234567", // Your SSI trading account
    })
    if err != nil {
        log.Fatal("FetchBalance failed:", err.Message)
    }
    fmt.Printf("💰 Available Cash: %.0f VND\n", balance.Free["VND"])
    
    // 5. Place a limit order (BUY 100 shares of VNM at 82,500 VND)
    order, err := exg.CreateOrder(
        "HOSE:VNM",           // Symbol
        banexg.OdTypeLimit,   // Order type
        banexg.OdSideBuy,     // Side (buy/sell)
        100,                  // Quantity (shares)
        82500,                // Price (VND)
        map[string]interface{}{
            "accountNo": "0001234567",
        },
    )
    if err != nil {
        log.Fatal("CreateOrder failed:", err.Message)
    }
    
    fmt.Printf("✅ Order placed successfully!\n")
    fmt.Printf("   Order ID: %s\n", order.ID)
    fmt.Printf("   Status: %s\n", order.Status)
    fmt.Printf("   Symbol: %s\n", order.Symbol)
    fmt.Printf("   Side: %s\n", order.Side)
    fmt.Printf("   Amount: %.0f shares\n", order.Amount)
    fmt.Printf("   Price: %.0f VND\n", order.Price)
}
```

**Run it:**

```bash
export SSI_CONSUMER_ID="your_id"
export SSI_CONSUMER_SECRET="your_secret"
go run main.go
```

---

## Complete Usage Examples

### 1. Initialize Exchange (Multiple Methods)

#### Method A: Environment Variables (Recommended for Production)

```go
import "os"

exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     os.Getenv("SSI_CONSUMER_ID"),
    "consumerSecret": os.Getenv("SSI_CONSUMER_SECRET"),
})
```

#### Method B: Configuration File

```go
import (
    "github.com/banbox/banexg/utils"
)

func loadConfig() (map[string]interface{}, error) {
    var config map[string]interface{}
    err := utils.ReadJsonFile("vietnam/local.json", &config, utils.JsonNumDefault)
    return config, err
}

func main() {
    config, err := loadConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    exg, err := vietnam.New(config)
    if err != nil {
        log.Fatal(err)
    }
}
```

#### Method C: Factory Pattern (Multi-Exchange Apps)

```go
import "github.com/banbox/banexg/bex"

// Works with any exchange (binance, vietnam, bybit, okx)
exg, err := bex.New("vietnam", map[string]interface{}{
    "consumerID":     "...",
    "consumerSecret": "...",
})
```

### 2. Market Data Operations

#### Load All Markets

```go
// Load all available markets (HOSE, HNX, UPCOM)
markets, err := exg.LoadMarkets(false, nil)
if err != nil {
    log.Fatal(err)
}

// Access specific market
vnmMarket := markets["HOSE:VNM"]
fmt.Printf("Symbol: %s\n", vnmMarket.Symbol)
fmt.Printf("Name: %s\n", vnmMarket.Info["stockName"])
fmt.Printf("Exchange: %s\n", vnmMarket.Info["exchange"])
fmt.Printf("Lot Size: %.0f\n", vnmMarket.Limits.Amount.Min) // 100 shares

// List all symbols
for symbol := range markets {
    fmt.Println(symbol)
}
// Output:
// HOSE:VNM
// HOSE:VCB
// HOSE:VHM
// HNX:SHB
// UPCOM:VNG
// ...
```

#### Fetch Single Ticker (Real-time Quote)

```go
ticker, err := exg.FetchTicker("HOSE:VNM", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Symbol: %s\n", ticker.Symbol)
fmt.Printf("Last Price: %.0f VND\n", ticker.Last)
fmt.Printf("Change: %.0f VND (%.2f%%)\n", ticker.Change, ticker.Percentage)
fmt.Printf("Open: %.0f VND\n", ticker.Open)
fmt.Printf("High: %.0f VND\n", ticker.High)
fmt.Printf("Low: %.0f VND\n", ticker.Low)
fmt.Printf("Volume: %.0f shares\n", ticker.BaseVolume)
fmt.Printf("Value: %.0f VND\n", ticker.QuoteVolume)
fmt.Printf("Time: %d\n", ticker.Timestamp)
```

#### Fetch Multiple Tickers

```go
symbols := []string{"HOSE:VNM", "HOSE:VCB", "HNX:SHB"}
tickers, err := exg.FetchTickers(symbols, nil)
if err != nil {
    log.Fatal(err)
}

for _, ticker := range tickers {
    fmt.Printf("%s: %.0f VND (%.2f%%)\n", 
        ticker.Symbol, ticker.Last, ticker.Percentage)
}
// Output:
// HOSE:VNM: 82500 VND (+1.23%)
// HOSE:VCB: 95000 VND (-0.52%)
// HNX:SHB: 12300 VND (+2.08%)
```

#### Fetch OHLCV Data (Candlesticks)

```go
// Get 30 days of daily candles for VNM
klines, err := exg.FetchOHLCV("HOSE:VNM", "1d", 0, 30, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Got %d candles\n", len(klines))
for _, k := range klines {
    fmt.Printf("Date: %s | O: %.0f | H: %.0f | L: %.0f | C: %.0f | V: %.0f\n",
        time.Unix(k.Time/1000, 0).Format("2006-01-02"),
        k.Open, k.High, k.Low, k.Close, k.Volume)
}

// Supported timeframes:
// - "1m", "5m", "15m", "30m" (intraday)
// - "1h", "4h" (intraday)
// - "1d" (daily)
```

### 3. Order Management

#### Place Limit Order

```go
// BUY 100 shares of VNM at 82,500 VND
order, err := exg.CreateOrder(
    "HOSE:VNM",           // Symbol (EXCHANGE:CODE format)
    banexg.OdTypeLimit,   // Order type
    banexg.OdSideBuy,     // Side
    100,                  // Amount (shares)
    82500,                // Price (VND)
    map[string]interface{}{
        "accountNo": "0001234567", // REQUIRED: Your trading account
    },
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order ID: %s\n", order.ID)
fmt.Printf("Status: %s\n", order.Status) // "open", "filled", "canceled"
```

#### Place Market Order (Session-Aware)

```go
// During trading hours (09:00-11:30, 13:00-14:30):
// Market order becomes MTL (Market-To-Limit)
order, err := exg.CreateOrder(
    "HOSE:VNM",
    banexg.OdTypeMarket,  // Automatically mapped to MTL/ATO/ATC
    banexg.OdSideBuy,
    100,
    0,                     // Price not needed for market orders
    map[string]interface{}{
        "accountNo": "0001234567",
    },
)

// Library automatically handles session-aware mapping:
// - Pre-open (08:30-09:00) → ATO (At The Opening)
// - Trading (09:00-14:30) → MTL (Market To Limit)
// - Closing (14:30-14:45) → ATC (At The Close)
```

#### Place ATO Order (At-The-Opening)

```go
// ATO orders only allowed during pre-open (08:30-09:00)
order, err := exg.CreateOrder(
    "HOSE:VNM",
    banexg.OdTypeMarket,  // Will be converted to ATO in pre-open session
    banexg.OdSideBuy,
    100,
    0,
    map[string]interface{}{
        "accountNo": "0001234567",
    },
)
```

#### Cancel Order

```go
cancelled, err := exg.CancelOrder(
    "ORDER_ID",           // Order ID from CreateOrder response
    "HOSE:VNM",           // Symbol
    map[string]interface{}{
        "accountNo": "0001234567",
    },
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order %s cancelled successfully\n", cancelled.ID)
```

#### Modify Order (Change Price/Quantity)

```go
// Change order to 200 shares at 83,000 VND
modified, err := exg.EditOrder(
    "HOSE:VNM",
    "ORDER_ID",
    "",                   // Side (empty = no change)
    200,                  // New amount
    83000,                // New price
    map[string]interface{}{
        "accountNo": "0001234567",
    },
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order modified: %s\n", modified.ID)
```

#### Fetch Order Details

```go
order, err := exg.FetchOrder(
    "ORDER_ID",
    "HOSE:VNM",
    map[string]interface{}{
        "accountNo": "0001234567",
    },
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order Status: %s\n", order.Status)
fmt.Printf("Filled: %.0f / %.0f shares\n", order.Filled, order.Amount)
fmt.Printf("Average Price: %.0f VND\n", order.Average)
```

#### Fetch Order History

```go
// Get last 50 orders (all symbols)
orders, err := exg.FetchOrders("", 0, 50, map[string]interface{}{
    "accountNo": "0001234567",
})
if err != nil {
    log.Fatal(err)
}

for _, order := range orders {
    fmt.Printf("%s | %s | %s | %.0f shares @ %.0f VND | Status: %s\n",
        time.Unix(order.Timestamp/1000, 0).Format("2006-01-02"),
        order.Symbol,
        order.Side,
        order.Amount,
        order.Price,
        order.Status)
}
```

#### Fetch Open Orders Only

```go
// Get all open orders for VNM
openOrders, err := exg.FetchOpenOrders("HOSE:VNM", 0, 0, map[string]interface{}{
    "accountNo": "0001234567",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("You have %d open orders for VNM\n", len(openOrders))
```

### 4. Account Management

#### Fetch Account Balance

```go
balance, err := exg.FetchBalance(map[string]interface{}{
    "accountNo": "0001234567",
})
if err != nil {
    log.Fatal(err)
}

// Cash balances
fmt.Printf("Available Cash: %.0f VND\n", balance.Free["VND"])
fmt.Printf("Total Cash: %.0f VND\n", balance.Total["VND"])

// Account summary (from Info field)
fmt.Printf("Total Assets: %.0f VND\n", balance.Info["totalAsset"])
fmt.Printf("Stock Value: %.0f VND\n", balance.Info["totalStockValue"])
fmt.Printf("Purchasing Power: %.0f VND\n", balance.Info["purchasingPower"])
```

#### Fetch Stock Positions

```go
positions, err := exg.FetchPositions(nil, map[string]interface{}{
    "accountNo": "0001234567",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("You hold %d stock positions\n\n", len(positions))

for _, pos := range positions {
    fmt.Printf("Symbol: %s\n", pos.Symbol)
    fmt.Printf("  Quantity: %.0f shares\n", pos.Contracts)
    fmt.Printf("  Available: %.0f shares\n", pos.ContractsAvailable)
    fmt.Printf("  Avg Cost: %.0f VND\n", pos.EntryPrice)
    fmt.Printf("  Market Price: %.0f VND\n", pos.MarkPrice)
    fmt.Printf("  Market Value: %.0f VND\n", pos.Notional)
    fmt.Printf("  P&L: %.0f VND (%.2f%%)\n", pos.UnrealizedPnl, pos.Percentage)
    fmt.Println()
}
```

### 5. Fee Calculation

```go
// Calculate fee for buying 100 shares of VNM at 82,500 VND
fee, err := exg.CalculateFee(
    "HOSE:VNM",
    banexg.OdTypeLimit,
    banexg.OdSideBuy,
    100,                  // Amount
    82500,                // Price
    true,                 // isMaker (true for limit orders)
    nil,
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Trading Fee: %.2f VND (%.2f%%)\n", fee.Cost, fee.Rate*100)
fmt.Printf("Total Cost: %.0f VND\n", 100*82500+fee.Cost)

// Default SSI fees:
// - Trading: 0.15% (both maker/taker)
// - VAT: 10% on trading fee
// - Transfer fee: Separate (exchange-specific)
```

---

## Configuration Patterns

### 1. Multi-Account Setup

```go
// If you have multiple trading accounts under one SSI credential
exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "YOUR_ID",
    "consumerSecret": "YOUR_SECRET",
})

// Use different accounts for different operations
exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, map[string]interface{}{
    "accountNo": "0001234567", // Account 1
})

exg.CreateOrder("HOSE:VCB", "limit", "buy", 100, 95000, map[string]interface{}{
    "accountNo": "0009876543", // Account 2
})
```

### 2. Sandbox vs Production

```go
// For testing with SSI sandbox environment
exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "SANDBOX_CONSUMER_ID",
    "consumerSecret": "SANDBOX_CONSUMER_SECRET",
    "sandbox":        true, // Use sandbox endpoints
})

// For production trading
exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "PROD_CONSUMER_ID",
    "consumerSecret": "PROD_CONSUMER_SECRET",
    "sandbox":        false, // Default
})
```

### 3. Proxy Configuration

```go
exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "YOUR_ID",
    "consumerSecret": "YOUR_SECRET",
    "proxy":          "http://127.0.0.1:7890", // HTTP proxy
})
```

---

## Trading Operations Guide

### Understanding Vietnam Stock Market

#### Trading Hours (ICT/UTC+7)

| Session | Time (ICT) | Order Types Allowed |
|---------|------------|---------------------|
| **Pre-Open** | 08:30-09:00 | ATO only |
| **Morning** | 09:00-11:30 | LO, MTL, MAK |
| **Lunch Break** | 11:30-13:00 | ❌ Market closed |
| **Afternoon** | 13:00-14:30 | LO, MTL, MAK |
| **Closing** | 14:30-14:45 | ATC only |

#### Order Types

| Type | Full Name | Description | When to Use |
|------|-----------|-------------|-------------|
| **ATO** | At-The-Opening | Auction order for opening price | Pre-open (08:30-09:00) |
| **ATC** | At-The-Close | Auction order for closing price | Closing (14:30-14:45) |
| **LO** | Limit Order | Standard limit order | Anytime during trading |
| **MTL** | Market-to-Limit | Execute at market, unfilled becomes limit | Continuous trading |
| **MAK** | Market At Kill | Market order, unfilled cancels | HNX only |

#### Symbol Format

- **Format**: `EXCHANGE:CODE`
- **Examples**:
  - `HOSE:VNM` - Vinamilk on Ho Chi Minh Stock Exchange
  - `HNX:SHB` - SHB Bank on Hanoi Stock Exchange
  - `UPCOM:VNG` - VNG on Unlisted Public Company Market

#### Price Rules

1. **Daily Price Limits**: ±7% from reference price
2. **Tick Size** (minimum price movement):
   - 0-10,000 VND: 10 VND
   - 10,000-50,000 VND: 50 VND
   - 50,000+ VND: 100 VND
3. **Lot Size**: 100 shares per lot (default minimum)

#### Settlement

- **Trade Date (T)**: Day you place the order
- **Settlement**: T+1.5 (transitioning to T+1 in 2026)
- **Cash Settlement**: T+2

### Session-Aware Order Placement

The library **automatically** handles session-aware order type mapping:

```go
// Example: Placing a "market" order

// If current time is 08:45 (pre-open)
exg.CreateOrder("HOSE:VNM", banexg.OdTypeMarket, ...) 
// → Converts to ATO (At The Opening)

// If current time is 10:00 (morning session)
exg.CreateOrder("HOSE:VNM", banexg.OdTypeMarket, ...)
// → Converts to MTL (Market To Limit)

// If current time is 14:35 (closing session)
exg.CreateOrder("HOSE:VNM", banexg.OdTypeMarket, ...)
// → Converts to ATC (At The Close)
```

### Price and Quantity Validation

The library automatically validates:

```go
// ❌ This will fail - price not at tick size
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82555, ...)
// Error: "price must be multiple of tick size (100 VND)"

// ✅ This will succeed - price at valid tick
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, ...)

// ❌ This will fail - quantity not multiple of lot size
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 50, 82500, ...)
// Error: "quantity must be multiple of lot size (100)"

// ✅ This will succeed - quantity is multiple of 100
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, ...)
```

---

## Common Patterns & Best Practices

### 1. Always Check Market Hours Before Trading

```go
import "time"

func isMarketOpen() bool {
    now := time.Now().In(time.FixedZone("ICT", 7*3600)) // Vietnam time
    hour, min, _ := now.Hour(), now.Minute(), 0
    
    // Morning: 09:00-11:30
    if hour == 9 || (hour == 10) || (hour == 11 && min <= 30) {
        return true
    }
    
    // Afternoon: 13:00-14:30
    if hour == 13 || (hour == 14 && min <= 30) {
        return true
    }
    
    return false
}

func main() {
    if !isMarketOpen() {
        log.Fatal("Market is closed")
    }
    
    // Place orders...
}
```

### 2. Handle Token Expiry Gracefully

```go
// The library auto-refreshes tokens, but you should handle errors:

ticker, err := exg.FetchTicker("HOSE:VNM", nil)
if err != nil {
    if err.Code == errs.CodeUnauthorized {
        log.Println("⚠️ Token expired, library will auto-refresh on next call")
        // Retry once
        ticker, err = exg.FetchTicker("HOSE:VNM", nil)
    }
}
```

### 3. Use Limit Orders with Conservative Pricing

```go
// Get current price
ticker, _ := exg.FetchTicker("HOSE:VNM", nil)

// Buy at 0.5% below current price (more likely to fill)
buyPrice := ticker.Last * 0.995
buyPrice = roundToTickSize(buyPrice, 100) // Round to nearest 100 VND

order, err := exg.CreateOrder(
    "HOSE:VNM",
    banexg.OdTypeLimit,
    banexg.OdSideBuy,
    100,
    buyPrice,
    map[string]interface{}{"accountNo": "0001234567"},
)
```

### 4. Always Specify accountNo

```go
// ❌ BAD - Missing accountNo
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, nil)
// Error: "accountNo is required"

// ✅ GOOD - accountNo provided
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, 
    map[string]interface{}{
        "accountNo": "0001234567",
    })
```

### 5. Monitor Open Orders Regularly

```go
import "time"

func monitorOrders(exg *vietnam.Vietnam, accountNo string) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        orders, err := exg.FetchOpenOrders("", 0, 0, map[string]interface{}{
            "accountNo": accountNo,
        })
        if err != nil {
            log.Printf("Error fetching orders: %v", err)
            continue
        }
        
        log.Printf("Open orders: %d", len(orders))
        for _, order := range orders {
            log.Printf("  %s | %s | %.0f @ %.0f | %s",
                order.Symbol, order.Side, order.Amount, order.Price, order.Status)
        }
    }
}
```

### 6. Calculate Total Cost Including Fees

```go
func calculateTotalCost(exg *vietnam.Vietnam, symbol string, quantity float64, price float64) (float64, error) {
    // Calculate fee
    fee, err := exg.CalculateFee(symbol, banexg.OdTypeLimit, banexg.OdSideBuy, quantity, price, true, nil)
    if err != nil {
        return 0, err
    }
    
    // Total cost = (quantity * price) + fee
    totalCost := (quantity * price) + fee.Cost
    
    fmt.Printf("Breakdown:\n")
    fmt.Printf("  Shares: %.0f @ %.0f VND = %.0f VND\n", quantity, price, quantity*price)
    fmt.Printf("  Fee: %.2f VND (%.2f%%)\n", fee.Cost, fee.Rate*100)
    fmt.Printf("  Total: %.0f VND\n", totalCost)
    
    return totalCost, nil
}

// Usage:
total, err := calculateTotalCost(exg, "HOSE:VNM", 100, 82500)
// Output:
// Breakdown:
//   Shares: 100 @ 82500 VND = 8250000 VND
//   Fee: 12375.00 VND (0.15%)
//   Total: 8262375 VND
```

---

## Error Handling

### Common Error Codes

```go
import "github.com/banbox/banexg/errs"

order, err := exg.CreateOrder(...)
if err != nil {
    switch err.Code {
    case errs.CodeParamRequired:
        // Missing required parameter (e.g., accountNo)
        log.Printf("Missing parameter: %s", err.Message)
        
    case errs.CodeParamInvalid:
        // Invalid parameter (price out of range, wrong quantity, etc.)
        log.Printf("Invalid parameter: %s", err.Message)
        
    case errs.CodeUnauthorized:
        // Authentication failed or token expired
        log.Printf("Auth error: %s", err.Message)
        
    case errs.CodeRunTime:
        // SSI API error (market closed, insufficient balance, etc.)
        log.Printf("SSI API error: %s", err.Message)
        
    case errs.CodeNotImplement:
        // Feature not yet implemented (e.g., WebSocket)
        log.Printf("Not implemented: %s", err.Message)
        
    default:
        log.Printf("Unknown error: %s", err.Message)
    }
}
```

### Retry Logic Example

```go
func createOrderWithRetry(exg *vietnam.Vietnam, symbol string, maxRetries int) (*banexg.Order, error) {
    var order *banexg.Order
    var err *errs.Error
    
    for i := 0; i < maxRetries; i++ {
        order, err = exg.CreateOrder(
            symbol,
            banexg.OdTypeLimit,
            banexg.OdSideBuy,
            100,
            82500,
            map[string]interface{}{"accountNo": "0001234567"},
        )
        
        if err == nil {
            return order, nil
        }
        
        // Don't retry on parameter errors
        if err.Code == errs.CodeParamRequired || err.Code == errs.CodeParamInvalid {
            return nil, err
        }
        
        log.Printf("Attempt %d/%d failed: %s", i+1, maxRetries, err.Message)
        time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
    }
    
    return nil, err
}
```

---

## Testing

### Unit Tests (No Credentials Required)

```bash
cd /home/manht/manexg/vietnam
go test -short -v
```

Output:
```
=== RUN   TestVietnamNew
--- PASS: TestVietnamNew (0.04s)
=== RUN   TestVietnamAuth
    base_test.go:54: Skipping integration test in short mode
--- SKIP: TestVietnamAuth (0.00s)
PASS
```

### Integration Tests (Requires Credentials)

1. Create `vietnam/local.json`:

```json
{
  "consumerID": "YOUR_CONSUMER_ID",
  "consumerSecret": "YOUR_CONSUMER_SECRET",
  "sandbox": false
}
```

2. Run full tests:

```bash
cd /home/manht/manexg/vietnam
go test -v
```

### Writing Your Own Tests

```go
package main

import (
    "testing"
    "github.com/banbox/banexg/vietnam"
)

func TestMyTrading(t *testing.T) {
    exg, err := vietnam.New(map[string]interface{}{
        "consumerID":     "test-id",
        "consumerSecret": "test-secret",
    })
    if err != nil {
        t.Fatal(err)
    }
    
    // Test market loading
    markets, err := exg.LoadMarkets(false, nil)
    if err != nil {
        t.Fatal(err)
    }
    
    if len(markets) == 0 {
        t.Error("Expected markets to be loaded")
    }
}
```

---

## Troubleshooting

### Issue 1: "consumerID and consumerSecret are required"

**Cause**: Missing credentials in initialization.

**Solution**:
```go
// ❌ Wrong
exg, err := vietnam.New(map[string]interface{}{})

// ✅ Correct
exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "YOUR_ID",
    "consumerSecret": "YOUR_SECRET",
})
```

### Issue 2: "empty access token received"

**Cause**: Invalid credentials or SSI service expired.

**Solutions**:
1. Verify credentials are correct
2. Check if SSI API service subscription is still active (renew if expired)
3. Test credentials on SSI iBoard portal first

### Issue 3: "accountNo is required"

**Cause**: Missing accountNo parameter in trading operations.

**Solution**:
```go
// ❌ Wrong
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, nil)

// ✅ Correct
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, 
    map[string]interface{}{
        "accountNo": "0001234567",
    })
```

### Issue 4: "Market is closed"

**Cause**: Trying to trade outside trading hours.

**Trading Hours (ICT/UTC+7)**:
- Monday-Friday: 09:00-11:30, 13:00-14:30
- Closed: Weekends, public holidays, 11:30-13:00 (lunch)

**Solution**: Check current time in ICT timezone before trading.

### Issue 5: "price must be multiple of tick size"

**Cause**: Price not at valid tick increment.

**Tick Size Rules**:
- 0-10,000 VND: 10 VND increments
- 10,000-50,000 VND: 50 VND increments
- 50,000+ VND: 100 VND increments

**Solution**:
```go
// ❌ Wrong - 82,555 not valid (should be multiple of 100)
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82555, ...)

// ✅ Correct - 82,500 is valid
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, ...)
```

### Issue 6: "quantity must be multiple of lot size"

**Cause**: Quantity not a multiple of 100 shares.

**Solution**:
```go
// ❌ Wrong
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 50, 82500, ...)

// ✅ Correct
order, err := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, ...)
```

### Issue 7: Token expired after 8 hours

**Cause**: SSI tokens expire after 8 hours.

**Solution**: The library auto-refreshes tokens. No action needed. If you get auth errors:
```go
// Just retry the operation - token will auto-refresh
ticker, err := exg.FetchTicker("HOSE:VNM", nil)
if err != nil && err.Code == errs.CodeUnauthorized {
    // Retry once
    ticker, err = exg.FetchTicker("HOSE:VNM", nil)
}
```

---

## FAQ

### Q1: Do I need a VPN to access SSI API?

**A**: No, SSI API is accessible globally. However, you can configure a proxy if needed:

```go
exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "...",
    "consumerSecret": "...",
    "proxy":          "http://127.0.0.1:7890",
})
```

### Q2: Can I use this library for paper trading / simulation?

**A**: SSI provides a sandbox environment. Set `"sandbox": true` in options:

```go
exg, err := vietnam.New(map[string]interface{}{
    "consumerID":     "SANDBOX_ID",
    "consumerSecret": "SANDBOX_SECRET",
    "sandbox":        true, // Use sandbox endpoints
})
```

### Q3: What's the difference between FetchPositions and FetchAccountPositions?

**A**: In this implementation, both return the same data (stock positions). Use either:

```go
positions, err := exg.FetchPositions(nil, map[string]interface{}{
    "accountNo": "0001234567",
})
```

### Q4: Can I trade on margin?

**A**: This library currently supports **cash trading only**. Margin trading requires additional SSI service subscription and is not yet implemented.

### Q5: How do I get historical data for backtesting?

```go
// Get 1 year of daily data (252 trading days)
klines, err := exg.FetchOHLCV("HOSE:VNM", "1d", 0, 252, nil)

// Process for backtesting
for _, k := range klines {
    // Your backtesting logic
    fmt.Printf("Date: %s | Close: %.0f\n", 
        time.Unix(k.Time/1000, 0).Format("2006-01-02"), k.Close)
}
```

### Q6: Does this support real-time WebSocket data?

**A**: WebSocket support is stubbed but not yet implemented. Currently, use REST API polling:

```go
import "time"

ticker := time.NewTicker(5 * time.Second)
for range ticker.C {
    ticker, err := exg.FetchTicker("HOSE:VNM", nil)
    if err != nil {
        log.Println(err)
        continue
    }
    fmt.Printf("VNM: %.0f VND\n", ticker.Last)
}
```

### Q7: How do I handle multiple accounts?

```go
accounts := []string{"0001234567", "0009876543"}

for _, accountNo := range accounts {
    balance, err := exg.FetchBalance(map[string]interface{}{
        "accountNo": accountNo,
    })
    if err != nil {
        log.Printf("Account %s error: %v", accountNo, err)
        continue
    }
    
    fmt.Printf("Account %s: %.0f VND\n", accountNo, balance.Free["VND"])
}
```

### Q8: What's the rate limit?

**A**: SSI doesn't publicly document rate limits. Best practices:
- Use WebSocket streaming when available (future)
- For REST API, limit to 1-2 requests per second per endpoint
- If you hit rate limits, contact SSI to increase quota

### Q9: How do I renew my API subscription?

**A**: SSI sends email notification 7 days before expiration. To renew:
1. Log into SSI iBoard portal
2. Navigate to API Service section
3. Click "Renew subscription"
4. Pay renewal fee (contact SSI for pricing)

Your existing credentials continue to work after renewal.

### Q10: Can I use this in production?

**A**: YES for REST API operations:
- ✅ Market data fetching
- ✅ Order management (Create, Cancel, Edit)
- ✅ Account management
- ✅ Fee calculation

NOT YET for:
- ❌ WebSocket real-time streams
- ❌ OrderBook depth data

---

## Additional Resources

### Official SSI Documentation
- **API Reference**: https://fc-data.ssi.com.vn/Help
- **Connection Guide**: https://guide.ssi.com.vn/ssi-products/fastconnect-trading/connection-guide
- **iBoard Portal**: https://iboard.ssi.com.vn/support/api-service/management

### SSI Official GitHub Examples
- **Python FC Data**: https://github.com/SSI-Securities-Corporation/python-fcdata
- **Python FC Trading**: https://github.com/SSI-Securities-Corporation/python-fctrading
- **Node.js FC Data**: https://github.com/SSI-Securities-Corporation/node-fcdata
- **Node.js FC Trading**: https://github.com/SSI-Securities-Corporation/node-fctrading

### BanExg Library
- **Main Repository**: https://github.com/banbox/banexg
- **Issues**: https://github.com/banbox/banexg/issues
- **Developer Guide**: [AGENTS.md](./AGENTS.md)

### Contact
- **SSI Support**: support@ssi.com.vn
- **SSI Hotline**: 1900 9088

---

## Appendix: Complete Working Example

Here's a complete, production-ready example application:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/banbox/banexg"
    "github.com/banbox/banexg/vietnam"
)

func main() {
    // Configuration
    consumerID := os.Getenv("SSI_CONSUMER_ID")
    consumerSecret := os.Getenv("SSI_CONSUMER_SECRET")
    accountNo := os.Getenv("SSI_ACCOUNT_NO")
    
    if consumerID == "" || consumerSecret == "" || accountNo == "" {
        log.Fatal("Required environment variables: SSI_CONSUMER_ID, SSI_CONSUMER_SECRET, SSI_ACCOUNT_NO")
    }
    
    // 1. Initialize
    log.Println("🔌 Connecting to Vietnam Stock Market...")
    exg, err := vietnam.New(map[string]interface{}{
        "consumerID":     consumerID,
        "consumerSecret": consumerSecret,
    })
    if err != nil {
        log.Fatal("❌ Init failed:", err.Message)
    }
    log.Println("✅ Connected successfully")
    
    // 2. Load Markets
    log.Println("\n📋 Loading markets...")
    markets, err := exg.LoadMarkets(false, nil)
    if err != nil {
        log.Fatal("❌ LoadMarkets failed:", err.Message)
    }
    log.Printf("✅ Loaded %d markets\n", len(markets))
    
    // 3. Check Account Balance
    log.Println("\n💰 Fetching account balance...")
    balance, err := exg.FetchBalance(map[string]interface{}{
        "accountNo": accountNo,
    })
    if err != nil {
        log.Fatal("❌ FetchBalance failed:", err.Message)
    }
    
    cash := balance.Free["VND"]
    totalAssets := balance.Info["totalAsset"].(float64)
    stockValue := balance.Info["totalStockValue"].(float64)
    
    log.Printf("✅ Account Balance:")
    log.Printf("   Cash: %.0f VND", cash)
    log.Printf("   Stock Value: %.0f VND", stockValue)
    log.Printf("   Total Assets: %.0f VND", totalAssets)
    
    // 4. Check Stock Positions
    log.Println("\n📊 Fetching stock positions...")
    positions, err := exg.FetchPositions(nil, map[string]interface{}{
        "accountNo": accountNo,
    })
    if err != nil {
        log.Fatal("❌ FetchPositions failed:", err.Message)
    }
    
    if len(positions) == 0 {
        log.Println("   No stock positions")
    } else {
        log.Printf("✅ You hold %d positions:", len(positions))
        for _, pos := range positions {
            log.Printf("   %s: %.0f shares @ %.0f VND (P&L: %.2f%%)",
                pos.Symbol, pos.Contracts, pos.EntryPrice, pos.Percentage)
        }
    }
    
    // 5. Fetch Market Data for VNM
    log.Println("\n📈 Fetching VNM market data...")
    ticker, err := exg.FetchTicker("HOSE:VNM", nil)
    if err != nil {
        log.Fatal("❌ FetchTicker failed:", err.Message)
    }
    
    log.Printf("✅ VNM Quote:")
    log.Printf("   Last Price: %.0f VND", ticker.Last)
    log.Printf("   Change: %.0f VND (%.2f%%)", ticker.Change, ticker.Percentage)
    log.Printf("   Open: %.0f | High: %.0f | Low: %.0f", ticker.Open, ticker.High, ticker.Low)
    log.Printf("   Volume: %.0f shares", ticker.BaseVolume)
    
    // 6. Check if market is open
    log.Println("\n🕐 Checking market hours...")
    now := time.Now().In(time.FixedZone("ICT", 7*3600))
    hour, min := now.Hour(), now.Minute()
    
    isOpen := (hour == 9 || hour == 10 || (hour == 11 && min <= 30)) ||
              (hour == 13 || (hour == 14 && min <= 30))
    
    if !isOpen {
        log.Printf("⏸️  Market is currently CLOSED (Current time: %02d:%02d ICT)", hour, min)
        log.Println("   Trading hours: 09:00-11:30, 13:00-14:30 (Mon-Fri)")
        log.Println("\n✅ All checks completed successfully!")
        return
    }
    
    log.Printf("✅ Market is OPEN (Current time: %02d:%02d ICT)", hour, min)
    
    // 7. Check open orders
    log.Println("\n📝 Fetching open orders...")
    openOrders, err := exg.FetchOpenOrders("", 0, 0, map[string]interface{}{
        "accountNo": accountNo,
    })
    if err != nil {
        log.Fatal("❌ FetchOpenOrders failed:", err.Message)
    }
    
    if len(openOrders) == 0 {
        log.Println("   No open orders")
    } else {
        log.Printf("✅ You have %d open orders:", len(openOrders))
        for _, order := range openOrders {
            log.Printf("   %s | %s | %.0f shares @ %.0f VND | Status: %s",
                order.Symbol, order.Side, order.Amount, order.Price, order.Status)
        }
    }
    
    // 8. Example: Place a limit order (commented out for safety)
    /*
    log.Println("\n🛒 Placing test order...")
    order, err := exg.CreateOrder(
        "HOSE:VNM",
        banexg.OdTypeLimit,
        banexg.OdSideBuy,
        100,
        ticker.Last * 0.99, // 1% below current price
        map[string]interface{}{
            "accountNo": accountNo,
        },
    )
    if err != nil {
        log.Printf("❌ CreateOrder failed: %s", err.Message)
    } else {
        log.Printf("✅ Order placed successfully!")
        log.Printf("   Order ID: %s", order.ID)
        log.Printf("   Status: %s", order.Status)
    }
    */
    
    log.Println("\n✅ All operations completed successfully!")
    log.Println("\n💡 Tip: Uncomment the CreateOrder section to place real orders")
}
```

**Run it:**

```bash
export SSI_CONSUMER_ID="your_consumer_id"
export SSI_CONSUMER_SECRET="your_consumer_secret"
export SSI_ACCOUNT_NO="0001234567"
go run main.go
```

**Expected Output:**

```
🔌 Connecting to Vietnam Stock Market...
✅ Connected successfully

📋 Loading markets...
✅ Loaded 1,247 markets

💰 Fetching account balance...
✅ Account Balance:
   Cash: 50000000 VND
   Stock Value: 123456789 VND
   Total Assets: 173456789 VND

📊 Fetching stock positions...
✅ You hold 3 positions:
   HOSE:VNM: 1000 shares @ 81500 VND (P&L: +1.23%)
   HOSE:VCB: 500 shares @ 94000 VND (P&L: +1.06%)
   HNX:SHB: 2000 shares @ 12000 VND (P&L: +2.50%)

📈 Fetching VNM market data...
✅ VNM Quote:
   Last Price: 82500 VND
   Change: 1000 VND (+1.23%)
   Open: 81500 | High: 82800 | Low: 81200
   Volume: 1234567 shares

🕐 Checking market hours...
✅ Market is OPEN (Current time: 10:30 ICT)

📝 Fetching open orders...
   No open orders

✅ All operations completed successfully!

💡 Tip: Uncomment the CreateOrder section to place real orders
```

---

**🎉 Congratulations!** You now have a complete guide to using the BanExg Vietnam module.

For questions or issues, please open an issue at: https://github.com/banbox/banexg/issues

**Happy Trading! 📈🇻🇳**
