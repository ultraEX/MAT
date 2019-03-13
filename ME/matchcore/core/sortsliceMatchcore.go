package core

import (
	"container/list"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"../../comm"
	"../../config"
	"../../db/use_mysql"
	"../../doctor"
	te "../../tickers"
	"../chansIO"
	chs "../chansIO"
	rt "../runtime"
)

const (
	TEST_DATA_SCALE int = 0 //100 * 10000

	INCHANNEL_BUFF_SIZE     int = 168
	INCHANNEL_POOL_SIZE     int = 3
	OUTCHANNEL_BUFF_SIZE    int = 68
	OUTCHANNEL_POOL_SIZE    int = 1
	CANCELCHANNEL_BUFF_SIZE int = 16
)

const (
	MODULE_NAME_SORTSLICE string = "[MatchCore-Sortslice]: "
)

type TradeControl int64

const (
	TradeControl_Work  TradeControl = 0
	TradeControl_Stop  TradeControl = 1
	TradeControl_Pause TradeControl = 2
)

func (p TradeControl) String() string {
	switch p {
	case TradeControl_Work:
		return "TradeControl Work"
	case TradeControl_Stop:
		return "TradeControl Stop"
	case TradeControl_Pause:
		return "TradeControl Pause"
	}
	return "<TradeControl UNSET>"
}

type ReturnStatus int64

const (
	ReturnStatus_OK    ReturnStatus = 0
	ReturnStatus_FAIL  ReturnStatus = 1
	ReturnStatus_RETRY ReturnStatus = 2
)

func (p ReturnStatus) String() string {
	switch p {
	case ReturnStatus_OK:
		return "Success"
	case ReturnStatus_FAIL:
		return "Fail"
	case ReturnStatus_RETRY:
		return "Retry"
	}
	return "<ReturnStatus UNSET>"
}

type InPutPool struct {
	inChannel [INCHANNEL_POOL_SIZE]chan *comm.Order
	cur       int
}

func newInPutPool() *InPutPool {
	o := new(InPutPool)
	for i := 0; i < INCHANNEL_POOL_SIZE; i++ {
		o.inChannel[i] = make(chan *comm.Order, INCHANNEL_BUFF_SIZE)
	}
	o.cur = 0
	return o
}

func (t *InPutPool) GetChannel() int {
	t.cur++
	if t.cur >= INCHANNEL_POOL_SIZE {
		t.cur = 0
	}
	return t.cur
}

func (t *InPutPool) ErrCheck(cur int) error {
	if cur < 0 || cur >= INCHANNEL_POOL_SIZE {
		return fmt.Errorf("InPutPool ErrCheck in cur(%d) fail.", cur)
	}
	return nil
}

func (t *InPutPool) In(order *comm.Order) {
	NO := t.GetChannel()
	t.inChannel[NO] <- order
	comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "InPutPool In order to channel(%d).\n", NO)
}

func (t *InPutPool) Out(cur int) (order *comm.Order, ok bool) {
	err := t.ErrCheck(cur)
	if err == nil {
		order, ok = <-t.inChannel[cur]
		comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "InPutPool Out order from  channel(%d).\n", cur)
		return order, ok
	} else {
		panic(err)
	}
}

func (t *InPutPool) Start(tp *TradePool) {
	for i := 0; i < INCHANNEL_POOL_SIZE; i++ {
		go tp.input(i)
	}
}

func (t *InPutPool) Size() int {
	return INCHANNEL_POOL_SIZE
}
func (t *InPutPool) Len() int {
	sum := 0
	for i := 0; i < INCHANNEL_POOL_SIZE; i++ {
		sum += len(t.inChannel[i])
	}
	return sum
}

type Channel struct {
	InChannelBlock chan *comm.Order
	InChannel      *InPutPool
	MultiChanOut   *chansIO.MultiChans_Out
	CancelChannel  chan *comm.Order
}

///------------------------------------------------------------------
type Lock struct {
	askPoolRWMutex *comm.DebugLock
	bidPoolRWMutex *comm.DebugLock
}

type Control struct {
	tradeControl TradeControl
}

///------------------------------------------------------------------
type TradePool struct {
	Symbol       string
	MarketConfig *config.MarketConfig
	MarketType   config.MarketType

	askPool        *list.List
	askPoolSlice   []*list.Element
	askPoolIDSlice []*list.Element

	bidPool        *list.List
	bidPoolSlice   []*list.Element
	bidPoolIDSlice []*list.Element

	latestPrice float64

	Channel
	Lock
	Control

	debug        *rt.DebugInfo
	doctor       *doctor.Doctor
	tickerEngine *te.TickerPool
	rt.OrderID
}

func NewTradePool(symbol string, mrketType config.MarketType, conf *config.MarketConfig, te *te.TickerPool) *TradePool {
	o := new(TradePool)
	o.doctor = doctor.NewDoctor()
	o.Symbol = symbol
	o.MarketConfig = conf
	o.MarketType = mrketType
	o.tickerEngine = te
	o.latestPrice = te.GetNewestPrice()

	s := strings.Split(symbol, "/")
	if len(s) != 2 {
		panic(fmt.Errorf("NewMEXCore.NewOrderID fail, as sym(%s) input illegal.", symbol))
	}
	vB, okB := config.GetCoinMapInt()[s[0]]
	vQ, okQ := config.GetCoinMapInt()[s[1]]
	if !okB || !okQ {
		panic(fmt.Errorf("NewMEXCore.NewOrderID to GetCoinMapInt(%s) fail.", symbol))
	}
	o.OrderID = *rt.NewOrderID((int(vB) & 0x2f) | (int(vQ) & 0x2f))

	///return o.init()
	return o.setup()
}

///------------------------------------------------------------------
type sortByAskPrice []*list.Element

func (I sortByAskPrice) Len() int {
	return len(I)
}

func (I sortByAskPrice) Less(i, j int) bool {
	return I[i].Value.(comm.Order).Price < I[j].Value.(comm.Order).Price
}

func (I sortByAskPrice) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

///------------------------------------------------------------------
type sortByBidPrice []*list.Element

func (I sortByBidPrice) Len() int {
	return len(I)
}

func (I sortByBidPrice) Less(i, j int) bool {
	return I[i].Value.(comm.Order).Price > I[j].Value.(comm.Order).Price
}

func (I sortByBidPrice) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

///------------------------------------------------------------------
type sortByTime []*list.Element

func (I sortByTime) Len() int {
	return len(I)
}

func (I sortByTime) Less(i, j int) bool {
	return I[i].Value.(comm.Order).Timestamp < I[j].Value.(comm.Order).Timestamp
}

func (I sortByTime) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

///------------------------------------------------------------------
type sortOrderByID []*list.Element

func (I sortOrderByID) Len() int {
	return len(I)
}

func (I sortOrderByID) Less(i, j int) bool {
	return I[i].Value.(comm.Order).ID < I[j].Value.(comm.Order).ID
}

func (I sortOrderByID) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

///------------------------------------------------------------------
type sortByID []*list.Element

func (I sortByID) Len() int {
	return len(I)
}

func (I sortByID) Less(i, j int) bool {
	return I[i].Value.(int64) < I[j].Value.(int64)
}

func (I sortByID) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}

