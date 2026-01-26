package vietnam

import (
	"context"
	"fmt"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/banbox/banexg/utils"
)

func (e *Vietnam) CreateOrder(symbol, odType, side string, amount, price float64, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	if !IsMarketOpen() {
		return nil, errs.NewMsg(errs.CodeRunTime, "market is closed")
	}

	session := GetCurrentSessionType()
	ssiOrderType := MapOrderTypeToSSI(odType, session)
	if !CanPlaceOrderType(ssiOrderType) {
		return nil, errs.NewMsg(errs.CodeParamInvalid, "order type %s not allowed in %s session", ssiOrderType, session)
	}

	quantity := int(amount)
	lotSize := DefaultLotSize
	if ls, ok := market.Info["lotSize"].(int); ok && ls > 0 {
		lotSize = ls
	}
	if err := ValidateQuantity(quantity, lotSize); err != nil {
		return nil, err
	}

	if ssiOrderType == OdTypeLO && price <= 0 {
		return nil, errs.NewMsg(errs.CodeParamRequired, "price required for limit orders")
	}

	if ssiOrderType == OdTypeLO {
		refPrice := market.Info["refPrice"].(float64)
		if err := ValidatePrice(price, refPrice, market.Info["exchange"].(string)); err != nil {
			return nil, err
		}
	}

	accountNo := utils.GetMapVal(args, "accountNo", "")
	if accountNo == "" && e.DefAccName != "" {
		if acc, ok := e.Accounts[e.DefAccName]; ok {
			accountNo = acc.Name
		}
	}
	if accountNo == "" {
		return nil, errs.NewMsg(errs.CodeParamRequired, "accountNo required")
	}

	clientOrderID := utils.GetMapVal(args, "clientOrderId", fmt.Sprintf("OMO_%d", time.Now().UnixMilli()))

	orderReq := SSIOrderReq{
		Symbol:    market.ID,
		OrderType: ssiOrderType,
		Side:      MapOrderSideToSSI(side),
		Quantity:  quantity,
		Price:     price,
		AccountNo: accountNo,
		RequestID: clientOrderID,
	}

	if validDate, ok := args["validDate"].(string); ok {
		orderReq.ValidDate = validDate
	}

	reqBody := utils.SafeParams(nil)
	reqBody["symbol"] = orderReq.Symbol
	reqBody["orderType"] = orderReq.OrderType
	reqBody["side"] = orderReq.Side
	reqBody["quantity"] = orderReq.Quantity
	reqBody["accountNo"] = orderReq.AccountNo
	reqBody["requestId"] = orderReq.RequestID
	if orderReq.Price > 0 {
		reqBody["price"] = orderReq.Price
	}
	if orderReq.ValidDate != "" {
		reqBody["validDate"] = orderReq.ValidDate
	}

	rsp := e.RequestApiRetry(context.Background(), MethodPlaceOrder, reqBody, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var orderRsp SSIOrderRsp
	if err := utils.UnmarshalString(rsp.Content, &orderRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse order response: %v", err)
	}

	if orderRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI order error: %s", orderRsp.Message)
	}

	return e.parseOrder(&orderRsp.Data, market), nil
}

func (e *Vietnam) CancelOrder(orderId, symbol string, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	accountNo := utils.GetMapVal(args, "accountNo", "")
	if accountNo == "" && e.DefAccName != "" {
		if acc, ok := e.Accounts[e.DefAccName]; ok {
			accountNo = acc.Name
		}
	}
	if accountNo == "" {
		return nil, errs.NewMsg(errs.CodeParamRequired, "accountNo required")
	}

	reqBody := utils.SafeParams(nil)
	reqBody["orderId"] = orderId
	reqBody["accountNo"] = accountNo

	rsp := e.RequestApiRetry(context.Background(), MethodCancelOrder, reqBody, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var orderRsp SSIOrderRsp
	if err := utils.UnmarshalString(rsp.Content, &orderRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse cancel response: %v", err)
	}

	if orderRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI cancel error: %s", orderRsp.Message)
	}

	return e.parseOrder(&orderRsp.Data, market), nil
}

func (e *Vietnam) EditOrder(symbol, orderId, side string, amount, price float64, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	accountNo := utils.GetMapVal(args, "accountNo", "")
	if accountNo == "" && e.DefAccName != "" {
		if acc, ok := e.Accounts[e.DefAccName]; ok {
			accountNo = acc.Name
		}
	}
	if accountNo == "" {
		return nil, errs.NewMsg(errs.CodeParamRequired, "accountNo required")
	}

	reqBody := utils.SafeParams(nil)
	reqBody["orderId"] = orderId
	reqBody["accountNo"] = accountNo

	if price > 0 {
		refPrice := market.Info["refPrice"].(float64)
		if err := ValidatePrice(price, refPrice, market.Info["exchange"].(string)); err != nil {
			return nil, err
		}
		reqBody["price"] = price
	}

	if amount > 0 {
		quantity := int(amount)
		lotSize := DefaultLotSize
		if ls, ok := market.Info["lotSize"].(int); ok && ls > 0 {
			lotSize = ls
		}
		if err := ValidateQuantity(quantity, lotSize); err != nil {
			return nil, err
		}
		reqBody["quantity"] = quantity
	}

	rsp := e.RequestApiRetry(context.Background(), MethodModifyOrder, reqBody, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var orderRsp SSIOrderRsp
	if err := utils.UnmarshalString(rsp.Content, &orderRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse modify response: %v", err)
	}

	if orderRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI modify error: %s", orderRsp.Message)
	}

	return e.parseOrder(&orderRsp.Data, market), nil
}

