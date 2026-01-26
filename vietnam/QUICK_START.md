# Vietnam Module - Quick Start Guide

**5-Minute Setup | Get Trading Fast**

---

## 📋 Prerequisites Checklist

- [ ] SSI trading account
- [ ] API credentials (consumerID + consumerSecret) from SSI
- [ ] Trading account number (e.g., "0001234567")
- [ ] Go 1.21+

**Don't have API credentials?** See [USER_GUIDE.md](./USER_GUIDE.md#getting-ssi-api-credentials)

---

## 🚀 Quick Setup (3 Steps)

### Step 1: Install

```bash
go get github.com/banbox/banexg
```

### Step 2: Set Environment Variables

```bash
export SSI_CONSUMER_ID="your_consumer_id"
export SSI_CONSUMER_SECRET="your_consumer_secret"
export SSI_ACCOUNT_NO="0001234567"
```

### Step 3: Copy & Run This Code

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
    // Initialize
    exg, err := vietnam.New(map[string]interface{}{
        "consumerID":     os.Getenv("SSI_CONSUMER_ID"),
        "consumerSecret": os.Getenv("SSI_CONSUMER_SECRET"),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Load markets
    markets, _ := exg.LoadMarkets(false, nil)
    fmt.Printf("✅ Connected! %d markets loaded\n", len(markets))
    
    // Get VNM price
    ticker, _ := exg.FetchTicker("HOSE:VNM", nil)
    fmt.Printf("📊 VNM: %.0f VND (%.2f%%)\n", ticker.Last, ticker.Percentage)
    
    // Get balance
    balance, _ := exg.FetchBalance(map[string]interface{}{
        "accountNo": os.Getenv("SSI_ACCOUNT_NO"),
    })
    fmt.Printf("💰 Cash: %.0f VND\n", balance.Free["VND"])
}
```

**Run:**
```bash
go run main.go
```

---

## 📖 Common Operations

### Market Data

```go
// Single ticker
ticker, _ := exg.FetchTicker("HOSE:VNM", nil)

// Multiple tickers
tickers, _ := exg.FetchTickers([]string{"HOSE:VNM", "HOSE:VCB"}, nil)

// Historical data (30 days)
klines, _ := exg.FetchOHLCV("HOSE:VNM", "1d", 0, 30, nil)
```

### Trading

```go
accountNo := "0001234567" // Your trading account

// Buy limit order
order, _ := exg.CreateOrder(
    "HOSE:VNM",           // Symbol
    banexg.OdTypeLimit,   // Type
    banexg.OdSideBuy,     // Side
    100,                  // Shares
    82500,                // Price (VND)
    map[string]interface{}{"accountNo": accountNo},
)

// Cancel order
exg.CancelOrder(order.ID, "HOSE:VNM", 
    map[string]interface{}{"accountNo": accountNo})
```

### Account

```go
// Balance
balance, _ := exg.FetchBalance(map[string]interface{}{
    "accountNo": accountNo,
})

// Positions
positions, _ := exg.FetchPositions(nil, map[string]interface{}{
    "accountNo": accountNo,
})

// Open orders
orders, _ := exg.FetchOpenOrders("", 0, 0, map[string]interface{}{
    "accountNo": accountNo,
})
```

---

## 🇻🇳 Vietnam Market Essentials

### Trading Hours (ICT)

| Session | Time | Order Types |
|---------|------|-------------|
| Pre-Open | 08:30-09:00 | ATO |
| Morning | 09:00-11:30 | LO, MTL |
| **BREAK** | **11:30-13:00** | **CLOSED** |
| Afternoon | 13:00-14:30 | LO, MTL |
| Closing | 14:30-14:45 | ATC |

### Symbol Format

```
HOSE:VNM    ✅ Correct
HOSE:VCB    ✅ Correct
VNM         ❌ Wrong (missing exchange)
vnm         ❌ Wrong (case sensitive)
```

### Price Rules

- **Daily Limit**: ±7% from reference
- **Tick Size**: 100 VND for stocks >50,000 VND
- **Lot Size**: 100 shares minimum

### Order Types

| Code | Name | When |
|------|------|------|
| **LO** | Limit | Anytime during trading |
| **MTL** | Market-to-Limit | Continuous trading |
| **ATO** | At-The-Opening | Pre-open only |
| **ATC** | At-The-Close | Closing only |

**💡 Tip**: Use `banexg.OdTypeLimit` for limit orders, library handles session mapping automatically.

---

## ⚠️ Common Mistakes

### 1. Missing accountNo

```go
// ❌ WRONG
order, _ := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, nil)

