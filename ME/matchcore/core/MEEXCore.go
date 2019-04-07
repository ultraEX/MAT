package core

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"

	"../../comm"
	"../../config"
	"../../db/use_mysql"
	te "../../tickers"
	chs "../chansIO"
	rt "../runtime"
	sm "./structmap"
	zs "./zset"
	zc "./zsetcluster"
)

const (
	MODULE_NAME_MEEXCORE string = "[MEEXCore]: "
)

type MEECoreInfo struct {
	Symbol       string
	MarketType   config.MarketType
	MarketConfig *config.MarketConfig

	*te.TickerPool
}

func NewMEECoreInfo(symbol string, mrketType config.MarketType, conf *config.MarketConfig, te *te.TickerPool) MEECoreInfo {
	return MEECoreInfo{Symbol: symbol, MarketType: mrketType, MarketConfig: conf, TickerPool: te}
}

type MEEXCore struct {
	comm.OrderContainerItf

	chs.MultiChans_In
	MultiChans_Out chs.MultiChans_OutGo

	MEECoreInfo
	rt.DebugInfo

	latestPrice float64
	rt.OrderID
}

func NewMEEXCore(sym string, mt config.MarketType, conf *config.MarketConfig, te *te.TickerPool) *MEEXCore {
	o := new(MEEXCore)

	if config.GetMEConfig().Algorithm == "heapmap" {
		o.OrderContainerItf = sm.NewOrderContainerBaseHp()
	} else if config.GetMEConfig().Algorithm == "skipmap1" {
		o.OrderContainerItf = sm.NewOrderCoitainerBase1Sk()
	} else if config.GetMEConfig().Algorithm == "skipmap2" {
		o.OrderContainerItf = sm.NewOrderCoitainerBase2Sk()
	} else if config.GetMEConfig().Algorithm == "zset" {
		o.OrderContainerItf = zs.NewOrderContainer()
	} else if config.GetMEConfig().Algorithm == "zsetcluster" {
		o.OrderContainerItf = zc.NewOrderContainer()
	} else {
		panic(fmt.Errorf("MEEXCore.NewMEEXCore fail, as Algorithm(%s) not support.", config.GetMEConfig().Algorithm))
	}

	o.MultiChans_In = *chs.NewMultiChans_In(o.multiChanInProc)
	o.MultiChans_Out = *chs.NewMultiChans_OutGo(o.multiChanOutProc)

	o.MEECoreInfo = NewMEECoreInfo(sym, mt, conf, te)
	o.DebugInfo = *rt.NewDebugInfo()

	o.latestPrice = te.GetNewestPrice()

	s := strings.Split(sym, "/")
	if len(s) != 2 {
		panic(fmt.Errorf("NewMEXCore.NewOrderID fail, as sym(%s) input illegal.", sym))
	}
	vB, okB := config.GetCoinMapInt()[s[0]]
	vQ, okQ := config.GetCoinMapInt()[s[1]]
	if !okB || !okQ {
		panic(fmt.Errorf("NewMEXCore.NewOrderID to GetCoinMapInt(%s) fail.", sym))
	}
	o.OrderID = *rt.NewOrderID((int(vB) & 0x2f) | (int(vQ) & 0x2f))

	go o.initHistoryOrder()

	return o
}

func (t *MEEXCore) EnOrder(order *comm.Order) error {
	t.MultiChans_In.InChannel(&chs.InElem{Type_: chs.InElemType_EnOrder, Order: order, Count: 0})
	return nil
}

func (t *MEEXCore) CancelOrder(id int64) error {
	t.MultiChans_In.InChannel(&chs.InElem{Type_: chs.InElemType_CancelID, CancelId: id, Count: 0})
	return nil
}

func (t *MEEXCore) CancelTheOrder(order *comm.Order) error {
	return nil
}

func (t *MEEXCore) init(order *comm.Order) error {
	// to do

	comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK, "init: .\n")
	return nil
}

