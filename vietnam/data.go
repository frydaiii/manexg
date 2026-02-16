package vietnam

import "github.com/banbox/banexg"

const (
	HostPublic = "public"
	HostStream = "stream"
)

const (
	MethodPublicPostMarketAccessToken    = "publicPostMarketAccessToken"
	MethodPublicPostMarketSecurities     = "publicPostMarketSecurities"
	MethodPublicPostMarketSecuritiesInfo = "publicPostMarketSecuritiesDetails"
	MethodPublicPostMarketDailyStock     = "publicPostMarketDailyStockPrice"
	MethodPublicPostMarketIntradayOHLC   = "publicPostMarketIntradayOhlc"
)

var (
	timeFrameMap = map[string]string{
		"1m":  "1P",
		"5m":  "5P",
		"15m": "15P",
		"30m": "30P",
		"1h":  "1H",
		"1d":  "1D",
	}

	marketBoards = []string{"HOSE", "HNX", "UPCOM"}

	statusSuccess = map[string]bool{
		"SUCCESS": true,
		"OK":      true,
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
