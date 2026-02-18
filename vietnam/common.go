package vietnam

import (
	"context"
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/log"
	"github.com/banbox/banexg/utils"
	"go.uber.org/zap"
)

func makeSign(e *Vietnam) banexg.FuncSign {
	return func(api *banexg.Entry, args map[string]interface{}) *banexg.HttpReq {
		params := utils.SafeParams(args)
		headers := http.Header{}
		headers.Set("Accept", "application/json")
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
			headers.Set("Content-Type", "application/json")
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
		if utils.GetMapVal(api.More, "auth", false) {
			token, err := e.getAccessToken()
			if err != nil {
				return &banexg.HttpReq{Error: err, Private: true}
			}
			headers.Set("Authorization", "Bearer "+token)
		}
		return &banexg.HttpReq{
			Url:     reqUrl,
			Method:  api.Method,
			Headers: headers,
			Body:    body,
			Private: utils.GetMapVal(api.More, "auth", false),
		}
	}
}

func (e *Vietnam) getAccessToken() (string, *errs.Error) {
	now := e.MilliSeconds()
	e.tokenLock.RLock()
	if e.token != "" && now < e.tokenExp-5*60*1000 {
		token := e.token
		e.tokenLock.RUnlock()
		return token, nil
	}
	e.tokenLock.RUnlock()

	e.tokenLock.Lock()
	defer e.tokenLock.Unlock()
	now = e.MilliSeconds()
	if e.token != "" && now < e.tokenExp-5*60*1000 {
		return e.token, nil
	}

	consumerID := strings.TrimSpace(utils.GetMapVal(e.Options, "ConsumerID", ""))
	if consumerID == "" {
		consumerID = strings.TrimSpace(utils.GetMapVal(e.Options, banexg.OptApiKey, ""))
	}
	consumerSecret := strings.TrimSpace(utils.GetMapVal(e.Options, "ConsumerSecret", ""))
	if consumerSecret == "" {
		consumerSecret = strings.TrimSpace(utils.GetMapVal(e.Options, banexg.OptApiSecret, ""))
	}
	if consumerID == "" || consumerSecret == "" {
		return "", errs.NewMsg(errs.CodeCredsRequired, "ConsumerID/ConsumerSecret required")
	}
	payload := map[string]interface{}{
		"consumerID":     consumerID,
		"consumerSecret": consumerSecret,
	}
	res := e.RequestApiRetryAdv(context.Background(), MethodPublicPostMarketAccessToken, payload, e.GetRetryNum("AccessToken", 1), false, false)
	if res.Error != nil {
		return "", res.Error
	}
	obj, err := decodeSSIMap(res.Content)
	if err != nil {
		return "", err
	}
	// Auth endpoint returns responseCode/token at top level, not status/data
	respCode := anyInt64(obj["responseCode"])
	if respCode != 0 {
		msg := strings.TrimSpace(anyString(obj["message"]))
		if msg == "" {
			msg = "auth failed"
		}
		return "", errs.NewMsg(errs.CodeRunTime, msg)
	}
	token := strings.TrimSpace(anyString(obj["token"]))
	if token == "" {
		// fallback: try nested data.accessToken for backward compat
		data := mapStringAny(obj["data"])
		token = strings.TrimSpace(anyString(data["accessToken"]))
	}
	if token == "" {
		token = strings.TrimSpace(anyString(obj["accessToken"]))
	}
	if token == "" {
		return "", errs.NewMsg(errs.CodeInvalidData, "empty access token")
	}
	expMS := parseJWTExpMS(token)
	if expMS == 0 {
		expMS = now + 2*60*60*1000
	}
	e.token = token
	e.tokenExp = expMS
	return token, nil
}

