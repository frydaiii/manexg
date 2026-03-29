package vci

import "github.com/banbox/banexg"

const (
	HostPublic = "public"
)

const (
	MethodPublicGetSymbolsAll  = "publicGetSymbolsAll"
	MethodPublicPostOhlcChart  = "publicPostOhlcChart"
)

type boardDefaults struct {
	priceTick float64
	lotSize   float64
}

var (
	timeFrameMap = map[string]string{
		"1m": "ONE_MINUTE",
		"1h": "ONE_HOUR",
		"1d": "ONE_DAY",
	}

	marketBoards = []string{"HOSE", "HNX", "UPCOM"}

	// indexSymbolMap maps our canonical index names to VCI API symbols
	indexSymbolMap = map[string]string{
		"VNINDEX":    "VNINDEX",
		"HNXINDEX":   "HNXIndex",
		"UPCOMINDEX": "HNXUpcomIndex",
	}

	// boardPrecision provides default tick sizes and lot sizes per board
	boardPrecision = map[string]*boardDefaults{
		"HOSE":  {priceTick: 10, lotSize: 100},
		"HNX":   {priceTick: 100, lotSize: 100},
		"UPCOM": {priceTick: 100, lotSize: 1},
	}

	defaultHas = map[string]map[string]int{
		"": {
			banexg.ApiFetchTicker:           banexg.HasFail,
			banexg.ApiFetchTickers:          banexg.HasFail,
			banexg.ApiFetchTickerPrice:      banexg.HasFail,
			banexg.ApiLoadLeverageBrackets:  banexg.HasFail,
			banexg.ApiFetchCurrencies:       banexg.HasFail,
			banexg.ApiGetLeverage:           banexg.HasFail,
			banexg.ApiFetchOHLCV:            banexg.HasOk,
			banexg.ApiFetchOrderBook:        banexg.HasFail,
			banexg.ApiFetchOrder:            banexg.HasFail,
			banexg.ApiFetchOrders:           banexg.HasFail,
			banexg.ApiFetchBalance:          banexg.HasFail,
			banexg.ApiFetchAccountPositions: banexg.HasFail,
			banexg.ApiFetchPositions:        banexg.HasFail,
			banexg.ApiFetchOpenOrders:       banexg.HasFail,
			banexg.ApiCreateOrder:           banexg.HasFail,
			banexg.ApiEditOrder:             banexg.HasFail,
			banexg.ApiCancelOrder:           banexg.HasFail,
			banexg.ApiSetLeverage:           banexg.HasFail,
			banexg.ApiCalcMaintMargin:       banexg.HasFail,
			banexg.ApiWatchOrderBooks:       banexg.HasFail,
			banexg.ApiUnWatchOrderBooks:     banexg.HasFail,
			banexg.ApiWatchOHLCVs:           banexg.HasFail,
			banexg.ApiUnWatchOHLCVs:         banexg.HasFail,
			banexg.ApiWatchMarkPrices:       banexg.HasFail,
			banexg.ApiUnWatchMarkPrices:     banexg.HasFail,
			banexg.ApiWatchTrades:           banexg.HasFail,
			banexg.ApiUnWatchTrades:         banexg.HasFail,
			banexg.ApiWatchMyTrades:         banexg.HasFail,
			banexg.ApiWatchBalance:          banexg.HasFail,
			banexg.ApiWatchPositions:        banexg.HasFail,
			banexg.ApiWatchAccountConfig:    banexg.HasFail,
		},
	}
)
