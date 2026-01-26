# Vietnam Stock Market Adaptation - Task Planning

## Project Overview

**Project:** Adapt BanExg SDK for Vietnam Stock Market  
**Target API:** SSI FastConnect API  
**Estimated Duration:** 7 weeks  
**Status:** Planning Phase

---

## Task Breakdown

### Phase 1: Project Setup & Core Infrastructure
**Duration:** Week 1-2  
**Goal:** Establish module skeleton and core authentication

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 1.1 | Create `vietnam/` directory structure | High | 2h | None | Directory with placeholder files |
| 1.2 | Define `vietnam/data.go` constants | High | 4h | 1.1 | Host constants, Method constants, Trading schedule |
| 1.3 | Define `vietnam/types.go` structures | High | 6h | 1.1 | Vietnam struct, SSI response types, Order types |
| 1.4 | Implement `vietnam/entry.go` constructor | High | 4h | 1.2, 1.3 | New() function, API mappings, ExgInfo |
| 1.5 | Register in `bex/entrys.go` | High | 1h | 1.4 | "vietnam" entry in newExgs map |
| 1.6 | Implement SSI token authentication | High | 6h | 1.4 | Token fetch, refresh, storage |
| 1.7 | Implement `vietnam/common.go` utilities | Medium | 4h | 1.3 | Time zone (ICT), symbol normalization |
| 1.8 | Create trading hours validation | High | 4h | 1.7 | isMarketOpen(), getSessionType() |
| 1.9 | Setup test infrastructure | Medium | 3h | 1.5 | Test helpers, mock data, local.json template |
| 1.10 | Write `vietnam/AGENTS.md` dev guide | Low | 2h | 1.4 | Module documentation |

**Phase 1 Total: ~36 hours**

---

### Phase 2: Market Data - Load Markets
**Duration:** Week 2 (continued)  
**Goal:** Implement LoadMarkets and market data structures

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 2.1 | Define SSI securities list response types | High | 2h | 1.3 | SSISecurityInfo, SSISecurityListRsp |
| 2.2 | Implement `LoadMarkets()` | High | 8h | 1.6, 2.1 | Fetch all securities from SSI |
| 2.3 | Implement market mapping logic | High | 4h | 2.2 | Map SSI format to banexg.Market |
| 2.4 | Handle exchange differentiation | Medium | 3h | 2.3 | HOSE/HNX/UPCOM market flags |
| 2.5 | Implement `MapMarket()` | Medium | 3h | 2.3 | Raw ID to symbol conversion |
| 2.6 | Implement `GetMarket()` | Medium | 2h | 2.2 | Symbol lookup with validation |
| 2.7 | Cache markets with expiry | Medium | 3h | 2.2 | Market cache refresh logic |
| 2.8 | Create `markets.yml` static fallback | Low | 4h | 2.3 | Embedded market data (like china/) |
| 2.9 | Write LoadMarkets unit tests | Medium | 3h | 2.2 | Test coverage for market loading |

**Phase 2 Total: ~32 hours**

---

### Phase 3: Market Data - Quotes & OHLCV
**Duration:** Week 3  
**Goal:** Implement ticker, OHLCV, and order book fetching

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 3.1 | Define SSI ticker response types | High | 2h | 1.3 | SSITickerRsp, SSIPriceData |
| 3.2 | Implement `FetchTicker()` | High | 4h | 2.2, 3.1 | Single symbol quote |
| 3.3 | Implement `FetchTickers()` | High | 4h | 3.2 | Multiple symbols batch |
| 3.4 | Implement `FetchTickerPrice()` | Medium | 2h | 3.2 | Price-only response |
| 3.5 | Define SSI OHLCV response types | High | 2h | 1.3 | SSIOHLCData, SSIDailyOHLCRsp |
| 3.6 | Implement `FetchOHLCV()` - Daily | High | 6h | 2.2, 3.5 | Daily candlesticks |
| 3.7 | Implement `FetchOHLCV()` - Intraday | High | 4h | 3.6 | Intraday candlesticks |
| 3.8 | Handle timeframe mapping | Medium | 2h | 3.6 | Map "1d", "1h" to SSI format |
| 3.9 | Define SSI order book response types | Medium | 2h | 1.3 | SSIOrderBookRsp, SSIBidAsk |
| 3.10 | Implement `FetchOrderBook()` | Medium | 4h | 2.2, 3.9 | Bid/Ask depth |
| 3.11 | Implement `FetchLastPrices()` | Low | 3h | 3.2 | Last traded prices |
| 3.12 | Write market data unit tests | Medium | 4h | 3.2-3.10 | Test coverage for all fetchers |

