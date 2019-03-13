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
	hm "./heapmap"
)

const (
	MODULE_NAME_HEAPMAP string = "[MatchCore-Heapmap]: "
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
	hm.TradeContainer

	chs.MultiChans_In
	chs.MultiChans_Out

	MECoreInfo
	rt.DebugInfo

	latestPrice float64
	rt.OrderID
}

func NewMEXCore(sym string, mt config.MarketType, conf *config.MarketConfig, te *te.TickerPool) *MEXCore {
	o := new(MEXCore)
	o.TradeContainer = *hm.NewTradeContainer()

	o.MultiChans_In = *chs.NewMultiChans_In(o.multiChanInProc)
	o.MultiChans_Out = *chs.NewMultiChans_Out(o.multiChanOutProc)

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
	go o.match()

	return o
}

func (t *MEXCore) EnOrder(order *comm.Order) error {
	t.MultiChans_In.InChannel(&chs.InElem{Type_: chs.InElemType_EnOrder, Order: order, Count: 0})

	if order.AorB == comm.TradeType_BID {
		t.DebugInfo_BidEnOrderNormalAdd()
	}
	if order.AorB == comm.TradeType_ASK {
		t.DebugInfo_AskEnOrderNormalAdd()
	}

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

	comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_TRACK, "init: .\n")
	return nil
}

/// Core process: ----------------------------------------------------------------------------------------
func (t *MEXCore) enOrder(order *comm.Order) error {
	t.TradeContainer.Push(order)

	comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_TRACK, "enOrder: id = %d.\n", order.ID)
	return nil
}

func (t *MEXCore) cancelOrder(id int64) error {

	order := t.TradeContainer.Get(id)
	if order != nil {
		order.Status = comm.ORDER_CANCELED
	} else {
		comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_TRACK, "cancelOrder id=%d not in MEXCore.\n", id)
	}

	t.MultiChans_Out.InChannel(&chs.OutElem{Trade: nil, CancelOrder: &chs.CanceledOrder{order}, Type_: chs.OUTPOOL_CANCELORDER, Count: 0})

	return nil
}

/// Output Process: ----------------------------------------------------------------------------------------
func (t *MEXCore) matchtradeOutput(bidTrade *comm.Trade, askTrade *comm.Trade) {
	/// debug:
	comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_TRACK,
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
	comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_TRACK,
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
reEnorder_:
	err, errCode := use_mysql.MEMySQLInstance().EnOrder(order)
	if err != nil {
		if errCode == use_mysql.ErrorCode_DupPrimateKey {
			comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_TRACK, "EnOrder fail, Retry to do it once more.\n")
			order.ID = time.Now().UnixNano()
			goto reEnorder_
		} else if errCode == use_mysql.ErrorCode_FundNoEnough {
			comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_ALWAYS, "No enough money, errMsg = %s.\n", err.Error())
			return
		} else {
			comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_ALWAYS, "EnOrder fail,  errMsg = %s.\n", err.Error())
			// panic(fmt.Errorf("EnOrder fail, check it !!! err = %s", err.Error()))
			return
		}
	}

	/// Add to trade pool to match
	order.Volume = order.TotalVolume
	t.enOrder(order)

	///debug===
	comm.DebugPrintf(MODULE_NAME_HEAPMAP, comm.LOG_LEVEL_TRACK,
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
func (t *MEXCore) trade() {
	var bidStatus, askStatus comm.TradeStatus

	orderAsk := t.TradeContainer.GetTop(comm.TradeType_ASK)
	if orderAsk == nil {
		return
	}
	orderBid := t.TradeContainer.GetTop(comm.TradeType_BID)
	if orderBid == nil {
		return
	}
	if orderAsk.Status == comm.ORDER_CANCELED {
		t.TradeContainer.Pop(comm.TradeType_ASK)
		return
	}
	if orderBid.Status == comm.ORDER_CANCELED {
		t.TradeContainer.Pop(comm.TradeType_BID)
		return
	}

	if orderValidatable(orderAsk) && orderValidatable(orderBid) {
		if orderBid.Price >= orderAsk.Price {
			TimeDot1 := time.Now().UnixNano()

			orderAsk = t.TradeContainer.Pop(comm.TradeType_ASK)
			orderBid = t.TradeContainer.Pop(comm.TradeType_BID)

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

			TimeDot2 := time.Now().UnixNano()
			t.DebugInfo_RecordCorePerform(TimeDot2 - TimeDot1)
			/// fmt.Printf("MEXCore.match trade performance(second this round): %.9f \n", float64(TimeDot2-TimeDot1)/float64(1*time.Second))
		}
	} else {
		fmt.Printf("[Trade]:Met Illegal Orders.\n\tBid Order: User(%s), ID(%d), Status(%s), Price(%f), Volume(%f)\n\tAsk Order: User(%s), ID(%d), Status(%s), Price(%f), Volume(%f)\n",
			orderBid.Who, orderBid.ID, orderBid.Status, orderBid.Price, orderBid.Volume, orderAsk.Who, orderAsk.ID, orderAsk.Status, orderAsk.Price, orderAsk.Volume)
		panic("trade data illegel, need process!========================!")
	}
}

func (t *MEXCore) match() {
	// to do
	for {

		/// To match trade and put it out to outchannels
		t.trade()

		time.Sleep(comm.MECORE_MATCH_DURATION)
		runtime.Gosched()
	}

}

/// ----------------------------------------------------------------------------------------