///------------------------------------------------------------------
func (t *TradePool) dumpBuff() (askCount int, askSliceCount int, askIDSliceCount int, bidCount int, bidSliceCount int, bidIDSliceCount int, strBuff string) {
	t.askPoolRWMutex.RLock("Dump ASK")
	t.bidPoolRWMutex.RLock("Dump BID")
	defer t.askPoolRWMutex.RUnlock("Dump ASK")
	defer t.bidPoolRWMutex.RUnlock("Dump BID")

	/// ask order info:
	strBuff = fmt.Sprintf("=================[%s-%s Ask Pool]=================\n", t.Symbol, t.MarketType.String())
	askCount = 0
	e := t.askPool.Front()
	for elem := e; elem != nil; elem = elem.Next() {
		strBuff = fmt.Sprintf(strBuff+"ASK ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		askCount++
	}
	strBuff = fmt.Sprintf(strBuff+"[askPool]: Total(%d)\n", askCount)

	strBuff = fmt.Sprintf(strBuff+"=================[%s-%s Ask PoolSlice]=================\n", t.Symbol, t.MarketType.String())
	askSliceCount = 0
	for _, elem := range t.askPoolSlice {
		strBuff = fmt.Sprintf(strBuff+"ASK ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		askSliceCount++
	}
	strBuff = fmt.Sprintf(strBuff+"[askPoolSlice]: Total(%d)\n", askSliceCount)

	strBuff = fmt.Sprintf(strBuff+"=================[%s-%s Ask PoolIDSlice]=================\n", t.Symbol, t.MarketType.String())
	askIDSliceCount = 0
	for _, elem := range t.askPoolIDSlice {
		strBuff = fmt.Sprintf(strBuff+"ASK ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		askIDSliceCount++
	}
	strBuff = fmt.Sprintf(strBuff+"[askPoolIDSlice]: Total(%d)\n", askIDSliceCount)

	/// bid order info:
	strBuff = fmt.Sprintf(strBuff+"=================[%s-%s Bid Pool]=================\n", t.Symbol, t.MarketType.String())
	bidCount = 0
	e = t.bidPool.Front()
	for elem := e; elem != nil; elem = elem.Next() {
		strBuff = fmt.Sprintf(strBuff+"BID ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		bidCount++
	}
	strBuff = fmt.Sprintf(strBuff+"[bidPool]: Total(%d)\n", bidCount)

	strBuff = fmt.Sprintf(strBuff+"=================[%s-%s Bid PoolSlice]=================\n", t.Symbol, t.MarketType.String())
	bidSliceCount = 0
	for _, elem := range t.bidPoolSlice {
		strBuff = fmt.Sprintf(strBuff+"BID ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		bidSliceCount++
	}
	strBuff = fmt.Sprintf(strBuff+"[bidPoolSlice]: Total(%d)\n", bidSliceCount)

	/// bid order id slice info:
	strBuff = fmt.Sprintf(strBuff+"=================[%s-%s Bid PoolIDSlice]=================\n", t.Symbol, t.MarketType.String())
	bidIDSliceCount = 0
	for _, elem := range t.bidPoolIDSlice {
		strBuff = fmt.Sprintf(strBuff+"BID ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		bidIDSliceCount++
	}
	strBuff = fmt.Sprintf(strBuff+"[bidPoolIDSlice]: Total(%d)\n", bidIDSliceCount)

	return askCount, askSliceCount, askIDSliceCount, bidCount, bidSliceCount, bidIDSliceCount, strBuff
}

///------------------------------------------------------------------
func (t *TradePool) dumpPrint() {
	var (
		askCount, askSliceCount, askIDSliceCount, bidIDSliceCount, bidCount, bidSliceCount int
	)
	t.askPoolRWMutex.RLock("Dump ASK")
	t.bidPoolRWMutex.RLock("Dump BID")
	defer t.askPoolRWMutex.RUnlock("Dump ASK")
	defer t.bidPoolRWMutex.RUnlock("Dump BID")

	/// ask order info:
	fmt.Printf("=================[%s-%s Ask Pool]=================\n", t.Symbol, t.MarketType.String())
	askCount = 0
	e := t.askPool.Front()
	for elem := e; elem != nil; elem = elem.Next() {
		fmt.Printf("ASK ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		askCount++
	}
	fmt.Printf("[askPool]: Total(%d)\n", askCount)

	fmt.Printf("=================[%s-%s Ask PoolSlice]=================\n", t.Symbol, t.MarketType.String())
	askSliceCount = 0
	for _, elem := range t.askPoolSlice {
		fmt.Printf("ASK ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		askSliceCount++
	}
	fmt.Printf("[askPoolSlice]: Total(%d)\n", askSliceCount)

	fmt.Printf("=================[%s-%s Ask PoolIDSlice]=================\n", t.Symbol, t.MarketType.String())
	askIDSliceCount = 0
	for _, elem := range t.askPoolIDSlice {
		fmt.Printf("ASK ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		askIDSliceCount++
	}
	fmt.Printf("[askPoolIDSlice]: Total(%d)\n", askIDSliceCount)

	/// bid order info:
	fmt.Printf("=================[%s-%s Bid Pool]=================\n", t.Symbol, t.MarketType.String())
	bidCount = 0
	e = t.bidPool.Front()
	for elem := e; elem != nil; elem = elem.Next() {
		fmt.Printf("BID ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		bidCount++
	}
	fmt.Printf("[bidPool]: Total(%d)\n", bidCount)

	fmt.Printf("=================[%s-%s Bid PoolSlice]=================\n", t.Symbol, t.MarketType.String())
	bidSliceCount = 0
	for _, elem := range t.bidPoolSlice {
		fmt.Printf("BID ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		bidSliceCount++
	}
	fmt.Printf("[bidPoolSlice]: Total(%d)\n", bidSliceCount)

	/// bid order id slice info:
	fmt.Printf("=================[%s-%s Bid PoolIDSlice]=================\n", t.Symbol, t.MarketType.String())
	bidIDSliceCount = 0
	for _, elem := range t.bidPoolIDSlice {
		fmt.Printf("BID ORDER=== symbol:%s, id:%d, user:%s, price:%f, time:%d, volume:%f\n",
			elem.Value.(comm.Order).Symbol, elem.Value.(comm.Order).ID, elem.Value.(comm.Order).Who, elem.Value.(comm.Order).Price, elem.Value.(comm.Order).Timestamp, elem.Value.(comm.Order).Volume,
		)
		bidIDSliceCount++
	}
	fmt.Printf("[bidPoolIDSlice]: Total(%d)\n", bidIDSliceCount)

}

func (t *TradePool) dump() {
	t.dumpPrint()
}

func (t *TradePool) DumpTradePoolPrint(detail bool) {
	fmt.Printf("==================[%s-%s Trade Pool Info]=====================\n", t.Symbol, t.MarketType.String())
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	fmt.Printf("Date Time: %s\n", time.Now().In(loc).Format(formate))

	bidCount := len(t.bidPoolSlice)
	askCount := len(t.askPoolSlice)
	if detail {
		t.dumpPrint()
	}
	fmt.Printf("Bid Order Total Num:\t %d\n", bidCount)
	fmt.Printf("Ask Order Total Num:\t %d\n", askCount)
	fmt.Printf("Order Total Num:\t %d\n", bidCount+askCount)

	fmt.Printf("=======================================================\n")
}

func (t *TradePool) DumpTradePool(detail bool) string {
	strBuff := fmt.Sprintf("==================[%s-%s Trade Pool Info]=====================\n", t.Symbol, t.MarketType.String())
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	strBuff = fmt.Sprintf(strBuff+"Date Time: %s\n", time.Now().In(loc).Format(formate))

	if detail {
		askCount, askSliceCount, askIDSliceCount, bidCount, bidSliceCount, bidIDSliceCount, strDetail := t.dumpBuff()
		strBuff = fmt.Sprintf(strBuff + strDetail)
		strBuff = fmt.Sprintf(strBuff+"Bid Order Total Num:\t %d\n", bidCount)
		strBuff = fmt.Sprintf(strBuff+"Bid Slice Total Num:\t %d\n", bidSliceCount)
		strBuff = fmt.Sprintf(strBuff+"Bid IDSlice Total Num:\t %d\n", bidIDSliceCount)
		strBuff = fmt.Sprintf(strBuff+"Ask Order Total Num:\t %d\n", askCount)
		strBuff = fmt.Sprintf(strBuff+"Ask Slice Total Num:\t %d\n", askSliceCount)
		strBuff = fmt.Sprintf(strBuff+"Ask IDSlice Total Num:\t %d\n", askIDSliceCount)
		strBuff = fmt.Sprintf(strBuff+"Order Total Num:\t %d\n", bidCount+askCount)
	} else {
		bidCount := len(t.bidPoolSlice)
		askCount := len(t.askPoolSlice)

		strBuff = fmt.Sprintf(strBuff+"Bid Order Total Num:\t %d\n", bidCount)
		strBuff = fmt.Sprintf(strBuff+"Ask Order Total Num:\t %d\n", askCount)
		strBuff = fmt.Sprintf(strBuff+"Order Total Num:\t %d\n", bidCount+askCount)
	}

	strBuff = fmt.Sprintf(strBuff + "=======================================================\n")

	fmt.Print(strBuff)
	return strBuff
}

func (t *TradePool) DumpBeatHeart() string {
	strBuff := fmt.Sprintf("==================[%s-%s Beatheart Infoo]=====================\n", t.Symbol, t.MarketType.String())
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	strBuff = fmt.Sprintf(strBuff+"Date Time: %s\n", time.Now().In(loc).Format(formate))

	strBuff += t.doctor.DumpAllBeatHeart()

	strBuff = fmt.Sprintf(strBuff + "=======================================================\n")

	fmt.Print(strBuff)
	return strBuff
}

func (t *TradePool) DumpChannel() string {
	strBuff := fmt.Sprintf("==================[%s-%s Channel Infoo]=====================\n", t.Symbol, t.MarketType.String())
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	strBuff += fmt.Sprintf("Date Time: %s\n", time.Now().In(loc).Format(formate))

	strBuff += fmt.Sprintf("Inchannel Status\t: num=%d * (cap=%d, len=%d)\n", INCHANNEL_POOL_SIZE, INCHANNEL_BUFF_SIZE, t.InChannel.Len())
	strBuff += fmt.Sprintf("CancelChannel Status\t: cap=%d, len=%d\n", CANCELCHANNEL_BUFF_SIZE, len(t.CancelChannel))

	strBuff = fmt.Sprintf(strBuff + "=======================================================\n")

	fmt.Print(strBuff)
	return strBuff
}

func (t *TradePool) DumpChanlsMap() {
	fmt.Printf("==================[%s-%s Channel Map Infoo]=====================\n", t.Symbol, t.MarketType.String())
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	fmt.Printf("Date Time: %s\n", time.Now().In(loc).Format(formate))
	t.MultiChanOut.Dump()
	fmt.Printf("=======================================================\n")
}

func (t *TradePool) GetChannel() Channel {

	return t.Channel
}

const FAULTY_DIAGNOSE_MIN_TASK_PROTECT int = 1

func (t *TradePool) IsFaulty() bool {
	//	isLaunchFaulty := t.doctor.IsLaunchFault()
	isEnorderFaulty := false
	if config.GetMEConfig().InPoolMode == "block" {
		isEnorderFaulty = t.doctor.IsEnorderFault() && len(t.InChannelBlock) >= FAULTY_DIAGNOSE_MIN_TASK_PROTECT
	} else {
		isEnorderFaulty = t.doctor.IsEnorderFault() && t.InChannel.Len() >= FAULTY_DIAGNOSE_MIN_TASK_PROTECT
	}

	isCancelFaulty := t.doctor.IsCancelOrderFault() && len(t.CancelChannel) >= FAULTY_DIAGNOSE_MIN_TASK_PROTECT
	isMatchFaulty := t.doctor.IsMatchCoreFault()
	return isEnorderFaulty ||
		isCancelFaulty ||
		isMatchFaulty

}

///------------------------------------------------------------------
func (t *TradePool) RestartDebuginfo() {
	t.debug.DebugInfo_RestartDebuginfo()
}

func (t *TradePool) ResetMatchCorePerform() {
	t.debug.DebugInfo_ResetMatchCorePerform()
}

func (t *TradePool) GetTradeCompleteRate() float64 {
	return t.debug.DebugInfo_GetTradeCompleteRate()
}

func (t *TradePool) GetAskPoolLen() int {
	return len(t.askPoolSlice)
}

func (t *TradePool) GetBidPoolLen() int {
	return len(t.bidPoolSlice)
}

func (t *TradePool) GetPoolLen() int {
	return len(t.bidPoolSlice) + len(t.askPoolSlice)
}

func (t *TradePool) GetAskLevelOrders(limit int64) []*comm.Order {
	var os []*comm.Order

	for c, elem := range t.askPoolSlice {
		if int64(c) >= limit {
			break
		}
		o := elem.Value.(comm.Order)
		os = append(os, &o)
	}
	return os
}

func (t *TradePool) GetBidLevelOrders(limit int64) []*comm.Order {
	var os []*comm.Order

	for c, elem := range t.bidPoolSlice {
		if int64(c) >= limit {
			break
		}
		o := elem.Value.(comm.Order)
		os = append(os, &o)
	}
	return os
}

func (t *TradePool) GetAskLevelsGroupByPrice(limit int64) []comm.OrderLevel {
	var (
		ols                []comm.OrderLevel /*= make([]OrderLevel, limit)*/
		ol                 comm.OrderLevel   = comm.OrderLevel{float64(0), float64(0), float64(0)}
		curPrice, prePrice float64           = float64(0), float64(0)
		levels             int64             = 0
		levelFull          bool              = false
	)

	/// do read protect
	for _, elem := range t.askPoolSlice {

		curPrice = elem.Value.(comm.Order).Price

		if curPrice != prePrice {
			if prePrice != 0 {
				ols = append(ols, ol)
				levels++
				if levels >= limit {
					levelFull = true
					break
				}
				ol = comm.OrderLevel{float64(0), float64(0), float64(0)}
			}
			prePrice = curPrice
			ol.Price = curPrice
			ol.Volume = elem.Value.(comm.Order).Volume
			ol.TotalVolume = elem.Value.(comm.Order).TotalVolume

		} else {
			ol.Volume += elem.Value.(comm.Order).Volume
			ol.TotalVolume += elem.Value.(comm.Order).TotalVolume
		}
	}

	if !levelFull && limit != 0 && ol.Price != float64(0) && ol.Volume != float64(0) && ol.TotalVolume != float64(0) {
		ols = append(ols, ol)
	}
	return ols
}

func (t *TradePool) GetBidLevelsGroupByPrice(limit int64) []comm.OrderLevel {
	var (
		ols                []comm.OrderLevel /*= make([]OrderLevel, limit)*/
		ol                 comm.OrderLevel   = comm.OrderLevel{float64(0), float64(0), float64(0)}
		curPrice, prePrice float64           = float64(0), float64(0)
		levels             int64             = 0
		levelFull          bool              = false
	)

	/// do read protect
	for _, elem := range t.bidPoolSlice {

		curPrice = elem.Value.(comm.Order).Price

		if curPrice != prePrice {
			if prePrice != 0 {
				ols = append(ols, ol)
				levels++
				if levels >= limit {
					levelFull = true
					break
				}
				ol = comm.OrderLevel{float64(0), float64(0), float64(0)}
			}
			prePrice = curPrice
			ol.Price = curPrice
			ol.Volume = elem.Value.(comm.Order).Volume
			ol.TotalVolume = elem.Value.(comm.Order).TotalVolume

		} else {
			ol.Volume += elem.Value.(comm.Order).Volume
			ol.TotalVolume += elem.Value.(comm.Order).TotalVolume
		}
	}

	if !levelFull && limit != 0 && ol.Price != float64(0) && ol.Volume != float64(0) && ol.TotalVolume != float64(0) {
		ols = append(ols, ol)
	}
	return ols
}

func (t *TradePool) Statics() string {
	strBuff := fmt.Sprintf("===============[Market %s-%s Trade Info]==============\n", t.Symbol, t.MarketType.String())
	strBuff = fmt.Sprintf(strBuff + "===================(User Input Order)====================\n")
	strBuff = fmt.Sprintf(strBuff+"ASK ORDERS		: %d\n", t.debug.DebugInfo_GetUserAskEnOrders())
	strBuff = fmt.Sprintf(strBuff+"BID ORDERS		: %d\n", t.debug.DebugInfo_GetUserBidEnOrders())

	strBuff = fmt.Sprintf(strBuff + "=====================(Add+QuickAdd)======================\n")
	strBuff = fmt.Sprintf(strBuff+"ORDER total		: %d\n", t.debug.DebugInfo_GetEnOrders())
	strBuff = fmt.Sprintf(strBuff+"ASK ORDERS		: %d\n", t.debug.DebugInfo_GetAskEnOrders())
	strBuff = fmt.Sprintf(strBuff+"BID ORDERS		: %d\n", t.debug.DebugInfo_GetBidEnOrders())

	strBuff = fmt.Sprintf(strBuff + "====================(Output+Complete)====================\n")
	strBuff = fmt.Sprintf(strBuff+"TRADE OUTPUTS	: %d\n", t.debug.DebugInfo_GetTradeOutputs())
	strBuff = fmt.Sprintf(strBuff+"TRADE COMPLETES	: %d\n", t.debug.DebugInfo_GetTradeCompletes())

	strBuff = fmt.Sprintf(strBuff + "=====================[Trade Statics]=====================\n")
	strBuff = fmt.Sprintf(strBuff+"Trade Complete Rate	: %f\n", t.debug.DebugInfo_GetTradeCompleteRate())
	strBuff = fmt.Sprintf(strBuff+"Trade Output Rate	: %f\n", t.debug.DebugInfo_GetTradeOutputRate())
	strBuff = fmt.Sprintf(strBuff+"Trade UserInput Rate	: %f\n", t.debug.DebugInfo_GetUserEnOrderRate())
	strBuff = fmt.Sprintf(strBuff + "----------------------------------------------------------\n")
	max, min, ave := t.debug.DebugInfo_GetCorePerform()
	strBuff = fmt.Sprintf(strBuff+"Match core performance(second/round): min=%.9f, max=%.9f, ave=%.9f\n", min, max, ave)
	strBuff = fmt.Sprintf(strBuff + "----------------------------------------------------------\n")

	strBuff = fmt.Sprintf(strBuff + "================[Trade Output Pool Status]===============\n")
	strBuff = fmt.Sprintf(strBuff+"Ask Pool Scale	:	%d\n", t.askPool.Len())
	strBuff = fmt.Sprintf(strBuff+"Bid Pool Scale	:	%d\n", t.bidPool.Len())
	strBuff = fmt.Sprintf(strBuff+"Newest Price		:	%f\n", t.latestPrice)

	strBuff = fmt.Sprintf(strBuff + "----------------------------------------------------------\n")
	strBuff = fmt.Sprintf(strBuff+"InChannel Pool Work Mode:	%s\n", config.GetMEConfig().InPoolMode)
	strBuff = fmt.Sprintf(strBuff + "----------------------------------------------------------\n")

	strBuff = fmt.Sprintf(strBuff+"InChannelBlock Cap	:	%d\n", INCHANNEL_BUFF_SIZE)   ///(chan buff size)
	strBuff = fmt.Sprintf(strBuff+"InChannelBlock Len	:	%d\n", len(t.InChannelBlock)) ///(buff total usage)

	strBuff = fmt.Sprintf(strBuff + "----------------------------------------------------------\n")
	strBuff = fmt.Sprintf(strBuff+"InChannel Pool Size	:	%d\n", INCHANNEL_POOL_SIZE)                    ///(channel num)
	strBuff = fmt.Sprintf(strBuff+"InChannel Buff Size	:	%d\n", INCHANNEL_BUFF_SIZE)                    ///(buff size per chan)
	strBuff = fmt.Sprintf(strBuff+"InChannel Pool Cap	:	%d\n", INCHANNEL_BUFF_SIZE*INCHANNEL_POOL_SIZE) ///(total buff size)
	strBuff = fmt.Sprintf(strBuff+"InChannel Pool Len	:	%d\n", t.InChannel.Len())                       ///(buff total usage)
	strBuff = fmt.Sprintf(strBuff+"InChannel Pool Current Channel:	%d\n", t.InChannel.GetChannel())     ///(order output serialize map scale)
	strBuff = fmt.Sprintf(strBuff + "----------------------------------------------------------\n")
	strBuff = fmt.Sprintf(strBuff + "----------------------------------------------------------\n")
	strBuff = fmt.Sprintf(strBuff+"MultiChansOut Status	:	Pool size = %d\n", t.MultiChanOut.Len())
	strBuff = fmt.Sprintf(strBuff+"MultiChansOut Status	:	Chan size = %d\n", t.MultiChanOut.ChanCap())
	IDs, CHs, chnums := t.MultiChanOut.GetChanUseStatus()
	strBuff = fmt.Sprintf(strBuff+"MultiChansOut Chans Usage status: %d; %d; %d\n", IDs, CHs, chnums)

	strBuff = fmt.Sprintf(strBuff + "=======================================================\n")

	fmt.Print(strBuff)
	return strBuff
}

func getSlice(l *list.List) []*list.Element {
	var slice []*list.Element
	for e := l.Front(); e != nil; e = e.Next() {
		slice = append(slice, e)
	}

	return slice
}

func SortByAskPrice(r sortByAskPrice) []*list.Element {
	///debug ==
	//	fmt.Println("poolsice Before sort by price:========================\n")
	//	for _, elem := range r {
	//		fmt.Print(elem.Value.(order).price, "; ", elem.Value.(order).timestamp, "\n")
	//	}
	///sort the pool by price
	sort.Sort(r)
	///debug ==
	//	fmt.Println("poolsice After sort by price:========================\n")
	//	for _, elem := range r {
	//		fmt.Println(elem.Value.(order).price, "; ", elem.Value.(order).timestamp, "\n")
	//	}
	return r
}

func SortByBidPrice(r sortByBidPrice) []*list.Element {
	///debug ==
	//	fmt.Println("poolsice Before sort by bid price:========================\n")
	//	for _, elem := range r {
	//		fmt.Print(elem.Value.(order).price, "; ", elem.Value.(order).timestamp, "\n")
	//	}
	///sort the pool by price
	sort.Sort(r)
	//	///debug ==
	//	fmt.Println("poolsice After sort by bid price:========================\n")
	//	for _, elem := range r {
	//		fmt.Println(elem.Value.(order).price, "; ", elem.Value.(order).timestamp, "\n")
	//	}
	return r
}

func SortOrderByID(r sortOrderByID) []*list.Element {
	///debug ==
	//	fmt.Println("poolsice Before sort by id:========================\n")
	//	for _, elem := range r {
	//		fmt.Print(elem.Value.(comm.Order).ID, "; ", elem.Value.(comm.Order).Timestamp, "\n")
	//	}

	///sort the pool by price
	sort.Sort(r)

	///debug ==
	//	fmt.Println("poolsice After sort by id:========================\n")
	//	for _, elem := range r {
	//		fmt.Println(elem.Value.(comm.Order).ID, "; ", elem.Value.(comm.Order).Timestamp, "\n")
	//	}
	return r
}

func SortByID(r sortByID) []*list.Element {
	///debug ==
	fmt.Println("cancelsice Before sort by ID:========================\n")
	for _, elem := range r {
		fmt.Print(elem.Value.(int64), "; ", elem.Value.(int64), "\n")
	}

	///sort the cancelsice by ID
	sort.Sort(r)

	///debug ==
	fmt.Println("poolsice After sort by price:========================\n")
	for _, elem := range r {
		fmt.Println(elem.Value.(int64), "; ", elem.Value.(int64), "\n")
	}
	return r
}

func SortByTime(r []*list.Element) []*list.Element {
	if len(r) == 0 {
		return make([]*list.Element, 0, 1)
	}

	///sort the pool by time
	sbt := sortByTime{}
	sbp := []*list.Element{}
	var preElem *list.Element = r[0]
	sbt = append(sbt, preElem)
	sCount := 1
	start := 0
	count := 1
	for _, elem := range r[1:] {
		if preElem.Value.(comm.Order).Price == elem.Value.(comm.Order).Price {
			sCount++
		} else {
			if sCount > 1 {
				sort.Sort(sbt)
				sbp = nil
				sbp = append(sbp, r[:start]...)
				sbp = append(sbp, sbt...)
				sbp = append(sbp, r[count:]...)
				r = sbp
			}
			sbt = nil
			sCount = 1
			start = count
		}
		sbt = append(sbt, elem)
		preElem = elem
		count++
	}
	if sCount > 1 {
		sort.Sort(sbt)
		sbp = nil
		sbp = append(sbp, r[:start]...)
		sbp = append(sbp, sbt...)
		sbp = append(sbp, r[count:]...)
		r = sbp
	}
	//	///debug ==
	//	fmt.Println("poolsice After sort by time:========================\n")
	//	for _, elem := range r {
	//		fmt.Println(elem.Value.(order).price, "; ", elem.Value.(order).timestamp, "\n")
	//	}

	return r
}

func (t *TradePool) sortPool(p *list.List, ab comm.TradeType) (sortPool *list.List, sortSlice []*list.Element, sortIDSlice []*list.Element) {
	r := getSlice(p)
	if len(r) == 0 {
		return p, make([]*list.Element, 0, 1), make([]*list.Element, 0, 1)
	}

	///sort the pool slice by price
	if ab == comm.TradeType_ASK {
		r = SortByAskPrice(r)
	} else {
		r = SortByBidPrice(r)
	}

	///sort the pool slice by time
	r = SortByTime(r)

	///sort the pool
	var preElem *list.Element
	for count, elem := range r {
		if count == 0 {
			p.MoveToFront(elem)
			preElem = p.Front()
		} else {
			p.MoveAfter(elem, preElem)
			preElem = elem
		}
	}
	//	///debug ==
	//	fmt.Println("pool After sort by time and price:========================\n")
	//	for elem := p.Front(); elem != nil; {
	//		fmt.Println(elem.Value.(order).price, "; ", elem.Value.(order).timestamp, "\n")
	//		elem = elem.Next()
	//	}
	sortPool = p
	sortSlice = r

	/// sort askPoolIDSlice
	idSlice := []*list.Element{}
	idSlice = append(idSlice, r[:]...)
	sortIDSlice = SortOrderByID(idSlice)
	return
}

func (t *TradePool) sortCancelPool(p *list.List) (sortPool *list.List, sortSlice []*list.Element) {
	r := getSlice(p)

	///sort the cancel pool slice by ID
	r = SortByID(r)

	///sort the cancel pool
	var preElem *list.Element
	for count, elem := range r {
		if count == 0 {
			p.MoveToFront(elem)
			preElem = p.Front()
		} else {
			p.MoveAfter(elem, preElem)
			preElem = elem
		}
	}
	//	///debug ==
	fmt.Println("pool After sort by time and price:========================\n")
	for elem := p.Front(); elem != nil; {
		fmt.Println(elem.Value.(int64), "; ", elem.Value.(int64), "\n")
		elem = elem.Next()
	}

	sortPool = p
	sortSlice = r
	return
}

func (t *TradePool) initHistoryOrder() (size int64, err error) {
	fmt.Printf("%s: Start to get history orders of %s-%s\n", MODULE_NAME_SORTSLICE, t.Symbol, t.MarketType.String())
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
	fmt.Printf("%s: History orders(%s-%s) scale(%d)\n", MODULE_NAME_SORTSLICE, t.Symbol, t.MarketType.String(), hsSize)

	/// Put them in ME
	fmt.Printf("%s: Start to loading orders(%s-%s) to Match Engine...\n", MODULE_NAME_SORTSLICE, t.Symbol, t.MarketType.String())
	for count, order := range hs {
		if order.AorB == comm.TradeType_BID {
			if order.Status == comm.ORDER_SUBMIT || order.Status == comm.ORDER_PARTIAL_FILLED {
				t.bidPool.PushBack(*order)
			} else {
				comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_ALWAYS, "[InitHistoryOrders]: Market(%s) met illeagal orders with incorrect status in the order duration storage! It should be remove from DS.\n\tOrder info: User(%s), ID(%d), Status(%s)\n",
					t.Symbol, order.Who, order.ID, order.Status)
				//				err := use_mysql.MEMySQLInstance().RmOrder(order.Who, order.ID, order.Symbol, nil)
				//				if err != nil {
				//					panic(fmt.Errorf("[InitHistoryOrders]: Met errors, should be fixed!"))
				//				}
			}
		} else if order.AorB == comm.TradeType_ASK {
			if order.Status == comm.ORDER_SUBMIT || order.Status == comm.ORDER_PARTIAL_FILLED {
				t.askPool.PushBack(*order)
			} else {
				comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_ALWAYS, "[InitHistoryOrders]: Market(%s) met illeagal orders with incorrect status in the order duration storage! It should be remove from DS.\n\tOrder info: User(%s), ID(%d), Status(%s)\n",
					t.Symbol, order.Who, order.ID, order.Status)
				//				err := use_mysql.MEMySQLInstance().RmOrder(order.Who, order.ID, order.Symbol, nil)
				//				if err != nil {
				//					panic(fmt.Errorf("[InitHistoryOrders]: Met errors, should be fixed!"))
				//				}
			}
		} else {
			fmt.Println("[InitHistoryOrders]: Market(%s) met illeagal orders with neith bid nor ask order! It would be remove from duration storage.\n", t.Symbol)
			err := use_mysql.MEMySQLInstance().RmOrder(order.Who, order.ID, order.Symbol, nil)
			if err != nil {
				panic(fmt.Errorf("[InitHistoryOrders]: Met errors, should be fixed!"))
			}
		}
		if count == 0 {
			fmt.Printf("%s: %s-%s Adding orders: \n", MODULE_NAME_SORTSLICE, t.Symbol, t.MarketType.String())
		}
		if count%1000 == 0 && count != 0 {
			fmt.Printf("+1000..")
			if count%10000 == 0 {
				fmt.Printf("\n%sPercent: %f%%\n", MODULE_NAME_SORTSLICE, float64(count+1)*100/float64(hsSize))
			}
		}
	}
	fmt.Printf("\n%s: Load %s-%s orders complete.\n", MODULE_NAME_SORTSLICE, t.Symbol, t.MarketType.String())
	return int64(hsSize), nil
}

func (t *TradePool) initTestData() {
	//// test data construction: bid + ask
	for i := 0; i < TEST_DATA_SCALE; i++ {
		volume := (10 + 10*float64(rand.Intn(10))/10)
		price := 1 + float64(rand.Intn(3))/10
		tmpBid := comm.Order{time.Now().UnixNano(), "Hunter", comm.TradeType_BID, t.Symbol,
			time.Now().UnixNano(), price, price, volume, volume, 0.001, comm.ORDER_SUBMIT, "localhost:IP"}
		t.bidPool.PushBack(tmpBid)
	}
	for i := 0; i < TEST_DATA_SCALE; i++ {
		volume := (10 + 10*float64(rand.Intn(10))/10)
		price := 1 + float64(rand.Intn(3))/10
		tmpAsk := comm.Order{time.Now().UnixNano(), "Hunter", comm.TradeType_ASK, t.Symbol,
			time.Now().UnixNano(), price, price, volume, volume, 0.001, comm.ORDER_SUBMIT, "localhost:IP"}
		t.askPool.PushBack(tmpAsk)
	}
}

func (t *TradePool) init() *TradePool {
	t.doctor.SetProgress(doctor.Progress_BeginInit)

	//// init trade pool
	t.askPool = list.New()
	t.bidPool = list.New()

	/// debug:
	TimeDot1 := time.Now().UnixNano()
	/// initTestData()

	/// debug:
	TimeDot2 := time.Now().UnixNano()

	//// History order init to Match Engine
	hsize, err := t.initHistoryOrder()
	if err != nil {
		panic(err)
	}
	/// debug:
	TimeDot3 := time.Now().UnixNano()

	//	fmt.Println("pool data init:========================\n")
	//	t.Dump()

	//// sort the trade pool: bid + ask
	t.askPool, t.askPoolSlice, t.askPoolIDSlice = t.sortPool(t.askPool, comm.TradeType_ASK)
	/// debug:
	TimeDot4 := time.Now().UnixNano()
	t.bidPool, t.bidPoolSlice, t.bidPoolIDSlice = t.sortPool(t.bidPool, comm.TradeType_BID)
	/// debug:
	TimeDot5 := time.Now().UnixNano()

	//	fmt.Println("pool data sorted:========================\n")
	//	t.Dump()

	///init Ticker
	t.latestPrice = float64(0)

	///init Channel
	t.InChannelBlock = make(chan *comm.Order, INCHANNEL_BUFF_SIZE)
	t.InChannel = newInPutPool()
	t.MultiChanOut = chs.NewMultiChans_Out(t.multiChanOutProc)
	t.CancelChannel = make(chan *comm.Order, CANCELCHANNEL_BUFF_SIZE)

	///init Lock
	t.askPoolRWMutex = comm.NewDebugLock("Init ASK")
	t.bidPoolRWMutex = comm.NewDebugLock("Init BID")

	///init DebugInfo
	t.debug = rt.NewDebugInfo()

	fmt.Println(
		"============================[Market ",
		t.Symbol, "-", t.MarketType.String(),
		"]==================================\n",
		"Trade Pool Init Time Log:\n",
		"Test data scale = ", TEST_DATA_SCALE, "\n",
		"Init test data to Pool = ", float64(TimeDot2-TimeDot1)/float64(1*time.Second), "s;\n",
		"History orders scale = ", hsize, "\n",
		"Init history orders = ", float64(TimeDot3-TimeDot2)/float64(1*time.Second), "s;\n",
		"SortPool(askPool) = ", float64(TimeDot4-TimeDot3)/float64(1*time.Second), "s;\n",
		"SortPool(bidPool) = ", float64(TimeDot5-TimeDot4)/float64(1*time.Second), "s;\n")

	///init tradeControl
	t.tradeControl = TradeControl_Work

	t.doctor.SetProgress(doctor.Progress_CompleteInit)
	return t
}

func (t *TradePool) run() {
	fmt.Println("=====================================================================")
	fmt.Printf("Run Match Engine %s...\n", t.Symbol)

	//	t.test(tp)
	go t.match()
	go t.cancel()
	go t.inputBlock()

	t.InChannel.Start(t)

	fmt.Printf("Start Match Engine %s complete.\n", t.Symbol)
	t.doctor.SetProgress(doctor.Progress_Working)
}

func (t *TradePool) setup() *TradePool {
	t.init()
	t.run()
	return t
}

type IEnOrder interface {
	Add(order_ comm.Order) bool
}

// 二分查找
func binarySearch(m []*list.Element, newPrice float64) (target int, res bool) {
	if len(m) == 0 {
		return -1, true
	}

	var left, right, mid int = 0, len(m) - 1, 0
	mid = 0
	for left <= right {
		mid = (left + right) / 2
		if m[mid].Value.(comm.Order).Price == newPrice {
			return mid, true
		}
		if newPrice < m[mid].Value.(comm.Order).Price {
			if left == right {
				return mid - 1, true
			} else {
				right = mid - 1
				target = right
			}
		} else if newPrice > m[mid].Value.(comm.Order).Price {
			if left == right {
				return mid, true
			} else {
				left = mid + 1
				target = left
			}
		}
	}

	return target, true
}

func binarySearchPriceAsc(m []*list.Element, newPrice float64) (target int, res bool) {
	if len(m) == 0 {
		return -1, true
	}

	target = sort.Search(len(m), func(i int) bool { return m[i].Value.(comm.Order).Price > newPrice })

	return target - 1, true
}

func binarySearchPriceDes(m []*list.Element, newPrice float64) (target int, res bool) {
	if len(m) == 0 {
		return -1, true
	}

	target = sort.Search(len(m), func(i int) bool { return m[i].Value.(comm.Order).Price < newPrice })

	return target - 1, true
}

func binarySearchOrderIDAsc(m []*list.Element, id int64) (target int, res bool) {
	if len(m) == 0 {
		return -1, true
	}

	target = sort.Search(len(m), func(i int) bool { return m[i].Value.(comm.Order).ID > id })

	return target - 1, true
}

func binarySearchOrderID(m []*list.Element, id int64) (target int, res bool) {
	if len(m) == 0 {
		return -1, false
	}

	var left, right, mid int = 0, len(m) - 1, 0
	mid = 0
	for left <= right {
		mid = (left + right) / 2
		if m[mid].Value.(comm.Order).ID == id {
			return mid, true
		}
		if id < m[mid].Value.(comm.Order).ID {
			right = mid - 1
		} else if id > m[mid].Value.(comm.Order).ID {
			left = mid + 1
		}
	}

	return -1, false
}

func binarySearchBidOrderPrice(m []*list.Element, price float64) (target int, res bool) {
	if len(m) == 0 {
		return -1, false
	}

	var left, right, mid int = 0, len(m) - 1, 0
	mid = 0
	for left <= right {
		mid = (left + right) / 2
		Price := m[mid].Value.(comm.Order).Price
		if Price == price {
			return mid, true
		} else if price < Price {
			left = mid + 1
		} else if price > Price {
			right = mid - 1
		}
	}

	return -1, false
}

func binarySearchAskOrderPrice(m []*list.Element, price float64) (target int, res bool) {
	if len(m) == 0 {
		return -1, false
	}

	var left, right, mid int = 0, len(m) - 1, 0
	mid = 0
	for left <= right {
		mid = (left + right) / 2
		Price := m[mid].Value.(comm.Order).Price
		if Price == price {
			return mid, true
		} else if price < Price {
			right = mid - 1
		} else if price > Price {
			left = mid + 1
		}
	}

	return -1, false
}

func binarySearchIDAsc(m []*list.Element, id int64) (target int, res bool) {

	target = sort.Search(len(m), func(i int) bool { return m[i].Value.(int64) > id })

	return target - 1, true
}

//// Sort By Time By timeAdjust
func timeAdjust(target int, s []*list.Element, elem *comm.Order) int {
	var obj *list.Element

	/// When target != 0
	for targetT := target; targetT >= 0 && targetT < len(s); targetT-- {
		obj = s[targetT]
		if obj.Value.(comm.Order).Price == elem.Price {
			if obj.Value.(comm.Order).Timestamp <= elem.Timestamp {
				break
			} else {
				target--
				///debug:
				comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_DEBUG, "Order(Symbol:%s, ID: %d) TimeAdjust- act: set target to %d\n", elem.Symbol, elem.ID, target)
				continue
			}
		} else {
			break
		}
	}

	/// adjust down:
	/// When target != 0
	for targetT := target + 1; targetT >= 0 && targetT < len(s); targetT++ {
		obj = s[targetT]
		if obj.Value.(comm.Order).Price == elem.Price {
			if obj.Value.(comm.Order).Timestamp >= elem.Timestamp {
				break
			} else {
				target++
				///debug:
				comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_DEBUG, "Order(Symbol:%s, ID: %d) TimeAdjust+ act: set target to %d\n", elem.Symbol, elem.ID, target)
				continue
			}
		} else {
			break
		}
	}

	return target
}

func (t *TradePool) addToAskPool(order_ *comm.Order) (*list.Element, bool) {

	if order_.AorB != comm.TradeType_ASK {
		return nil, false
	}

	var elem *list.Element = nil
	target, success := binarySearchPriceAsc(t.askPoolSlice, order_.Price)
	if success {
		/// Sort by time
		target = timeAdjust(target, t.askPoolSlice, order_)

		if target == -1 {
			elem = t.askPool.PushFront(*order_)

			askPoolSlice := []*list.Element{}
			askPoolSlice = append(askPoolSlice, t.askPool.Front())
			askPoolSlice = append(askPoolSlice, t.askPoolSlice[:]...)
			t.askPoolSlice = askPoolSlice

		} else if target == len(t.askPoolSlice) {
			elem = t.askPool.PushBack(*order_)

			t.askPoolSlice = append(t.askPoolSlice, t.askPool.Back())
		} else {
			elem = t.askPool.InsertAfter(*order_, t.askPoolSlice[target])

			var askPoolSlice []*list.Element
			askPoolSlice = append(askPoolSlice, t.askPoolSlice[:target+1]...)
			askPoolSlice = append(askPoolSlice, elem)
			askPoolSlice = append(askPoolSlice, t.askPoolSlice[target+1:]...)
			t.askPoolSlice = askPoolSlice
		}
	} else {
		return nil, false
	}

	///debug info
	t.debug.DebugInfo_AskEnOrderNormalAdd()

	return elem, true
}

func (t *TradePool) addToBidPool(order_ *comm.Order) (*list.Element, bool) {

	if order_.AorB != comm.TradeType_BID {
		return nil, false
	}

	var elem *list.Element = nil
	target, success := binarySearchPriceDes(t.bidPoolSlice, order_.Price)
	if success {
		/// Sort by time
		target = timeAdjust(target, t.bidPoolSlice, order_)

		if target == -1 {
			elem = t.bidPool.PushFront(*order_)

			bidPoolSlice := []*list.Element{}
			bidPoolSlice = append(bidPoolSlice, t.bidPool.Front())
			bidPoolSlice = append(bidPoolSlice, t.bidPoolSlice[:]...)
			t.bidPoolSlice = bidPoolSlice
		} else if target == len(t.bidPoolSlice) {
			elem = t.bidPool.PushBack(*order_)

			t.bidPoolSlice = append(t.bidPoolSlice, t.bidPool.Back())
		} else {
			elem = t.bidPool.InsertAfter(*order_, t.bidPoolSlice[target])

			bidPoolSlice := []*list.Element{}
			bidPoolSlice = append(bidPoolSlice, t.bidPoolSlice[:target+1]...)
			bidPoolSlice = append(bidPoolSlice, elem)
			bidPoolSlice = append(bidPoolSlice, t.bidPoolSlice[target+1:]...)
			t.bidPoolSlice = bidPoolSlice
		}
	} else {
		return nil, false
	}

	///debug info
	t.debug.DebugInfo_BidEnOrderNormalAdd()

	return elem, true
}

func (t *TradePool) addToIdSlice(elem *list.Element, slice []*list.Element) ([]*list.Element, bool) {

	targetID, suc := binarySearchOrderIDAsc(slice, elem.Value.(comm.Order).ID)

	if suc {
		s := []*list.Element{}
		if targetID == -1 {
			s = append(s, elem)
			s = append(s, slice[:]...)
		} else if targetID == len(slice) {
			s = slice
			s = append(s, elem)
		} else {
			s = append(s, slice[:targetID+1]...)
			s = append(s, elem)
			s = append(s, slice[targetID+1:]...)
		}
		return s, true
	}

	return nil, false
}

func (t *TradePool) removeFromIdSlice(elem *list.Element, slice []*list.Element) ([]*list.Element, bool) {
	targetID, suc := binarySearchOrderID(slice, elem.Value.(comm.Order).ID)
	if suc {
		s := []*list.Element{}
		if targetID == 0 {
			s = slice[1:]
		} else {
			s = append(s, slice[:targetID]...)
			s = append(s, slice[targetID+1:]...)
		}
		return s, true
	}

	return nil, false
}

func (t *TradePool) insertBefore(order_ *comm.Order, target int) *list.Element {
	var elem *list.Element = nil
	if order_.AorB == comm.TradeType_ASK {
		if target == 0 {
			elem = t.askPool.PushFront(*order_)

			askPoolSlice := []*list.Element{}
			askPoolSlice = append(askPoolSlice, t.askPool.Front())
			askPoolSlice = append(askPoolSlice, t.askPoolSlice[:]...)
			t.askPoolSlice = askPoolSlice
		} else if target == len(t.askPoolSlice) {
			elem = t.askPool.PushBack(*order_)

			t.askPoolSlice = append(t.askPoolSlice, t.askPool.Back())
		} else {
			elem = t.askPool.InsertBefore(*order_, t.askPoolSlice[target])

			askPoolSlice := []*list.Element{}
			askPoolSlice = append(askPoolSlice, t.askPoolSlice[:target]...)
			askPoolSlice = append(askPoolSlice, elem)
			askPoolSlice = append(askPoolSlice, t.askPoolSlice[target:]...)
			t.askPoolSlice = askPoolSlice
		}
	} else if order_.AorB == comm.TradeType_BID {
		if target == 0 {
			elem = t.bidPool.PushFront(*order_)

			bidPoolSlice := []*list.Element{}
			bidPoolSlice = append(bidPoolSlice, t.bidPool.Front())
			bidPoolSlice = append(bidPoolSlice, t.bidPoolSlice[:]...)
			t.bidPoolSlice = bidPoolSlice
		} else if target == len(t.bidPoolSlice) {
			elem = t.bidPool.PushBack(*order_)

			t.bidPoolSlice = append(t.bidPoolSlice, t.bidPool.Back())
		} else {

			elem = t.bidPool.InsertBefore(*order_, t.bidPoolSlice[target])

			bidPoolSlice := []*list.Element{}
			bidPoolSlice = append(bidPoolSlice, t.bidPoolSlice[:target]...)
			bidPoolSlice = append(bidPoolSlice, elem)
			bidPoolSlice = append(bidPoolSlice, t.bidPoolSlice[target:]...)
			t.bidPoolSlice = bidPoolSlice
		}
	}

	return elem
}

///quick add is used to insert order while partly trade, the precondition is the pool is sorted
func (t *TradePool) addQuickToAskPool(order_ *comm.Order) *list.Element {

	var target int = 0
	var elem, e *list.Element
	if order_.AorB != comm.TradeType_ASK {
		return nil
	}

	for _, e = range t.askPoolSlice {
		if order_.Price < e.Value.(comm.Order).Price {
			elem = t.insertBefore(order_, target)
			break
		} else if order_.Price == e.Value.(comm.Order).Price {
			if order_.Timestamp <= e.Value.(comm.Order).Timestamp {
				elem = t.insertBefore(order_, target)
				break
			} else {
				target++
				continue
			}
		} else {
			target++
			continue
		}
	}

	if elem == nil {
		if target == 0 || target >= len(t.askPoolSlice) {
			elem = t.insertBefore(order_, len(t.askPoolSlice))
		} else {
			panic(fmt.Errorf("addQuickToAskPool met Logic error!"))
		}
	}

	///debug info
	t.debug.DebugInfo_AskEnOrderQuickAdd()

	return elem
}

///quick add is used to insert order while partly trade, the precondition is the pool is sorted
func (t *TradePool) addQuickToBidPool(order_ *comm.Order) *list.Element {

	var target int = 0
	var elem, e *list.Element
	if order_.AorB != comm.TradeType_BID {
		return nil
	}

	for _, e = range t.bidPoolSlice {
		if order_.Price > e.Value.(comm.Order).Price {
			elem = t.insertBefore(order_, target)
			break
		} else if order_.Price == e.Value.(comm.Order).Price {
			if order_.Timestamp <= e.Value.(comm.Order).Timestamp {
				elem = t.insertBefore(order_, target)
				break
			} else {
				target++
				continue
			}
		} else {
			target++
			continue
		}
	}

	if elem == nil {
		if target == 0 || target >= len(t.bidPoolSlice) {
			elem = t.insertBefore(order_, len(t.bidPoolSlice))
		} else {
			panic(fmt.Errorf("addQuickToBidPool met Logic error!"))
		}
	}

	///debug info
	t.debug.DebugInfo_BidEnOrderQuickAdd()

	return elem
}

func (t *TradePool) add(order_ *comm.Order) (*list.Element, bool) {
	comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s TradePool Add Order ID(%d), Time(%d)\n", t.Symbol, t.MarketType.String(), order_.ID, order_.Timestamp)
	if order_.Symbol != t.Symbol {
		fmt.Printf("%s Add illegal order with symbol(%s) to %s-%s Match Engine", t.Symbol, order_.Symbol, t.Symbol, t.MarketType.String())
		return nil, false
	}

	if order_.AorB == comm.TradeType_ASK {
		t.askPoolRWMutex.Lock("Add ASK")
		defer t.askPoolRWMutex.Unlock("Add ASK")

		/// debug:
		TimeDot1 := time.Now().UnixNano()
		elem, suc := t.addToAskPool(order_)
		if suc {
			t.askPoolIDSlice, _ = t.addToIdSlice(elem, t.askPoolIDSlice)
		} else {
			panic(fmt.Errorf("%s Add addTo askPoolIDSlice met error!", t.Symbol))
		}

		/// debug:
		TimeDot2 := time.Now().UnixNano()
		comm.DebugPrintln(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_DEBUG, "-",
			t.Symbol, t.MarketType.String(),
			" TradePool Add ask order==",
			"id: ", order_.ID,
			"; user: ", order_.Who,
			"; type_: ", order_.AorB.String(),
			"; time: ", order_.Timestamp,
			"; price: ", order_.Price,
			"; volume: ", order_.Volume,
			"****USE_TIME: ", float64(TimeDot2-TimeDot1)/float64(1*time.Second),
			"\n",
		)
		return elem, true
	} else if order_.AorB == comm.TradeType_BID {
		t.bidPoolRWMutex.Lock("Add BID")
		defer t.bidPoolRWMutex.Unlock("Add BID")

		/// debug:
		TimeDot1 := time.Now().UnixNano()
		elem, suc := t.addToBidPool(order_)
		if suc {
			t.bidPoolIDSlice, _ = t.addToIdSlice(elem, t.bidPoolIDSlice)
		} else {
			panic("Add addTo bidPoolIDSlice met error!")
		}

		/// debug:
		TimeDot2 := time.Now().UnixNano()
		comm.DebugPrintln(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_DEBUG, "-",
			t.Symbol, t.MarketType.String(),
			" TradePool Add bid order==",
			"id: ", order_.ID,
			"; user: ", order_.Who,
			"; type_: ", order_.AorB.String(),
			"; time: ", order_.Timestamp,
			"; price: ", order_.Price,
			"; volume: ", order_.Volume,
			"****USE_TIME: ", float64(TimeDot2-TimeDot1)/float64(1*time.Second),
			"\n",
		)
		return elem, true
	} else {
		return nil, false
	}
}

///quick add is used to insert order while partly trade, the precondition is the pool is sorted
func (t *TradePool) addQuick(order_ *comm.Order) (*list.Element, bool) {
	comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s TradePool addQuick Order ID(%d), Time(%d)\n", t.Symbol, order_.ID, order_.Timestamp)
	if order_ == nil {
		return nil, false
	}

	var elem *list.Element = nil
	if order_.AorB == comm.TradeType_ASK {
		t.askPoolRWMutex.Lock("addQuick ASK")
		defer t.askPoolRWMutex.Unlock("addQuick ASK")

		/// debug:
		TimeDot1 := time.Now().UnixNano()
		elem = t.addQuickToAskPool(order_)
		if elem != nil {
			t.askPoolIDSlice, _ = t.addToIdSlice(elem, t.askPoolIDSlice)
		} else {
			return nil, false
		}

		/// debug:
		TimeDot2 := time.Now().UnixNano()
		comm.DebugPrintln(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_DEBUG, "-",
			t.Symbol, t.MarketType.String(),
			" TradePool Quick add ask order==",
			"id: ", order_.ID,
			"user: ", order_.Who,
			"; type_: ", order_.AorB.String(),
			"; time: ", order_.Timestamp,
			"; price: ", order_.Price,
			"; volume: ", order_.Volume,
			"****USE_TIME: ", float64(TimeDot2-TimeDot1)/float64(1*time.Second),
			"\n",
		)
	} else if order_.AorB == comm.TradeType_BID {
		t.bidPoolRWMutex.Lock("addQuick BID")
		defer t.bidPoolRWMutex.Unlock("addQuick BID")

		/// debug:
		TimeDot1 := time.Now().UnixNano()
		elem = t.addQuickToBidPool(order_)
		if elem != nil {
			t.bidPoolIDSlice, _ = t.addToIdSlice(elem, t.bidPoolIDSlice)
		} else {
			return nil, false
		}

		/// debug:
		TimeDot2 := time.Now().UnixNano()
		comm.DebugPrintln(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_DEBUG, "-",
			t.Symbol, t.MarketType.String(),
			" TradePool Quick add bid order==",
			"id: ", order_.ID,
			"user: ", order_.Who,
			"; type_: ", order_.AorB.String(),
			"; time: ", order_.Timestamp,
			"; price: ", order_.Price,
			"; volume: ", order_.Volume,
			"****USE_TIME: ", float64(TimeDot2-TimeDot1)/float64(1*time.Second),
			"\n",
		)
	}

	return elem, true
}

type ITrade interface {
	GetTop(type_ string) (order_ comm.Order, res bool)
	PopTop(type_ string) (order_ []comm.Order, num uint)
	Trade()
}

func (t *TradePool) getTop(type_ comm.TradeType) (order_ *comm.Order, res bool) {

	if type_ == comm.TradeType_BID {
		t.bidPoolRWMutex.RLock("GetTop BID")
		defer t.bidPoolRWMutex.RUnlock("GetTop BID")

		if t.bidPool.Len() <= 0 || len(t.bidPoolSlice) <= 0 {
			return nil, false
		}

		/// do business:
		order := t.bidPool.Front().Value.(comm.Order)
		return &order, true
	} else if type_ == comm.TradeType_ASK {
		t.askPoolRWMutex.RLock("GetTop ASK")
		defer t.askPoolRWMutex.RUnlock("GetTop ASK")

		if t.askPool.Len() <= 0 || len(t.askPoolSlice) <= 0 {
			return nil, false
		}

		/// do business:
		order := t.askPool.Front().Value.(comm.Order)
		return &order, true
	} else {
		return nil, false
	}
}

func (t *TradePool) popTops(type_ comm.TradeType) (order_ []comm.Order, num uint) {
	if type_ == comm.TradeType_BID {
		t.bidPoolRWMutex.Lock("PopTops BID")
		defer t.bidPoolRWMutex.Unlock("PopTops BID")

		if t.bidPool.Len() <= 0 || len(t.bidPoolSlice) <= 0 {
			return nil, 0
		}

		/// do business:
		var tmp *list.Element
		var out []comm.Order
		top := t.bidPool.Front()
		price := top.Value.(comm.Order).Price
		for num = 0; top.Value.(comm.Order).Price == price; {
			out = append(out, top.Value.(comm.Order))
			tmp = top
			top = top.Next()

			///remove from index and pool
			t.bidPoolSlice = t.bidPoolSlice[1:]
			t.bidPoolIDSlice, _ = t.removeFromIdSlice(tmp, t.bidPoolIDSlice)
			if t.bidPoolIDSlice == nil {
				panic("PopTops met bidPoolIDSlice nil!")
			}
			t.bidPool.Remove(tmp)

			num++
		}

		return out, num
	} else if type_ == comm.TradeType_ASK {
		t.askPoolRWMutex.Lock("PopTops ASK")
		defer t.askPoolRWMutex.Unlock("PopTops ASK")

		if t.askPool.Len() <= 0 || len(t.askPoolSlice) <= 0 {
			return nil, 0
		}

		/// do business:
		var tmp *list.Element
		var out []comm.Order
		top := t.askPool.Front()
		price := top.Value.(comm.Order).Price
		for num = 0; top.Value.(comm.Order).Price == price; {
			out = append(out, top.Value.(comm.Order))
			tmp = top
			top = top.Next()

			///remove from index and pool
			t.askPoolSlice = t.askPoolSlice[1:]
			t.askPoolIDSlice, _ = t.removeFromIdSlice(tmp, t.askPoolIDSlice)
			if t.askPoolIDSlice == nil {
				panic("PopTops met askPoolIDSlice nil!")
			}
			t.askPool.Remove(tmp)

			num++
		}

		return out, num
	} else {
		return nil, 0
	}
}

func (t *TradePool) popTop(type_ comm.TradeType) (order_ *comm.Order, res bool) {
	if type_ == comm.TradeType_BID {
		t.bidPoolRWMutex.Lock("PopTop BID")
		defer t.bidPoolRWMutex.Unlock("PopTop BID")

		if t.bidPool.Len() <= 0 || len(t.bidPoolSlice) <= 0 {
			/// panic("GetTop t.bidPool.Len() <= 0 or len(bidPoolSlice) <= 0!")
			return nil, false
		}

		/// do business:
		top := t.bidPool.Front()
		order := top.Value.(comm.Order)

		///remove from index and pool
		t.bidPoolSlice = t.bidPoolSlice[1:]
		t.bidPoolIDSlice, _ = t.removeFromIdSlice(top, t.bidPoolIDSlice)
		if t.bidPoolIDSlice == nil {
			panic("PopTop met bidPoolIDSlice nil!")
		}
		t.bidPool.Remove(top)

		return &order, true
	} else if type_ == comm.TradeType_ASK {
		t.askPoolRWMutex.Lock("PopTop ASK")
		defer t.askPoolRWMutex.Unlock("PopTop ASK")

		if t.askPool.Len() <= 0 || len(t.askPoolSlice) <= 0 {
			/// panic("GetTop t.askPool.Len() <= 0 or len(askPoolSlice) <= 0!")
			return nil, false
		}

		/// do business:
		top := t.askPool.Front()
		order := top.Value.(comm.Order)

		///remove from index and pool
		t.askPoolSlice = t.askPoolSlice[1:]
		t.askPoolIDSlice, _ = t.removeFromIdSlice(top, t.askPoolIDSlice)
		if t.askPoolIDSlice == nil {
			panic("PopTop met askPoolIDSlice nil!")
		}
		t.askPool.Remove(top)

		return &order, true
	} else {
		return nil, false
	}
}

func removeFromSlice(target int, slice []*list.Element) []*list.Element {
	s := []*list.Element{}
	if target == 0 {
		s = slice[1:]
	} else {
		s = append(s, slice[:target]...)
		s = append(s, slice[target+1:]...)
	}

	return s
}

/// should be protect by PoolRWMutex.RLock()
func (t *TradePool) orderCheck(order *comm.Order) (target int, res bool) {
	if order == nil {
		return -1, false
	}

	/// find Order by ID
	var (
		targetID int           = -1
		suc      bool          = false
		elem     *list.Element = nil
	)
	if order.AorB == comm.TradeType_BID {
		targetID, suc = binarySearchOrderID(t.bidPoolIDSlice, order.ID)
		if suc {
			elem = t.bidPoolIDSlice[targetID]
		}
	} else if order.AorB == comm.TradeType_ASK {
		targetID, suc = binarySearchOrderID(t.askPoolIDSlice, order.ID)
		if suc {
			elem = t.askPoolIDSlice[targetID]
		}
	} else {
		fmt.Println("Illegal order type!\n")
		return -1, false
	}

	if suc {
		if elem.Value.(comm.Order).Who == order.Who &&
			elem.Value.(comm.Order).Price == order.Price &&
			///elem.Value.(comm.Order).Timestamp == order.Timestamp {
			elem.Value.(comm.Order).EnOrderPrice == order.EnOrderPrice &&
			elem.Value.(comm.Order).TotalVolume == order.TotalVolume {
			return targetID, true
		} else {
			return targetID, false
		}
	} else {
		comm.DebugPrintln(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "OrderCheck in ME trade pool fail! Not a trading order.\n")
		return -1, false
	}
}

/// should be protect by PoolRWMutex.Lock()
func (t *TradePool) rmBidOrderByTarget(target int) bool {
	if target < 0 || target >= len(t.bidPoolIDSlice) {
		fmt.Println("RmBidOrderByTarget input target(", target, ") error!\n")
		return false
	}

	obj := t.bidPoolIDSlice[target]

	/// remove from bidPoolIDSlice
	t.bidPoolIDSlice = removeFromSlice(target, t.bidPoolIDSlice)

	/// remove from bidPoolSlice
	targetP, sucP := binarySearchBidOrderPrice(t.bidPoolSlice, obj.Value.(comm.Order).Price)
	if sucP {
		var objP *list.Element
		for targetP >= 0 && targetP < len(t.bidPoolSlice) {
			objP = t.bidPoolSlice[targetP]
			var operDirection int = 0
			if objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price {
				if objP.Value.(comm.Order).Timestamp < obj.Value.(comm.Order).Timestamp {
					targetP++
					operDirection = 1
					///debug:
					if objP.Value.(comm.Order).Price != obj.Value.(comm.Order).Price {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmBidOrderByTarget+ occur exception when at bidPoolSlice search Target(ID:%d,Price:%f,Time:%d), SearchedObj(ID:%d,Price:%f,Time:%d)\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}
					///debug:
					if targetP < len(t.askPoolSlice) && t.bidPoolSlice[targetP].Value.(comm.Order).Timestamp > obj.Value.(comm.Order).Timestamp {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmBidOrderByTarget+ occur Pingpong exception: Target(ID:%d,Price:%f,Time:%d), TargetP(ID:%d,Price:%f,Time:%d), TargetP=%d\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
							targetP,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}

					continue
				} else if objP.Value.(comm.Order).Timestamp > obj.Value.(comm.Order).Timestamp {
					targetP--
					operDirection = -1
					///debug:
					if objP.Value.(comm.Order).Price != obj.Value.(comm.Order).Price {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmBidOrderByTarget- occur exception when at bidPoolSlice search Target(ID:%d,Price:%f,Time:%d), SearchedObj(ID:%d,Price:%f,Time:%d)\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}
					///debug:
					if targetP >= 0 && t.bidPoolSlice[targetP].Value.(comm.Order).Timestamp < obj.Value.(comm.Order).Timestamp {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmBidOrderByTarget- occur Pingpong exception: Target(ID:%d,Price:%f,Time:%d), TargetP(ID:%d,Price:%f,Time:%d), TargetP=%d\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
							targetP,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}

					continue
				} else {
					if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
						break
					} else {
						bFind := false
						if operDirection != 0 {
							for targetP = targetP + operDirection; targetP >= 0 && targetP < len(t.bidPoolSlice); targetP = targetP + operDirection {
								objP = t.bidPoolSlice[targetP]
								if (objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price) && (objP.Value.(comm.Order).Timestamp == obj.Value.(comm.Order).Timestamp) {
									if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
										bFind = true
										break
									}
								} else {
									panic(fmt.Errorf("Core Algorithm Bug!"))
								}
							}
						} else {
							targetPP := targetP
							operDirection = 1
							for targetP = targetP + operDirection; targetP >= 0 && targetP < len(t.bidPoolSlice); targetP = targetP + operDirection {
								objP = t.bidPoolSlice[targetP]
								if (objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price) && (objP.Value.(comm.Order).Timestamp == obj.Value.(comm.Order).Timestamp) {
									if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
										bFind = true
										break
									}
								} else {
									break
								}
							}
							if bFind {
								break
							}
							targetP = targetPP
							operDirection = -1
							for targetP = targetP + operDirection; targetP >= 0 && targetP < len(t.bidPoolSlice); targetP = targetP + operDirection {
								objP = t.bidPoolSlice[targetP]
								if (objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price) && (objP.Value.(comm.Order).Timestamp == obj.Value.(comm.Order).Timestamp) {
									if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
										bFind = true
										break
									}
								} else {
									panic(fmt.Errorf("Core Algorithm Bug!"))
								}
							}
						}
						if bFind {
							break
						} else {
							panic("RmOrder bsearch order from poolslide fail, can not find target order in bidPoolSlice, Core Algorithm Bug!")
						}
					}
				}
			} else {
				panic(fmt.Errorf("Core Algorithm Bug!"))
			}
		}
		if targetP >= 0 && targetP < len(t.bidPoolSlice) {
			t.bidPoolSlice = removeFromSlice(targetP, t.bidPoolSlice)
		} else {
			t.dump()
			panic(fmt.Errorf("%s-%s RmBidOrderByTarget order id(%d) not found at bidPoolSlice", t.Symbol, t.MarketType.String(), obj.Value.(comm.Order).ID))
		}
	} else {
		panic(fmt.Errorf("%s-%s RmBidOrderByTarget remove order from bidPoolSlice fail! data in bidPoolSlice and bidPoolIDSlice not sync!", t.Symbol, t.MarketType.String()))
	}

	/// remove from pool
	t.bidPool.Remove(obj)

	return true
}

func (t *TradePool) rmBidOrderByID(id int64) bool {
	t.bidPoolRWMutex.Lock("RmBidOrderByID BID")
	defer t.bidPoolRWMutex.Unlock("RmBidOrderByID BID")

	if t.bidPool.Len() <= 0 || len(t.bidPoolSlice) <= 0 {
		return false
	}

	/// find Order by ID
	targetID, suc := binarySearchOrderID(t.bidPoolIDSlice, id)

	if suc {
		return t.rmBidOrderByTarget(targetID)
	} else {
		fmt.Printf("Order(ID=%d) not found at %s-%s bidPoolIDSlice, it perhaps not exist or had been processed!\n", id, t.Symbol, t.MarketType.String())
		return false
	}
}

func (t *TradePool) rmAskOrderByTarget(target int) bool {
	if target < 0 || target >= len(t.askPoolIDSlice) {
		fmt.Printf("%s-%s RmAskOrderByTarget input target(%d) error!\n", t.Symbol, t.MarketType.String(), target)
		return false
	}

	obj := t.askPoolIDSlice[target]

	/// remove from askPoolIDSlice
	t.askPoolIDSlice = removeFromSlice(target, t.askPoolIDSlice)

	/// remove from askPoolSlice
	targetP, sucP := binarySearchAskOrderPrice(t.askPoolSlice, obj.Value.(comm.Order).Price)
	if sucP {
		var objP *list.Element
		for targetP >= 0 && targetP < len(t.askPoolSlice) {
			objP = t.askPoolSlice[targetP]
			var operDirection int = 0
			if objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price {
				if objP.Value.(comm.Order).Timestamp < obj.Value.(comm.Order).Timestamp {
					targetP++
					operDirection = 1
					///debug:
					if objP.Value.(comm.Order).Price != obj.Value.(comm.Order).Price {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmAskOrderByTarget+ occur exception when at askPoolSlice search Target(ID:%d,Price:%f,Time:%d), SearchedObj(ID:%d,Price:%f,Time:%d)\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}
					///debug:
					if targetP < len(t.askPoolSlice) && t.askPoolSlice[targetP].Value.(comm.Order).Timestamp > obj.Value.(comm.Order).Timestamp {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmAskOrderByTarget+ occur Pingpong exception: Target(ID:%d,Price:%f,Time:%d), TargetP(ID:%d,Price:%f,Time:%d), TargetP=%d\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
							targetP,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}

					continue
				} else if objP.Value.(comm.Order).Timestamp > obj.Value.(comm.Order).Timestamp {
					targetP--
					operDirection = -1
					///debug:
					if objP.Value.(comm.Order).Price != obj.Value.(comm.Order).Price {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmAskOrderByTarget- occur exception when at askPoolSlice search Target(ID:%d,Price:%f,Time:%d), SearchedObj(ID:%d,Price:%f,Time:%d)\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}
					///debug:
					if targetP >= 0 && t.askPoolSlice[targetP].Value.(comm.Order).Timestamp < obj.Value.(comm.Order).Timestamp {
						comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_FATAL, "%s-%s RmAskOrderByTarget- occur Pingpong exception: Target(ID:%d,Price:%f,Time:%d), TargetP(ID:%d,Price:%f,Time:%d), TargetP=%d\n",
							t.Symbol, t.MarketType.String(),
							obj.Value.(comm.Order).ID,
							obj.Value.(comm.Order).Price,
							obj.Value.(comm.Order).Timestamp,
							objP.Value.(comm.Order).ID,
							objP.Value.(comm.Order).Price,
							objP.Value.(comm.Order).Timestamp,
							targetP,
						)
						t.dump()
						panic(fmt.Errorf("Core Algorithm Bug!"))
					}

					continue
				} else {
					if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
						break
					} else {
						bFind := false
						if operDirection != 0 {
							for targetP = targetP + operDirection; targetP >= 0 && targetP < len(t.askPoolSlice); targetP = targetP + operDirection {
								objP = t.askPoolSlice[targetP]
								if (objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price) && (objP.Value.(comm.Order).Timestamp == obj.Value.(comm.Order).Timestamp) {
									if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
										bFind = true
										break
									}
								} else {
									panic(fmt.Errorf("Core Algorithm Bug!"))
								}
							}
						} else {
							targetPP := targetP
							operDirection = 1
							for targetP = targetP + operDirection; targetP >= 0 && targetP < len(t.askPoolSlice); targetP = targetP + operDirection {
								objP = t.askPoolSlice[targetP]
								if (objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price) && (objP.Value.(comm.Order).Timestamp == obj.Value.(comm.Order).Timestamp) {
									if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
										bFind = true
										break
									}
								} else {
									break
								}
							}
							if bFind {
								break
							}
							targetP = targetPP
							operDirection = -1
							for targetP = targetP + operDirection; targetP >= 0 && targetP < len(t.askPoolSlice); targetP = targetP + operDirection {
								objP = t.askPoolSlice[targetP]
								if (objP.Value.(comm.Order).Price == obj.Value.(comm.Order).Price) && (objP.Value.(comm.Order).Timestamp == obj.Value.(comm.Order).Timestamp) {
									if objP.Value.(comm.Order).ID == obj.Value.(comm.Order).ID {
										bFind = true
										break
									}
								} else {
									panic(fmt.Errorf("Core Algorithm Bug!"))
								}
							}
						}
						if bFind {
							break
						} else {
							panic("RmOrder bsearch order from poolslide fail, can not find target order in askPoolSlice, Core Algorithm Bug!")
						}
					}
				}
			} else {
				panic(fmt.Errorf("Core Algorithm Bug!"))
			}
		}
		if targetP >= 0 && targetP < len(t.askPoolSlice) {
			t.askPoolSlice = removeFromSlice(targetP, t.askPoolSlice)
		} else {
			t.dump()
			panic(fmt.Errorf("RmAskOrderByTarget order id(%d) not found at askPoolSlice", obj.Value.(comm.Order).ID))
		}
	} else {
		panic("RmAskOrderByTarget remove order from askPoolSlice fail! data in askPoolSlice and askPoolIDSlice not sync!")
	}

	/// remove from pool
	t.askPool.Remove(obj)

	return true
}