func requestSSI[T any](e *Vietnam, endpoint string, payload map[string]interface{}, retryNum int) (T, *errs.Error) {
	var zero T
	api := e.Apis[endpoint]
	method, url := "", ""
	if api != nil {
		method = api.Method
		url = api.Url
	}
	log.Info("vietnam ssi request",
		zap.String("endpoint", endpoint),
		zap.String("method", method),
		zap.String("url", url),
		zap.Any("payload", sanitizePayload(payload)),
	)
	res := e.RequestApiRetryAdv(context.Background(), endpoint, payload, retryNum, false, false)
	if res.Error != nil {
		log.Warn("vietnam ssi request failed",
			zap.String("endpoint", endpoint),
			zap.String("method", method),
			zap.String("url", url),
			zap.Error(res.Error),
		)
		return zero, res.Error
	}
	obj, err := decodeSSIMap(res.Content)
	if err != nil {
		log.Warn("vietnam ssi decode failed",
			zap.String("endpoint", endpoint),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("body_len", len(res.Content)),
			zap.Error(err),
		)
		return zero, err
	}
	if e2 := ensureSSISuccess(obj); e2 != nil {
		log.Warn("vietnam ssi response not success",
			zap.String("endpoint", endpoint),
			zap.String("method", method),
			zap.String("url", url),
			zap.Any("status", obj["status"]),
			zap.String("message", strings.TrimSpace(anyString(obj["message"]))),
			zap.Int("body_len", len(res.Content)),
		)
		return zero, e2
	}
	// Spec uses "dataList" for most endpoints, "data" for Securities
	raw := obj["dataList"]
	if raw == nil {
		raw = obj["data"]
	}
	log.Info("vietnam ssi response",
		zap.String("endpoint", endpoint),
		zap.String("method", method),
		zap.String("url", url),
		zap.Any("status", obj["status"]),
		zap.String("message", strings.TrimSpace(anyString(obj["message"]))),
		zap.Int("items", rawItemCount(raw)),
		zap.Int("body_len", len(res.Content)),
	)
	text, err2 := utils.MarshalString(raw)
	if err2 != nil {
		return zero, errs.New(errs.CodeMarshalFail, err2)
	}
	var out T
	if err3 := utils.UnmarshalString(text, &out, utils.JsonNumDefault); err3 != nil {
		return zero, errs.New(errs.CodeUnmarshalFail, err3)
	}
	return out, nil
}

func rawItemCount(v interface{}) int {
	switch items := v.(type) {
	case []interface{}:
		return len(items)
	case []map[string]interface{}:
		return len(items)
	case nil:
		return 0
	default:
		return 1
	}
}

func sanitizePayload(payload map[string]interface{}) map[string]interface{} {
	if len(payload) == 0 {
		return map[string]interface{}{}
	}
	res := make(map[string]interface{}, len(payload))
	for k, v := range payload {
		low := strings.ToLower(strings.TrimSpace(k))
		if strings.Contains(low, "secret") || strings.Contains(low, "token") || strings.Contains(low, "password") {
			res[k] = "***"
			continue
		}
		res[k] = v
	}
	return res
}

func decodeSSIMap(content string) (map[string]interface{}, *errs.Error) {
	obj := map[string]interface{}{}
	if err := utils.UnmarshalString(content, &obj, utils.JsonNumDefault); err != nil {
		return nil, errs.New(errs.CodeUnmarshalFail, err)
	}
	return obj, nil
}

func ensureSSISuccess(obj map[string]interface{}) *errs.Error {
	if obj == nil {
		return errs.NewMsg(errs.CodeInvalidData, "empty response")
	}
	status := obj["status"]
	switch v := status.(type) {
	case int64:
		if v == 200 {
			return nil
		}
	case float64:
		if int64(v) == 200 {
			return nil
		}
	case string:
		if statusSuccess[strings.ToUpper(strings.TrimSpace(v))] {
			return nil
		}
	}
	msg := strings.TrimSpace(anyString(obj["message"]))
	if msg == "" {
		msg = "ssi api failed"
	}
	return errs.NewMsg(errs.CodeRunTime, msg)
}

func parseJWTExpMS(token string) int64 {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return 0
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0
	}
	obj := map[string]interface{}{}
	if err = utils.Unmarshal(payload, &obj, utils.JsonNumDefault); err != nil {
		return 0
	}
	exp := anyInt64(obj["exp"])
	if exp <= 0 {
		return 0
	}
	return exp * 1000
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

func mapStringAny(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{}
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func sortKlinesAsc(rows []*banexg.Kline) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Time < rows[j].Time
	})
}

func msToSSIDate(ms int64) string {
	loc := time.FixedZone("ICT", 7*60*60)
	return time.UnixMilli(ms).In(loc).Format("02/01/2006")
}

func parseSSIDateTimeMs(dateText, timeText string) int64 {
	loc := time.FixedZone("ICT", 7*60*60)
	dateText = strings.TrimSpace(dateText)
	if dateText == "" {
		return 0
	}
	if strings.TrimSpace(timeText) == "" {
		t, err := time.ParseInLocation("02/01/2006", dateText, loc)
		if err != nil {
			return 0
		}
		return t.UnixMilli()
	}
	t, err := time.ParseInLocation("02/01/2006 15:04:05", dateText+" "+strings.TrimSpace(timeText), loc)
	if err != nil {
		t2, err2 := time.ParseInLocation("02/01/2006", dateText, loc)
		if err2 != nil {
			return 0
		}
		return t2.UnixMilli()
	}
	return t.UnixMilli()
}
