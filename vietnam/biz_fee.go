package vietnam

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

func (e *Vietnam) CalculateFee(symbol, odType, side string, amount float64, price float64, isMaker bool, params map[string]interface{}) (*banexg.Fee, *errs.Error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return nil, err
	}

	var rate float64
	if isMaker {
		rate = market.Maker
	} else {
		rate = market.Taker
	}

	cost := amount * price
	feeCost := cost * rate

	return &banexg.Fee{
		IsMaker:  isMaker,
		Currency: "VND",
		Cost:     feeCost,
		Rate:     rate,
	}, nil
}
