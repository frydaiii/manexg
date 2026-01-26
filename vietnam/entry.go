package vietnam

import (
	"context"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/utils"
)

func New(Options map[string]interface{}) (*Vietnam, *errs.Error) {
	exg := &Vietnam{
		Exchange: &banexg.Exchange{
			ExgInfo: &banexg.ExgInfo{
				ID:        "vietnam",
				Name:      "Vietnam Stock Market",
				Countries: []string{"VN"},
				FullDay:   false,
				NoHoliday: false,
			},
			RateLimit: 100,
			Options:   Options,
			Hosts: &banexg.ExgHosts{
				Prod: map[string]string{
					HostDataAPI:    "https://fc-data.ssi.com.vn",
					HostTradingAPI: "https://fc-tradeapi.ssi.com.vn",
					HostWS:         "wss://fc-datahub.ssi.com.vn",
				},
				Www: "https://www.ssi.com.vn",
				Doc: []string{
					"https://guide.ssi.com.vn/ssi-products",
					"https://fc-data.ssi.com.vn/",
				},
				Fees: "https://www.ssi.com.vn/en/brokerage-fee",
			},
			Fees: &banexg.ExgFee{
				Main: &banexg.TradeFee{
					FeeSide:    "quote",
					TierBased:  false,
					Percentage: true,
					Taker:      0.0015,
					Maker:      0.0015,
				},
			},
			Apis: map[string]*banexg.Entry{
				MethodGetAccessToken:       {Path: "api/v2/Market/AccessToken", Host: HostDataAPI, Method: "POST", Cost: 1},
				MethodGetSecuritiesList:    {Path: "api/Market/GetSecuritiesList", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetSecuritiesDetails: {Path: "api/Market/GetSecuritiesDetails", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetDailyOHLC:         {Path: "api/Market/GetDailyOHLC", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetIntradayOHLC:      {Path: "api/Market/GetIntradayOHLC", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetDailyStockPrice:   {Path: "api/Market/GetDailyStockPrice", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetIndexComponents:   {Path: "api/Market/GetIndexComponents", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetIndexList:         {Path: "api/Market/GetIndexList", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetIndexSeries:       {Path: "api/Market/GetIndexSeries", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodPlaceOrder:           {Path: "api/Trading/NewOrder", Host: HostTradingAPI, Method: "POST", Cost: 1},
				MethodCancelOrder:          {Path: "api/Trading/CancelOrder", Host: HostTradingAPI, Method: "POST", Cost: 1},
				MethodModifyOrder:          {Path: "api/Trading/ModifyOrder", Host: HostTradingAPI, Method: "POST", Cost: 1},
				MethodGetOrderHistory:      {Path: "api/Trading/GetOrderHistory", Host: HostTradingAPI, Method: "GET", Cost: 1},
				MethodGetAccountBalance:    {Path: "api/Account/GetAccountBalance", Host: HostTradingAPI, Method: "GET", Cost: 1},
				MethodGetStockPosition:     {Path: "api/Account/GetStockPosition", Host: HostTradingAPI, Method: "GET", Cost: 1},
				MethodGetOrderDetail:       {Path: "api/Trading/GetOrderDetail", Host: HostTradingAPI, Method: "GET", Cost: 1},
				MethodGetCompanyInfo:       {Path: "api/Market/GetCompanyInfo", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetFinancialReport:   {Path: "api/Market/GetFinancialReport", Host: HostDataAPI, Method: "GET", Cost: 1},
				MethodGetTradingHolidays:   {Path: "api/Market/GetTradingHolidays", Host: HostDataAPI, Method: "GET", Cost: 1},
			},
			Has: map[string]map[string]int{
				"": {
					banexg.ApiFetchTicker:           banexg.HasOk,
					banexg.ApiFetchTickers:          banexg.HasOk,
					banexg.ApiFetchTickerPrice:      banexg.HasOk,
					banexg.ApiFetchOHLCV:            banexg.HasOk,
					banexg.ApiFetchOrderBook:        banexg.HasFail,
					banexg.ApiFetchOrder:            banexg.HasOk,
					banexg.ApiFetchOrders:           banexg.HasOk,
					banexg.ApiFetchBalance:          banexg.HasOk,
					banexg.ApiFetchOpenOrders:       banexg.HasOk,
					banexg.ApiCreateOrder:           banexg.HasOk,
					banexg.ApiCancelOrder:           banexg.HasOk,
					banexg.ApiEditOrder:             banexg.HasOk,
					banexg.ApiSetLeverage:           banexg.HasFail,
					banexg.ApiFetchPositions:        banexg.HasOk,
					banexg.ApiFetchAccountPositions: banexg.HasOk,
					banexg.ApiLoadLeverageBrackets:  banexg.HasFail,
					banexg.ApiGetLeverage:           banexg.HasFail,
					banexg.ApiCalcMaintMargin:       banexg.HasFail,
					banexg.ApiWatchOrderBooks:       banexg.HasFail,
					banexg.ApiWatchTrades:           banexg.HasFail,
					banexg.ApiWatchOHLCVs:           banexg.HasFail,
					banexg.ApiWatchMyTrades:         banexg.HasFail,
					banexg.ApiWatchBalance:          banexg.HasFail,
					banexg.ApiWatchPositions:        banexg.HasFail,
					banexg.ApiWatchAccountConfig:    banexg.HasFail,
				},
			},
		},
	}

	consumerID := utils.GetMapVal(Options, "consumerID", "")
	consumerSecret := utils.GetMapVal(Options, "consumerSecret", "")
	if consumerID == "" || consumerSecret == "" {
		return nil, errs.NewMsg(errs.CodeParamRequired, "consumerID and consumerSecret are required for Vietnam exchange")
	}

	exg.ConsumerID = consumerID
	exg.ConsumerSecret = consumerSecret

	err := exg.Init()
	if err != nil {
		return nil, err
	}

	return exg, nil
}

func NewExchange(Options map[string]interface{}) (banexg.BanExchange, *errs.Error) {
	return New(Options)
}

func (e *Vietnam) Init() *errs.Error {
	return e.Exchange.Init()
}

func (e *Vietnam) ensureValidToken() *errs.Error {
	if e.AccessToken != "" && time.Now().UnixMilli() < e.TokenExpiry-60000 {
		return nil
	}

	args := map[string]interface{}{
		"consumerID":     e.ConsumerID,
		"consumerSecret": e.ConsumerSecret,
	}

	rsp := e.Exchange.RequestApiRetry(context.Background(), MethodGetAccessToken, args, 1)
	if rsp.Error != nil {
		return rsp.Error
	}

	var authResp SSIAuthResponse
	if err := utils.UnmarshalString(rsp.Content, &authResp, utils.JsonNumDefault); err != nil {
		return errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse auth response: %v", err)
	}

	if authResp.AccessToken == "" {
		return errs.NewMsg(errs.CodeUnauthorized, "empty access token received")
	}

	e.AccessToken = authResp.AccessToken
	e.TokenExpiry = time.Now().UnixMilli() + authResp.ExpiresIn*1000

	return nil
}

func (e *Vietnam) RequestApiRetry(ctx context.Context, method string, args map[string]interface{}, retry int) *banexg.HttpRes {
	if method != MethodGetAccessToken {
		err := e.ensureValidToken()
		if err != nil {
			return &banexg.HttpRes{Error: err}
		}

		if e.AccessToken != "" {
			if args == nil {
				args = make(map[string]interface{})
			}
			args["Authorization"] = "Bearer " + e.AccessToken
		}
	}

	return e.Exchange.RequestApiRetry(ctx, method, args, retry)
}

func (e *Vietnam) Close() *errs.Error {
	return e.Exchange.Close()
}
