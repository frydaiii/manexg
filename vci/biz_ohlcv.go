package vci

import (
	"strconv"
	"strings"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	"github.com/banbox/banexg/utils"
	"go.uber.org/zap"
)

// vciOHLCData matches the VCI gap-chart API response.
// Note: the API returns timestamps as JSON strings (e.g. "1774396800"),
// so T is []string and parsed to int64 in parseVCIKlines.
type vciOHLCData struct {
	Symbol string    `json:"symbol"`
	T      []string  `json:"t"`
	O      []float64 `json:"o"`
	H      []float64 `json:"h"`
	L      []float64 `json:"l"`
	C      []float64 `json:"c"`
	V      []float64 `json:"v"`
}

func (e *VCI) FetchOHLCV(symbol, timeframe string, since int64, limit int, params map[string]interface{}) ([]*banexg.Kline, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	// Determine the VCI API symbol
	vciSymbol := ""
	if vs, ok := market.Info["vciSymbol"]; ok {
		vciSymbol = strings.TrimSpace(anyString(vs))
	}
	if vciSymbol == "" {
		vciSymbol = strings.ToUpper(strings.TrimSpace(anyString(market.Info["ticker"])))
	}
	if vciSymbol == "" {
		vciSymbol = market.Base
	}
	if vciSymbol == "" {
		return nil, errs.NewMsg(errs.CodeInvalidData, "empty market ticker")
	}

	// Resolve timeframe — only native resolutions supported in v1
	apiTF, ok := timeFrameMap[strings.ToLower(timeframe)]
	if !ok {
		return nil, errs.NewMsg(errs.CodeInvalidTimeFrame, "unsupported timeframe: %s (only 1m, 1h, 1d supported)", timeframe)
	}

	if limit <= 0 {
		limit = 200
	}

	until := utils.PopMapVal(args, banexg.ParamUntil, int64(0))
	if until <= 0 {
		until = e.MilliSeconds()
	}

	// countBack: overfetch with 1.5x multiplier, then trim to limit
	countBack := int(float64(limit) * 1.5)
	if since > 0 {
		estimated := estimateCountBack(since, until, apiTF)
		if estimated > countBack {
			countBack = estimated
		}
	}

	// Convert until from ms to seconds for VCI API
	toSec := until / 1000

	payload := map[string]interface{}{
		"timeFrame": apiTF,
		"symbols":   []string{vciSymbol},
		"to":        toSec,
		"countBack": countBack,
	}

	log.Info("vci ohlcv request",
		zap.String("symbol", vciSymbol),
		zap.String("timeFrame", apiTF),
		zap.Int64("to", toSec),
		zap.Int("countBack", countBack),
	)

	data, err := requestVCI[[]vciOHLCData](e, MethodPublicPostOhlcChart, payload, e.GetRetryNum("FetchOHLCV", 1))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []*banexg.Kline{}, nil
	}

	klines := parseVCIKlines(data[0])
	sortKlinesAsc(klines)
	return normalizeKlineRange(klines, since, until, limit), nil
}

func parseVCIKlines(data vciOHLCData) []*banexg.Kline {
	// Validate array lengths — use minimum safe length
	minLen := len(data.T)
	for _, arrLen := range []int{len(data.O), len(data.H), len(data.L), len(data.C), len(data.V)} {
		if arrLen < minLen {
			minLen = arrLen
		}
	}

	if minLen == 0 {
		return []*banexg.Kline{}
	}

	if minLen != len(data.T) || minLen != len(data.O) || minLen != len(data.H) ||
		minLen != len(data.L) || minLen != len(data.C) || minLen != len(data.V) {
		log.Warn("vci ohlcv column length mismatch, using safe minimum",
			zap.Int("t", len(data.T)),
			zap.Int("o", len(data.O)),
			zap.Int("h", len(data.H)),
			zap.Int("l", len(data.L)),
			zap.Int("c", len(data.C)),
			zap.Int("v", len(data.V)),
			zap.Int("safe_len", minLen),
		)
	}

	out := make([]*banexg.Kline, 0, minLen)
	for i := 0; i < minLen; i++ {
		ts, err := strconv.ParseInt(data.T[i], 10, 64)
		if err != nil || ts == 0 {
			continue
		}
		k := &banexg.Kline{
			Time:   ts * 1000, // seconds → milliseconds
			Open:   data.O[i],
			High:   data.H[i],
			Low:    data.L[i],
			Close:  data.C[i],
			Volume: data.V[i],
		}
		out = append(out, k)
	}
	return out
}

// estimateCountBack estimates the number of candles between since and until
// based on business days and the VN trading session schedule.
func estimateCountBack(since, until int64, apiTimeFrame string) int {
	if since <= 0 || until <= 0 || until <= since {
		return 0
	}
	calendarDays := int((until - since) / (1000 * 86400))
	if calendarDays <= 0 {
		calendarDays = 1
	}
	businessDays := calendarDays * 5 / 7
	if businessDays <= 0 {
		businessDays = 1
	}

	switch apiTimeFrame {
	case "ONE_DAY":
		return businessDays + 5
	case "ONE_HOUR":
		return businessDays*5 + 5 // ~5 trading hours/day
	case "ONE_MINUTE":
		return businessDays*255 + 100 // 255 trading minutes/day
	default:
		return businessDays + 5
	}
}
