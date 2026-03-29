package vci

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	"github.com/banbox/banexg/utils"
	"go.uber.org/zap"
)

func makeSign(e *VCI) banexg.FuncSign {
	return func(api *banexg.Entry, args map[string]interface{}) *banexg.HttpReq {
		params := utils.SafeParams(args)
		headers := http.Header{}
		headers.Set("Accept", "application/json, text/plain, */*")
		headers.Set("Content-Type", "application/json")
		headers.Set("Referer", "https://trading.vietcap.com.vn/")
		headers.Set("Origin", "https://trading.vietcap.com.vn")
		headers.Set("DNT", "1")
		headers.Set("sec-ch-ua-platform", `"Windows"`)
		headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")

		reqUrl := api.Url
		body := ""
		isGet := strings.EqualFold(api.Method, "GET")
		if isGet {
			if len(params) > 0 {
				qs := utils.UrlEncodeMap(params, true)
				if strings.Contains(reqUrl, "?") {
					reqUrl += "&" + qs
				} else {
					reqUrl += "?" + qs
				}
			}
		} else {
			if len(params) > 0 {
				text, err := utils.MarshalString(params)
				if err != nil {
					return &banexg.HttpReq{Error: errs.New(errs.CodeMarshalFail, err)}
				}
				body = text
			} else {
				body = "{}"
			}
		}
		return &banexg.HttpReq{
			Url:     reqUrl,
			Method:  api.Method,
			Headers: headers,
			Body:    body,
		}
	}
}

// requestVCI is a generic HTTP request helper for the VCI API.
// It handles both bare array responses and {"data": ...} envelope format.
func requestVCI[T any](e *VCI, endpoint string, payload map[string]interface{}, retryNum int) (T, *errs.Error) {
	var zero T
	api := e.Apis[endpoint]
	method, url := "", ""
	if api != nil {
		method = api.Method
		url = api.Url
	}
	log.Info("vci request",
		zap.String("endpoint", endpoint),
		zap.String("method", method),
		zap.String("url", url),
	)

	res := e.RequestApiRetryAdv(context.Background(), endpoint, payload, retryNum, false, false)
	if res.Error != nil {
		log.Warn("vci request failed",
			zap.String("endpoint", endpoint),
			zap.Error(res.Error),
		)
		return zero, res.Error
	}

	log.Info("vci response",
		zap.String("endpoint", endpoint),
		zap.Int("body_len", len(res.Content)),
	)

	// Try bare unmarshal first (e.g. `[{...}]`)
	var out T
	if err := utils.UnmarshalString(res.Content, &out, utils.JsonNumDefault); err == nil {
		return out, nil
	} else {
		log.Warn("vci bare unmarshal failed, trying envelope",
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
	}

	// Try envelope format: {"data": <T>}
	envelope := map[string]interface{}{}
	if err := utils.UnmarshalString(res.Content, &envelope, utils.JsonNumDefault); err != nil {
		return zero, errs.New(errs.CodeUnmarshalFail, err)
	}
	raw := envelope["data"]
	if raw == nil {
		return zero, errs.NewMsg(errs.CodeInvalidData, "vci: no data in response")
	}
	text, err2 := utils.MarshalString(raw)
	if err2 != nil {
		return zero, errs.New(errs.CodeMarshalFail, err2)
	}
	if err := utils.UnmarshalString(text, &out, utils.JsonNumDefault); err != nil {
		return zero, errs.New(errs.CodeUnmarshalFail, err)
	}
	return out, nil
}

func anyString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func anyInt64(v interface{}) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	case string:
		val, _ := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return val
	default:
		return 0
	}
}

func anyFloat(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case string:
		val, _ := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return val
	default:
		return 0
	}
}

func sortKlinesAsc(rows []*banexg.Kline) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Time < rows[j].Time
	})
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
