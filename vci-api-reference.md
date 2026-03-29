# VCI (VietCap) REST & GraphQL API Reference

> Reverse-engineered from the [vnstock](https://github.com/thinh-vu/vnstock) Python library.
> All endpoints are **unauthenticated** — no API keys or tokens required.
> Last updated: 2026-03-28

---

## Table of Contents

- [Base URLs](#base-urls)
- [Common Headers](#common-headers)
- [REST Endpoints](#rest-endpoints)
  - [1. OHLC Price History](#1-ohlc-price-history)
  - [2. Intraday Matched Orders (Tick Data)](#2-intraday-matched-orders-tick-data)
  - [3. All Listed Symbols](#3-all-listed-symbols)
  - [4. Symbols by Market Group](#4-symbols-by-market-group)
  - [5. Price Board (Real-Time Multi-Symbol Quote)](#5-price-board-real-time-multi-symbol-quote)
- [GraphQL Endpoint](#graphql-endpoint)
  - [6. Symbols by Industries (ICB Classification)](#6-symbols-by-industries-icb-classification)
  - [7. ICB Industry Code Reference](#7-icb-industry-code-reference)
  - [8. Company Profile (Mega Query)](#8-company-profile-mega-query)
  - [9. Financial Ratio Field Dictionary](#9-financial-ratio-field-dictionary)
  - [10. Financial Statements & Ratios](#10-financial-statements--ratios)
- [Reference Tables](#reference-tables)
  - [Interval Mapping](#interval-mapping)
  - [Market Group Codes](#market-group-codes)
  - [Index Symbol Mapping](#index-symbol-mapping)
  - [Financial Field Codes](#financial-field-codes)
  - [Company Type Codes](#company-type-codes)

---

## Base URLs

| Name | URL | Used For |
|------|-----|----------|
| REST API | `https://trading.vietcap.com.vn/api/` | Price data, symbols, order book |
| GraphQL API | `https://trading.vietcap.com.vn/data-mt/graphql` | Financials, company info, industry classification |

---

## Common Headers

All requests should include these headers to mimic the VietCap web platform:

```http
Accept: application/json, text/plain, */*
Accept-Language: en-US,en;q=0.9,vi-VN;q=0.8,vi;q=0.7
Content-Type: application/json
Connection: keep-alive
Cache-Control: no-cache
Pragma: no-cache
DNT: 1
Referer: https://trading.vietcap.com.vn/
Origin: https://trading.vietcap.com.vn
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: same-site
sec-ch-ua-mobile: ?0
sec-ch-ua-platform: "Windows"
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 ...
```

> **Note:** POST bodies must be JSON-serialized strings sent as raw `body` (not form-encoded). Ensure `Content-Type: application/json` is set.

---

## REST Endpoints

### 1. OHLC Price History

Fetch historical candlestick (OHLC + volume) data.

| Field | Value |
|-------|-------|
| **URL** | `https://trading.vietcap.com.vn/api/chart/OHLCChart/gap-chart` |
| **Method** | `POST` |

#### Request Body

```json
{
  "timeFrame": "ONE_DAY",
  "symbols": ["VCI"],
  "to": 1743120000,
  "countBack": 252
}
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `timeFrame` | `string` | Yes | Candle resolution. See [Interval Mapping](#interval-mapping) |
| `symbols` | `string[]` | Yes | Array of ticker symbols (typically one element) |
| `to` | `int64` | Yes | End timestamp — Unix epoch in **seconds** |
| `countBack` | `int` | Yes | Number of candles to return, counting backwards from `to` |

#### `timeFrame` Values

| User Interval | API `timeFrame` Value |
|---------------|----------------------|
| `1m`, `5m`, `15m`, `30m` | `"ONE_MINUTE"` |
| `1H` | `"ONE_HOUR"` |
| `1D`, `1W`, `1M` | `"ONE_DAY"` |

> For `5m`, `15m`, `30m`, `1W`, `1M` — the API returns the **base resolution** data (`ONE_MINUTE` or `ONE_DAY`), and client-side resampling produces the desired interval. See [Interval Mapping](#interval-mapping) for resample rules.

#### `countBack` Calculation Guide

| Interval | Formula |
|----------|---------|
| `1D` (daily) | Number of business days in date range + 1 |
| `1H` (hourly) | Business days × 5 (VN market has ~5 trading hours/day) + 1 |
| `1m` (minute) | Business days × 255 (150 min morning + 105 min afternoon) + 1 |

#### Response

```json
[
  {
    "t": [1609459200, 1609545600, 1609632000],
    "o": [45.20, 45.50, 46.00],
    "h": [46.00, 46.10, 46.30],
    "l": [44.80, 45.00, 45.50],
    "c": [45.80, 45.60, 46.20],
    "v": [1200000, 980000, 1100000]
  }
]
```

> Response may be a bare array `[{...}]` or wrapped as `{"data": [{...}]}` — handle both.

| Field | Type | Description |
|-------|------|-------------|
| `t` | `int64[]` | Unix timestamps (seconds). For minute data: exact minute timestamps |
| `o` | `float64[]` | Open prices |
| `h` | `float64[]` | High prices |
| `l` | `float64[]` | Low prices |
| `c` | `float64[]` | Close prices |
| `v` | `int64[]` | Volume |

#### Index Symbols

When querying index data (VNINDEX, HNXINDEX, UPCOMINDEX), use these mapped symbols:

| Input | API Symbol |
|-------|-----------|
| `VNINDEX` | `VNINDEX` |
| `HNXINDEX` | `HNXIndex` |
| `UPCOMINDEX` | `HNXUpcomIndex` |

---

### 2. Intraday Matched Orders (Tick Data)

Fetch today's individual matched trades (tick-by-tick).

| Field | Value |
|-------|-------|
| **URL** | `https://trading.vietcap.com.vn/api/market-watch/LEData/getAll` |
| **Method** | `POST` |

#### Request Body

```json
{
  "symbol": "VCI",
  "limit": 100,
  "truncTime": null
}
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbol` | `string` | Yes | Single ticker symbol |
| `limit` | `int` | Yes | Number of matched orders to return (default: 100) |
| `truncTime` | `int64 \| null` | No | Pagination cursor — Unix epoch in **milliseconds**. Pass `null` for latest data. Use the earliest `truncTime` from a previous response to fetch older records |

> **Not supported for index symbols** (VNINDEX, HNXINDEX, UPCOMINDEX).

#### Response

```json
[
  {
    "truncTime": 1711609200000,
    "matchPrice": 46.5,
    "matchVol": 3000,
    "matchType": "ATO",
    "id": "abc123"
  },
  {
    "truncTime": 1711609260000,
    "matchPrice": 46.6,
    "matchVol": 1500,
    "matchType": "BU",
    "id": "def456"
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `truncTime` | `int64` | Match timestamp — Unix epoch in **milliseconds** |
| `matchPrice` | `float64` | Matched price |
| `matchVol` | `int64` | Matched volume (shares) |
| `matchType` | `string` | Match type: `"ATO"` (opening), `"ATC"` (closing), `"BU"` (buy up), `"SD"` (sell down), `"UD"` (undefined) |
| `id` | `string` | Unique trade ID |

---

### 3. All Listed Symbols

Fetch all listed symbols with exchange info.

| Field | Value |
|-------|-------|
| **URL** | `https://trading.vietcap.com.vn/api/price/symbols/getAll` |
| **Method** | `GET` |

#### Query Parameters

None.

#### Response

```json
[
  {
    "id": 1,
    "symbol": "VCI",
    "board": "HOSE",
    "type": "STOCK",
    "organName": "Công ty Chứng khoán Bản Việt",
    "enOrganName": "VietCap Securities",
    "organShortName": "VCI",
    "enOrganShortName": "VCI"
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | `string` | Ticker symbol |
| `board` | `string` | Exchange: `"HOSE"`, `"HNX"`, `"UPCOM"` |
| `type` | `string` | Instrument type: `"STOCK"`, `"ETF"`, `"BOND"`, `"CW"` (covered warrant), etc. |
| `organName` | `string` | Company name (Vietnamese) |
| `enOrganName` | `string` | Company name (English) |
| `organShortName` | `string` | Short name (Vietnamese) |
| `enOrganShortName` | `string` | Short name (English) |

---

### 4. Symbols by Market Group

Fetch symbols belonging to a specific market group/index.

| Field | Value |
|-------|-------|
| **URL** | `https://trading.vietcap.com.vn/api/price/symbols/getByGroup` |
| **Method** | `GET` |

#### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group` | `string` | Yes | Market group code. See [Market Group Codes](#market-group-codes) |

#### Example

```
GET https://trading.vietcap.com.vn/api/price/symbols/getByGroup?group=VN30
```

#### Response

```json
[
  { "symbol": "VCI" },
  { "symbol": "TCB" },
  { "symbol": "VNM" }
]
```

Returns an array of objects with at least a `symbol` field.

---

### 5. Price Board (Real-Time Multi-Symbol Quote)

Fetch real-time price board data (bid/ask levels, last match) for multiple symbols.

| Field | Value |
|-------|-------|
| **URL** | `https://trading.vietcap.com.vn/api/price/symbols/getList` |
| **Method** | `POST` |

#### Request Body

```json
{
  "symbols": ["VCI", "TCB", "VNM"]
}
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbols` | `string[]` | Yes | Array of ticker symbols |

#### Response

```json
[
  {
    "listingInfo": {
      "symbol": "VCI",
      "board": "HOSE",
      "organName": "Công ty Chứng khoán Bản Việt",
      "enOrganName": "VietCap Securities",
      "organShortName": "VCI",
      "enOrganShortName": "VCI",
      "ticker": "VCI",
      "exercisePrice": null,
      "exerciseRatio": null,
      "maturityDate": null,
      "underlyingSymbol": null,
      "issuerName": null
    },
    "bidAsk": {
      "symbol": "VCI",
      "session": "ATO",
      "time": "2024-03-28T09:15:00",
      "bidPrices": [
        { "price": 45.00, "volume": 5000 },
        { "price": 44.90, "volume": 3000 },
        { "price": 44.80, "volume": 2000 }
      ],
      "askPrices": [
        { "price": 45.10, "volume": 4000 },
        { "price": 45.20, "volume": 2500 },
        { "price": 45.30, "volume": 1500 }
      ]
    },
    "matchPrice": {
      "symbol": "VCI",
      "session": "ATO",
      "time": "2024-03-28T09:15:00",
      "matchPrice": 45.05,
      "matchVolume": 1200
    }
  }
]
```

| Top-Level Key | Description |
|---------------|-------------|
| `listingInfo` | Static symbol metadata (name, exchange, warrant info) |
| `bidAsk` | Order book with 3 best bid and 3 best ask levels |
| `matchPrice` | Last matched trade price and volume |

---

## GraphQL Endpoint

**URL:** `https://trading.vietcap.com.vn/data-mt/graphql`
**Method:** `POST`
**Content-Type:** `application/json`

All GraphQL queries use the same endpoint. Request format:

```json
{
  "query": "...",
  "variables": { ... }
}
```

---

### 6. Symbols by Industries (ICB Classification)

Fetch all companies with their ICB industry classification.

#### Query

```graphql
{
  CompaniesListingInfo {
    ticker
    organName
    enOrganName
    icbName3
    enIcbName3
    icbName2
    enIcbName2
    icbName4
    enIcbName4
    comTypeCode
    icbCode1
    icbCode2
    icbCode3
    icbCode4
    __typename
  }
}
```

#### Variables

```json
{}
```

#### Response Path

`data.CompaniesListingInfo[]`

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `ticker` | `string` | Stock symbol |
| `organName` | `string` | Company name (Vietnamese) |
| `enOrganName` | `string` | Company name (English) |
| `icbName2` | `string` | ICB Sector name level 2 (Vietnamese) |
| `enIcbName2` | `string` | ICB Sector name level 2 (English) |
| `icbName3` | `string` | ICB Sub-sector name level 3 (Vietnamese) |
| `enIcbName3` | `string` | ICB Sub-sector name level 3 (English) |
| `icbName4` | `string` | ICB Sub-sector name level 4 (Vietnamese) |
| `enIcbName4` | `string` | ICB Sub-sector name level 4 (English) |
| `comTypeCode` | `string` | Company type: `"CT"` (general), `"CK"` (securities), `"NH"` (bank), `"BH"` (insurance) |
| `icbCode1`–`icbCode4` | `string` | ICB numeric codes at levels 1–4 |

---

### 7. ICB Industry Code Reference

Fetch the ICB industry code master table plus per-company assignments.

#### Query

```graphql
query Query {
  ListIcbCode {
    icbCode
    level
    icbName
    enIcbName
    __typename
  }
  CompaniesListingInfo {
    ticker
    icbCode1
    icbCode2
    icbCode3
    icbCode4
    __typename
  }
}
```

#### Variables

```json
{}
```

#### Response Paths

- `data.ListIcbCode[]` — Industry code reference table
- `data.CompaniesListingInfo[]` — Per-company ICB assignments

#### `ListIcbCode` Fields

| Field | Type | Description |
|-------|------|-------------|
| `icbCode` | `string` | ICB classification code |
| `level` | `int` | Hierarchy level (1–4) |
| `icbName` | `string` | Industry name (Vietnamese) |
| `enIcbName` | `string` | Industry name (English) |

---

### 8. Company Profile (Mega Query)

A single compound query that returns all company data in one round-trip.

#### Query

```graphql
query Query($ticker: String!, $lang: String!) {

  AnalysisReportFiles(ticker: $ticker, langCode: $lang) {
    date
    description
    link
    name
    __typename
  }

  News(ticker: $ticker, langCode: $lang) {
    id
    organCode
    ticker
    newsTitle
    newsSubTitle
    friendlySubTitle
    newsImageUrl
    newsSourceLink
    createdAt
    publicDate
    updatedAt
    langCode
    newsId
    newsShortContent
    newsFullContent
    closePrice
    referencePrice
    floorPrice
    ceilingPrice
    percentPriceChange
    __typename
  }

  TickerPriceInfo(ticker: $ticker) {
    financialRatio {
      yearReport
      lengthReport
      updateDate
      revenue
      revenueGrowth
      netProfit
      netProfitGrowth
      ebitMargin
      roe
      roic
      roa
      pe
      pb
      eps
      currentRatio
      cashRatio
      quickRatio
      interestCoverage
      ae
      fae
      netProfitMargin
      grossMargin
      ev
      issueShare
      ps
      pcf
      bvps
      evPerEbitda
      at
      fat
      acp
      dso
      dpo
      epsTTM
      charterCapital
      RTQ4
      charterCapitalRatio
      RTQ10
      dividend
      ebitda
      ebit
      le
      de
      ccc
      RTQ17
      __typename
    }
    ticker
    exchange
    ev
    ceilingPrice
    floorPrice
    referencePrice
    openPrice
    matchPrice
    closePrice
    priceChange
    percentPriceChange
    highestPrice
    lowestPrice
    totalVolume
    highestPrice1Year
    lowestPrice1Year
    percentLowestPriceChange1Year
    percentHighestPriceChange1Year
    foreignTotalVolume
    foreignTotalRoom
    averageMatchVolume2Week
    foreignHoldingRoom
    currentHoldingRatio
    maxHoldingRatio
    __typename
  }

  Subsidiary(ticker: $ticker) {
    id
    organCode
    subOrganCode
    percentage
    subOrListingInfo {
      enOrganName
      organName
      __typename
    }
    __typename
  }

  Affiliate(ticker: $ticker) {
    id
    organCode
    subOrganCode
    percentage
    subOrListingInfo {
      enOrganName
      organName
      __typename
    }
    __typename
  }

  CompanyListingInfo(ticker: $ticker) {
    id
    issueShare
    en_History
    history
    en_CompanyProfile
    companyProfile
    icbName3
    enIcbName3
    icbName2
    enIcbName2
    icbName4
    enIcbName4
    financialRatio {
      id
      ticker
      issueShare
      charterCapital
      __typename
    }
    __typename
  }

  OrganizationManagers(ticker: $ticker) {
    id
    ticker
    fullName
    positionName
    positionShortName
    en_PositionName
    en_PositionShortName
    updateDate
    percentage
    quantity
    __typename
  }

  OrganizationShareHolders(ticker: $ticker) {
    id
    ticker
    ownerFullName
    en_OwnerFullName
    quantity
    percentage
    updateDate
    __typename
  }

  OrganizationResignedManagers(ticker: $ticker) {
    id
    ticker
    fullName
    positionName
    positionShortName
    en_PositionName
    en_PositionShortName
    updateDate
    percentage
    quantity
    __typename
  }

  OrganizationEvents(ticker: $ticker) {
    id
    organCode
    ticker
    eventTitle
    en_EventTitle
    publicDate
    issueDate
    sourceUrl
    eventListCode
    ratio
    value
    recordDate
    exrightDate
    eventListName
    en_EventListName
    __typename
  }
}
```

#### Variables

```json
{
  "ticker": "VCI",
  "lang": "vi"
}
```

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `ticker` | `String!` | Yes | Stock symbol |
| `lang` | `String!` | Yes | Language: `"vi"` (Vietnamese) or `"en"` (English) |

#### Response Sub-Queries

| Response Key | Description |
|--------------|-------------|
| `CompanyListingInfo` | Company profile, history, ICB industry |
| `OrganizationManagers` | Current board members & executives |
| `OrganizationResignedManagers` | Former/resigned officers |
| `OrganizationShareHolders` | Major shareholders with stake % |
| `Subsidiary` | Subsidiary companies with ownership % |
| `Affiliate` | Affiliate companies with ownership % |
| `OrganizationEvents` | Corporate events (dividends, rights issues, etc.) |
| `TickerPriceInfo` | Current price data + key financial ratios |
| `TickerPriceInfo.financialRatio` | Summary financial ratios (nested) |
| `News` | News articles related to the ticker |
| `AnalysisReportFiles` | Analyst report PDF links |

---

### 9. Financial Ratio Field Dictionary

Fetch the metadata/schema for all financial report fields. Use this to map raw field codes (e.g., `BSA1`, `ISA2`) to human-readable names.

#### Query

```graphql
query Query {
  ListFinancialRatio {
    id
    type
    name
    unit
    isDefault
    fieldName
    en_Type
    en_Name
    tagName
    comTypeCode
    order
    __typename
  }
}
```

#### Variables

```json
{}
```

#### Response Path

`data.ListFinancialRatio[]`

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int` | Unique record ID |
| `type` | `string` | Report category (Vietnamese), e.g. `"Chi tieu can doi ke toan"` |
| `en_Type` | `string` | Report category (English), e.g. `"Balance Sheet"` |
| `name` | `string` | Field name (Vietnamese) |
| `en_Name` | `string` | Field name (English) |
| `fieldName` | `string` | Raw code used in financial data: `"BSA1"`, `"ISA2"`, `"CFA3"`, etc. |
| `unit` | `string` | Unit: `"BILLION"`, `"PERCENT"`, `"INDEX"`, `"MILLION"` |
| `comTypeCode` | `string` | Company type filter: `"CT"`, `"CK"`, `"NH"`, `"BH"` |
| `isDefault` | `bool` | Whether this is a default display field |
| `tagName` | `string` | Tag identifier |
| `order` | `int` | Display sort order |

#### Field Code Naming Convention

| Prefix | Report Type |
|--------|-------------|
| `BSA*`, `BSB*` | Balance Sheet |
| `ISA*`, `ISB*`, `ISS*`, `ISI*` | Income Statement |
| `CFA*`, `CFB*`, `CFS*` | Cash Flow Statement |

The letter suffix (`A`, `B`, `S`, `I`) corresponds to company sub-type variants. Numbers indicate line item order.

---

### 10. Financial Statements & Ratios

Fetch full financial data (balance sheet, income statement, cash flow, ratios).

#### Query

The query uses a GraphQL **fragment** to request ~200+ financial fields:

```graphql
fragment Ratios on CompanyFinancialRatio {
  ticker
  yearReport
  lengthReport
  updateDate
  revenue
  revenueGrowth
  netProfit
  netProfitGrowth
  ebitMargin
  roe
  roic
  roa
  pe
  pb
  eps
  currentRatio
  cashRatio
  quickRatio
  interestCoverage
  ae
  netProfitMargin
  grossMargin
  ev
  issueShare
  ps
  pcf
  bvps
  evPerEbitda
  at
  fat
  acp
  dso
  dpo
  ccc
  de
  le
  ebitda
  ebit
  dividend
  RTQ10
  charterCapitalRatio
  RTQ4
  epsTTM
  charterCapital
  fae
  RTQ17
  # Balance Sheet fields
  BSA1
  BSA2
  BSA5
  BSA8
  BSA10
  BSA159
  BSA16
  BSA22
  BSA23
  BSA24
  BSA162
  BSA27
  BSA29
  BSA43
  BSA46
  BSA50
  BSA209
  BSA53
  BSA54
  BSA55
  BSA56
  BSA58
  BSA67
  BSA71
  BSA173
  BSA78
  BSA79
  BSA80
  BSA175
  BSA86
  BSA90
  BSA96
  BSB97
  BSB98
  BSB99
  BSB100
  BSB101
  BSB102
  BSB103
  BSB104
  BSB110
  BSB111
  BSB112
  BSB113
  BSB114
  BSB115
  BSB117
  BSB118
  BSB119
  BSB121
  # Income Statement fields
  ISA1
  ISA2
  ISA5
  ISA6
  ISA7
  ISA8
  ISA10
  ISA14
  ISA16
  ISA18
  ISA19
  ISA22
  ISA23
  ISB25
  ISB26
  ISB27
  ISB28
  ISB29
  ISB30
  ISB33
  ISB34
  ISB36
  ISB40
  ISB41
  ISS141
  ISS142
  ISS143
  ISS144
  ISS148
  ISS149
  ISS150
  ISS151
  ISS152
  ISI64
  ISI87
  ISI97
  # Cash Flow fields
  CFA1
  CFA2
  CFA3
  CFA4
  CFA5
  CFA6
  CFA7
  CFA8
  CFA9
  CFA10
  CFA11
  CFA16
  CFA17
  CFA18
  CFA19
  CFA21
  CFA22
  CFA24
  CFA25
  CFA26
  CFA27
  CFA30
  CFA33
  CFA35
  CFA36
  CFA38
  CFB64
  CFB65
  CFB66
  CFB67
  CFB68
  CFB69
  CFB70
  CFB71
  CFB72
  CFB73
  CFB74
  CFB80
  CFS191
  CFS192
  CFS193
  CFS194
  CFS195
  CFS196
  CFS200
  CFS201
  CFS202
  CFS203
  CFS210
  __typename
}

query Query($ticker: String!, $period: String!) {
  CompanyFinancialRatio(ticker: $ticker, period: $period) {
    ratio {
      ...Ratios
      __typename
    }
    period
    __typename
  }
}
```

#### Variables

```json
{
  "ticker": "VCI",
  "period": "Q"
}
```

| Variable | Type | Required | Values | Description |
|----------|------|----------|--------|-------------|
| `ticker` | `String!` | Yes | Any stock symbol | Target company |
| `period` | `String!` | Yes | `"Q"` or `"Y"` | `"Q"` = quarterly, `"Y"` = yearly |

#### Response Path

`data.CompanyFinancialRatio.ratio[]`

Each element is one reporting period with all the fields from the `Ratios` fragment.

#### Key Response Fields

| Group | Example Fields | Description |
|-------|---------------|-------------|
| **Meta** | `ticker`, `yearReport`, `lengthReport`, `updateDate` | Period identifiers (`lengthReport`: 1-4 for quarters) |
| **Valuation** | `pe`, `pb`, `ps`, `pcf`, `ev`, `evPerEbitda`, `bvps`, `eps`, `epsTTM` | Price & valuation multiples |
| **Profitability** | `roe`, `roa`, `roic`, `grossMargin`, `netProfitMargin`, `ebitMargin` | Margin & return ratios |
| **Liquidity** | `currentRatio`, `cashRatio`, `quickRatio`, `interestCoverage` | Liquidity ratios |
| **Leverage** | `de`, `le`, `ae`, `fae` | Debt & leverage ratios |
| **Efficiency** | `at`, `fat`, `acp`, `dso`, `dpo`, `ccc` | Asset turnover & cycle metrics |
| **Growth** | `revenueGrowth`, `netProfitGrowth` | Growth rates |
| **Balance Sheet** | `BSA1`–`BSB121` | Raw balance sheet line items (in billions VND) |
| **Income Statement** | `ISA1`–`ISI97` | Raw income statement line items |
| **Cash Flow** | `CFA1`–`CFS210` | Raw cash flow line items |

> Use the [Financial Ratio Field Dictionary](#9-financial-ratio-field-dictionary) (query #9) to resolve `BSA*`, `ISA*`, `CFA*` codes to human-readable names. Field mapping varies by company type (`comTypeCode`).

---

## Reference Tables

### Interval Mapping

How user-facing intervals map to API `timeFrame` values and client-side resampling:

| User Interval | API `timeFrame` | Client Resample Rule | Description |
|---------------|-----------------|---------------------|-------------|
| `1m` | `ONE_MINUTE` | None (native) | 1-minute candles |
| `5m` | `ONE_MINUTE` | Resample to `5min` | 5-minute candles |
| `15m` | `ONE_MINUTE` | Resample to `15min` | 15-minute candles |
| `30m` | `ONE_MINUTE` | Resample to `30min` | 30-minute candles |
| `1H` | `ONE_HOUR` | None (native) | 1-hour candles |
| `1D` | `ONE_DAY` | None (native) | Daily candles |
| `1W` | `ONE_DAY` | Resample to `1W` | Weekly candles |
| `1M` | `ONE_DAY` | Resample to `ME` (month-end) | Monthly candles |

### Market Group Codes

Valid `group` values for the [Symbols by Market Group](#4-symbols-by-market-group) endpoint:

| Group Code | Description |
|------------|-------------|
| `HOSE` | Ho Chi Minh Stock Exchange (all) |
| `VN30` | VN30 Index constituents |
| `VNMidCap` | VN Mid-Cap Index |
| `VNSmallCap` | VN Small-Cap Index |
| `VNAllShare` | VN All-Share Index |
| `VN100` | VN100 Index |
| `ETF` | Exchange-Traded Funds |
| `HNX` | Hanoi Stock Exchange (all) |
| `HNX30` | HNX30 Index constituents |
| `HNXCon` | HNX Construction |
| `HNXFin` | HNX Finance |
| `HNXLCap` | HNX Large Cap |
| `HNXMSCap` | HNX Mid/Small Cap |
| `HNXMan` | HNX Manufacturing |
| `UPCOM` | Unlisted Public Company Market |
| `FU_INDEX` | Index Futures |
| `FU_BOND` | Bond Futures |
| `BOND` | Bonds |
| `CW` | Covered Warrants |

### Index Symbol Mapping

| Standard Name | VCI API Symbol |
|--------------|----------------|
| `VNINDEX` | `VNINDEX` |
| `HNXINDEX` | `HNXIndex` |
| `UPCOMINDEX` | `HNXUpcomIndex` |

### Financial Field Codes

| Prefix | Report | Example |
|--------|--------|---------|
| `BSA*` | Balance Sheet (general companies) | `BSA1` = Total current assets |
| `BSB*` | Balance Sheet (banks/securities) | `BSB97` = Total liabilities |
| `ISA*` | Income Statement (general) | `ISA1` = Revenue |
| `ISB*` | Income Statement (banks/securities) | `ISB25` = Interest income |
| `ISS*` | Income Statement (securities firms) | `ISS141` = Brokerage revenue |
| `ISI*` | Income Statement (insurance) | `ISI64` = Insurance premium |
| `CFA*` | Cash Flow (general) | `CFA1` = Operating cash flow |
| `CFB*` | Cash Flow (banks/securities) | `CFB64` = Investing cash flow |
| `CFS*` | Cash Flow (securities firms) | `CFS191` = Financing cash flow |

> Use query [#9 (Financial Ratio Field Dictionary)](#9-financial-ratio-field-dictionary) to get the full mapping of codes to names for a specific `comTypeCode`.

### Company Type Codes

| Code | Type | Description |
|------|------|-------------|
| `CT` | General | Manufacturing, services, real estate, etc. |
| `CK` | Securities | Securities brokerages |
| `NH` | Banking | Banks, asset management |
| `BH` | Insurance | Life & non-life insurance, reinsurance |

Each company type uses a different subset of financial field codes in the [Financial Statements query](#10-financial-statements--ratios). The `comTypeCode` is determined from the company's ICB level-4 industry classification.

---

## Go Implementation Notes

1. **POST body serialization:** All POST requests send JSON as a raw string body (not form-encoded). Use `json.Marshal()` and set `Content-Type: application/json`.

2. **Response parsing:** The OHLC endpoint may return either `[{...}]` or `{"data": [{...}]}`. Use `json.RawMessage` and try both.

3. **Timestamps:** OHLC uses Unix seconds (`int64`). Intraday tick data uses Unix milliseconds (`int64`). Handle both.

4. **GraphQL:** All GraphQL queries go to a single endpoint. Use a simple HTTP POST with `{"query": "...", "variables": {...}}`. No need for a GraphQL client library.

5. **Rate limiting:** No explicit rate limits documented, but the `Referer` and `Origin` headers are important for access. Consider adding reasonable delays between requests.

6. **Resampling:** For `5m`, `15m`, `30m`, `1W`, `1M` intervals, you receive base-resolution data and must resample client-side. Implement OHLCV resampling: `open` = first, `high` = max, `low` = min, `close` = last, `volume` = sum.