// ✅ CORRECT
order, _ := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500,
    map[string]interface{}{"accountNo": "0001234567"})
```

### 2. Wrong Symbol Format

```go
// ❌ WRONG
ticker, _ := exg.FetchTicker("VNM", nil)

// ✅ CORRECT
ticker, _ := exg.FetchTicker("HOSE:VNM", nil)
```

### 3. Invalid Quantity

```go
// ❌ WRONG (must be multiple of 100)
order, _ := exg.CreateOrder("HOSE:VNM", "limit", "buy", 50, 82500, ...)

// ✅ CORRECT
order, _ := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, ...)
```

### 4. Invalid Price

```go
// ❌ WRONG (not at tick size 100)
order, _ := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82555, ...)

// ✅ CORRECT (multiple of 100)
order, _ := exg.CreateOrder("HOSE:VNM", "limit", "buy", 100, 82500, ...)
```

---

## 🆘 Quick Troubleshooting

| Error | Solution |
|-------|----------|
| "consumerID and consumerSecret are required" | Set env vars or pass in options |
| "accountNo is required" | Add `"accountNo": "xxx"` to params |
| "Market is closed" | Check trading hours (09:00-14:30 ICT, Mon-Fri) |
| "empty access token received" | Check credentials, ensure SSI subscription active |
| "price must be multiple of tick size" | Round to nearest 100 VND |
| "quantity must be multiple of lot size" | Use multiples of 100 shares |

---

## 🧪 Testing Without Real Trading

### Option 1: Sandbox Environment

```go
exg, _ := vietnam.New(map[string]interface{}{
    "consumerID":     "SANDBOX_ID",
    "consumerSecret": "SANDBOX_SECRET",
    "sandbox":        true, // Use test endpoints
})
```

### Option 2: Read-Only Operations

```go
// These don't affect your account:
exg.LoadMarkets(false, nil)           // Safe
exg.FetchTicker("HOSE:VNM", nil)      // Safe
exg.FetchOHLCV("HOSE:VNM", "1d", ...) // Safe
exg.FetchBalance(...)                 // Safe (read-only)

// These DO affect your account:
exg.CreateOrder(...)                  // ⚠️ Places real order
exg.CancelOrder(...)                  // ⚠️ Cancels real order
```

---

## 📚 Learn More

- **Complete Guide**: [USER_GUIDE.md](./USER_GUIDE.md) - Everything in detail
- **Developer Guide**: [AGENTS.md](./AGENTS.md) - For contributors
- **API Reference**: [README.md](./README.md) - Method documentation
- **SSI Docs**: https://fc-data.ssi.com.vn/Help

---

## 🎯 Next Steps

1. ✅ Run the quick start code above
2. ✅ Read [USER_GUIDE.md](./USER_GUIDE.md) for complete examples
3. ✅ Test with sandbox environment first
4. ✅ Start with small orders in production
5. ✅ Monitor positions regularly

---

## 💡 Pro Tips

1. **Always check market hours** before placing orders
2. **Use limit orders** for better price control
3. **Monitor your open orders** regularly
4. **Calculate fees** before trading (0.15% default)
5. **Keep credentials secure** (use env vars in production)

---

## 📞 Support

- **Issues**: https://github.com/banbox/banexg/issues
- **SSI Support**: support@ssi.com.vn | 1900 9088
- **Documentation**: https://guide.ssi.com.vn/ssi-products

---

**Ready to trade? Start with the code above! 🚀🇻🇳**
