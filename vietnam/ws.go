package vietnam

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

func (e *Vietnam) WatchOrderBooks(symbols []string, limit int, params map[string]interface{}) (chan *banexg.OrderBook, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchOrderBooks not yet implemented for Vietnam")
}

func (e *Vietnam) UnWatchOrderBooks(symbols []string, params map[string]interface{}) *errs.Error {
	return errs.NewMsg(errs.CodeNotImplement, "UnWatchOrderBooks not yet implemented for Vietnam")
}

func (e *Vietnam) WatchOHLCVs(jobs [][2]string, params map[string]interface{}) (chan *banexg.PairTFKline, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchOHLCVs not yet implemented for Vietnam")
}

func (e *Vietnam) UnWatchOHLCVs(jobs [][2]string, params map[string]interface{}) *errs.Error {
	return errs.NewMsg(errs.CodeNotImplement, "UnWatchOHLCVs not yet implemented for Vietnam")
}

func (e *Vietnam) WatchMarkPrices(symbols []string, params map[string]interface{}) (chan map[string]float64, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchMarkPrices not applicable for Vietnam stock market")
}

func (e *Vietnam) UnWatchMarkPrices(symbols []string, params map[string]interface{}) *errs.Error {
	return errs.NewMsg(errs.CodeNotImplement, "UnWatchMarkPrices not applicable for Vietnam stock market")
}

func (e *Vietnam) WatchTrades(symbols []string, params map[string]interface{}) (chan *banexg.Trade, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchTrades not yet implemented for Vietnam")
}

func (e *Vietnam) UnWatchTrades(symbols []string, params map[string]interface{}) *errs.Error {
	return errs.NewMsg(errs.CodeNotImplement, "UnWatchTrades not yet implemented for Vietnam")
}

func (e *Vietnam) WatchMyTrades(params map[string]interface{}) (chan *banexg.MyTrade, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchMyTrades not yet implemented for Vietnam")
}

func (e *Vietnam) WatchBalance(params map[string]interface{}) (chan *banexg.Balances, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchBalance not yet implemented for Vietnam")
}

func (e *Vietnam) WatchPositions(params map[string]interface{}) (chan []*banexg.Position, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchPositions not yet implemented for Vietnam")
}

func (e *Vietnam) WatchAccountConfig(params map[string]interface{}) (chan *banexg.AccountConfig, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "WatchAccountConfig not yet implemented for Vietnam")
}
