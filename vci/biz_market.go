package vci

import (
	"fmt"
	"strings"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	"go.uber.org/zap"
)

func (e *VCI) LoadMarkets(reload bool, params map[string]interface{}) (banexg.MarketMap, *errs.Error) {
	if e.Markets != nil && !reload {
		return e.Markets, nil
	}

	markets := make(banexg.MarketMap)
	marketsByID := make(banexg.MarketArrMap)
	lookupByRaw := make(map[string]*banexg.Market)
	lookupByTicker := make(map[string][]*banexg.Market)

	// Fetch all symbols from VCI API
	rows, err := requestVCI[[]map[string]interface{}](e, MethodPublicGetSymbolsAll, nil, e.GetRetryNum("LoadMarkets", 1))
	if err != nil {
		return nil, err
	}

	log.Info("vci load markets response",
		zap.Int("total_symbols", len(rows)),
	)

	for _, row := range rows {
		symType := strings.ToUpper(strings.TrimSpace(anyString(row["type"])))
		if symType != "STOCK" && symType != "ETF" {
			continue
		}
		ticker := strings.ToUpper(strings.TrimSpace(anyString(row["symbol"])))
		if ticker == "" {
			continue
		}
		board := strings.ToUpper(strings.TrimSpace(anyString(row["board"])))
		if board == "" {
			continue
		}

		market := newStockMarket(board, ticker, row)
		markets[market.Symbol] = market
		marketsByID[market.ID] = []*banexg.Market{market}
		lookupByRaw[market.ID] = market
		lookupByTicker[ticker] = append(lookupByTicker[ticker], market)
	}

	// Add index pseudo-markets
	for canonical, vciSymbol := range indexSymbolMap {
		market := newIndexMarket(canonical, vciSymbol)
		markets[market.Symbol] = market
		marketsByID[market.ID] = []*banexg.Market{market}
		lookupByRaw[market.ID] = market
		lookupByTicker[canonical] = append(lookupByTicker[canonical], market)
	}

	e.MarketsLock.Lock()
	e.MarketsByIdLock.Lock()
	e.Markets = markets
	e.MarketsById = marketsByID
	e.MarketsByIdLock.Unlock()
	e.MarketsLock.Unlock()

	e.lookupLock.Lock()
	e.marketsByRawID = lookupByRaw
	e.marketsByTicker = lookupByTicker
	e.lookupLock.Unlock()

	log.Info("vci load markets done",
		zap.Int("markets", len(markets)),
		zap.Int("raw_lookup", len(lookupByRaw)),
		zap.Int("ticker_lookup", len(lookupByTicker)),
	)

	return markets, nil
}

func newStockMarket(board, ticker string, info map[string]interface{}) *banexg.Market {
	rawID := board + ":" + ticker
	symbol := makeMarketSymbol(board, ticker)

	bd := boardPrecision[board]
	priceTick := float64(0)
	minLot := float64(100)
	if bd != nil {
		priceTick = bd.priceTick
		minLot = bd.lotSize
	}

	modePrice := banexg.PrecModeDecimalPlace
	if priceTick > 0 {
		modePrice = banexg.PrecModeTickSize
	}

	market := &banexg.Market{
		ID:          rawID,
		LowercaseID: strings.ToLower(rawID),
		Symbol:      symbol,
		Base:        ticker,
		Quote:       "VND",
		Type:        banexg.MarketSpot,
		Spot:        true,
		Active:      true,
		FeeSide:     "quote",
		Precision: &banexg.Precision{
			Amount:     0,
			Price:      priceTick,
			ModeAmount: banexg.PrecModeDecimalPlace,
			ModePrice:  modePrice,
		},
		Limits: &banexg.MarketLimits{
			Amount: &banexg.LimitRange{Min: minLot},
		},
		Info: info,
	}
	market.Info["board"] = board
	market.Info["ticker"] = ticker
	market.Info["rawId"] = rawID
	return market
}

func newIndexMarket(canonical, vciSymbol string) *banexg.Market {
	rawID := "INDEX:" + canonical
	symbol := fmt.Sprintf("INDEX:%s/VND", canonical)
	info := map[string]interface{}{
		"board":     "INDEX",
		"ticker":    canonical,
		"rawId":     rawID,
		"isIndex":   true,
		"vciSymbol": vciSymbol,
	}
	return &banexg.Market{
		ID:          rawID,
		LowercaseID: strings.ToLower(rawID),
		Symbol:      symbol,
		Base:        canonical,
		Quote:       "VND",
		Type:        banexg.MarketSpot,
		Spot:        true,
		Active:      true,
		FeeSide:     "quote",
		Precision: &banexg.Precision{
			Amount:     0,
			Price:      0,
			ModeAmount: banexg.PrecModeDecimalPlace,
			ModePrice:  banexg.PrecModeDecimalPlace,
		},
		Limits: &banexg.MarketLimits{
			Amount: &banexg.LimitRange{Min: 0},
		},
		Info: info,
	}
}

func makeMarketSymbol(board, ticker string) string {
	return fmt.Sprintf("%s:%s/VND", strings.ToUpper(strings.TrimSpace(board)), strings.ToUpper(strings.TrimSpace(ticker)))
}

func splitRawMarketID(rawID string) (string, string) {
	parts := strings.Split(strings.TrimSpace(rawID), ":")
	if len(parts) != 2 {
		return "", ""
	}
	board := strings.ToUpper(strings.TrimSpace(parts[0]))
	ticker := strings.ToUpper(strings.TrimSpace(parts[1]))
	if board == "" || ticker == "" {
		return "", ""
	}
	return board, ticker
}

func (e *VCI) MapMarket(rawID string, year int) (*banexg.Market, *errs.Error) {
	_, err := e.LoadMarkets(false, nil)
	if err != nil {
		return nil, err
	}
	_ = year
	rawID = strings.TrimSpace(rawID)
	if rawID == "" {
		return nil, errs.NewMsg(errs.CodeParamRequired, "raw market id required")
	}
	board, ticker := splitRawMarketID(rawID)
	e.lookupLock.RLock()
	defer e.lookupLock.RUnlock()
	if board != "" {
		if mar, ok := e.marketsByRawID[board+":"+ticker]; ok {
			return mar, nil
		}
		return nil, errs.NewMsg(errs.CodeNoMarketForPair, "no market found: %s", rawID)
	}
	ticker = strings.ToUpper(rawID)
	matches := e.marketsByTicker[ticker]
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return nil, errs.NewMsg(errs.CodeParamInvalid, "ambiguous ticker without board: %s", rawID)
	}
	return nil, errs.NewMsg(errs.CodeNoMarketForPair, "no market found: %s", rawID)
}
