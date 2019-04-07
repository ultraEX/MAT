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
	MODULE_NAME_MEXCORE   string = "[MatchCore]: "
	MXECORE_GOROUTINE_NUM        = 10
)

type MECoreInfo struct {
	Symbol       string
	MarketType   config.MarketType
	MarketConfig *config.MarketConfig

	*te.TickerPool
}

func NewMECoreInfo(symbol string, mrketType config.MarketType, conf *config.MarketConfig, te *te.TickerPool) MECoreInfo {
	return MECoreInfo{Symbol: symbol, MarketType: mrketType, MarketConfig: conf, TickerPool: te}
}

type MEXCore struct {
	comm.OrderContainerItf

	chs.MultiChans_In
	MultiChans_Out chs.MultiChans_OutGo

	MECoreInfo
	rt.DebugInfo

	latestPrice float64
	rt.OrderID
}

func NewMEXCore(sym string, mt config.MarketType, conf *config.MarketConfig, te *te.TickerPool) *MEXCore {
	o := new(MEXCore)

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
		panic(fmt.Errorf("NewMEXCore.NewMEXCore fail, as Algorithm(%s) not support.", config.GetMEConfig().Algorithm))
	}

	o.MultiChans_In = *chs.NewMultiChans_In(o.multiChanInProc)
	o.MultiChans_Out = *chs.NewMultiChans_OutGo(o.multiChanOutProc)

	o.MECoreInfo = NewMECoreInfo(sym, mt, conf, te)
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

	/// to improve with match thread pool!!!
	for i := 0; i < MXECORE_GOROUTINE_NUM; i++ {
		go o.match()
	}

	go o.initHistoryOrder()

	return o
}

func (t *MEXCore) EnOrder(order *comm.Order) error {
	t.MultiChans_In.InChannel(&chs.InElem{Type_: chs.InElemType_EnOrder, Order: order, Count: 0})
	return nil
}

func (t *MEXCore) CancelOrder(id int64) error {
	t.MultiChans_In.InChannel(&chs.InElem{Type_: chs.InElemType_CancelID, CancelId: id, Count: 0})
	return nil
}

func (t *MEXCore) CancelTheOrder(order *comm.Order) error {
	return nil
}

func (t *MEXCore) init(order *comm.Order) error {
	// to do

	comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_TRACK, "init: .\n")
	return nil
}

/// Core process: ----------------------------------------------------------------------------------------
func (t *MEXCore) enOrder(order *comm.Order) error {
	t.OrderContainerItf.Push(order)

	comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_TRACK, "enOrder: id = %d.\n", order.ID)
	return nil
}

func (t *MEXCore) cancelOrder(id int64) error {

	order := t.OrderContainerItf.Get(id)
	if order != nil {
		order.Status = comm.ORDER_CANCELED
		t.MultiChans_Out.InChannel(&chs.OutElem{Trade: nil, CancelOrder: &chs.CanceledOrder{Order: order}, Type_: chs.OUTPOOL_CANCELORDER, Count: 0})
		return nil
	} else {
		comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_TRACK, "cancelOrder id=%d not in MEXCore.\n", id)
		return fmt.Errorf("cancelOrder id=%d not in MEXCore, may have traded!", id)
	}
}

/// Output Process: ----------------------------------------------------------------------------------------
func (t *MEXCore) matchtradeOutput(bidTrade *comm.Trade, askTrade *comm.Trade) {
	/// debug:
	comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_TRACK,
		"%s-%s MatchTrade(bid:%d,ask:%d) Output to channel=======================>>>>>>>>>>\n",
		t.Symbol, t.MarketType.String(), bidTrade.ID, askTrade.ID)

	//// Update bid and ask trade output to ds:
	err, _ := use_mysql.MEMySQLInstance().UpdateTrade(bidTrade, askTrade)
	if err != nil {
		panic(err)
	}

	//// Update tickers infomation
	tradePair := te.TradePair{bidTrade, askTrade}
	t.MECoreInfo.UpdateTicker(&tradePair)
}