func (e *Vietnam) FetchOrder(orderId, symbol string, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	accountNo := utils.GetMapVal(args, "accountNo", "")
	if accountNo == "" && e.DefAccName != "" {
		if acc, ok := e.Accounts[e.DefAccName]; ok {
			accountNo = acc.Name
		}
	}

	args["orderId"] = orderId
	if accountNo != "" {
		args["accountNo"] = accountNo
	}

	rsp := e.RequestApiRetry(context.Background(), MethodGetOrderDetail, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var orderRsp SSIOrderRsp
	if err := utils.UnmarshalString(rsp.Content, &orderRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse order detail: %v", err)
	}

	if orderRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", orderRsp.Message)
	}

	return e.parseOrder(&orderRsp.Data, market), nil
}

func (e *Vietnam) FetchOrders(symbol string, since int64, limit int, params map[string]interface{}) ([]*banexg.Order, *errs.Error) {
	args, market, err := e.LoadArgsMarket(symbol, params)
	if err != nil {
		return nil, err
	}

	accountNo := utils.GetMapVal(args, "accountNo", "")
	if accountNo == "" && e.DefAccName != "" {
		if acc, ok := e.Accounts[e.DefAccName]; ok {
			accountNo = acc.Name
		}
	}

	if accountNo != "" {
		args["accountNo"] = accountNo
	}

	if symbol != "" {
		args["symbol"] = market.ID
	}

	if since > 0 {
		fromDate := time.UnixMilli(since).In(VietnamLocation).Format("2006-01-02")
		args["fromDate"] = fromDate
	}

	if limit > 0 {
		args["pageSize"] = limit
	}

	rsp := e.RequestApiRetry(context.Background(), MethodGetOrderHistory, args, 1)
	if rsp.Error != nil {
		return nil, rsp.Error
	}

	var historyRsp SSIOrderHistoryRsp
	if err := utils.UnmarshalString(rsp.Content, &historyRsp, utils.JsonNumDefault); err != nil {
		return nil, errs.NewMsg(errs.CodeUnmarshalFail, "failed to parse order history: %v", err)
	}

	if historyRsp.Status != 200 {
		return nil, errs.NewMsg(errs.CodeRunTime, "SSI API error: %s", historyRsp.Message)
	}

	orders := make([]*banexg.Order, 0, len(historyRsp.Data))
	for _, orderInfo := range historyRsp.Data {
		orderMarket := market
		if market == nil {
			orderMarket, _ = e.GetMarket(orderInfo.Symbol)
		}
		order := e.parseOrder(&orderInfo, orderMarket)
		if order != nil {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (e *Vietnam) FetchOpenOrders(symbol string, since int64, limit int, params map[string]interface{}) ([]*banexg.Order, *errs.Error) {
	orders, err := e.FetchOrders(symbol, since, limit, params)
	if err != nil {
		return nil, err
	}

	openOrders := make([]*banexg.Order, 0)
	for _, order := range orders {
		if order.Status == banexg.OdStatusOpen || order.Status == banexg.OdStatusPartFilled {
			openOrders = append(openOrders, order)
		}
	}

	return openOrders, nil
}

func (e *Vietnam) parseOrder(data *SSIOrderInfo, market *banexg.Market) *banexg.Order {
	if data == nil {
		return nil
	}

	var symbol string
	if market != nil {
		symbol = market.Symbol
	} else {
		symbol = data.Symbol
	}

	createTime := time.Now().UnixMilli()
	if data.CreateTime != "" {
		if t, err := parseVietnamTime(data.CreateTime); err == nil {
			createTime = t.UnixMilli()
		}
	}

	lastModified := createTime
	if data.LastModified != "" {
		if t, err := parseVietnamTime(data.LastModified); err == nil {
			lastModified = t.UnixMilli()
		}
	}

	status := MapSSIOrderStatus(data.Status)
	odType := MapSSIOrderTypeToBanExg(data.OrderType)
	side := MapSSIOrderSideToBanExg(data.Side)

	amount := float64(data.Quantity)
	filled := float64(data.FilledQty)
	remaining := amount - filled

	var cost float64
	var average float64
	if data.FilledQty > 0 && data.AvgPrice > 0 {
		cost = float64(data.FilledQty) * data.AvgPrice
		average = data.AvgPrice
	}

	return &banexg.Order{
		ID:                 data.OrderID,
		ClientOrderID:      data.RequestID,
		Timestamp:          createTime,
		LastTradeTimestamp: lastModified,
		Symbol:             symbol,
		Type:               odType,
		Side:               side,
		Price:              data.Price,
		Amount:             amount,
		Filled:             filled,
		Remaining:          remaining,
		Cost:               cost,
		Average:            average,
		Status:             status,
		Info:               map[string]interface{}{"raw": data},
	}
}
