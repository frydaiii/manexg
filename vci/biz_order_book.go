package vci

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

func (e *VCI) FetchOrderBook(symbol string, limit int, params map[string]interface{}) (*banexg.OrderBook, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}