func (t *MEEXCore) getTradePrice(bidPrice, askPrice float64) float64 {
	tradePrice := float64(0)
	if t.latestPrice <= askPrice { ///如果前一笔成交价低于或等于卖出价，则最新成交价就是卖出价
		tradePrice = askPrice
	} else if t.latestPrice >= bidPrice { ///如果前一笔成交价高于或等于买入价，则最新成交价就是买入价
		tradePrice = bidPrice
	} else { ///如果前一笔成交价在卖出价与买入价之间，则最新成交价就是前一笔的成交价
		tradePrice = t.latestPrice
	}
	priceValidatable(tradePrice)
	t.latestPrice = tradePrice
	return tradePrice
}

func (t *MEEXCore) bidOrderRecurseMatch(bid *comm.Order) {
	TimeDot1 := time.Now().UnixNano()
	defer func() {
		TimeDot2 := time.Now().UnixNano()
		t.DebugInfo_RecordCorePerform(TimeDot2 - TimeDot1)
	}()

	var bidStatus, askStatus comm.TradeStatus
	isBreak := false
	orderBid := bid
	orderAsk := t.OrderContainerItf.GetTop(comm.TradeType_ASK)
	if orderAsk != nil && orderBid.Price >= orderAsk.Price {
		for {
			orderAsk = t.OrderContainerItf.Pop(comm.TradeType_ASK)

			if orderValidatable(orderAsk) && orderValidatable(orderBid) {
				if orderBid.Price >= orderAsk.Price {

					comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK,
						`=======>>>%s-%s Bid Order Matching<<<======
		BID order == symbol:%s, id:%d, user:%s, time:%d, price:%f, volume:%f
		ASK order == symbol:%s, id:%d, user:%s, time:%d, price:%f, volume:%f
		`,
						t.Symbol, t.MarketType.String(),
						orderBid.Symbol, orderBid.ID, orderBid.Who, orderBid.Timestamp, orderBid.Price, orderBid.Volume,
						orderAsk.Symbol, orderAsk.ID, orderAsk.Who, orderAsk.Timestamp, orderAsk.Price, orderAsk.Volume,
					)
					///trade price
					tradePrice := t.getTradePrice(orderBid.Price, orderAsk.Price)

					///trade volume
					tradeVolume := math.Min(orderAsk.Volume, orderBid.Volume)

					///trade amount
					tradeAmount := tradePrice * tradeVolume
					tradeBidAmount := tradeVolume * (1 - orderBid.Fee)
					tradeAskAmount := tradeAmount * (1 - orderAsk.Fee)

					///update the order
					if orderBid.Volume == orderAsk.Volume {
						/// updeate order status
						bidStatus = comm.ORDER_FILLED
						askStatus = comm.ORDER_FILLED

						///debug info
						t.DebugInfo_AskTradeOutputAdd()
						t.DebugInfo_BidTradeOutputAdd()
						t.DebugInfo_AskTradeCompleteAdd()
						t.DebugInfo_BidTradeCompleteAdd()

						isBreak = true
					} else {
						if tradeVolume < orderBid.Volume {
							orderBid.Volume -= tradeVolume
							orderBid.Status = comm.ORDER_PARTIAL_FILLED
							orderAsk.Status = comm.ORDER_FILLED

							/// updeate order status
							bidStatus = comm.ORDER_PARTIAL_FILLED
							askStatus = comm.ORDER_FILLED
							///debug info
							t.DebugInfo_BidTradeOutputAdd()
							t.DebugInfo_AskTradeCompleteAdd()
						} else {
							orderAsk.Volume -= tradeVolume
							orderBid.Status = comm.ORDER_FILLED
							orderAsk.Status = comm.ORDER_PARTIAL_FILLED
							t.OrderContainerItf.Push(orderAsk)

							/// updeate order status
							bidStatus = comm.ORDER_FILLED
							askStatus = comm.ORDER_PARTIAL_FILLED
							///debug info
							t.DebugInfo_BidTradeCompleteAdd()
							t.DebugInfo_AskTradeOutputAdd()

							isBreak = true
						}
					}

					///trade output
					tradeBid := comm.Trade{
						Order: comm.Order{
							ID:           orderBid.ID,
							Who:          orderBid.Who,
							AorB:         comm.TradeType_BID,
							Symbol:       orderBid.Symbol,
							Timestamp:    orderBid.Timestamp,
							EnOrderPrice: orderBid.EnOrderPrice,
							Price:        tradePrice,
							Volume:       tradeVolume,
							TotalVolume:  orderBid.TotalVolume,
							Fee:          orderBid.Fee,
							Status:       bidStatus,
							IPAddr:       orderBid.IPAddr,
						},
						Amount:    tradeBidAmount,
						TradeTime: time.Now().UnixNano(),
						FeeCost:   tradeVolume * orderBid.Fee,
					}
					tradeAsk := comm.Trade{
						Order: comm.Order{
							ID:           orderAsk.ID,
							Who:          orderAsk.Who,
							AorB:         comm.TradeType_ASK,
							Symbol:       orderAsk.Symbol,
							Timestamp:    orderAsk.Timestamp,
							EnOrderPrice: orderAsk.EnOrderPrice,
							Price:        tradePrice,
							Volume:       tradeVolume,
							TotalVolume:  orderAsk.TotalVolume,
							Fee:          orderAsk.Fee,
							Status:       askStatus,
							IPAddr:       orderAsk.IPAddr,
						},
						Amount:    tradeAskAmount,
						TradeTime: time.Now().UnixNano(),
						FeeCost:   tradeAmount * orderAsk.Fee,
					}

					///To do: put to channel to send to database
					t.MultiChans_Out.InChannel(&chs.OutElem{Trade: &chs.MatchTrade{BidTrade: &tradeBid, AskTrade: &tradeAsk}, CancelOrder: nil, Type_: chs.OUTPOOL_MATCHTRADE, Count: 0})

				} else {
					t.OrderContainerItf.Push(orderBid)
					t.OrderContainerItf.Push(orderAsk)
					isBreak = true
				}
			} else {
				if orderAsk == nil {
					t.OrderContainerItf.Push(orderBid)
					return
				}
				if orderAsk.Status == comm.ORDER_CANCELED {
					continue
				}
				if orderBid.Status == comm.ORDER_CANCELED {
					isBreak = true
				}
				panic(fmt.Errorf("bidOrderRecurseMatch Met invalid order, check it !!!"))
			}

			if isBreak {
				break
			}
		}
	} else {
		t.OrderContainerItf.Push(orderBid)
	}
}

