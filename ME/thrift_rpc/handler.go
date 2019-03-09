package thrift_rpc

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"context"
	"fmt"
	"time"

	"strconv"

	"../config"
	"../db/use_mysql"
	. "../itf"
	"../markets"
	"./gen-go/rpc_order"
)

type IOrderHandler struct {
	marks *markets.Markets
}

func NewIOrderHandler(marks *markets.Markets) *IOrderHandler {
	return &IOrderHandler{marks: marks}
}

func (p *IOrderHandler) MarketSelect(symbol string, userID string) (config.MarketType, error) {
	/// Market Type Dispatch
	marketType := config.MarketType_Num
	matchEng, _ := p.marks.GetMatchEngine(symbol)
	if matchEng != nil {
		conf := matchEng.GetConfig()
		iValue, _ := strconv.ParseInt(userID, 10, 64)

		if conf.Market_MixHR {
			marketType = config.MarketType_MixHR
		} else {
			if conf.RobotSet.Contains(iValue) {
				marketType = config.MarketType_Robot
			} else {
				marketType = config.MarketType_Human
			}
		}
		return marketType, nil
	} else {
		return marketType, fmt.Errorf("MarketSelect MatchEngine not exist.")
	}
}

func (p *IOrderHandler) SelectRobotMarket(symbol string) (config.MarketType, bool) {
	/// Market Type Dispatch
	matchEng, _ := p.marks.GetMatchEngine(symbol)
	if matchEng != nil {
		conf := matchEng.GetConfig()

		if conf.Market_MixHR {
			return config.MarketType_MixHR, true
		} else {
			if conf.Market_Robot {
				return config.MarketType_Robot, true
			} else {
				return config.MarketType_Num, false
			}
		}
	} else {
		return config.MarketType_Num, false
	}

}

// Parameters:
//  - Order
func (p *IOrderHandler) EnOrder(ctx context.Context, order *rpc_order.Order) (r *rpc_order.ReturnInfo, err error) {
	DebugPrintln(rpc_order.MODULE_NAME, LOG_LEVEL_TRACK, "Receive RPC EnOrder request, will to process")
	var (
		aorb    TradeType           = TradeType_UNSET
		errCode use_mysql.ErrorCode = use_mysql.ErrorCode_Fail
	)

	/// Order input parameter protect
	if order.Who == "" ||
		(order.Aorb != rpc_order.TradeType_BID && order.Aorb != rpc_order.TradeType_ASK) ||
		order.Price <= 0 ||
		order.Volume <= 0 ||
		order.Fee < 0 {
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "EnOrder Invalid Order met", Order: nil}, nil
	}

	/// Construct input order from user
	if order.Aorb == rpc_order.TradeType_BID {
		aorb = TradeType_BID
	} else if order.Aorb == rpc_order.TradeType_ASK {
		aorb = TradeType_ASK
	} else {
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "EnOrder Invalid Order met", Order: nil}, nil
	}

	meOrder := Order{
		ID:           time.Now().UnixNano(),
		Who:          order.Who,
		AorB:         aorb,
		Symbol:       order.Symbol,
		Timestamp:    time.Now().UnixNano(),
		EnOrderPrice: order.Price,
		Price:        order.Price,
		Volume:       order.Volume,
		TotalVolume:  order.Volume,
		Fee:          order.Fee,
		Status:       ORDER_SUBMIT,
		IPAddr:       order.IpAddr,
	}

	/// Market Type Dispatch
	marketType, errMarketType := p.MarketSelect(meOrder.Symbol, meOrder.Who)
	/// Symbol validation check
	if errMarketType != nil || !p.marks.MarketTypeCheck(meOrder.Symbol, marketType) {
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "EnOrder symbol not support.", Order: nil}, fmt.Errorf("EnOrder symbol(%s) not support.", meOrder.Symbol)
	}

	if config.GetMEConfig().InPoolMode == "block" {
		//// Enorder to Match Engine Duration Storage
	reEnorder_:
		err, errCode = use_mysql.MEMySQLInstance().EnOrder(&meOrder)
		if err != nil {
			DebugPrintln(rpc_order.MODULE_NAME, LOG_LEVEL_ALWAYS, err)
			if errCode == use_mysql.ErrorCode_DupPrimateKey {
				DebugPrintf(rpc_order.MODULE_NAME, LOG_LEVEL_FATAL, "EnOrder fail, Retry to do it once more.\n")
				meOrder.ID = time.Now().UnixNano()
				///err, errCode = use_mysql.MEMySQLInstance().EnOrder(&meOrder)
				goto reEnorder_
				//				if err != nil {
				//					if errCode == use_mysql.ErrorCode_DupPrimateKey {
				//						DebugPrintf(rpc_order.MODULE_NAME, LOG_LEVEL_FATAL, "EnOrder fail, Retry to do it twice more.\n")
				//						meOrder.ID = time.Now().UnixNano()
				//						err, _ = use_mysql.MEMySQLInstance().EnOrder(&meOrder)
				//						if err != nil {
				//							return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "EnOrder fail. Please retry.", Order: nil}, err
				//						}
				//					}
				//				}
			} else if errCode == use_mysql.ErrorCode_FundNoEnough {
				return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "EnOrder fail. Fund not enough.", Order: nil}, err
			} else {
				return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "EnOrder fail. Please retry.", Order: nil}, err
			}
		}
		/// Put the order in the match engine to process
		meOrder.Volume = meOrder.TotalVolume
		matchEng, _ := p.marks.GetMatchEngine(meOrder.Symbol)
		if matchEng != nil {
			matchEng.GetTradePool(marketType).InChannelBlock <- &meOrder
		} else {
			panic(fmt.Errorf("EnOrder unbelievable error occur."))
		}

	} else {
		matchEng, _ := p.marks.GetMatchEngine(meOrder.Symbol)
		if matchEng != nil {
			matchEng.GetTradePool(marketType).InChannel.In(&meOrder)
		} else {
			panic(fmt.Errorf("EnOrder unbelievable error occur."))
		}
	}

	/// Construct Return Info:
	id := meOrder.ID
	timestamp := meOrder.Timestamp
	status := rpc_order.OrderStatus_ORDER_SUBMIT
	tVolume := float64(0)
	order.ID = &id
	order.Timestamp = &timestamp
	order.Status = &status
	order.TradedVolume = &tVolume

	return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_SUCC, Info: "EnOrder Operate Complete Success.", Order: order}, nil
}