func (t *TradePool) rmAskOrderByID(id int64) bool {
	t.askPoolRWMutex.Lock("RmAskOrderByID  ASK")
	defer t.askPoolRWMutex.Unlock("RmAskOrderByID  ASK")

	if t.askPool.Len() <= 0 || len(t.askPoolSlice) <= 0 {
		return false
	}

	/// find Order by ID
	targetID, suc := binarySearchOrderID(t.askPoolIDSlice, id)
	if suc {
		return t.rmAskOrderByTarget(targetID)
	} else {
		fmt.Println("Order(ID=%d) not found at %s-%s askPoolSlice, it perhaps not exist or had been processed!\n", id, t.Symbol, t.MarketType.String())
		return false
	}
}

func orderValidatable(o *comm.Order) bool {
	if o.Price <= 0 || o.Volume <= 0 {
		return false
	}

	return true
}

func priceValidatable(v float64) {
	if v <= 0 {
		panic("priceValidatable fail!")
	}
}
func volumeValidatable(v float64) {
	if v <= 0 {
		panic("volumeValidatable fail!")
	}
}

func (t *TradePool) trade() {

	orderAsk, sucAsk := t.getTop(comm.TradeType_ASK)
	orderBid, sucBid := t.getTop(comm.TradeType_BID)
	var bidStatus, askStatus comm.TradeStatus

	if sucAsk && sucBid {
		if orderValidatable(orderAsk) && orderValidatable(orderBid) {
			if orderBid.Price >= orderAsk.Price {
				TimeDot1 := time.Now().UnixNano()

				orderAsk, sucAsk = t.popTop(comm.TradeType_ASK)
				orderBid, sucBid = t.popTop(comm.TradeType_BID)
				if sucAsk && sucBid {
					/// debug:
					comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, `=======>>>%s-%s Order Matching<<<======
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
						t.debug.DebugInfo_AskTradeOutputAdd()
						t.debug.DebugInfo_BidTradeOutputAdd()
						t.debug.DebugInfo_AskTradeCompleteAdd()
						t.debug.DebugInfo_BidTradeCompleteAdd()
					} else {
						if tradeVolume < orderBid.Volume {
							orderBid.Volume -= tradeVolume
							t.addQuick(orderBid)

							/// updeate order status
							bidStatus = comm.ORDER_PARTIAL_FILLED
							askStatus = comm.ORDER_FILLED
							///debug info
							t.debug.DebugInfo_BidTradeOutputAdd()
							t.debug.DebugInfo_AskTradeCompleteAdd()
						} else {
							orderAsk.Volume -= tradeVolume
							t.addQuick(orderAsk)

							/// updeate order status
							bidStatus = comm.ORDER_FILLED
							askStatus = comm.ORDER_PARTIAL_FILLED
							///debug info
							t.debug.DebugInfo_BidTradeCompleteAdd()
							t.debug.DebugInfo_AskTradeOutputAdd()
						}
					}

					///trade output
					orderTemp := comm.Order{orderBid.ID, orderBid.Who, comm.TradeType_BID, orderBid.Symbol, orderBid.Timestamp, orderBid.EnOrderPrice, tradePrice, tradeVolume, orderBid.TotalVolume, orderBid.Fee, bidStatus, orderBid.IPAddr}
					tradeBid := comm.Trade{orderTemp, tradeBidAmount, time.Now().UnixNano(), tradeVolume * orderBid.Fee}
					orderTemp = comm.Order{orderAsk.ID, orderAsk.Who, comm.TradeType_ASK, orderAsk.Symbol, orderAsk.Timestamp, orderAsk.EnOrderPrice, tradePrice, tradeVolume, orderAsk.TotalVolume, orderAsk.Fee, askStatus, orderAsk.IPAddr}
					tradeAsk := comm.Trade{orderTemp, tradeAskAmount, time.Now().UnixNano(), tradeAmount * orderAsk.Fee}
					///To do: put to channel to send to database
					t.MultiChanOut.InChannel(&chs.OutElem{Trade: &chs.MatchTrade{&tradeBid, &tradeAsk}, CancelOrder: nil, Type_: chs.OUTPOOL_MATCHTRADE, Count: 0})

				} else {
					if !sucAsk && sucBid {
						t.addQuick(orderBid)
						fmt.Println(t.Symbol, "-", t.MarketType.String(), ": Something occur: Trade routine, when bid gettop ok but poptop fail because of the order cancel operate concurrence")
					}
					if !sucBid && sucAsk {
						t.addQuick(orderAsk)
						fmt.Println(t.Symbol, "-", t.MarketType.String(), ": Something occur: Trade routine, when bid gettop ok but poptop fail because of the order cancel operate concurrence")
					}
					if !sucAsk && !sucBid {
						fmt.Println(t.Symbol, "-", t.MarketType.String(), ":Something occur: Trade routine, when bid&ask gettop ok but poptop fail because of the order cancel operate concurrence")
					}
					///panic("Trade process logic error, need process!========================!")
				}

				TimeDot2 := time.Now().UnixNano()
				t.debug.DebugInfo_RecordCorePerform(TimeDot2 - TimeDot1)
				// fmt.Printf("MEXCore.match trade performance(second this round): %.9f \n", float64(TimeDot2-TimeDot1)/float64(1*time.Second))
			} else {
				//fmt.Print("trade no match order to trade... ", "\n")
			}
		} else {
			fmt.Printf("[Trade]:Met Illegal Orders.\n\tBid Order: User(%s), ID(%d), Status(%s), Price(%f), Volume(%f)\n\tAsk Order: User(%s), ID(%d), Status(%s), Price(%f), Volume(%f)\n",
				orderBid.Who, orderBid.ID, orderBid.Status, orderBid.Price, orderBid.Volume, orderAsk.Who, orderAsk.ID, orderAsk.Status, orderAsk.Price, orderAsk.Volume)
			panic("pool data illegel, need process!========================!")
		}
	}
}

func (t *TradePool) match() {
	for {
		///tradeControl
		switch t.tradeControl {
		case TradeControl_Stop:
			fallthrough
		case TradeControl_Pause:
			time.Sleep(1 * time.Second)
			runtime.Gosched()
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s Match Routing on idle...\n", t.Symbol, t.MarketType.String())
			continue
		case TradeControl_Work:

		default:
		}

		/// To match trade and put it out to outchannels
		t.trade()

		/// Record beatheart
		t.doctor.RecordBeatHeart(doctor.RunningType_Match)
		time.Sleep(comm.MECORE_MATCH_DURATION)
		runtime.Gosched()
	}
}

func (t *TradePool) input(cur int) {
	var (
		v  *comm.Order = nil
		ok bool        = false
	)

	for {
		///tradeControl
		switch t.tradeControl {
		case TradeControl_Stop:
			time.Sleep(1 * time.Second)
			runtime.Gosched()
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s Input Routing on idle...\n", t.Symbol, t.MarketType.String())
			continue
		case TradeControl_Work:
		case TradeControl_Pause:
		default:
		}

		///trade process
		//v, ok = <-t.InChannel
		v, ok = t.InChannel.Out(cur)
		if ok {
			///debug===
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, `=======================>>>>>>>>>>%s-%s Order Input[%d]
	Order id:%d, user:%s, type:%s, symbol:%s, time:%d, price:%f, volume:%f, tatalVolume:%f, fee:%f
	Get from Inchannel(cap:%d, len:%d)
`,
				t.Symbol, t.MarketType.String(), cur,
				v.ID, v.Who, v.AorB.String(), v.Symbol, v.Timestamp, v.Price, v.Volume, v.TotalVolume, v.Fee,
				INCHANNEL_BUFF_SIZE*INCHANNEL_POOL_SIZE, t.InChannel.Len(),
			)

			//// Enorder to Match Engine Duration Storage
		reEnorder_:
			err, errCode := use_mysql.MEMySQLInstance().EnOrder(v)
			if err != nil {
				if errCode == use_mysql.ErrorCode_DupPrimateKey {
					comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "EnOrder fail, Retry to do it once more.\n")
					v.ID = time.Now().UnixNano()
					goto reEnorder_
				} else {
					comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "EnOrder fail, errCode = %s.\n", errCode.String())
					continue
				}
			}

			/// Add to trade pool to match
			v.Volume = v.TotalVolume
			t.add(v)
		} else {
			panic(fmt.Errorf("%s-%s Input Routine InChannel exception occur!", t.Symbol, t.MarketType.String()))
		}

		t.doctor.RecordBeatHeart(doctor.RunningType_Enorder)
		runtime.Gosched()
	}
}

func (t *TradePool) inputBlock() {
	var (
		v  *comm.Order = nil
		ok bool        = false
	)

	for {
		///tradeControl
		switch t.tradeControl {
		case TradeControl_Stop:
			time.Sleep(1 * time.Second)
			runtime.Gosched()
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s Input Routing on idle...\n", t.Symbol, t.MarketType.String())
			continue
		case TradeControl_Work:
		case TradeControl_Pause:
		default:
		}

		///trade process
		v, ok = <-t.InChannelBlock
		if ok {
			///debug===
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, `=======================>>>>>>>>>>%s-%s Order Input[Block mode]
	Order id:%d, user:%s, type:%s, symbol:%s, time:%d, price:%f, volume:%f, tatalVolume:%f, fee:%f
	Get from Inchannel(cap:%d, len:%d)
`,
				t.Symbol, t.MarketType.String(),
				v.ID, v.Who, v.AorB.String(), v.Symbol, v.Timestamp, v.Price, v.Volume, v.TotalVolume, v.Fee,
				INCHANNEL_BUFF_SIZE, len(t.InChannelBlock),
			)

			/// Add to trade pool to match
			v.Volume = v.TotalVolume
			t.add(v)
		} else {
			panic(fmt.Errorf("%s-%s Input Routine InChannel exception occur!", t.Symbol, t.MarketType.String()))
		}

		t.doctor.RecordBeatHeart(doctor.RunningType_Enorder)
		runtime.Gosched()
	}
}

func (t *TradePool) matchtradeOutput(bidTrade *comm.Trade, askTrade *comm.Trade) {

	/// debug:
	comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s MatchTrade(bid:%d,ask:%d) Output to channel=======================>>>>>>>>>>\n",
		t.Symbol, t.MarketType.String(), bidTrade.ID, askTrade.ID)

	//// Update bid and ask trade output to ds:
	err, _ := use_mysql.MEMySQLInstance().UpdateTrade(bidTrade, askTrade)
	if err != nil {
		panic(err)
	}

	//// Update tickers infomation
	tradePair := te.TradePair{bidTrade, askTrade}
	t.tickerEngine.UpdateTicker(&tradePair)
}

func (t *TradePool) cancelOrderOutput(order *comm.Order) {
	/// debug:
	comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s CancelOrder(id:%d) Output to channel=======================>>>>>>>>>>\n",
		t.Symbol, t.MarketType.String(), order.ID)

	/// 2.1: Settle fund and remove from duration storage
	err, _ := use_mysql.MEMySQLInstance().CancelOrder(order)
	if err != nil {
		panic(err)
	}
}

func (t *TradePool) multiChanOutProc(chNO int) {
	var (
		v  *chs.OutElem = nil
		ok bool         = false
	)

	for {
		v, ok = t.MultiChanOut.OutChannel(chNO)
		if ok {
			switch v.Type_ {
			case chs.OUTPOOL_MATCHTRADE:
				t.matchtradeOutput(v.Trade.BidTrade, v.Trade.AskTrade)
			case chs.OUTPOOL_CANCELORDER:
				t.cancelOrderOutput(v.CancelOrder.Order)
			}
		}

		t.doctor.RecordBeatHeart(doctor.RunningType_Outpool)
		runtime.Gosched()
	}
}

func (t *TradePool) cancel() {
	var (
		v  *comm.Order = nil
		ok bool        = false
	)

	for {
		v, ok = <-t.CancelChannel
		if ok {
			/// debug:
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, `=======================>>>>>>>>>>%s-%s Cancel Order Input
	User:%s, Symbol:%s, ID:%d, Type: %s, Price:%f, Timestamp:%d
	Get from CancelChannel(cap:%d, len:%d)
`,
				t.Symbol, t.MarketType.String(),
				v.Who, v.Symbol, v.ID, v.AorB.String(), v.Price, v.Timestamp,
				CANCELCHANNEL_BUFF_SIZE, len(t.CancelChannel),
			)

			/// 1: Remove order from ME trade pool
			if ok, orderCorpse := t.cancelProcess(v); ok {
				/// 2: Settle fund and remove from duration storage
				t.MultiChanOut.InChannel(&chs.OutElem{Trade: nil, CancelOrder: &chs.CanceledOrder{&orderCorpse}, Type_: chs.OUTPOOL_CANCELORDER, Count: 0})
			} else {
				comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s To Cancel Order(ID=%d) not in trade pool, Perhaps had been matched out, or input order illegal.\n", t.Symbol, t.MarketType.String(), v.ID)
				continue
			}
		} else {
			panic(fmt.Errorf("%s-%s CancelChannel exception occur!", t.Symbol, t.MarketType.String()))
		}

		t.doctor.RecordBeatHeart(doctor.RunningType_CancelOrder)
		time.Sleep(comm.MECORE_CANCEL_DURATION)
		runtime.Gosched()
	}
}

/// cancel order function
func (t *TradePool) cancelProcess(order *comm.Order) (bool, comm.Order) {
	comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK, "%s-%s TradePool CancelProcess Order ID(%d), Time(%d)\n", t.Symbol, t.MarketType.String(), order.ID, order.Timestamp)
	if order == nil {
		fmt.Printf("%s-%s Cancel input order==nil error!\n", t.Symbol, t.MarketType.String())
		return false, comm.Order{}
	}
	if order.Symbol != t.Symbol {
		fmt.Printf("Market(%s-%s) cancelProcess illegal order with symbol(%s) to %s Match Engine", t.Symbol, t.MarketType.String(), order.Symbol, t.Symbol)
		return false, comm.Order{}
	}

	if order.AorB == comm.TradeType_BID {
		t.bidPoolRWMutex.Lock("CancelProcess BID")
		defer t.bidPoolRWMutex.Unlock("CancelProcess BID")

		/// debug:
		TimeDot1 := time.Now().UnixNano()
		target, suc := t.orderCheck(order)
		if !suc {
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK,
				"%s-%s Cancel OrderCheck ID(%d) fail! Order perhaps not exist or had been processed or input order illegal!\n",
				t.Symbol, t.MarketType.String(), order.ID)
			return false, comm.Order{}
		}

		/// Output the order corpse for DS operate use
		orderCorpse := t.bidPoolIDSlice[target].Value.(comm.Order)
		bRet := t.rmBidOrderByTarget(target)

		/// debug:
		TimeDot2 := time.Now().UnixNano()
		comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK,
			`%s-%s TradePool CancelProcess bid order complete==
	user:%s, id:%d, type:%s, time:%d, price:%f, volume:%f	, ****USE_TIME:%f
`,
			t.Symbol, t.MarketType.String(),
			orderCorpse.Who, orderCorpse.ID, orderCorpse.AorB.String(), orderCorpse.Timestamp, orderCorpse.Price, orderCorpse.Volume, float64(TimeDot2-TimeDot1)/float64(1*time.Second),
		)
		return bRet, orderCorpse
	} else if order.AorB == comm.TradeType_ASK {
		t.askPoolRWMutex.Lock("CancelProcess ASK")
		defer t.askPoolRWMutex.Unlock("CancelProcess ASK")

		/// debug:
		TimeDot1 := time.Now().UnixNano()
		target, suc := t.orderCheck(order)
		if !suc {
			comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK,
				"%s-%s Cancel OrderCheck ID(%d) fail! Order perhaps not exist or had been processed or input order illegal!\n",
				t.Symbol, t.MarketType.String(), order.ID)
			return false, comm.Order{}
		}

		/// Output the order corpse for DS operate use
		orderCorpse := t.askPoolIDSlice[target].Value.(comm.Order)
		bRet := t.rmAskOrderByTarget(target)

		/// debug:
		TimeDot2 := time.Now().UnixNano()
		comm.DebugPrintf(MODULE_NAME_SORTSLICE, comm.LOG_LEVEL_TRACK,
			`%s-%s TradePool CancelProcess ask order complete==
	user:%s, id:%d, type:%s, time:%d, price:%f, volume:%f	, ****USE_TIME:%f
`,
			t.Symbol, t.MarketType.String(),
			orderCorpse.Who, orderCorpse.ID, orderCorpse.AorB.String(), orderCorpse.Timestamp, orderCorpse.Price, orderCorpse.Volume, float64(TimeDot2-TimeDot1)/float64(1*time.Second),
		)
		return bRet, orderCorpse
	} else {
		fmt.Printf("%s-%s CancelProcess illegal order type!", t.Symbol, t.MarketType.String())
		return false, comm.Order{}
	}
}

/// cancel test
func (t *TradePool) test_CancelOrder(user string, id int64, symbol string) {
	if t.Symbol != symbol {
		fmt.Printf("Test_CancelOrder not the corresponding symbol ME.\n")
		return
	}

	order, err := use_mysql.MEMySQLInstance().GetOrder(user, id, symbol, nil)
	if err != nil {
		fmt.Printf("%s-%s Test_CancelOrder GetOrder fail! Order may not in the order duration storage, Test_CancelOrder aborted.\n", t.Symbol, t.MarketType.String())
		return
	}
	t.CancelChannel <- order
}

//	debug *DebugInfo
/// test pool
func (t *TradePool) PrintHealth() {
	fmt.Printf("=================[%s-%s Pool Health Start]=================\n", t.Symbol, t.MarketType.String())
	fmt.Println("AskPool=============")
	fmt.Println("askPool		==", "Length:", t.askPool.Len(), ";\tTopElem:", t.askPool.Front().Value.(comm.Order).ID, ";\tEndElem:", t.askPool.Back().Value.(comm.Order).ID)
	fmt.Println("askPoolSlice	==", "Length:", len(t.askPoolSlice), ";\tTopElem:", t.askPoolSlice[0].Value.(comm.Order).ID, ";\tEndElem:", t.askPoolSlice[len(t.askPoolSlice)-1].Value.(comm.Order).ID)
	fmt.Println("askPoolIDSlice	==", "Length:", len(t.askPoolIDSlice), ";\tTopElem:", t.askPoolIDSlice[0].Value.(comm.Order).ID, ";\tEndElem:", t.askPoolIDSlice[len(t.askPoolIDSlice)-1].Value.(comm.Order).ID)

	fmt.Println("BidPool=============")
	fmt.Println("bidPool		==", "Length:", t.bidPool.Len(), ";\tTopElem:", t.bidPool.Front().Value.(comm.Order).ID, ";\tEndElem:", t.bidPool.Back().Value.(comm.Order).ID)
	fmt.Println("bidPoolSlice	==", "Length:", len(t.bidPoolSlice), ";\tTopElem:", t.bidPoolSlice[0].Value.(comm.Order).ID, ";\tEndElem:", t.bidPoolSlice[len(t.bidPoolSlice)-1].Value.(comm.Order).ID)
	fmt.Println("bidPoolIDSlice	==", "Length:", len(t.bidPoolIDSlice), ";\tTopElem:", t.bidPoolIDSlice[0].Value.(comm.Order).ID, ";\tEndElem:", t.bidPoolIDSlice[len(t.bidPoolIDSlice)-1].Value.(comm.Order).ID)

	fmt.Println("Channel=============")
	for i := 0; i < INCHANNEL_POOL_SIZE; i++ {
		in, ok := t.InChannel.Out(i)
		if ok {
			fmt.Println("InChannel 		==", "Working", ";\tTopElem:", in.ID)
			t.InChannel.In(in)
		} else {
			fmt.Println("InChannel 		==", "IsClosed")
		}
	}

	can, ok := <-t.CancelChannel
	if ok {
		fmt.Println("CancelChannel	==", "Working", ";\tTopElem:", can.ID)
		t.CancelChannel <- can
	} else {
		fmt.Println("CancelChannel 	==", "IsClosed")
	}

	fmt.Println("Lock	=============")
	/// test askPoolRWMutex RLock
	c := make(chan bool, 1)
askPoolRWMutex_RLock:
	go func() {
		/// test lock
		t.askPoolRWMutex.RLock("LockTest ASK")
		t.askPoolRWMutex.RUnlock("LockTest ASK")
		c <- true
	}()
	select {
	case res := <-c:
		/// Test Pass
		fmt.Println("askPoolRWMutex	==", "RLock() Not Locked")
		_ = res
	case <-time.After(time.Second * 1):
		/// Test Fail
		fmt.Println("askPoolRWMutex	==", "RLock() Be Locked!", " Now try to unlock it and retry")
		t.askPoolRWMutex.RUnlock("LockTest ASK")
		goto askPoolRWMutex_RLock
	}
	close(c)

	/// test askPoolRWMutex Lock
	c = make(chan bool, 1)
askPoolRWMutex_Lock:
	go func() {
		/// test lock
		t.askPoolRWMutex.Lock("LockTest ASK")
		t.askPoolRWMutex.Unlock("LockTest ASK")
		c <- true
	}()
	select {
	case res := <-c:
		/// Test Pass
		fmt.Println("askPoolRWMutex	==", "Lock() Not Locked")
		_ = res
	case <-time.After(time.Second * 1):
		/// Test Fail
		fmt.Println("askPoolRWMutex	==", "Lock() Be Locked!", " Now try to unlock it and retry")
		t.askPoolRWMutex.Unlock("LockTest ASK")
		goto askPoolRWMutex_Lock
	}
	close(c)

	/// test bidPoolRWMutex RLock
	c = make(chan bool, 1)
bidPoolRWMutex_RLock:
	go func() {
		/// test lock
		t.bidPoolRWMutex.RLock("LockTest BID")
		t.bidPoolRWMutex.RUnlock("LockTest BID")
		c <- true
	}()
	select {
	case res := <-c:
		/// Test Pass
		fmt.Println("bidPoolRWMutex	==", "RLock() Not Locked")
		_ = res
	case <-time.After(time.Second * 1):
		/// Test Fail
		fmt.Println("bidPoolRWMutex	==", "RLock() Be Locked!", " Now try to unlock it and retry")
		t.bidPoolRWMutex.RUnlock("LockTest BID")
		goto bidPoolRWMutex_RLock
	}
	close(c)

	/// test bidPoolRWMutex Lock
	c = make(chan bool, 1)
bidPoolRWMutex_Lock:
	go func() {
		/// test lock
		t.bidPoolRWMutex.Lock("LockTest BID")
		t.bidPoolRWMutex.Unlock("LockTest BID")
		c <- true
	}()
	select {
	case res := <-c:
		/// Test Pass
		fmt.Println("bidPoolRWMutex	==", "Lock() Not Locked")
		_ = res
	case <-time.After(time.Second * 1):
		/// Test Fail
		fmt.Println("bidPoolRWMutex	==", "Lock() Be Locked!", " Now try to unlock it and retry")
		t.bidPoolRWMutex.Unlock("LockTest BID")
		goto bidPoolRWMutex_Lock
	}
	close(c)

	fmt.Println("Control=============")
	fmt.Println("Trade Control	==", t.tradeControl)

	fmt.Println("=================[Pool Health End ]=================\n")
}

func (t *TradePool) test_Pool(p ...interface{}) {
	ope, _ := p[0].(string)
	switch ope {
	case "health":
		t.PrintHealth()
	}
}

func (t *TradePool) Test(u string, p ...interface{}) {
	switch u {
	case "cancel":
		who, _ := p[0].(string)
		id, _ := strconv.ParseInt(p[1].(string), 10, 64)
		symbol, _ := p[2].(string)
		t.test_CancelOrder(who, id, symbol)

	case "pool":
		t.test_Pool(p...)
	default:

	}
}

func (t *TradePool) TradeCommand(u string, p ...interface{}) {
	switch u {
	case "start":
		t.tradeControl = TradeControl_Work
	case "stop":
		t.tradeControl = TradeControl_Stop
	case "pause":
		t.tradeControl = TradeControl_Pause
	case "resume":
		t.tradeControl = TradeControl_Work
	default:

	}
}

func (t *TradePool) EnOrder(order *comm.Order) error {
	if config.GetMEConfig().InPoolMode == "block" {
		t.InChannelBlock <- order
	} else {
		t.InChannel.In(order)
	}

	return nil
}

func (t *TradePool) CancelOrder(id int64) error {
	return nil
}

func (t *TradePool) CancelTheOrder(order *comm.Order) error {
	t.CancelChannel <- order
	return nil
}
