package vietnam

import (
	"strings"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

// Location for ICT (Indochina Time = UTC+7)
var LocICT *time.Location
var VietnamLocation *time.Location

func init() {
	var err error
	LocICT, err = time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		// Fallback to fixed offset if location not available
		LocICT = time.FixedZone("ICT", 7*60*60)
	}
	VietnamLocation = LocICT
}

// NormalizeSymbol parses symbol and returns exchange and code
// Format: "EXCHANGE:SYMBOL" or just "SYMBOL" (defaults to HOSE)
// Examples: "HOSE:VNM", "HNX:SHB", "VNM" (-> HOSE:VNM)
func NormalizeSymbol(symbol string) (exchange, code string) {
	parts := strings.Split(symbol, ":")
	if len(parts) == 2 {
		return strings.ToUpper(parts[0]), strings.ToUpper(parts[1])
	}
	// Default to HOSE if no exchange specified
	return ExchangeHOSE, strings.ToUpper(symbol)
}

// BuildSymbol creates standard symbol format "EXCHANGE:CODE"
func BuildSymbol(exchange, code string) string {
	return strings.ToUpper(exchange) + ":" + strings.ToUpper(code)
}

// GetCurrentSessionType returns the current trading session type
func GetCurrentSessionType() string {
	return GetSessionTypeAt(time.Now())
}

// GetSessionTypeAt returns the trading session type at a specific time
func GetSessionTypeAt(t time.Time) string {
	ict := t.In(LocICT)

	// Check if weekend
	if ict.Weekday() == time.Saturday || ict.Weekday() == time.Sunday {
		return SessionClosed
	}

	// Convert time to HHMM format
	timeVal := ict.Hour()*100 + ict.Minute()

	// Check against trading schedule
	for session, times := range TradingSchedule {
		if timeVal >= times[0] && timeVal < times[1] {
			return session
		}
	}

	// Before market open or after close
	if timeVal < TradingSchedule[SessionPreOpen][0] {
		return SessionClosed
	}
	if timeVal >= TradingSchedule[SessionAfterHours][1] {
		return SessionClosed
	}

	return SessionClosed
}

// IsMarketOpen checks if market is currently open for trading
func IsMarketOpen() bool {
	return IsMarketOpenAt(time.Now())
}

// IsMarketOpenAt checks if market is open at a specific time
func IsMarketOpenAt(t time.Time) bool {
	session := GetSessionTypeAt(t)
	// Market is open during pre-open, morning, afternoon, closing, and after-hours
	// (lunch is closed for trading)
	return session != SessionClosed && session != SessionLunch
}

// CanPlaceOrderType checks if an order type can be placed in current session
func CanPlaceOrderType(orderType string) bool {
	return CanPlaceOrderTypeAt(orderType, time.Now())
}

// CanPlaceOrderTypeAt checks if an order type can be placed at a specific time
func CanPlaceOrderTypeAt(orderType string, t time.Time) bool {
	session := GetSessionTypeAt(t)
	allowedTypes, ok := SessionOrderTypes[session]
	if !ok {
		return false
	}

	for _, allowed := range allowedTypes {
		if allowed == orderType {
			return true
		}
	}
	return false
}

// GetPriceTickSize returns the tick size for a given price and exchange
func GetPriceTickSize(price float64, exchange string) float64 {
	for _, rule := range PriceTickRules {
		if rule.Exchange == exchange && price >= rule.MinPrice && price < rule.MaxPrice {
			return rule.TickSize
		}
	}
	// Default to 100 VND if no rule matches
	return 100
}

// RoundPrice rounds a price to the nearest valid tick size
func RoundPrice(price float64, exchange string) float64 {
	tickSize := GetPriceTickSize(price, exchange)
	if tickSize == 0 {
		return price
	}
	return float64(int(price/tickSize+0.5)) * tickSize
}

// CalculatePriceLimits calculates ceiling and floor prices based on reference price
func CalculatePriceLimits(refPrice float64, exchange string) (ceiling, floor float64) {
	limitPct := DefaultPriceLimitPct
	ceiling = RoundPrice(refPrice*(1+limitPct), exchange)
	floor = RoundPrice(refPrice*(1-limitPct), exchange)
	return ceiling, floor
}

