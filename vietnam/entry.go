package vietnam

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

func New(options map[string]interface{}) (*Vietnam, *errs.Error) {
	api := func(path, host, method string, auth bool) *banexg.Entry {
		return &banexg.Entry{
			Path:   path,
			Host:   host,
			Method: method,
			Cost:   1,
			More: map[string]interface{}{
				"auth": auth,
			},
		}
	}
	exg := &Vietnam{
		Exchange: &banexg.Exchange{
			ExgInfo: &banexg.ExgInfo{
				ID:        "vietnam",
				Name:      "Vietnam",
				Countries: []string{"VN"},
			},
			RateLimit:  50,
			Options:    options,
			TimeFrames: timeFrameMap,
			Hosts: &banexg.ExgHosts{
				Test: map[string]string{
					HostPublic: "https://fc-data.ssi.com.vn",
					HostStream: "wss://fc-datahub.ssi.com.vn/v2.0",
				},
				Prod: map[string]string{
					HostPublic: "https://fc-data.ssi.com.vn",
					HostStream: "wss://fc-datahub.ssi.com.vn/v2.0",
				},
				Www: "https://www.ssi.com.vn",
				Doc: []string{
					"https://fc-data.ssi.com.vn/Help",
				},
			},
			Apis: map[string]*banexg.Entry{
				MethodPublicPostMarketAccessToken:    api("api/v2/Market/AccessToken", HostPublic, "POST", false),
				MethodPublicPostMarketSecurities:     api("api/v2/Market/Securities", HostPublic, "GET", true),
				MethodPublicPostMarketSecuritiesInfo: api("api/v2/Market/SecuritiesDetails", HostPublic, "GET", true),
				MethodPublicPostMarketDailyStock:     api("api/v2/Market/DailyStockPrice", HostPublic, "GET", true),
				MethodPublicPostMarketIntradayOHLC:   api("api/v2/Market/IntradayOhlc", HostPublic, "POST", true),
				MethodPublicGetMarketDailyOhlc:       api("api/v2/Market/DailyOhlc", HostPublic, "GET", true),
				MethodPublicGetMarketIndexList:       api("api/v2/Market/IndexList", HostPublic, "GET", true),
				MethodPublicGetMarketIndexComponents: api("api/v2/Market/IndexComponents", HostPublic, "GET", true),
				MethodPublicGetMarketDailyIndex:      api("api/v2/Market/DailyIndex", HostPublic, "GET", true),
			},
			Has: defaultHas,
		},
		marketsByRawID:  map[string]*banexg.Market{},
		marketsByTicker: map[string][]*banexg.Market{},
	}
	exg.Sign = makeSign(exg)
	err := exg.Init()
	if err != nil {
		return nil, err
	}
	return exg, nil
}

func NewExchange(options map[string]interface{}) (banexg.BanExchange, *errs.Error) {
	return New(options)
}