func (t *MEEXCore) askOrderRecurseMatch(ask *comm.Order) {
	TimeDot1 := time.Now().UnixNano()
	defer func() {
		TimeDot2 := time.Now().UnixNano()
		t.DebugInfo_RecordCorePerform(TimeDot2 - TimeDot1)
	}()

	var bidStatus, askStatus comm.TradeStatus
	isBreak := false
	orderAsk := ask
	orderBid := t.OrderContainerItf.GetTop(comm.TradeType_BID)
	if orderBid != nil && orderBid.Price >= orderAsk.Price {

		for {
			orderBid = t.OrderContainerItf.Pop(comm.TradeType_BID)

			if orderValidatable(orderAsk) && orderValidatable(orderBid) {
				if orderBid.Price >= orderAsk.Price {

					comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK,
						`=======>>>%s-%s Ask Order Matching<<<======
		BID order == symbol:%s, id:%d, user:%s, time:%d, price:%f, volume:%f
		ASK order == symbol:%s, id:%d, user:%s, time:%d, price:%f, volume:%f
		`,
						t.Symbol, t.MarketType.String(),
						orderBid.Symbol, orderBid.ID, orderBid.Who, orderBid.Timestamp, orderBid.Price, orderBid.Volume,
						orderAsk.Symbol, orderAsk.ID, orderAsk.Who, orderAsk.Timestamp, orderAsk.Price, orderAsk.Volume,
					)
					///trade price
					tradePrice := t.getTradePrice(orderBid.Price, orderAsk.Price)

					///trade volume
					tradeVolume := math.Min(orderAsk.Volume, orderBid.Volume)

					///trade amount
					tradeAmount := tradePrice * tradeVolume
					tradeBidAmount := tradeVolume * (1 - orderBid.Fee)
					tradeAskAmount := tradeAmount * (1 - orderAsk.Fee)

					///update the order
					if orderBid.Volume == orderAsk.Volume {
						/// updeate order status
						bidStatus = comm.ORDER_FILLED
						askStatus = comm.ORDER_FILLED

						///debug info
						t.DebugInfo_AskTradeOutputAdd()
						t.DebugInfo_BidTradeOutputAdd()
						t.DebugInfo_AskTradeCompleteAdd()
						t.DebugInfo_BidTradeCompleteAdd()

						isBreak = true
					} else {
						if tradeVolume < orderBid.Volume {
							orderBid.Volume -= tradeVolume
							orderBid.Status = comm.ORDER_PARTIAL_FILLED
							orderAsk.Status = comm.ORDER_FILLED
							t.OrderContainerItf.Push(orderBid)

							/// updeate order status
							bidStatus = comm.ORDER_PARTIAL_FILLED
							askStatus = comm.ORDER_FILLED
							///debug info
							t.DebugInfo_BidTradeOutputAdd()
							t.DebugInfo_AskTradeCompleteAdd()

							isBreak = true
						} else {
							orderAsk.Volume -= tradeVolume
							orderBid.Status = comm.ORDER_FILLED
							orderAsk.Status = comm.ORDER_PARTIAL_FILLED

							/// updeate order status
							bidStatus = comm.ORDER_FILLED
							askStatus = comm.ORDER_PARTIAL_FILLED
							///debug info
							t.DebugInfo_BidTradeCompleteAdd()
							t.DebugInfo_AskTradeOutputAdd()

						}
					}

					///trade output
					tradeBid := comm.Trade{
						Order: comm.Order{
							ID:           orderBid.ID,
							Who:          orderBid.Who,
							AorB:         comm.TradeType_BID,
							Symbol:       orderBid.Symbol,
							Timestamp:    orderBid.Timestamp,
							EnOrderPrice: orderBid.EnOrderPrice,
							Price:        tradePrice,
							Volume:       tradeVolume,
							TotalVolume:  orderBid.TotalVolume,
							Fee:          orderBid.Fee,
							Status:       bidStatus,
							IPAddr:       orderBid.IPAddr,
						},
						Amount:    tradeBidAmount,
						TradeTime: time.Now().UnixNano(),
						FeeCost:   tradeVolume * orderBid.Fee,
					}
					tradeAsk := comm.Trade{
						Order: comm.Order{
							ID:           orderAsk.ID,
							Who:          orderAsk.Who,
							AorB:         comm.TradeType_ASK,
							Symbol:       orderAsk.Symbol,
							Timestamp:    orderAsk.Timestamp,
							EnOrderPrice: orderAsk.EnOrderPrice,
							Price:        tradePrice,
							Volume:       tradeVolume,
							TotalVolume:  orderAsk.TotalVolume,
							Fee:          orderAsk.Fee,
							Status:       askStatus,
							IPAddr:       orderAsk.IPAddr,
						},
						Amount:    tradeAskAmount,
						TradeTime: time.Now().UnixNano(),
						FeeCost:   tradeAmount * orderAsk.Fee,
					}

					///To do: put to channel to send to database
					t.MultiChans_Out.InChannel(&chs.OutElem{Trade: &chs.MatchTrade{BidTrade: &tradeBid, AskTrade: &tradeAsk}, CancelOrder: nil, Type_: chs.OUTPOOL_MATCHTRADE, Count: 0})

				} else {
					t.OrderContainerItf.Push(orderBid)
					t.OrderContainerItf.Push(orderAsk)
					isBreak = true
				}
			} else {
				if orderBid == nil {
					t.OrderContainerItf.Push(orderAsk)
					return
				}
				if orderBid.Status == comm.ORDER_CANCELED {
					continue
				}
				if orderAsk.Status == comm.ORDER_CANCELED {
					isBreak = true
				}
				panic(fmt.Errorf("askOrderRecurseMatch Met invalid order, check it !!!"))
			}

			if isBreak {
				break
			}
		}
	} else {
		t.OrderContainerItf.Push(orderAsk)
	}
}