// ValidatePrice checks if a price is within valid range (floor to ceiling)
func ValidatePrice(price, refPrice float64, exchange string) *errs.Error {
	ceiling, floor := CalculatePriceLimits(refPrice, exchange)
	if price < floor {
		return errs.NewMsg(errs.CodeParamInvalid, "price %.0f below floor %.0f", price, floor)
	}
	if price > ceiling {
		return errs.NewMsg(errs.CodeParamInvalid, "price %.0f above ceiling %.0f", price, ceiling)
	}
	return nil
}

// ValidateQuantity checks if quantity is a valid multiple of lot size
func ValidateQuantity(quantity int, lotSize int) *errs.Error {
	if lotSize == 0 {
		lotSize = DefaultLotSize
	}
	if quantity <= 0 {
		return errs.NewMsg(errs.CodeParamInvalid, "quantity must be positive")
	}
	if quantity%lotSize != 0 {
		return errs.NewMsg(errs.CodeParamInvalid, "quantity %d must be multiple of lot size %d", quantity, lotSize)
	}
	return nil
}

// MapOrderTypeToSSI maps standard BanExg order type to SSI format
func MapOrderTypeToSSI(banexgType string, session string) string {
	switch banexgType {
	case banexg.OdTypeMarket:
		// Market orders depend on session
		if session == SessionPreOpen {
			return OdTypeATO
		} else if session == SessionClosing {
			return OdTypeATC
		}
		return OdTypeMTL // Market-to-Limit during continuous trading
	case banexg.OdTypeLimit:
		return OdTypeLO
	default:
		return OdTypeLO
	}
}

// MapSSIOrderTypeToBanExg maps SSI order type to standard BanExg format
func MapSSIOrderTypeToBanExg(ssiType string) string {
	switch ssiType {
	case OdTypeATO, OdTypeATC, OdTypeMTL, OdTypeMAK, OdTypeMP:
		return banexg.OdTypeMarket
	case OdTypeLO:
		return banexg.OdTypeLimit
	default:
		return banexg.OdTypeLimit
	}
}

// MapOrderSideToSSI maps standard BanExg side to SSI format
func MapOrderSideToSSI(banexgSide string) string {
	switch strings.ToLower(banexgSide) {
	case "buy":
		return "BUY"
	case "sell":
		return "SELL"
	default:
		return "BUY"
	}
}

// MapSSIOrderSideToBanExg maps SSI side to standard BanExg format
func MapSSIOrderSideToBanExg(ssiSide string) string {
	switch strings.ToUpper(ssiSide) {
	case "BUY":
		return banexg.OdSideBuy
	case "SELL":
		return banexg.OdSideSell
	default:
		return banexg.OdSideBuy
	}
}

// MapSSIOrderStatus maps SSI order status to standard BanExg format
func MapSSIOrderStatus(ssiStatus string) string {
	switch strings.ToUpper(ssiStatus) {
	case "NEW", "PENDING":
		return banexg.OdStatusOpen
	case "PARTIALLY_FILLED":
		return banexg.OdStatusPartFilled
	case "FILLED", "MATCHED":
		return banexg.OdStatusFilled
	case "CANCELLED", "CANCELED":
		return banexg.OdStatusCanceled
	case "REJECTED", "FAILED":
		return banexg.OdStatusRejected
	case "EXPIRED":
		return banexg.OdStatusExpired
	default:
		return banexg.OdStatusOpen
	}
}

// FormatDateSSI formats time to SSI API date format YYYYMMDD
func FormatDateSSI(t time.Time) string {
	return t.Format("20060102")
}

// ParseDateSSI parses SSI date format YYYYMMDD to time
func ParseDateSSI(dateStr string) (time.Time, error) {
	return time.ParseInLocation("20060102", dateStr, LocICT)
}

// FormatDateTimeSSI formats time to SSI API datetime format
func FormatDateTimeSSI(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseDateTimeSSI parses SSI datetime format to time
func ParseDateTimeSSI(dateTimeStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", dateTimeStr, LocICT)
}