**Phase 3 Total: ~39 hours**

---

### Phase 4: Trading - Order Management
**Duration:** Week 3-4  
**Goal:** Implement order creation, modification, and cancellation

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 4.1 | Define SSI order request/response types | High | 4h | 1.3 | SSIOrderReq, SSIOrderRsp, SSIOrderStatus |
| 4.2 | Implement order type mapping | High | 4h | 4.1 | Map Limit/Market to LO/MTL/ATO/ATC |
| 4.3 | Implement trading session detection | High | 3h | 1.8 | Pre-open, Continuous, Closing detection |
| 4.4 | Implement `CreateOrder()` | High | 8h | 4.1-4.3 | Place new order with all types |
| 4.5 | Validate order within price limits | High | 3h | 4.4 | ±7% ceiling/floor check |
| 4.6 | Implement `CancelOrder()` | High | 4h | 4.1 | Cancel pending order |
| 4.7 | Implement `EditOrder()` | Medium | 4h | 4.4 | Modify LO order price/quantity |
| 4.8 | Implement `FetchOrder()` | High | 3h | 4.1 | Get single order details |
| 4.9 | Implement `FetchOrders()` | High | 4h | 4.1 | Order history with filters |
| 4.10 | Implement `FetchOpenOrders()` | High | 3h | 4.9 | Active orders only |
| 4.11 | Handle order status mapping | Medium | 2h | 4.1 | SSI status to banexg status |
| 4.12 | Implement order rejection handling | Medium | 3h | 4.4 | Parse SSI error codes |
| 4.13 | Write trading unit tests | High | 6h | 4.4-4.10 | Order lifecycle tests |

**Phase 4 Total: ~51 hours**

---

### Phase 5: Account Management
**Duration:** Week 4 (continued)  
**Goal:** Implement balance and position tracking

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 5.1 | Define SSI account response types | High | 3h | 1.3 | SSIBalanceRsp, SSIPositionRsp |
| 5.2 | Implement `FetchBalance()` | High | 4h | 5.1 | Cash balance (VND) |
| 5.3 | Handle buying power calculation | Medium | 2h | 5.2 | Available for trading |
| 5.4 | Implement `FetchPositions()` | High | 5h | 5.1 | Stock holdings |
| 5.5 | Implement `FetchAccountPositions()` | Medium | 2h | 5.4 | Alias for FetchPositions |
| 5.6 | Map stock positions to banexg.Position | Medium | 3h | 5.4 | Field mapping |
| 5.7 | Implement margin info (if available) | Low | 3h | 5.1 | Margin balance/usage |
| 5.8 | Handle T+1.5 settlement tracking | Medium | 4h | 5.2 | Settled vs pending cash |
| 5.9 | Write account unit tests | Medium | 3h | 5.2-5.6 | Balance/position tests |

**Phase 5 Total: ~29 hours**

---

### Phase 6: WebSocket Real-time Data
**Duration:** Week 5  
**Goal:** Implement WebSocket connections for real-time feeds

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 6.1 | Research SSI WebSocket protocol | High | 4h | None | Protocol documentation |
| 6.2 | Implement `vietnam/ws_client.go` | High | 8h | 6.1 | WS connection management |
| 6.3 | Implement WS authentication | High | 4h | 6.2, 1.6 | Token-based WS auth |
| 6.4 | Implement WS message routing | High | 4h | 6.2 | Message type dispatcher |
| 6.5 | Implement `WatchOrderBooks()` | Medium | 5h | 6.4 | Real-time order book |
| 6.6 | Implement `UnWatchOrderBooks()` | Medium | 2h | 6.5 | Unsubscribe |
| 6.7 | Implement `WatchTrades()` | Medium | 4h | 6.4 | Real-time trades |
| 6.8 | Implement `UnWatchTrades()` | Medium | 2h | 6.7 | Unsubscribe |
| 6.9 | Implement `WatchOHLCVs()` | Low | 4h | 6.4 | Real-time candles |
| 6.10 | Implement `UnWatchOHLCVs()` | Low | 2h | 6.9 | Unsubscribe |
| 6.11 | Implement WS reconnection logic | Medium | 4h | 6.2 | Auto-reconnect on disconnect |
| 6.12 | Handle market hours WS disconnect | Medium | 2h | 6.11, 1.8 | Graceful close outside hours |
| 6.13 | Write WebSocket unit tests | Medium | 4h | 6.5-6.10 | WS functionality tests |