/// Core process: ----------------------------------------------------------------------------------------
func (t *MEEXCore) enOrder(order *comm.Order) error {

	/// to check and match
	if order.AorB == comm.TradeType_BID {
		t.bidOrderRecurseMatch(order)
	}
	if order.AorB == comm.TradeType_ASK {
		t.askOrderRecurseMatch(order)
	}

	comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK, "enOrder: id = %d.\n", order.ID)
	return nil
}

func (t *MEEXCore) cancelOrder(id int64) error {

	order := t.OrderContainerItf.Get(id)
	if order != nil {
		order.Status = comm.ORDER_CANCELED
		t.MultiChans_Out.InChannel(&chs.OutElem{Trade: nil, CancelOrder: &chs.CanceledOrder{Order: order}, Type_: chs.OUTPOOL_CANCELORDER, Count: 0})
		return nil
	} else {
		comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK, "cancelOrder id=%d not in MEEXCore.\n", id)
		return fmt.Errorf("cancelOrder id=%d not in MEEXCore, may have traded!", id)
	}
}

/// Output Process: ----------------------------------------------------------------------------------------
func (t *MEEXCore) matchtradeOutput(bidTrade *comm.Trade, askTrade *comm.Trade) {
	/// debug:
	comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK,
		"%s-%s MatchTrade(bid:%d,ask:%d) Output to channel=======================>>>>>>>>>>\n",
		t.Symbol, t.MarketType.String(), bidTrade.ID, askTrade.ID)

	//// Update bid and ask trade output to ds:
	err, _ := use_mysql.MEMySQLInstance().UpdateTrade(bidTrade, askTrade)
	if err != nil {
		panic(err)
	}

	//// Update tickers infomation
	tradePair := te.TradePair{bidTrade, askTrade}
	t.MEECoreInfo.UpdateTicker(&tradePair)
}

