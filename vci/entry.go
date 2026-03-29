package vci

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

func New(options map[string]interface{}) (*VCI, *errs.Error) {
	api := func(path, host, method string) *banexg.Entry {
		return &banexg.Entry{
			Path:   path,
			Host:   host,
			Method: method,
			Cost:   1,
			More: map[string]interface{}{
				"auth": false,
			},
		}
	}
	exg := &VCI{
		Exchange: &banexg.Exchange{
			ExgInfo: &banexg.ExgInfo{
				ID:        "vci",
				Name:      "VietCap",
				Countries: []string{"VN"},
			},
			RateLimit:  500, // guess — no documented limit, expect production tuning
			Options:    options,
			TimeFrames: timeFrameMap,
			Hosts: &banexg.ExgHosts{
				Test: map[string]string{
					HostPublic: "https://trading.vietcap.com.vn/api",
				},
				Prod: map[string]string{
					HostPublic: "https://trading.vietcap.com.vn/api",
				},
				Www: "https://trading.vietcap.com.vn",
				Doc: []string{
					"https://github.com/thinh-vu/vnstock",
				},
			},
			Apis: map[string]*banexg.Entry{
				MethodPublicGetSymbolsAll: api("price/symbols/getAll", HostPublic, "GET"),
				MethodPublicPostOhlcChart: api("chart/OHLCChart/gap-chart", HostPublic, "POST"),
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
