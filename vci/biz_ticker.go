package vci

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

func (e *VCI) FetchTicker(symbol string, params map[string]interface{}) (*banexg.Ticker, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *VCI) FetchTickers(symbols []string, params map[string]interface{}) ([]*banexg.Ticker, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}
