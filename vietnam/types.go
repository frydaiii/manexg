package vietnam

import (
	"github.com/banbox/banexg"
)

// Vietnam exchange struct embedding base Exchange
type Vietnam struct {
	*banexg.Exchange

	// SSI API credentials
	ConsumerID     string
	ConsumerSecret string
	AccessToken    string
	TokenExpiry    int64 // Unix timestamp in milliseconds

	// Current trading session state
	SessionState string
}

// SSI API Authentication Response
type SSIAuthResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"` // Seconds until expiry
}

// SSI Securities List Response
type SSISecurityInfo struct {
	Symbol           string  `json:"symbol"`
	StockName        string  `json:"stockName"`
	StockNameEn      string  `json:"stockNameEn"`
	Exchange         string  `json:"exchange"`
	Type             string  `json:"type"` // STOCK, ETF, BOND, etc.
	LotSize          int     `json:"lotSize"`
	Sector           string  `json:"sector"`
	Industry         string  `json:"industry"`
	ListingDate      string  `json:"listingDate"`
	ISIN             string  `json:"isin"`
	OutstandingShare float64 `json:"outstandingShare"`
	IssueShare       float64 `json:"issueShare"`
}

type SSISecuritiesListRsp struct {
	Status      int               `json:"status"`
	Message     string            `json:"message"`
	TotalRecord int               `json:"totalRecord"`
	Data        []SSISecurityInfo `json:"data"`
}

// SSI Securities Details Response
type SSISecurityDetail struct {
	Symbol      string  `json:"symbol"`
	Ceiling     float64 `json:"ceiling"`  // Price ceiling (ref + 7%)
	Floor       float64 `json:"floor"`    // Price floor (ref - 7%)
	RefPrice    float64 `json:"refPrice"` // Reference price
	LastPrice   float64 `json:"lastPrice"`
	LastVolume  float64 `json:"lastVolume"`
	BidPrice1   float64 `json:"bidPrice1"`
	BidVolume1  float64 `json:"bidVolume1"`
	AskPrice1   float64 `json:"askPrice1"`
	AskVolume1  float64 `json:"askVolume1"`
	TotalVolume float64 `json:"totalVolume"`
	TotalValue  float64 `json:"totalValue"`
	OpenPrice   float64 `json:"openPrice"`
	HighPrice   float64 `json:"highPrice"`
	LowPrice    float64 `json:"lowPrice"`
	AvgPrice    float64 `json:"avgPrice"`
	Time        string  `json:"time"`
}

type SSISecurityDetailRsp struct {
	Status  int               `json:"status"`
	Message string            `json:"message"`
	Data    SSISecurityDetail `json:"data"`
}

// SSI OHLC Data Response
type SSIOHLCData struct {
	TradingDate string  `json:"tradingDate"`
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	Volume      float64 `json:"volume"`
	Value       float64 `json:"value"`
}

type SSIOHLCRsp struct {
	Status      int           `json:"status"`
	Message     string        `json:"message"`
	TotalRecord int           `json:"totalRecord"`
	Data        []SSIOHLCData `json:"data"`
}

// SSI Daily Stock Price Response
type SSIDailyStockPrice struct {
	Symbol          string  `json:"symbol"`
	TradingDate     string  `json:"tradingDate"`
	OpenPrice       float64 `json:"openPrice"`
	HighestPrice    float64 `json:"highestPrice"`
	LowestPrice     float64 `json:"lowestPrice"`
	ClosePrice      float64 `json:"closePrice"`
	PriorClosePrice float64 `json:"priorClosePrice"`
	TotalVolume     float64 `json:"totalVolume"`
	TotalValue      float64 `json:"totalValue"`
}

type SSIDailyStockPriceRsp struct {
	Status  int                  `json:"status"`
	Message string               `json:"message"`
	Data    []SSIDailyStockPrice `json:"data"`
}

// SSI Order Book Response
type SSIOrderBookLevel struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}

type SSIOrderBook struct {
	Symbol    string              `json:"symbol"`
	Bids      []SSIOrderBookLevel `json:"bids"`
	Asks      []SSIOrderBookLevel `json:"asks"`
	Timestamp int64               `json:"timestamp"`
}

type SSIOrderBookRsp struct {
	Status  int          `json:"status"`
	Message string       `json:"message"`
	Data    SSIOrderBook `json:"data"`
}

// SSI Order Request
type SSIOrderReq struct {
	Symbol    string  `json:"symbol"`
	OrderType string  `json:"orderType"` // LO, ATO, ATC, MTL, MAK
	Side      string  `json:"side"`      // BUY, SELL
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"` // Required for LO
	AccountNo string  `json:"accountNo"`
	ValidDate string  `json:"validDate"` // Optional: YYYYMMDD
	RequestID string  `json:"requestId"` // Client order ID
}

