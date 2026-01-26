package vietnam

import (
	"context"

	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/utils"
)

func (e *Vietnam) GetCompanyInfo(symbol string, params map[string]interface{}) (*CompanyInfo, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	args["symbol"] = market.ID

	rsp := e.RequestApiRetry(context.Background(), MethodGetCompanyInfo, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var result struct {
		Status  int         `json:"status"`
		Message string      `json:"message"`
		Data    CompanyInfo `json:"data"`
	}

	if err := utils.UnmarshalString(rsp.Content, &result, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse company info: %v", err)
	}

	if result.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", result.Message)
	}

	return &result.Data, nil
}

func (e *Vietnam) GetFinancialReport(symbol, reportType, period string, year int, params map[string]interface{}) (*FinancialReport, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	args["symbol"] = market.ID
	args["reportType"] = reportType
	args["period"] = period
	args["year"] = year

	rsp := e.RequestApiRetry(context.Background(), MethodGetFinancialReport, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var result struct {
		Status  int             `json:"status"`
		Message string          `json:"message"`
		Data    FinancialReport `json:"data"`
	}

	if err := utils.UnmarshalString(rsp.Content, &result, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse financial report: %v", err)
	}

	if result.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", result.Message)
	}

	return &result.Data, nil
}

func (e *Vietnam) GetTradingHolidays(year int, params map[string]interface{}) ([]TradingHoliday, *errs.Error) {
	args := utils.SafeParams(params)
	args["year"] = year

	rsp := e.RequestApiRetry(context.Background(), MethodGetTradingHolidays, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var result struct {
		Status  int              `json:"status"`
		Message string           `json:"message"`
		Data    []TradingHoliday `json:"data"`
	}

	if err := utils.UnmarshalString(rsp.Content, &result, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse holidays: %v", err)
	}

	if result.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", result.Message)
	}

	return result.Data, nil
}

func IsHoliday(date string, holidays []TradingHoliday) bool {
	for _, holiday := range holidays {
		if holiday.Date == date {
			return true
		}
	}
	return false
}
