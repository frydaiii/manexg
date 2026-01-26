package vietnam

import (
	"context"
	"strings"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	"github.com/banbox/banexg/utils"
	"go.uber.org/zap"
)

func (e *Vietnam) LoadMarkets(reload bool, params map[string]interface{}) (banexg.MarketMap, *errs.Error) {
	if e.Markets != nil && !reload {
		return e.Markets, nil
	}

	args := utils.SafeParams(params)
	exchange := utils.GetMapVal(args, "exchange", "")

	rsp := e.RequestApiRetry(context.Background(), MethodGetSecuritiesList, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var securitiesRsp SSISecuritiesListRsp
	if err := utils.UnmarshalString(rsp.Content, &securitiesRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse securities list: %v", err)
	}

	if securitiesRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", securitiesRsp.Message)
	}

	newMarkets := make(banexg.MarketMap)
	newMarketsById := make(banexg.MarketArrMap)

	for _, sec := range securitiesRsp.Data {
		if exchange != "" && sec.Exchange != exchange {
			continue
		}

		market := e.parseMarket(&sec)
		if market == nil {
			continue
		}

		newMarkets[market.Symbol] = market
		if arr, ok := newMarketsById[market.ID]; ok {
			newMarketsById[market.ID] = append(arr, market)
		} else {
			newMarketsById[market.ID] = []*banexg.Market{market}
		}
	}

	e.MarketsLock.Lock()
	e.MarketsByIdLock.Lock()
	e.Markets = newMarkets
	e.MarketsById = newMarketsById
	e.MarketsByIdLock.Unlock()
	e.MarketsLock.Unlock()

	log.Info("loaded vietnam markets", zap.Int("count", len(newMarkets)))
	return newMarkets, nil
}

func (e *Vietnam) parseMarket(sec *SSISecurityInfo) *banexg.Market {
	if sec.Symbol == "" || sec.Exchange == "" {
		return nil
	}

	symbol := BuildSymbol(sec.Exchange, sec.Symbol)

	active := true

	marketType := banexg.MarketSpot
	if sec.Type != "" {
		switch strings.ToUpper(sec.Type) {
		case "STOCK", "ETF":
			marketType = banexg.MarketSpot
		default:
			marketType = banexg.MarketSpot
		}
	}

	tickSize := GetPriceTickSize(10000.0, sec.Exchange)

	minLotSize := sec.LotSize
	if minLotSize == 0 {
		minLotSize = 100
	}

	minCost := float64(minLotSize) * tickSize

	market := &banexg.Market{
		ID:       sec.Symbol,
		Symbol:   symbol,
		Base:     sec.Symbol,
		Quote:    "VND",
		Settle:   "",
		BaseID:   sec.Symbol,
		QuoteID:  "VND",
		SettleID: "",
		Type:     marketType,
		Spot:     true,
		Margin:   false,
		Swap:     false,
		Future:   false,
		Option:   false,
		Active:   active,
		Contract: false,
		Linear:   false,
		Inverse:  false,
		Precision: &banexg.Precision{
			ModePrice:  banexg.PrecModeTickSize,
			ModeAmount: banexg.PrecModeDecimalPlace,
			Price:      tickSize,
			Amount:     0,
			Base:       0,
			Quote:      0,
		},
		Limits: &banexg.MarketLimits{
			Leverage: &banexg.LimitRange{Min: 1, Max: 1},
			Amount: &banexg.LimitRange{
				Min: float64(minLotSize),
				Max: 0,
			},
			Price: &banexg.LimitRange{
				Min: tickSize,
				Max: 0,
			},
			Cost: &banexg.LimitRange{
				Min: minCost,
				Max: 0,
			},
		},
		Info: make(map[string]interface{}),
	}

	return market
}

func (e *Vietnam) MapMarket(exgSymbol string, year int) (*banexg.Market, *errs.Error) {
	_, err := e.LoadMarkets(false, nil)
	if err != nil {
		return nil, err
	}

	mar := e.GetMarketById(exgSymbol, "")
	if mar != nil {
		return mar, nil
	}

	exchange, code := NormalizeSymbol(exgSymbol)
	normalized := BuildSymbol(exchange, code)

	e.MarketsLock.Lock()
	defer e.MarketsLock.Unlock()

	if mar, ok := e.Markets[normalized]; ok {
		return mar, nil
	}

	return nil, errs.NewMsg(errs.CodeUnsupportMarket, "market not found: %s", exgSymbol)
}