func (t *MEEXCore) cancelOrderOutput(order *comm.Order) {
	comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK,
		"%s-%s CancelOrder(id:%d) Output to channel=======================>>>>>>>>>>\n",
		t.Symbol, t.MarketType.String(), order.ID)

	/// Settle fund and remove from duration storage
	err, _ := use_mysql.MEMySQLInstance().CancelOrder(order)
	if err != nil {
		panic(err)
	}
}

func (t *MEEXCore) multiChanOutProc(chNO int) {
	var (
		v  *chs.OutElem = nil
		ok bool         = false
	)

	for {
		v, ok = t.MultiChans_Out.OutChannel(chNO)
		if ok {
			switch v.Type_ {
			case chs.OUTPOOL_MATCHTRADE:
				t.matchtradeOutput(v.Trade.BidTrade, v.Trade.AskTrade)
			case chs.OUTPOOL_CANCELORDER:
				t.cancelOrderOutput(v.CancelOrder.Order)
			}
		}
		runtime.Gosched()
	}
}

/// Input Process: ----------------------------------------------------------------------------------------
func (t *MEEXCore) orderInput(order *comm.Order) {
	///debug===
	comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_TRACK,
		`=======================>>>>>>>>>>%s-%s Order Input (to ds and container)
		Order id:%d, user:%s, type:%s, symbol:%s, time:%d, price:%f, volume:%f, tatalVolume:%f, fee:%f
		Get from Inchannel(cap:%d, len:%d)
		`,
		t.Symbol, t.MarketType.String(),
		order.ID, order.Who, order.AorB.String(), order.Symbol, order.Timestamp, order.Price, order.Volume, order.TotalVolume, order.Fee,
		INCHANNEL_BUFF_SIZE*INCHANNEL_POOL_SIZE, t.MultiChans_In.Len(),
	)

	//// Enorder to Match Engine Duration Storage
	err, errCode := use_mysql.MEMySQLInstance().EnOrder(order)
	if err != nil {
		if errCode == use_mysql.ErrorCode_DupPrimateKey {
			panic(fmt.Errorf("MEEXCore.orderInput: duplicate order id should not occur, please check it !!!"))
		} else if errCode == use_mysql.ErrorCode_FundNoEnough {
			comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_ALWAYS, "No enough money, errMsg = %s.\n", err.Error())
			return
		} else {
			comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_ALWAYS, "EnOrder fail,  errMsg = %s.\n", err.Error())
			return
		}
	}

	/// Add to trade pool to match
	order.Volume = order.TotalVolume
	t.enOrder(order)

	/// statics
	if order.AorB == comm.TradeType_BID {
		t.DebugInfo_BidEnOrderNormalAdd()
	}
	if order.AorB == comm.TradeType_ASK {
		t.DebugInfo_AskEnOrderNormalAdd()
	}

}

