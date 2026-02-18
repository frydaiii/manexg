package vietnam

import (
	"strings"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/utils"
)

func (e *Vietnam) FetchOHLCV(symbol, timeframe string, since int64, limit int, params map[string]interface{}) ([]*banexg.Kline, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}
	ticker := strings.ToUpper(strings.TrimSpace(anyString(market.Info["ticker"])))
	if ticker == "" {
		ticker = market.Base
	}
	if ticker == "" {
		return nil, errs.NewMsg(errs.CodeInvalidData, "empty market ticker")
	}
	if limit <= 0 {
		limit = 200
	}
	until := utils.PopMapVal(args, banexg.ParamUntil, int64(0))
	if until <= 0 {
		until = e.MilliSeconds()
	}
	from := since
	if from <= 0 {
		from = inferSinceByLimit(timeframe, until, limit)
	}

	if strings.EqualFold(timeframe, "1d") {
		return e.fetchDailyOHLCV(ticker, from, until, limit)
	}
	res := e.GetTimeFrame(timeframe)
	if res == "" || strings.EqualFold(res, "1D") {
		return nil, errs.NewMsg(errs.CodeInvalidTimeFrame, "invalid timeframe: %s", timeframe)
	}
	return e.fetchIntradayOHLCV(ticker, res, from, until, limit)
}

func inferSinceByLimit(timeframe string, until int64, limit int) int64 {
	if limit <= 0 {
		limit = 200
	}
	secs := utils.TFToSecs(timeframe)
	if secs <= 0 {
		secs = 60
	}
	return until - int64(limit*secs*1000)
}

func (e *Vietnam) fetchDailyOHLCV(ticker string, since, until int64, limit int) ([]*banexg.Kline, *errs.Error) {
	payload := map[string]interface{}{
		"Symbol":    ticker,
		"FromDate":  msToSSIDate(since),
		"ToDate":    msToSSIDate(until),
		"PageIndex": 1,
		"PageSize":  max(limit, 100),
		"ascending": true,
	}
	rows, err := requestSSI[[]map[string]interface{}](e, MethodPublicGetMarketDailyOhlc, payload, e.GetRetryNum("FetchOHLCV", 1))
	if err != nil {
		return nil, err
	}
	out := parseSSIKlines(rows, true)
	return normalizeKlineRange(out, since, until, limit), nil
}

func (e *Vietnam) fetchIntradayOHLCV(ticker, resolution string, since, until int64, limit int) ([]*banexg.Kline, *errs.Error) {
	payload := map[string]interface{}{
		"Symbol":     ticker,
		"FromDate":   msToSSIDate(since),
		"ToDate":     msToSSIDate(until),
		"PageIndex":  1,
		"PageSize":   max(limit, 1000),
		"resolution": resolution,
		"ascending":  true,
	}
	rows, err := requestSSI[[]map[string]interface{}](e, MethodPublicPostMarketIntradayOHLC, payload, e.GetRetryNum("FetchOHLCV", 1))
	if err != nil {
		return nil, err
	}
	out := parseSSIKlines(rows, false)
	return normalizeKlineRange(out, since, until, limit), nil
}

func parseSSIKlines(rows []map[string]interface{}, daily bool) []*banexg.Kline {
	out := make([]*banexg.Kline, 0, len(rows))
	for _, row := range rows {
		dateText := anyString(row["tradingDate"])
		if dateText == "" {
			dateText = anyString(row["TradingDate"])
		}
		timeText := ""
		if !daily {
			timeText = anyString(row["time"])
			if timeText == "" {
				timeText = anyString(row["Time"])
			}
		}
		stamp := parseSSIDateTimeMs(dateText, timeText)
		if stamp == 0 {
			continue
		}
		k := &banexg.Kline{
			Time:   stamp,
			Open:   firstFloat(row, "open", "Open", "openPrice", "OpenPrice"),
			High:   firstFloat(row, "high", "High", "highestPrice", "HighestPrice"),
			Low:    firstFloat(row, "low", "Low", "lowestPrice", "LowestPrice"),
			Close:  firstFloat(row, "close", "Close", "closePrice", "ClosePrice"),
			Volume: firstFloat(row, "volume", "Volume", "totalMatchVol", "TotalMatchVol", "totalTradedVol", "TotalTradedVol"),
			Quote:  firstFloat(row, "value", "Value", "totalMatchVal", "TotalMatchVal", "totalTradedValue", "TotalTradedValue"),
		}
		out = append(out, k)
	}
	sortKlinesAsc(out)
	return out
}

func firstFloat(row map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if val, ok := row[key]; ok {
			return anyFloat(val)
		}
	}
	return 0
}

func normalizeKlineRange(rows []*banexg.Kline, since, until int64, limit int) []*banexg.Kline {
	if len(rows) == 0 {
		return rows
	}
	out := make([]*banexg.Kline, 0, len(rows))
	for _, row := range rows {
		if since > 0 && row.Time < since {
			continue
		}
		if until > 0 && row.Time > until {
			continue
		}
		out = append(out, row)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseDate(dateText string) time.Time {
	loc := time.FixedZone("ICT", 7*60*60)
	t, _ := time.ParseInLocation("02/01/2006", dateText, loc)
	return t
}
