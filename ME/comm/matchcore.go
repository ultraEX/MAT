package comm

import (
	"fmt"
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