func (t *MEEXCore) multiChanInProc(chNO int) {
	var (
		v  *chs.InElem = nil
		ok bool        = false
	)

	for {
		///trade process
		v, ok = t.MultiChans_In.OutChannel(chNO)
		if ok {
			switch v.Type_ {
			case chs.InElemType_EnOrder:
				t.orderInput(v.Order)
			case chs.InElemType_CancelID:
				t.cancelOrder(v.CancelId)
			}
		} else {
			panic(fmt.Errorf("%s-%s multiChanInProcMultiChans_In.OutChannel(%d) exception occur!", t.Symbol, t.MarketType.String(), chNO))
		}
		runtime.Gosched()
	}
}

/// ----------------------------------------------------------------------------------------
func (t *MEEXCore) initHistoryOrder() (size int64, err error) {
	fmt.Printf("%s: Start to get history orders of %s-%s\n", MODULE_NAME_MEEXCORE, t.Symbol, t.MarketType.String())
	var (
		hs []*comm.Order
	)
	/// get orders from duration storage
	switch t.MarketType {
	case config.MarketType_Human:
		hs, err = use_mysql.MEMySQLInstance().GetAllHumanOrder(t.Symbol, t.MarketConfig.RobotSet.Elements())
	case config.MarketType_Robot:
		hs, err = use_mysql.MEMySQLInstance().GetAllRobotOrder(t.Symbol, t.MarketConfig.RobotSet.Elements())
	case config.MarketType_MixHR:
		hs, err = use_mysql.MEMySQLInstance().GetAllOrder(t.Symbol)
	default:
		panic(fmt.Errorf("initHistoryOrder met illeagal marketType(%s)", t.MarketType.String()))
	}

	if err != nil {
		panic(err)
	}
	hsSize := len(hs)
	fmt.Printf("%s: History orders(%s-%s) scale(%d)\n", MODULE_NAME_MEEXCORE, t.Symbol, t.MarketType.String(), hsSize)

	/// Put them in ME
	fmt.Printf("%s: Start to loading orders(%s-%s) to Match Engine...\n", MODULE_NAME_MEEXCORE, t.Symbol, t.MarketType.String())
	for count, order := range hs {
		if order.AorB == comm.TradeType_BID {
			if (order.Status == comm.ORDER_SUBMIT || order.Status == comm.ORDER_PARTIAL_FILLED) && order.Volume != 0 {
				t.OrderContainerItf.Push(order)
			} else {
				comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_ALWAYS,
					"[InitHistoryOrders]: Market(%s) met illeagal orders with incorrect status in the order duration storage! It should be remove from DS.\n\tOrder info: User(%s), ID(%d), Status(%s)\n",
					t.Symbol, order.Who, order.ID, order.Status,
				)
			}
		} else if order.AorB == comm.TradeType_ASK {
			if (order.Status == comm.ORDER_SUBMIT || order.Status == comm.ORDER_PARTIAL_FILLED) && order.Volume != 0 {
				t.OrderContainerItf.Push(order)
			} else {
				comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_ALWAYS,
					"[InitHistoryOrders]: Market(%s) met illeagal orders with incorrect status in the order duration storage! It should be remove from DS.\n\tOrder info: User(%s), ID(%d), Status(%s)\n",
					t.Symbol, order.Who, order.ID, order.Status,
				)
			}
		} else {
			comm.DebugPrintf(MODULE_NAME_MEEXCORE, comm.LOG_LEVEL_ALWAYS, "[InitHistoryOrders]: Market(%s) met illeagal orders with neith bid nor ask order! It would be remove from duration storage.\n", t.Symbol)
			panic(fmt.Errorf("[InitHistoryOrders]: Market(%s) met illeagal orders with neith bid nor ask order! It would be remove from duration storage.\n", t.Symbol))
		}

		if count == 0 {
			fmt.Printf("%s: %s-%s Adding orders: \n", MODULE_NAME_MEEXCORE, t.Symbol, t.MarketType.String())
		}
		if count%1000 == 0 && count != 0 {
			fmt.Printf("+1000..")
			if count%10000 == 0 {
				fmt.Printf("\n%sPercent: %f%%\n", MODULE_NAME_MEEXCORE, float64(count+1)*100/float64(hsSize))
			}
		}
	}
	fmt.Printf("\n%s: Load %s-%s orders complete.\n", MODULE_NAME_MEEXCORE, t.Symbol, t.MarketType.String())
	return int64(hsSize), nil
}

/// ----------------------------------------------------------------------------------------
