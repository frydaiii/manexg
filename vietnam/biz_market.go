package vietnam

import (
	"fmt"
	"strings"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	"go.uber.org/zap"
)

func (e *Vietnam) LoadMarkets(reload bool, params map[string]interface{}) (banexg.MarketMap, *errs.Error) {
	if e.Markets != nil && !reload {
		return e.Markets, nil
	}
	markets := make(banexg.MarketMap)
	marketsByID := make(banexg.MarketArrMap)
	lookupByRaw := make(map[string]*banexg.Market)
	lookupByTicker := make(map[string][]*banexg.Market)

	for _, board := range marketBoards {
		boardStart := e.MilliSeconds()
		rows, err := e.fetchSecuritiesRows(board)
		if err != nil {
			return nil, err
		}
		detailRows, _ := e.fetchSecuritiesDetailRows(board)
		log.Info("vietnam load markets board",
			zap.String("board", board),
			zap.Int("securities", len(rows)),
			zap.Int("details", len(detailRows)),
			zap.Int64("cost_ms", e.MilliSeconds()-boardStart),
		)
		detailMap := make(map[string]map[string]interface{}, len(detailRows))
		for _, row := range detailRows {
			ticker := parseTicker(row)
			if ticker != "" {
				detailMap[ticker] = row
			}
		}
		for _, row := range rows {
			ticker := parseTicker(row)
			if ticker == "" {
				continue
			}
			merged := map[string]interface{}{}
			for k, v := range row {
				merged[k] = v
			}
			if detail, ok := detailMap[ticker]; ok {
				for k, v := range detail {
					merged[k] = v
				}
			}
			market := newStockMarket(board, ticker, merged)
			markets[market.Symbol] = market
			marketsByID[market.ID] = []*banexg.Market{market}
			lookupByRaw[market.ID] = market
			lookupByTicker[ticker] = append(lookupByTicker[ticker], market)
		}
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

	log.Info("vietnam load markets done",
		zap.Int("markets", len(markets)),
		zap.Int("raw_lookup", len(lookupByRaw)),
		zap.Int("ticker_lookup", len(lookupByTicker)),
	)

	return markets, nil
}

func (e *Vietnam) fetchSecuritiesRows(board string) ([]map[string]interface{}, *errs.Error) {
	rows := make([]map[string]interface{}, 0)
	for page := 1; page <= 10; page++ {
		payload := map[string]interface{}{
			"Market":    board,
			"Pageindex": page,
			"Pagesize":  1000,
		}
		log.Info("vietnam securities request",
			zap.String("board", board),
			zap.Int("page", page),
			zap.Any("payload", payload),
		)
		data, err := requestSSI[[]map[string]interface{}](e, MethodPublicPostMarketSecurities, payload, e.GetRetryNum("LoadMarkets", 1))
		if err != nil {
			return nil, err
		}
		log.Info("vietnam securities response",
			zap.String("board", board),
			zap.Int("page", page),
			zap.Int("rows", len(data)),
		)
		if len(data) == 0 {
			break
		}
		rows = append(rows, data...)
		if len(data) < 1000 {
			break
		}
	}
	return rows, nil
}

func (e *Vietnam) fetchSecuritiesDetailRows(board string) ([]map[string]interface{}, *errs.Error) {
	rows := make([]map[string]interface{}, 0)
	for page := 1; page <= 10; page++ {
		payload := map[string]interface{}{
			"Market":                 board,
			"pageIndex":              page,
			"lookupRequest.pageSize": 1000,
		}
		log.Info("vietnam securities details request",
			zap.String("board", board),
			zap.Int("page", page),
			zap.Any("payload", payload),
		)
		// SecuritiesDetails returns dataList[].repeatedinfoList[] nested structure
		data, err := requestSSI[[]map[string]interface{}](e, MethodPublicPostMarketSecuritiesInfo, payload, e.GetRetryNum("LoadMarkets", 1))
		if err != nil {
			return rows, err
		}
		log.Info("vietnam securities details response",
			zap.String("board", board),
			zap.Int("page", page),
			zap.Int("rows", len(data)),
		)
		if len(data) == 0 {
			break
		}
		// Flatten: extract repeatedinfoList items from each dataList entry
		for _, entry := range data {
			infoList := getInterfaceSlice(entry, "repeatedinfoList")
			for _, item := range infoList {
				if m, ok := item.(map[string]interface{}); ok {
					rows = append(rows, m)
				}
			}
		}
		if len(data) < 1000 {
			break
		}
	}
	return rows, nil
}

func newStockMarket(board, ticker string, info map[string]interface{}) *banexg.Market {
	rawID := board + ":" + ticker
	symbol := makeMarketSymbol(board, ticker)
	priceTick := anyFloat(info["tickIncrement1"])
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
		Info: info,
	}
	market.Info["board"] = board
	market.Info["ticker"] = ticker
	market.Info["rawId"] = rawID
	return market
}

func makeMarketSymbol(board, ticker string) string {
	return fmt.Sprintf("%s:%s/VND", strings.ToUpper(strings.TrimSpace(board)), strings.ToUpper(strings.TrimSpace(ticker)))
}

func parseTicker(row map[string]interface{}) string {
	if len(row) == 0 {
		return ""
	}
	for _, key := range []string{"symbol", "stockSymbol", "ticker", "code"} {
		if val := strings.ToUpper(strings.TrimSpace(anyString(getAnyByLowerKey(row, key)))); val != "" {
			return val
		}
	}
	return ""
}

func getAnyByLowerKey(m map[string]interface{}, expect string) interface{} {
	if len(m) == 0 {
		return nil
	}
	expect = strings.ToLower(strings.TrimSpace(expect))
	for k, v := range m {
		if strings.ToLower(strings.TrimSpace(k)) == expect {
			return v
		}
	}
	return nil
}

func getInterfaceSlice(m map[string]interface{}, key string) []interface{} {
	val := getAnyByLowerKey(m, key)
	if val == nil {
		return nil
	}
	if arr, ok := val.([]interface{}); ok {
		return arr
	}
	return nil
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

func (e *Vietnam) MapMarket(rawID string, year int) (*banexg.Market, *errs.Error) {
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