**Phase 6 Total: ~49 hours**

---

### Phase 7: Vietnam-Specific Features
**Duration:** Week 5-6  
**Goal:** Implement features unique to Vietnam market

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 7.1 | Define financial report types | Medium | 3h | 1.3 | FinancialReport, BalanceSheet, etc. |
| 7.2 | Implement `FetchFinancials()` | Medium | 5h | 7.1 | Company financial reports |
| 7.3 | Implement `FetchCompanyInfo()` | Medium | 4h | 7.1 | Company profile/fundamentals |
| 7.4 | Implement index data fetching | Medium | 4h | 2.2 | VN-Index, VN30, HNX-Index |
| 7.5 | Implement `FetchIndexComponents()` | Low | 3h | 7.4 | VN30 constituent list |
| 7.6 | Handle Vietnamese holidays | Medium | 4h | 1.8 | Holiday calendar integration |
| 7.7 | Implement price limit helpers | Medium | 2h | 3.2 | GetCeiling(), GetFloor() |
| 7.8 | Implement lot size validation | Medium | 2h | 4.4 | Validate 100-share lots |
| 7.9 | Add foreign ownership tracking | Low | 3h | 7.3 | Room availability for foreigners |

**Phase 7 Total: ~30 hours**

---

### Phase 8: Fee Calculation
**Duration:** Week 6 (continued)  
**Goal:** Implement Vietnam-specific fee calculation

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 8.1 | Research Vietnam trading fees | High | 2h | None | Fee structure documentation |
| 8.2 | Define fee structure constants | High | 2h | 8.1 | Brokerage, tax, exchange fees |
| 8.3 | Implement `CalculateFee()` | High | 4h | 8.2 | Total fee calculation |
| 8.4 | Handle sell-side tax (0.1%) | Medium | 2h | 8.3 | Capital gains tax on sell |
| 8.5 | Implement fee tier support | Low | 2h | 8.3 | Volume-based discounts |
| 8.6 | Write fee calculation tests | Medium | 2h | 8.3-8.5 | Fee accuracy tests |

**Phase 8 Total: ~14 hours**

---

### Phase 9: Error Handling & Edge Cases
**Duration:** Week 6 (continued)  
**Goal:** Robust error handling and edge case management

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 9.1 | Map SSI error codes to banexg errors | High | 4h | All | Comprehensive error mapping |
| 9.2 | Handle network failures gracefully | High | 3h | All | Retry logic, timeout handling |
| 9.3 | Handle token expiry mid-request | Medium | 2h | 1.6 | Token refresh and retry |
| 9.4 | Handle market closed errors | Medium | 2h | 1.8, 4.4 | User-friendly error messages |
| 9.5 | Handle rate limiting | Medium | 3h | All | Backoff and queue |
| 9.6 | Validate all input parameters | Medium | 4h | All | Input sanitization |
| 9.7 | Handle partial order fills | Medium | 3h | 4.4 | Accurate fill tracking |
| 9.8 | Log and trace debugging | Low | 2h | All | Debug logging support |

**Phase 9 Total: ~23 hours**

---

### Phase 10: Testing & Quality Assurance
**Duration:** Week 7  
**Goal:** Comprehensive testing and documentation