func (t *MEXCore) cancelOrderOutput(order *comm.Order) {
	comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_TRACK,
		"%s-%s CancelOrder(id:%d) Output to channel=======================>>>>>>>>>>\n",
		t.Symbol, t.MarketType.String(), order.ID)

	/// Settle fund and remove from duration storage
	err, _ := use_mysql.MEMySQLInstance().CancelOrder(order)
	if err != nil {
		panic(err)
	}
}

func (t *MEXCore) multiChanOutProc(chNO int) {
	var (
		v  *chs.OutElem = nil
		ok bool         = false
	)
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Printf("Panic Error Occur at multiChanOutProc chan NO. %d\n", chNO)
	// 		fmt.Println(err)
	// 		fmt.Printf("Current MECore Dump:\n")
	// 		t.OrderContainerItf.Dump()
	// 	}
	// 	fmt.Printf("multiChanOutProc chan NO. %d process exited!!!\n", chNO)
	// }()

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
func (t *MEXCore) orderInput(order *comm.Order) {
	//// Enorder to Match Engine Duration Storage
	err, errCode := use_mysql.MEMySQLInstance().EnOrder(order)
	if err != nil {
		if errCode == use_mysql.ErrorCode_DupPrimateKey {
			panic(fmt.Errorf("MEXCore.orderInput: duplicate order id should not occur, please check it !!!"))
		} else if errCode == use_mysql.ErrorCode_FundNoEnough {
			comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_ALWAYS, "No enough money, errMsg = %s.\n", err.Error())
			return
		} else {
			comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_ALWAYS, "EnOrder fail,  errMsg = %s.\n", err.Error())
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

	///debug===
	comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_TRACK,
		`=======================>>>>>>>>>>%s-%s Order Input
Order id:%d, user:%s, type:%s, symbol:%s, time:%d, price:%f, volume:%f, tatalVolume:%f, fee:%f
Get from Inchannel(cap:%d, len:%d)
`,
		t.Symbol, t.MarketType.String(),
		order.ID, order.Who, order.AorB.String(), order.Symbol, order.Timestamp, order.Price, order.Volume, order.TotalVolume, order.Fee,
		INCHANNEL_BUFF_SIZE*INCHANNEL_POOL_SIZE, t.MultiChans_In.Len(),
	)
}

