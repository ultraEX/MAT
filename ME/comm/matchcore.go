package comm

import (
	"fmt"
	"strconv"
	"sync"
)

type ExInterface interface {

	//---------------------------------------------------------------------------------------------------
	EnOrder(order *Order) error
	CancelOrder(id int64) error
	CancelTheOrder(order *Order) error
	CreateID() int64
	//---------------------------------------------------------------------------------------------------

}

type MxInterface interface {

	//---------------------------------------------------------------------------------------------------
	Start()
	Stop()
	Destroy()
	//---------------------------------------------------------------------------------------------------

}

type MTInterface interface {

	//---------------------------------------------------------------------------------------------------
	GetAskLevelOrders(limit int64) []*Order
	GetBidLevelOrders(limit int64) []*Order
	GetAskLevelsGroupByPrice(limit int64) []OrderLevel
	GetBidLevelsGroupByPrice(limit int64) []OrderLevel
	//---------------------------------------------------------------------------------------------------

}

type MonitorInterface interface {
	//---------------------------------------------------------------------------------------------------
	DumpTradePool(detail bool) string
	DumpTradePoolPrint(detail bool)
	PumpTradePoolPrint()
	DumpBeatHeart() string
	DumpChannel() string
	DumpChanlsMap()
	IsFaulty() bool
	RestartDebuginfo()
	ResetMatchCorePerform()
	Statics() string
	PrintHealth()
	Test(u string, p ...interface{})
	TradeCommand(u string, p ...interface{})

	GetTradeCompleteRate() float64
	GetAskPoolLen() int
	GetBidPoolLen() int
	GetPoolLen() int
	//---------------------------------------------------------------------------------------------------
}

type MatchCoreItf interface {
	//---------------------------------------------------------------------------------------------------
	ExInterface
	MonitorInterface
	MTInterface
	//---------------------------------------------------------------------------------------------------
}

type OrderContainerItf interface {
	//---------------------------------------------------------------------------------------------------
	Push(order *Order)
	Pop(aorb TradeType) (order *Order)
	GetTop(aorb TradeType) (order *Order)
	Get(id int64) *Order
	Dump()
	Pump()
	BidSize() int64
	AskSize() int64
	TheSize() int64
	//---------------------------------------------------------------------------------------------------
}

///------------------------------------------------------------------
type IDOrderMap struct {
	m sync.Map
}

func NewIDOrderMap() *IDOrderMap {
	o := new(IDOrderMap)
	return o
}

func (t *IDOrderMap) Set(id int64, value *Order) {
	t.m.Store(id, value)
}

func (t *IDOrderMap) Get(id int64) (*Order, bool) {
	if orderItf, ok := t.m.Load(id); ok {
		return orderItf.(*Order), ok
	} else {
		return nil, ok
	}
}

func (t *IDOrderMap) Remove(id int64) {
	t.m.Delete(id)
}

func (t *IDOrderMap) Len() int64 {
	return LenOfSyncMap(&t.m)
}

func (t *IDOrderMap) Dump() {
	fmt.Printf("======================Dump IDOrderMap==========================\n")
	fmt.Printf("len = %d\n", t.Len())
	t.m.Range(func(id, order interface{}) bool {
		fmt.Printf("[%d]: id = %d, type = %s, price = %f, volume = %f, time = %d;\n",
			id,
			order.(*Order).ID,
			order.(*Order).AorB,
			order.(*Order).Price,
			order.(*Order).Volume,
			order.(*Order).Timestamp,
		)
		return true
	})
	fmt.Printf("===================================================================\n")
}

///------------------------------------------------------------------
/// price time priority algorithm
const (
	PRICE_MULTI_FACTOR float64 = 1000000000000000000
	PRICE_MAX_VALUE    float64 = 1000000000000000000
	PRICE_MAX_PRECISE  int64   = 1000000000
	TIME_DIV_FACTOR    int64   = 1000000000
)

func BidScore(price float64, timestamp int64) float64 {
	return (PRICE_MAX_VALUE - price*PRICE_MULTI_FACTOR) + float64(timestamp)
}

func AskScore(price float64, timestamp int64) float64 {
	return (price * PRICE_MULTI_FACTOR) + float64(timestamp)
}

func BidKey(price float64, timestamp int64, id int64) float64 {
	return (PRICE_MAX_VALUE - price*PRICE_MULTI_FACTOR) + float64(timestamp/TIME_DIV_FACTOR*TIME_DIV_FACTOR) + float64(id%TIME_DIV_FACTOR)
}

func AskKey(price float64, timestamp int64, id int64) float64 {
	return (price * PRICE_MULTI_FACTOR) + float64(timestamp/TIME_DIV_FACTOR*TIME_DIV_FACTOR) + float64(id%TIME_DIV_FACTOR)
}

func BidKeyStr(price float64, id int64) string {
	return strconv.FormatFloat(PRICE_MAX_VALUE-price*float64(PRICE_MAX_PRECISE), 'f', 10, 64) + strconv.FormatInt(id, 10)
}

func AskKeyStr(price float64, id int64) string {
	return strconv.FormatFloat(price, 'f', 10, 64) + strconv.FormatInt(id, 10)
}