| ID | Task | Priority | Effort | Dependencies | Deliverables |
|----|------|----------|--------|--------------|--------------|
| 10.1 | Create mock SSI server | Medium | 6h | All | Test server for unit tests |
| 10.2 | Write unit tests (>80% coverage) | High | 8h | All | Comprehensive test suite |
| 10.3 | Write integration tests | High | 6h | All | End-to-end test scenarios |
| 10.4 | Test with SSI sandbox | High | 4h | All | Real API testing |
| 10.5 | Performance testing | Medium | 3h | All | Latency, throughput tests |
| 10.6 | Test edge cases | Medium | 4h | 9.1-9.8 | Error scenario testing |
| 10.7 | Create testdata fixtures | Medium | 3h | 10.1 | JSON response samples |
| 10.8 | Document all public APIs | High | 4h | All | GoDoc comments |
| 10.9 | Update project AGENTS.md | Medium | 2h | All | Add Vietnam to main docs |
| 10.10 | Create usage examples | Medium | 3h | All | Example code snippets |
| 10.11 | Code review & cleanup | High | 4h | All | Final code quality check |

**Phase 10 Total: ~47 hours**

---

## Task Summary by Phase

| Phase | Description | Tasks | Hours | Weeks |
|-------|-------------|-------|-------|-------|
| 1 | Project Setup & Core | 10 | 36h | 1-2 |
| 2 | Load Markets | 9 | 32h | 2 |
| 3 | Quotes & OHLCV | 12 | 39h | 3 |
| 4 | Order Management | 13 | 51h | 3-4 |
| 5 | Account Management | 9 | 29h | 4 |
| 6 | WebSocket | 13 | 49h | 5 |
| 7 | Vietnam-Specific | 9 | 30h | 5-6 |
| 8 | Fee Calculation | 6 | 14h | 6 |
| 9 | Error Handling | 8 | 23h | 6 |
| 10 | Testing & QA | 11 | 47h | 7 |
| **Total** | | **100** | **350h** | **7** |

---

## Priority Matrix

### P0 - Critical (Must Have)
| ID | Task | Phase |
|----|------|-------|
| 1.4 | entry.go constructor | 1 |
| 1.5 | Register in bex | 1 |
| 1.6 | SSI token authentication | 1 |
| 1.8 | Trading hours validation | 1 |
| 2.2 | LoadMarkets() | 2 |
| 3.2 | FetchTicker() | 3 |
| 3.6 | FetchOHLCV() | 3 |
| 4.4 | CreateOrder() | 4 |
| 4.6 | CancelOrder() | 4 |
| 5.2 | FetchBalance() | 5 |
| 5.4 | FetchPositions() | 5 |

### P1 - High (Should Have)
| ID | Task | Phase |
|----|------|-------|
| 3.3 | FetchTickers() | 3 |
| 3.10 | FetchOrderBook() | 3 |
| 4.7 | EditOrder() | 4 |
| 4.8-4.10 | Order queries | 4 |
| 6.2 | WS client | 6 |
| 8.3 | CalculateFee() | 8 |
| 9.1 | Error mapping | 9 |
| 10.2-10.3 | Unit/Integration tests | 10 |

### P2 - Medium (Nice to Have)
| ID | Task | Phase |
|----|------|-------|
| 6.5-6.10 | WatchOrderBooks, Trades, OHLCV | 6 |
| 7.2-7.3 | Financials, CompanyInfo | 7 |
| 7.6 | Holiday handling | 7 |
| 5.8 | T+1.5 settlement tracking | 5 |

### P3 - Low (Future Enhancement)
| ID | Task | Phase |
|----|------|-------|
| 2.8 | markets.yml fallback | 2 |
| 7.5 | FetchIndexComponents | 7 |
| 7.9 | Foreign ownership tracking | 7 |
| 8.5 | Fee tier support | 8 |

---

## Dependencies Graph