func (t *MEXCore) multiChanInProc(chNO int) {
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

/// Matching: ----------------------------------------------------------------------------------------
func (t *MEXCore) inspect() bool {
	orderBid := t.OrderContainerItf.GetTop(comm.TradeType_BID)
	orderAsk := t.OrderContainerItf.GetTop(comm.TradeType_ASK)
	if orderBid != nil && orderAsk != nil {
		if orderBid.Status == comm.ORDER_CANCELED || orderAsk.Status == comm.ORDER_CANCELED {
			return true
		}
		if orderBid.Price >= orderAsk.Price {
			return true
		}
		return false
	}

	if orderBid != nil && orderBid.Status == comm.ORDER_CANCELED {
		return true
	}
	if orderAsk != nil && orderAsk.Status == comm.ORDER_CANCELED {
		return true
	}
	return false

}

func (t *MEXCore) deal(orderBid, orderAsk *comm.Order) {
	var bidStatus, askStatus comm.TradeStatus

	if orderValidatable(orderAsk) && orderValidatable(orderBid) {
		if orderBid.Price >= orderAsk.Price {

			comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_TRACK,
				`=======>>>%s-%s Orders Matching<<<======
BID order == symbol:%s, id:%d, user:%s, time:%d, price:%f, volume:%f
ASK order == symbol:%s, id:%d, user:%s, time:%d, price:%f, volume:%f
`,
				t.Symbol, t.MarketType.String(),
				orderBid.Symbol, orderBid.ID, orderBid.Who, orderBid.Timestamp, orderBid.Price, orderBid.Volume,
				orderAsk.Symbol, orderAsk.ID, orderAsk.Who, orderAsk.Timestamp, orderAsk.Price, orderAsk.Volume,
			)
			///trade price
			tradePrice := float64(0)
			if t.latestPrice <= orderAsk.Price { ///如果前一笔成交价低于或等于卖出价，则最新成交价就是卖出价
				tradePrice = orderAsk.Price
			} else if t.latestPrice >= orderBid.Price { ///如果前一笔成交价高于或等于买入价，则最新成交价就是买入价
				tradePrice = orderBid.Price
			} else { ///如果前一笔成交价在卖出价与买入价之间，则最新成交价就是前一笔的成交价
				tradePrice = t.latestPrice
			}
			priceValidatable(tradePrice)
			t.latestPrice = tradePrice

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
			} else {
				if tradeVolume < orderBid.Volume {
					orderBid.Volume -= tradeVolume
					t.enOrder(orderBid)

					/// updeate order status
					bidStatus = comm.ORDER_PARTIAL_FILLED
					askStatus = comm.ORDER_FILLED
					///debug info
					t.DebugInfo_BidTradeOutputAdd()
					t.DebugInfo_AskTradeCompleteAdd()
				} else {
					orderAsk.Volume -= tradeVolume
					t.enOrder(orderAsk)

					/// updeate order status
					bidStatus = comm.ORDER_FILLED
					askStatus = comm.ORDER_PARTIAL_FILLED
					///debug info
					t.DebugInfo_BidTradeCompleteAdd()
					t.DebugInfo_AskTradeOutputAdd()
				}
			}

			///trade output
			orderTemp := comm.Order{orderBid.ID, orderBid.Who, comm.TradeType_BID, orderBid.Symbol, orderBid.Timestamp, orderBid.EnOrderPrice, tradePrice, tradeVolume, orderBid.TotalVolume, orderBid.Fee, bidStatus, orderBid.IPAddr}
			tradeBid := comm.Trade{orderTemp, tradeBidAmount, time.Now().UnixNano(), tradeVolume * orderBid.Fee}
			orderTemp = comm.Order{orderAsk.ID, orderAsk.Who, comm.TradeType_ASK, orderAsk.Symbol, orderAsk.Timestamp, orderAsk.EnOrderPrice, tradePrice, tradeVolume, orderAsk.TotalVolume, orderAsk.Fee, askStatus, orderAsk.IPAddr}
			tradeAsk := comm.Trade{orderTemp, tradeAskAmount, time.Now().UnixNano(), tradeAmount * orderAsk.Fee}

			///To do: put to channel to send to database
			t.MultiChans_Out.InChannel(&chs.OutElem{Trade: &chs.MatchTrade{&tradeBid, &tradeAsk}, CancelOrder: nil, Type_: chs.OUTPOOL_MATCHTRADE, Count: 0})

		} else {
			t.OrderContainerItf.Push(orderAsk)
			t.OrderContainerItf.Push(orderBid)
		}
	} else {
		fmt.Printf("[Trade]:Met Illegal Orders.\n\tBid Order: User(%s), ID(%d), Status(%s), Price(%f), Volume(%f)\n\tAsk Order: User(%s), ID(%d), Status(%s), Price(%f), Volume(%f)\n",
			orderBid.Who, orderBid.ID, orderBid.Status, orderBid.Price, orderBid.Volume, orderAsk.Who, orderAsk.ID, orderAsk.Status, orderAsk.Price, orderAsk.Volume)
		panic("trade data illegel, need process!========================!")
	}
}

