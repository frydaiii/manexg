package vietnam

// Host constants
const (
	HostDataAPI    = "dataApi"
	HostTradingAPI = "tradingApi"
	HostWS         = "ws"
)

// Method constants - Authentication
const (
	MethodGetAccessToken = "getAccessToken"
)

// Method constants - Market Data
const (
	MethodGetSecuritiesList    = "getSecuritiesList"
	MethodGetSecuritiesDetails = "getSecuritiesDetails"
	MethodGetDailyOHLC         = "getDailyOHLC"
	MethodGetIntradayOHLC      = "getIntradayOHLC"
	MethodGetDailyStockPrice   = "getDailyStockPrice"
	MethodGetIndexComponents   = "getIndexComponents"
	MethodGetIndexList         = "getIndexList"
	MethodGetIndexSeries       = "getIndexSeries"
)

// Method constants - Trading
const (
	MethodPlaceOrder        = "placeOrder"
	MethodCancelOrder       = "cancelOrder"
	MethodModifyOrder       = "modifyOrder"
	MethodGetOrderHistory   = "getOrderHistory"
	MethodGetAccountBalance = "getAccountBalance"
	MethodGetStockPosition  = "getStockPosition"
	MethodGetOrderDetail    = "getOrderDetail"
)

// Method constants - Vietnam-Specific Features
const (
	MethodGetCompanyInfo     = "getCompanyInfo"
	MethodGetFinancialReport = "getFinancialReport"
	MethodGetTradingHolidays = "getTradingHolidays"
)

// Vietnam-specific order types
const (
	OdTypeATO = "ATO" // At-the-Opening
	OdTypeATC = "ATC" // At-the-Close
	OdTypeLO  = "LO"  // Limit Order
	OdTypeMTL = "MTL" // Market-to-Limit (replaced MP in 2025)
	OdTypeMAK = "MAK" // Market to Match and Cancel (HNX only)
	OdTypeMP  = "MP"  // Market Price (legacy, now MTL)
)

// Stock exchange identifiers
const (
	ExchangeHOSE  = "HOSE"
	ExchangeHNX   = "HNX"
	ExchangeUPCOM = "UPCOM"
)

// Trading session types
const (
	SessionPreOpen    = "pre_open"
	SessionMorning    = "morning"
	SessionLunch      = "lunch"
	SessionAfternoon  = "afternoon"
	SessionClosing    = "closing"
	SessionAfterHours = "after_hours"
	SessionClosed     = "closed"
)

// Trading schedule (ICT = UTC+7)
// Map session name to [start_time, end_time] in HHMM format
var TradingSchedule = map[string][2]int{
	SessionPreOpen:    {900, 915},   // 09:00-09:15
	SessionMorning:    {915, 1130},  // 09:15-11:30
	SessionLunch:      {1130, 1300}, // 11:30-13:00 (no trading)
	SessionAfternoon:  {1300, 1430}, // 13:00-14:30
	SessionClosing:    {1430, 1445}, // 14:30-14:45
	SessionAfterHours: {1445, 1500}, // 14:45-15:00 (negotiated deals)
}

// Order types allowed per session
var SessionOrderTypes = map[string][]string{
	SessionPreOpen:    {OdTypeATO, OdTypeLO},
	SessionMorning:    {OdTypeLO, OdTypeMTL},
	SessionAfternoon:  {OdTypeLO, OdTypeMTL},
	SessionClosing:    {OdTypeATC, OdTypeLO},
	SessionAfterHours: {OdTypeLO}, // Negotiated deals only
}

// Price tick sizes by exchange and price range
type PriceTickRule struct {
	Exchange string
	MinPrice float64
	MaxPrice float64
	TickSize float64
}

var PriceTickRules = []PriceTickRule{
	// HOSE tick sizes
	{Exchange: ExchangeHOSE, MinPrice: 0, MaxPrice: 10000, TickSize: 10},
	{Exchange: ExchangeHOSE, MinPrice: 10000, MaxPrice: 50000, TickSize: 50},
	{Exchange: ExchangeHOSE, MinPrice: 50000, MaxPrice: 999999999, TickSize: 100},
	// HNX tick sizes
	{Exchange: ExchangeHNX, MinPrice: 0, MaxPrice: 999999999, TickSize: 100},
	// UPCOM tick sizes
	{Exchange: ExchangeUPCOM, MinPrice: 0, MaxPrice: 999999999, TickSize: 100},
}

// Default lot size (shares per lot)
const DefaultLotSize = 100

// Default price limit percentage (±7%)
const DefaultPriceLimitPct = 0.07