```
Phase 1 (Core Setup)
    │
    ├──► 1.1 Directory Structure
    │         │
    │         ├──► 1.2 data.go ──────────────────┐
    │         │                                   │
    │         └──► 1.3 types.go ─────────────────┤
    │                                             │
    │                     ┌───────────────────────┘
    │                     ▼
    │              1.4 entry.go
    │                     │
    │         ┌───────────┼───────────┐
    │         ▼           ▼           ▼
    │      1.5 bex    1.6 Auth    1.7 common.go
    │      register      │            │
    │                    │            ▼
    │                    │        1.8 Trading Hours
    │                    │
    ▼                    ▼
Phase 2              Phase 3-8 (All require auth)
(Markets)
    │
    ├──► 2.2 LoadMarkets ◄──── Required by all data fetchers
    │         │
    ▼         ▼
Phase 3   Phase 4-5
(Data)    (Trading)
    │         │
    ▼         ▼
Phase 6   Phase 7
(WS)      (Vietnam-specific)
    │         │
    └────┬────┘
         ▼
     Phase 9-10
     (QA & Docs)
```

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| SSI API changes | High | Medium | Version lock, monitor changelog |
| SSI sandbox unavailable | Medium | Low | Use mock server for dev |
| Rate limiting issues | Medium | Medium | Implement backoff, caching |
| WebSocket protocol undocumented | High | Medium | Research vnstock, contact SSI |
| Trading hours edge cases | Medium | High | Comprehensive time zone testing |
| Token expiry during trading | High | Low | Proactive refresh, retry logic |
| Order type mapping errors | High | Medium | Thorough testing per session |

---

## Milestones & Checkpoints

| Milestone | Target Date | Criteria |
|-----------|-------------|----------|
| **M1: Module Skeleton** | End Week 1 | entry.go compiles, registered in bex |
| **M2: Authentication Working** | End Week 2 | Token fetch/refresh works |
| **M3: Markets Loading** | End Week 2 | LoadMarkets returns valid data |
| **M4: Data Fetching Complete** | End Week 3 | Ticker, OHLCV, OrderBook work |
| **M5: Basic Trading** | End Week 4 | CreateOrder, CancelOrder work |
| **M6: Account Complete** | End Week 4 | Balance, Positions work |
| **M7: WebSocket MVP** | End Week 5 | WatchOrderBooks works |
| **M8: Vietnam Features** | End Week 6 | Trading hours, validation complete |
| **M9: Testing Complete** | End Week 7 | >80% coverage, integration tests pass |
| **M10: Production Ready** | End Week 7 | All docs complete, code reviewed |

---

## Resource Requirements

### Development
- 1 Go developer (full-time, 7 weeks)
- SSI FastConnect API credentials (production + sandbox)
- Test trading account with SSI

### Infrastructure
- Go 1.23+ development environment
- Mock server for testing
- CI/CD pipeline for automated tests

### Documentation
- SSI API documentation access
- vnstock source code (reference)
- Vietnam market trading rules

---

## Definition of Done

A task is considered **DONE** when:

1. ✅ Code is written and compiles without errors
2. ✅ Unit tests written and passing
3. ✅ Code follows project conventions (see AGENTS.md)
4. ✅ GoDoc comments added for public APIs
5. ✅ No lsp_diagnostics errors on changed files
6. ✅ Manual testing completed (where applicable)
7. ✅ Code reviewed (for Phase 10)

---

## Appendix: File Creation Checklist

### New Files to Create
```
vietnam/
├── entry.go           # Phase 1
├── data.go            # Phase 1
├── types.go           # Phase 1
├── common.go          # Phase 1
├── biz.go             # Phase 2
├── biz_market.go      # Phase 3
├── biz_ticker.go      # Phase 3
├── biz_order.go       # Phase 4
├── biz_order_create.go # Phase 4
├── biz_account.go     # Phase 5
├── biz_financial.go   # Phase 7
├── ws_client.go       # Phase 6
├── ws_biz.go          # Phase 6
├── markets.yml        # Phase 2 (optional)
├── AGENTS.md          # Phase 1
├── local.json         # Phase 1 (template)
├── base_test.go       # Phase 10
├── biz_test.go        # Phase 10
├── biz_market_test.go # Phase 10
├── biz_order_test.go  # Phase 10
└── testdata/          # Phase 10
    ├── securities_list.json
    ├── ticker.json
    ├── ohlcv.json
    └── order.json
```

### Files to Modify
```
bex/entrys.go          # Phase 1 - Add "vietnam" entry
AGENTS.md              # Phase 10 - Document Vietnam module
```