func (t *MEXCore) trade() {
	TimeDot1 := time.Now().UnixNano()
	defer func() {
		TimeDot2 := time.Now().UnixNano()
		t.DebugInfo_RecordCorePerform(TimeDot2 - TimeDot1)
	}()

	orderBid := t.OrderContainerItf.Pop(comm.TradeType_BID)
	orderAsk := t.OrderContainerItf.Pop(comm.TradeType_ASK)

	if orderBid != nil && orderAsk != nil {
		if orderBid.Status != comm.ORDER_CANCELED && orderAsk.Status != comm.ORDER_CANCELED {
			t.deal(orderBid, orderAsk)
			return
		}
		if orderBid.Status == comm.ORDER_CANCELED {
			if orderAsk.Status != comm.ORDER_CANCELED {
				t.OrderContainerItf.Push(orderAsk)
			}
			return
		}
		if orderAsk.Status == comm.ORDER_CANCELED {
			if orderBid.Status != comm.ORDER_CANCELED {
				t.OrderContainerItf.Push(orderBid)
			}
			return
		}
	}

	if orderBid != nil && orderBid.Status != comm.ORDER_CANCELED {
		t.OrderContainerItf.Push(orderBid)
	}
	if orderAsk != nil && orderAsk.Status != comm.ORDER_CANCELED {
		t.OrderContainerItf.Push(orderAsk)
	}

	return
}

func (t *MEXCore) match() {
	// to do
	for {
		if t.inspect() {
			t.trade()
		}

		time.Sleep(comm.MECORE_MATCH_DURATION)
	}

}

/// ----------------------------------------------------------------------------------------
func (t *MEXCore) initHistoryOrder() (size int64, err error) {
	fmt.Printf("%s: Start to get history orders of %s-%s\n", MODULE_NAME_MEXCORE, t.Symbol, t.MarketType.String())
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
	fmt.Printf("%s: History orders(%s-%s) scale(%d)\n", MODULE_NAME_MEXCORE, t.Symbol, t.MarketType.String(), hsSize)

	/// Put them in ME
	fmt.Printf("%s: Start to loading orders(%s-%s) to Match Engine...\n", MODULE_NAME_MEXCORE, t.Symbol, t.MarketType.String())
	for count, order := range hs {
		if order.AorB == comm.TradeType_BID {
			if (order.Status == comm.ORDER_SUBMIT || order.Status == comm.ORDER_PARTIAL_FILLED) && order.Volume != 0 {
				t.OrderContainerItf.Push(order)
			} else {
				comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_ALWAYS,
					"[InitHistoryOrders]: Market(%s) met illeagal orders with incorrect status in the order duration storage! It should be remove from DS.\n\tOrder info: User(%s), ID(%d), Status(%s)\n",
					t.Symbol, order.Who, order.ID, order.Status,
				)
			}
		} else if order.AorB == comm.TradeType_ASK {
			if (order.Status == comm.ORDER_SUBMIT || order.Status == comm.ORDER_PARTIAL_FILLED) && order.Volume != 0 {
				t.OrderContainerItf.Push(order)
			} else {
				comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_ALWAYS,
					"[InitHistoryOrders]: Market(%s) met illeagal orders with incorrect status in the order duration storage! It should be remove from DS.\n\tOrder info: User(%s), ID(%d), Status(%s)\n",
					t.Symbol, order.Who, order.ID, order.Status,
				)
			}
		} else {
			comm.DebugPrintf(MODULE_NAME_MEXCORE, comm.LOG_LEVEL_ALWAYS, "[InitHistoryOrders]: Market(%s) met illeagal orders with neith bid nor ask order! It would be remove from duration storage.\n", t.Symbol)
			panic(fmt.Errorf("[InitHistoryOrders]: Market(%s) met illeagal orders with neith bid nor ask order! It would be remove from duration storage.\n", t.Symbol))
		}

		if count == 0 {
			fmt.Printf("%s: %s-%s Adding orders: \n", MODULE_NAME_MEXCORE, t.Symbol, t.MarketType.String())
		}
		if count%1000 == 0 && count != 0 {
			fmt.Printf("+1000..")
			if count%10000 == 0 {
				fmt.Printf("\n%sPercent: %f%%\n", MODULE_NAME_MEXCORE, float64(count+1)*100/float64(hsSize))
			}
		}
	}
	fmt.Printf("\n%s: Load %s-%s orders complete.\n", MODULE_NAME_MEXCORE, t.Symbol, t.MarketType.String())
	return int64(hsSize), nil
}

/// ----------------------------------------------------------------------------------------
