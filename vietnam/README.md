# Vietnam Stock Market Module

This module provides integration with the Vietnam Stock Market (HOSE, HNX, UPCOM) via the SSI FastConnect API.

## 📚 Documentation

**→ See [INDEX.md](./INDEX.md) for complete documentation index**

**Quick links:**
- 🚀 **[QUICK_START.md](./QUICK_START.md)** - Get started in 5 minutes
- 📖 **[USER_GUIDE.md](./USER_GUIDE.md)** - Complete guide with all features
- 🔄 **[CRYPTO_VS_VIETNAM.md](./CRYPTO_VS_VIETNAM.md)** - For crypto traders
- 🔧 **[AGENTS.md](./AGENTS.md)** - Developer guide

**New to this library?** → Start with [QUICK_START.md](./QUICK_START.md)

---

## Features

### ✅ Implemented (REST API)
- **Market Data**
  - ✅ LoadMarkets - Fetch all available securities
  - ✅ FetchTicker - Get real-time quote for a single security
  - ✅ FetchTickers - Get multiple securities' daily prices
  - ✅ FetchOHLCV - Get candlestick/OHLC data (intraday + daily)
  - ❌ FetchOrderBook - Not yet implemented

- **Order Management**
  - ✅ CreateOrder - Place new orders (ATO, ATC, LO, MTL, MAK)
  - ✅ CancelOrder - Cancel existing orders
  - ✅ EditOrder - Modify pending orders
  - ✅ FetchOrder - Get order details
  - ✅ FetchOrders - Get order history
  - ✅ FetchOpenOrders - Get active orders

- **Account Management**
  - ✅ FetchBalance - Get account balance (cash + stock value)
  - ✅ FetchPositions - Get stock positions

- **Fee Calculation**
  - ✅ CalculateFee - Calculate trading fees (0.15% both maker/taker)

### ⏳ Pending Implementation
- ❌ WebSocket real-time data streams
- ❌ Vietnam-specific features (Financial reports, company info)
- ❌ Holiday calendar integration

## Installation

```go
import (
    "github.com/banbox/banexg/vietnam"
)
```

## Configuration