/// toCancelOrder
func (p *IOrderHandler) toCancelOrder(symbol string, user string, id int64) (order *Order, err error) {
	/// meOrder input parameter protect
	if user == "" || symbol == "" || id < 0 {
		return nil, fmt.Errorf("toCancelOrder invalid params(symbol:%s, user:%s, id:%d) input", symbol, user, id)
	}

	/// Market Type Dispatch
	marketType, errMarketType := p.MarketSelect(symbol, user)
	/// Symbol validation check
	if errMarketType != nil || !p.marks.MarketTypeCheck(symbol, marketType) {
		return nil, fmt.Errorf("toCancelOrder symbol(%s) not support.", symbol)
	}

	meOrder, err := use_mysql.MEMySQLInstance().GetOrder(user, id, symbol, nil)
	if err != nil {
		return nil, fmt.Errorf("toCancelOrder symbol(%s) GetOrder(user:%s, id:%d) fail. Error Info:%s", symbol, user, id, err.Error())
	}
	if meOrder.Status != ORDER_SUBMIT && meOrder.Status != ORDER_PARTIAL_FILLED {
		return nil, fmt.Errorf("toCancelOrder symbol(%s) GetOrder(user:%s, id:%d) status(%s) illegal.", symbol, user, id, meOrder.Status.String())
	}

	/// use cancel chan to process cancel meOrder
	meOrder.Status = ORDER_CANCELING
	matchEng, _ := p.marks.GetMatchEngine(symbol)
	if matchEng != nil {
		matchEng.GetTradePool(marketType).CancelChannel <- meOrder
	} else {
		panic(fmt.Errorf("toCancelOrder unbelievable error occur."))
	}

	return meOrder, nil
}

/// CancelOrder
func (p *IOrderHandler) CancelOrder(ctx context.Context, user string, id int64, symbol string) (r *rpc_order.ReturnInfo, err error) {
	DebugPrintln(rpc_order.MODULE_NAME, LOG_LEVEL_TRACK, "Receive RPC CancelOrder request, Order.id=", id)

	meOrder, err := p.toCancelOrder(symbol, user, id)
	if err != nil {
		DebugPrintln(rpc_order.MODULE_NAME, LOG_LEVEL_ALWAYS, "CancelOrder toCancelOrder with error:", err)
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: err.Error(), Order: nil}, nil
	}

	/// construct return msg for caller
	order := rpc_order.Order{
		Aorb:   tradeTypeConvertTo(meOrder.AorB),
		Who:    user,
		Symbol: symbol,
		Price:  meOrder.Price,
		Volume: meOrder.Volume,
		Fee:    meOrder.Fee,
	}
	timestamp := meOrder.Timestamp
	status := rpc_order.OrderStatus_ORDER_CANCELING
	tradeVolume := meOrder.TotalVolume - meOrder.Volume

	order.ID = &id
	order.Timestamp = &timestamp
	order.Status = &status
	order.TradedVolume = &tradeVolume

	return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_SUCC, Info: "CancelOrder Operate Complete Success.", Order: &order}, nil
}