// SSI Order Response
type SSIOrderInfo struct {
	OrderID      string  `json:"orderId"`
	RequestID    string  `json:"requestId"`
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"`
	OrderType    string  `json:"orderType"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	FilledQty    int     `json:"filledQty"`
	AvgPrice     float64 `json:"avgPrice"`
	Status       string  `json:"status"` // NEW, PARTIALLY_FILLED, FILLED, CANCELLED, REJECTED
	CreateTime   string  `json:"createTime"`
	LastModified string  `json:"lastModified"`
	AccountNo    string  `json:"accountNo"`
	Message      string  `json:"message"` // Error message if rejected
}

type SSIOrderRsp struct {
	Status  int          `json:"status"`
	Message string       `json:"message"`
	Data    SSIOrderInfo `json:"data"`
}

// SSI Order History Response
type SSIOrderHistoryRsp struct {
	Status      int            `json:"status"`
	Message     string         `json:"message"`
	TotalRecord int            `json:"totalRecord"`
	Data        []SSIOrderInfo `json:"data"`
}

// SSI Account Balance Response
type SSIBalanceInfo struct {
	AccountNo       string  `json:"accountNo"`
	TotalCash       float64 `json:"totalCash"`       // Total cash
	AvailableCash   float64 `json:"availableCash"`   // Available for trading
	BuyingPower     float64 `json:"buyingPower"`     // Buying power (with margin)
	TotalStockValue float64 `json:"totalStockValue"` // Total stock holdings value
	TotalAsset      float64 `json:"totalAsset"`      // Total assets
	Debt            float64 `json:"debt"`            // Margin debt
	MarginRatio     float64 `json:"marginRatio"`     // Current margin ratio
	SettledCash     float64 `json:"settledCash"`     // T+0 settled cash
	PendingCash     float64 `json:"pendingCash"`     // Pending settlement
	Currency        string  `json:"currency"`        // VND
}

type SSIBalanceRsp struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Data    SSIBalanceInfo `json:"data"`
}

// SSI Stock Position Response
type SSIPositionInfo struct {
	Symbol          string  `json:"symbol"`
	Quantity        int     `json:"quantity"`        // Total quantity
	AvailableQty    int     `json:"availableQty"`    // Available to sell
	AvgPrice        float64 `json:"avgPrice"`        // Average purchase price
	MarketPrice     float64 `json:"marketPrice"`     // Current market price
	MarketValue     float64 `json:"marketValue"`     // Current value
	CostValue       float64 `json:"costValue"`       // Cost basis
	UnrealizedPL    float64 `json:"unrealizedPL"`    // Unrealized P&L
	UnrealizedPLPct float64 `json:"unrealizedPLPct"` // Unrealized P&L %
	AccountNo       string  `json:"accountNo"`
}

type SSIPositionRsp struct {
	Status      int               `json:"status"`
	Message     string            `json:"message"`
	TotalRecord int               `json:"totalRecord"`
	Data        []SSIPositionInfo `json:"data"`
}

// Financial Report (Vietnam-specific feature)
type FinancialReport struct {
	Symbol     string                 `json:"symbol"`
	ReportType string                 `json:"reportType"` // BALANCE_SHEET, INCOME_STATEMENT, CASH_FLOW
	Period     string                 `json:"period"`     // Q1, Q2, Q3, Q4, YEAR
	Year       int                    `json:"year"`
	Quarter    int                    `json:"quarter"`
	Data       map[string]interface{} `json:"data"`
	Currency   string                 `json:"currency"`
	Unit       string                 `json:"unit"` // Million VND, Billion VND
}

// Company Info (Vietnam-specific feature)
type CompanyInfo struct {
	Symbol              string  `json:"symbol"`
	CompanyName         string  `json:"companyName"`
	CompanyNameEn       string  `json:"companyNameEn"`
	Exchange            string  `json:"exchange"`
	Sector              string  `json:"sector"`
	Industry            string  `json:"industry"`
	Website             string  `json:"website"`
	ListingDate         string  `json:"listingDate"`
	CharterCapital      float64 `json:"charterCapital"`
	OutstandingShares   float64 `json:"outstandingShares"`
	IssuedShares        float64 `json:"issuedShares"`
	ForeignOwnership    float64 `json:"foreignOwnership"`    // Current foreign ownership %
	ForeignOwnershipMax float64 `json:"foreignOwnershipMax"` // Maximum allowed foreign ownership %
	RoomAvailable       float64 `json:"roomAvailable"`       // Room available for foreign investors
}

// Index data
type IndexInfo struct {
	IndexCode   string  `json:"indexCode"`
	IndexName   string  `json:"indexName"`
	IndexValue  float64 `json:"indexValue"`
	Change      float64 `json:"change"`
	ChangePct   float64 `json:"changePct"`
	TotalVolume float64 `json:"totalVolume"`
	TotalValue  float64 `json:"totalValue"`
	TradingDate string  `json:"tradingDate"`
	Time        string  `json:"time"`
}

type IndexSeriesData struct {
	TradingDate string  `json:"tradingDate"`
	IndexValue  float64 `json:"indexValue"`
	Change      float64 `json:"change"`
	ChangePct   float64 `json:"changePct"`
	Volume      float64 `json:"volume"`
	Value       float64 `json:"value"`
}

type TradingHoliday struct {
	Date        string `json:"date"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
