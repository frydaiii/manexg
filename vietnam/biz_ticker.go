package vietnam

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

func (e *Vietnam) FetchTicker(symbol string, params map[string]interface{}) (*banexg.Ticker, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Vietnam) FetchTickers(symbols []string, params map[string]interface{}) ([]*banexg.Ticker, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}
