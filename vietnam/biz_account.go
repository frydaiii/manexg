package vietnam

import (
	"context"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/utils"
)

func (e *Vietnam) FetchBalance(params map[string]interface{}) (*banexg.Balances, *errs.Error) {
	args := utils.SafeParams(params)

	accountNo := utils.GetMapVal(args, "accountNo", "")
	if accountNo == "" && e.DefAccName != "" {
		if acc, ok := e.Accounts[e.DefAccName]; ok {
			accountNo = acc.Name
		}
	}
	if accountNo == "" {
		return nil, errs.NewMsg(errs.CodeParamRequired, "accountNo required")
	}

	args["accountNo"] = accountNo

	rsp := e.RequestApiRetry(context.Background(), MethodGetAccountBalance, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var balanceRsp SSIBalanceRsp
	if err := utils.UnmarshalString(rsp.Content, &balanceRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse balance: %v", err)
	}

	if balanceRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", balanceRsp.Message)
	}

	return e.parseBalance(&balanceRsp.Data), nil
}

func (e *Vietnam) FetchPositions(symbols []string, params map[string]interface{}) ([]*banexg.Position, *errs.Error) {
	args := utils.SafeParams(params)

	accountNo := utils.GetMapVal(args, "accountNo", "")
	if accountNo == "" && e.DefAccName != "" {
		if acc, ok := e.Accounts[e.DefAccName]; ok {
			accountNo = acc.Name
		}
	}
	if accountNo == "" {
		return nil, errs.NewMsg(errs.CodeParamRequired, "accountNo required")
	}

	args["accountNo"] = accountNo

	rsp := e.RequestApiRetry(context.Background(), MethodGetStockPosition, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var positionRsp SSIPositionRsp
	if err := utils.UnmarshalString(rsp.Content, &positionRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse positions: %v", err)
	}

	if positionRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", positionRsp.Message)
	}

	positions := make([]*banexg.Position, 0, len(positionRsp.Data))
	for _, posInfo := range positionRsp.Data {
		if len(symbols) > 0 {
			found := false
			for _, sym := range symbols {
				if market, _ := e.GetMarket(sym); market != nil && market.ID == posInfo.Symbol {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		position := e.parsePosition(&posInfo)
		if position != nil {
			positions = append(positions, position)
		}
	}

	return positions, nil
}

func (e *Vietnam) parseBalance(data *SSIBalanceInfo) *banexg.Balances {
	if data == nil {
		return nil
	}

	currency := "VND"
	if data.Currency != "" {
		currency = data.Currency
	}

	free := map[string]float64{
		currency: data.AvailableCash,
	}

	used := map[string]float64{
		currency: data.TotalCash - data.AvailableCash,
	}

	total := map[string]float64{
		currency: data.TotalCash,
	}

	assets := map[string]*banexg.Asset{
		currency: {
			Code:  currency,
			Free:  data.AvailableCash,
			Used:  data.TotalCash - data.AvailableCash,
			Total: data.TotalCash,
			Debt:  data.Debt,
		},
	}

	return &banexg.Balances{
		TimeStamp: time.Now().UnixMilli(),
		Free:      free,
		Used:      used,
		Total:     total,
		Assets:    assets,
		Info: map[string]interface{}{
			"raw":         data,
			"totalAsset":  data.TotalAsset,
			"stockValue":  data.TotalStockValue,
			"buyingPower": data.BuyingPower,
			"settledCash": data.SettledCash,
			"pendingCash": data.PendingCash,
			"marginRatio": data.MarginRatio,
		},
	}
}

func (e *Vietnam) parsePosition(data *SSIPositionInfo) *banexg.Position {
	if data == nil {
		return nil
	}

	market, _ := e.GetMarket(data.Symbol)
	symbol := data.Symbol
	if market != nil {
		symbol = market.Symbol
	}

	var side string
	if data.Quantity > 0 {
		side = banexg.PosSideLong
	} else if data.Quantity < 0 {
		side = banexg.PosSideShort
	} else {
		return nil
	}

	contracts := float64(data.Quantity)
	if contracts < 0 {
		contracts = -contracts
	}

	percentage := 0.0
	if data.CostValue > 0 {
		percentage = data.UnrealizedPLPct
	}

	return &banexg.Position{
		Symbol:        symbol,
		TimeStamp:     time.Now().UnixMilli(),
		Side:          side,
		Contracts:     contracts,
		ContractSize:  1.0,
		EntryPrice:    data.AvgPrice,
		MarkPrice:     data.MarketPrice,
		Notional:      data.MarketValue,
		UnrealizedPnl: data.UnrealizedPL,
		Percentage:    percentage,
		Info: map[string]interface{}{
			"raw":          data,
			"availableQty": data.AvailableQty,
			"costValue":    data.CostValue,
			"accountNo":    data.AccountNo,
		},
	}
}