// Parameters:
//  - Symbol
//  - Durations
func (p *IOrderHandler) CancelRobotOverTimeOrder(ctx context.Context, symbol string, durations int64) (r *rpc_order.ReturnInfo, err error) {
	DebugPrintf(rpc_order.MODULE_NAME, LOG_LEVEL_TRACK, "Receive RPC CancelRobotOverTimeOrder request, symbol=%s, over duration=%ds", symbol, durations)

	/// meOrder input parameter protect
	if symbol == "" || durations < 0 {
		fmt.Printf("CancelRobotOverTimeOrder Invalid params(symbol:%s, durations:%d) input", symbol, durations)
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "CancelRobotOverTimeOrder Invalid params input.", Order: nil}, nil
	}

	/// Market Type Dispatch
	marketType, ok := p.SelectRobotMarket(symbol)
	if !ok {
		fmt.Printf("CancelRobotOverTimeOrder SelectRobotMarket(symbol:%s) fail. Please check MEconfig.json.\n", symbol)
		///config.GetMarket(config.Struct(symbol)).Dump()
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "CancelRobotOverTimeOrder SelectRobotMarket fail.", Order: nil}, nil
	}

	/// Symbol validation check
	if !p.marks.MarketTypeCheck(symbol, marketType) {
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "CancelRobotOverTimeOrder symbol not support.", Order: nil}, fmt.Errorf("CancelRobotOverTimeOrder symbol(%s) not support.", symbol)
	}

	matchEng, _ := p.marks.GetMatchEngine(symbol)
	conf := matchEng.GetConfig()
	orderList, err := use_mysql.MEMySQLInstance().GetOnesOverTimeOrder(symbol, conf.RobotSet.Elements(), durations)
	if err != nil {
		panic(err)
	}

	go func() {
		for _, order := range orderList {
			_, err := p.toCancelOrder(order.Symbol, order.Who, order.ID)
			if err != nil {
				DebugPrintf(rpc_order.MODULE_NAME, LOG_LEVEL_ALWAYS, "CancelRobotOverTimeOrder toCancelOrder(id:%d), with error:%s\n", order.ID, err.Error())
			}
		}
	}()

	toCancelOrderSize := len(orderList)
	info := fmt.Sprintf("CancelRobotOverTimeOrder Operate Complete Success, %d orders will be canceled.", toCancelOrderSize)
	return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_SUCC, Info: info, Order: nil}, nil
}

// Parameters:
//  - ID
func (p *IOrderHandler) GetOrder(ctx context.Context, user string, id int64, symbol string) (r *rpc_order.ReturnInfo, err error) {
	DebugPrintln(rpc_order.MODULE_NAME, LOG_LEVEL_TRACK, "Receive RPC GetOrder request, Order.id=", id)
	/// meOrder input parameter protect
	if user == "" || symbol == "" || id < 0 {
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "GetOrder Invalid Order met!", Order: nil}, nil
	}

	/// Market Type Dispatch
	marketType, errMarketType := p.MarketSelect(symbol, user)
	/// Symbol validation check
	if errMarketType != nil || !p.marks.MarketTypeCheck(symbol, marketType) {
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "GetOrder symbol not support.", Order: nil}, fmt.Errorf("GetOrder symbol(%s) not support.", symbol)
	}

	meOrder, err := use_mysql.MEMySQLInstance().GetOrder(user, id, symbol, nil)
	if err != nil {
		DebugPrintln(rpc_order.MODULE_NAME, LOG_LEVEL_TRACK, "IOrderHandler GetOrder fail! It may not a valid order.\n")
		return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_FAIL, Info: "IOrderHandler GetOrder fail", Order: nil}, nil
	}

	/// construct return msg for caller
	order := rpc_order.Order{
		Aorb:   tradeTypeConvertTo(meOrder.AorB),
		Who:    user,
		Symbol: symbol,
		Price:  meOrder.Price,
		Volume: meOrder.Volume,
		Fee:    meOrder.Fee,
	}
	timestamp := meOrder.Timestamp
	status := tradeStatusConvertTo(meOrder.Status)
	tradeVolume := meOrder.TotalVolume - meOrder.Volume

	order.ID = &id
	order.Timestamp = &timestamp
	order.Status = &status
	order.TradedVolume = &tradeVolume

	return &rpc_order.ReturnInfo{Status: rpc_order.RetunStatus_SUCC, Info: "GetOrder Operate Complete Success.", Order: &order}, nil
}

/// Data Adapter
func tradeTypeConvertTo(t TradeType) rpc_order.TradeType {
	switch t {
	case TradeType_BID:
		return rpc_order.TradeType_BID
	case TradeType_ASK:
		return rpc_order.TradeType_ASK
	}
	return rpc_order.TradeType(-1)
}

func tradeStatusConvertTo(s TradeStatus) rpc_order.OrderStatus {
	switch s {
	case ORDER_SUBMIT:
		return rpc_order.OrderStatus_ORDER_SUBMIT
	case ORDER_FILLED:
		return rpc_order.OrderStatus_ORDER_FILLED
	case ORDER_PARTIAL_FILLED:
		return rpc_order.OrderStatus_ORDER_PARTIAL_FILLED
	case ORDER_PARTIAL_CANCEL:
		return rpc_order.OrderStatus_ORDER_PARTIAL_CANCEL
	case ORDER_CANCELED:
		return rpc_order.OrderStatus_ORDER_CANCELED
	case ORDER_CANCELING:
		return rpc_order.OrderStatus_ORDER_CANCELING
	}
	return rpc_order.OrderStatus_ORDER_STATUSUNKNOW
}