### 1. Create SSI Account
1. Register at [SSI](https://www.ssi.com.vn)
2. Request API access to get `consumerID` and `consumerSecret`
3. Save credentials in `local.json`

### 2. Configuration File (`local.json`)
```json
{
  "consumerID": "YOUR_CONSUMER_ID",
  "consumerSecret": "YOUR_CONSUMER_SECRET",
  "sandbox": false
}
```

## Usage Examples

### Initialize Exchange
```go
package main

import (
    "github.com/banbox/banexg/vietnam"
    "log"
)

func main() {
    exg, err := vietnam.New(map[string]interface{}{
        "consumerID":     "your_consumer_id",
        "consumerSecret": "your_consumer_secret",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### Load Markets
```go
markets, err := exg.LoadMarkets([]string{"spot"}, nil)
if err != nil {
    log.Fatal(err)
}

// Access specific market
vnmMarket := markets["HOSE:VNM"]
fmt.Printf("Symbol: %s, Name: %s\n", vnmMarket.Symbol, vnmMarket.Info["stockName"])
```

### Fetch Market Data
```go
// Single ticker
ticker, err := exg.FetchTicker("HOSE:VNM", nil)
fmt.Printf("Last: %.2f, Change: %.2f%%\n", ticker.Last, ticker.Percentage)

// Multiple tickers
tickers, err := exg.FetchTickers([]string{"HOSE:VNM", "HOSE:VCB"}, nil)

// OHLCV data
klines, err := exg.FetchOHLCV("HOSE:VNM", "1d", 0, 30, nil)
for _, k := range klines {
    fmt.Printf("Time: %d, Open: %.2f, Close: %.2f\n", k.Time, k.Open, k.Close)
}
```

### Place Orders
```go
// Limit Order (LO)
order, err := exg.CreateOrder(
    "HOSE:VNM",           // symbol
    banexg.OdTypeLimit,   // order type
    banexg.OdSideBuy,     // side
    100,                  // amount (quantity in shares)
    82500,                // price (VND)
    map[string]interface{}{
        "accountNo": "0001234567", // Your trading account number
    },
)

// Market Order (converts to MTL during trading hours)
order, err := exg.CreateOrder(
    "HOSE:VNM",
    banexg.OdTypeMarket,
    banexg.OdSideBuy,
    100,
    0, // price not needed for market orders
    map[string]interface{}{
        "accountNo": "0001234567",
    },
)

// ATO Order (At-The-Opening, pre-open session 08:30-09:00)
order, err := exg.CreateOrder(
    "HOSE:VNM",
    banexg.OdTypeMarket, // Will be converted to ATO during pre-open
    banexg.OdSideBuy,
    100,
    0,
    params,
)
```

### Cancel/Edit Orders
```go
// Cancel order
cancelled, err := exg.CancelOrder("ORDER_ID", "HOSE:VNM", map[string]interface{}{
    "accountNo": "0001234567",
})

// Edit order (modify price/quantity)
modified, err := exg.EditOrder(
    "HOSE:VNM",
    "ORDER_ID",
    "",      // side (not changed)
    200,     // new amount
    83000,   // new price
    map[string]interface{}{
        "accountNo": "0001234567",
    },
)
```

### Fetch Orders
```go
// Get order detail
order, err := exg.FetchOrder("ORDER_ID", "HOSE:VNM", map[string]interface{}{
    "accountNo": "0001234567",
})

// Get order history (all orders)
orders, err := exg.FetchOrders("", 0, 50, map[string]interface{}{
    "accountNo": "0001234567",
})

// Get open orders only
openOrders, err := exg.FetchOpenOrders("HOSE:VNM", 0, 0, map[string]interface{}{
    "accountNo": "0001234567",
})
```

### Account Management
```go
// Get balance
balance, err := exg.FetchBalance(map[string]interface{}{
    "accountNo": "0001234567",
})
fmt.Printf("Cash: %.0f VND\n", balance.Free["VND"])
fmt.Printf("Total Assets: %.0f VND\n", balance.Info["totalAsset"])

// Get positions
positions, err := exg.FetchPositions(nil, map[string]interface{}{
    "accountNo": "0001234567",
})
for _, pos := range positions {
    fmt.Printf("%s: %.0f shares, P&L: %.2f%%\n",
        pos.Symbol, pos.Contracts, pos.Percentage)
}
```

### Calculate Fees
```go
fee, err := exg.CalculateFee(
    "HOSE:VNM",
    banexg.OdTypeLimit,
    banexg.OdSideBuy,
    100,     // amount
    82500,   // price
    true,    // isMaker
    nil,
)
fmt.Printf("Fee: %.2f VND (%.2f%%)\n", fee.Cost, fee.Rate*100)
```

## Vietnam Stock Market Characteristics

### Trading Hours (ICT/UTC+7)
| Session | Time | Order Types Allowed |
|---------|------|---------------------|
| Pre-Open | 08:30-09:00 | ATO only |
| Morning | 09:00-11:30 | LO, MTL, MAK |
| Lunch Break | 11:30-13:00 | Market closed |
| Afternoon | 13:00-14:30 | LO, MTL, MAK |
| Closing | 14:30-14:45 | ATC only |

### Order Types
| Type | Description | When to Use |
|------|-------------|-------------|
| **ATO** | At-The-Opening | Pre-open auction (08:30-09:00) |
| **ATC** | At-The-Close | Closing auction (14:30-14:45) |
| **LO** | Limit Order | Standard limit order |
| **MTL** | Market-to-Limit | Executes at market, unfilled becomes limit |
| **MAK** | Market At Kill | Market order, unfilled cancels (HNX only) |

### Symbol Format
- **Format:** `EXCHANGE:CODE`
- **Examples:** 
  - `HOSE:VNM` (Vinamilk on HOSE)
  - `HNX:SHB` (SHB Bank on HNX)
  - `UPCOM:VNG` (VNG on UPCOM)

### Price Rules
- **Daily Limits:** ±7% from reference price
- **Tick Size:** Varies by price range
  - 0-10,000 VND: 10 VND
  - 10,000-50,000 VND: 50 VND
  - 50,000+ VND: 100 VND
- **Lot Size:** 100 shares per lot (default)

### Settlement
- **T+1.5** (transitioning to T+1 in Q3/2026)
- Cash settlement: T+2

### Fees (SSI Default)
- **Trading Fee:** 0.15% (both maker/taker)
- **VAT:** 10% on trading fee
- **Transfer Fee:** Separate, exchange-specific

## Important Notes

### 1. Trading Session Awareness
The module automatically handles session-aware order type mapping:
```go
// During pre-open (08:30-09:00)
OdTypeMarket → ATO

// During continuous trading (09:00-11:30, 13:00-14:30)
OdTypeMarket → MTL

// During closing (14:30-14:45)
OdTypeMarket → ATC
```

### 2. Price Validation
All limit orders are validated against:
- Reference price ±7% limits
- Tick size rules for the exchange

### 3. Quantity Validation
- Must be multiples of lot size (default 100)
- Validated before order placement

### 4. Account Number Requirement
All trading operations require `accountNo` parameter:
```go
params := map[string]interface{}{
    "accountNo": "0001234567", // Your SSI trading account
}
```

### 5. Time Zone
All timestamps are in **ICT (UTC+7)**, not UTC.

## Error Handling

```go
order, err := exg.CreateOrder(...)
if err != nil {
    switch err.Code {
    case errs.CodeParamRequired:
        // Missing required parameter
    case errs.CodeParamInvalid:
        // Invalid parameter (price out of range, etc.)
    case errs.CodeRunTime:
        // SSI API error
    case errs.CodeNotImplement:
        // Feature not yet implemented
    default:
        // Other errors
    }
    log.Printf("Error: %s", err.Message)
}
```

## Testing

### Unit Tests
```bash
cd vietnam/
go test -short -v
```

### Integration Tests (requires credentials)
```bash
# Create local.json with your credentials first
go test -v
```

## Development Status

**Current Version:** 0.1.0 (Alpha)

**Production Ready:**
- ✅ Market data fetching
- ✅ Order management (Create, Cancel, Edit)
- ✅ Account management
- ✅ Fee calculation

**Not Production Ready:**
- ❌ WebSocket streams
- ❌ OrderBook depth data
- ❌ Vietnam-specific features

## Contributing

See [AGENTS.md](./AGENTS.md) for detailed development guidelines.

## License

Same as parent project (banexg).

## Support

- **SSI API Documentation:** https://fc-data.ssi.com.vn/
- **SSI Support:** support@ssi.com.vn
- **Issues:** https://github.com/banbox/banexg/issues

---

**Disclaimer:** This is an unofficial integration. Always test thoroughly in a sandbox environment before live trading.
