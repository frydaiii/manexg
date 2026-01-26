package vietnam

import (
	"context"
	"fmt"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/utils"
)

func (e *Vietnam) FetchTicker(symbol string, params map[string]interface{}) (*banexg.Ticker, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	args["symbol"] = market.ID

	rsp := e.RequestApiRetry(context.Background(), MethodGetSecuritiesDetails, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var detailRsp SSISecurityDetailRsp
	if err := utils.UnmarshalString(rsp.Content, &detailRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse security detail: %v", err)
	}

	if detailRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", detailRsp.Message)
	}

	return e.parseTicker(&detailRsp.Data, market), nil
}

func (e *Vietnam) FetchTickers(symbols []string, params map[string]interface{}) ([]*banexg.Ticker, *errs.Error) {
	if len(symbols) == 0 {
		return nil, errs.NewMsg(errs.CodeParamRequired, "symbols required for FetchTickers")
	}

	args := utils.SafeParams(params)
	symbolsStr := ""
	for i, sym := range symbols {
		market, err := e.GetMarket(sym)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			symbolsStr += ","
		}
		symbolsStr += market.ID
	}
	args["symbols"] = symbolsStr

	rsp := e.RequestApiRetry(context.Background(), MethodGetDailyStockPrice, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var pricesRsp SSIDailyStockPriceRsp
	if err := utils.UnmarshalString(rsp.Content, &pricesRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse daily prices: %v", err)
	}

	if pricesRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", pricesRsp.Message)
	}

	tickers := make([]*banexg.Ticker, 0, len(pricesRsp.Data))
	for _, data := range pricesRsp.Data {
		market, err := e.GetMarket(data.Symbol)
		if err != nil {
			continue
		}
		ticker := e.parseDailyPriceToTicker(&data, market)
		if ticker != nil {
			tickers = append(tickers, ticker)
		}
	}

	return tickers, nil
}

func (e *Vietnam) FetchOHLCV(symbol, timeframe string, since int64, limit int, params map[string]interface{}) ([]*banexg.Kline, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	resolution := timeframeToSSIResolution(timeframe)
	if resolution == "" {
		return nil, errs.NewMsg(errs.CodeParamInvalid, "unsupported timeframe: %s", timeframe)
	}

	args["symbol"] = market.ID
	args["resolution"] = resolution

	if since > 0 {
		fromDate := time.UnixMilli(since).In(VietnamLocation).Format("2006-01-02")
		args["fromDate"] = fromDate
	}

	if limit > 0 {
		args["pageSize"] = limit
	}

	method := MethodGetIntradayOHLC
	if resolution == "D" {
		method = MethodGetDailyOHLC
	}

	rsp := e.RequestApiRetry(context.Background(), method, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var ohlcRsp SSIOHLCRsp
	if err := utils.UnmarshalString(rsp.Content, &ohlcRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse OHLC: %v", err)
	}

	if ohlcRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", ohlcRsp.Message)
	}

	klines := make([]*banexg.Kline, 0, len(ohlcRsp.Data))
	for _, data := range ohlcRsp.Data {
		kline := e.parseOHLC(&data, market)
		if kline != nil {
			klines = append(klines, kline)
		}
	}

	return klines, nil
}

func (e *Vietnam) FetchOrderBook(symbol string, limit int, params map[string]interface{}) (*banexg.OrderBook, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "FetchOrderBook not yet implemented for Vietnam")
}

func (e *Vietnam) parseTicker(data *SSISecurityDetail, market *banexg.Market) *banexg.Ticker {
	timestamp := time.Now().UnixMilli()
	if data.Time != "" {
		if t, err := parseVietnamTime(data.Time); err == nil {
			timestamp = t.UnixMilli()
		}
	}

	var percentage float64
	if data.RefPrice > 0 && data.LastPrice > 0 {
		percentage = ((data.LastPrice - data.RefPrice) / data.RefPrice) * 100
	}

	return &banexg.Ticker{
		Symbol:        market.Symbol,
		TimeStamp:     timestamp,
		High:          data.HighPrice,
		Low:           data.LowPrice,
		Open:          data.OpenPrice,
		Close:         data.LastPrice,
		Last:          data.LastPrice,
		Bid:           data.BidPrice1,
		BidVolume:     data.BidVolume1,
		Ask:           data.AskPrice1,
		AskVolume:     data.AskVolume1,
		BaseVolume:    data.TotalVolume,
		QuoteVolume:   data.TotalValue,
		PreviousClose: data.RefPrice,
		Change:        data.LastPrice - data.RefPrice,
		Percentage:    percentage,
		Average:       data.AvgPrice,
		Info:          map[string]interface{}{"raw": data},
	}
}

func (e *Vietnam) parseDailyPriceToTicker(data *SSIDailyStockPrice, market *banexg.Market) *banexg.Ticker {
	timestamp := time.Now().UnixMilli()
	if data.TradingDate != "" {
		if t, err := parseVietnamTime(data.TradingDate); err == nil {
			timestamp = t.UnixMilli()
		}
	}

	var percentage float64
	if data.PriorClosePrice > 0 && data.ClosePrice > 0 {
		percentage = ((data.ClosePrice - data.PriorClosePrice) / data.PriorClosePrice) * 100
	}

	return &banexg.Ticker{
		Symbol:        market.Symbol,
		TimeStamp:     timestamp,
		High:          data.HighestPrice,
		Low:           data.LowestPrice,
		Open:          data.OpenPrice,
		Close:         data.ClosePrice,
		Last:          data.ClosePrice,
		BaseVolume:    data.TotalVolume,
		QuoteVolume:   data.TotalValue,
		PreviousClose: data.PriorClosePrice,
		Change:        data.ClosePrice - data.PriorClosePrice,
		Percentage:    percentage,
		Info:          map[string]interface{}{"raw": data},
	}
}

func (e *Vietnam) parseOHLC(data *SSIOHLCData, market *banexg.Market) *banexg.Kline {
	timestamp, err := parseVietnamTime(data.TradingDate)
	if err != nil {
		return nil
	}

	return &banexg.Kline{
		Time:   timestamp.UnixMilli(),
		Open:   data.Open,
		High:   data.High,
		Low:    data.Low,
		Close:  data.Close,
		Volume: data.Volume,
		Info:   data.Value,
	}
}

func timeframeToSSIResolution(timeframe string) string {
	switch timeframe {
	case "1m":
		return "1"
	case "3m":
		return "3"
	case "5m":
		return "5"
	case "15m":
		return "15"
	case "30m":
		return "30"
	case "1h":
		return "60"
	case "1d":
		return "D"
	default:
		return ""
	}
}

func parseVietnamTime(timeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006 15:04:05",
		"02/01/2006",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, timeStr, VietnamLocation); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}
