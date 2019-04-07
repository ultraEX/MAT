package structmap

import (
	"fmt"

	"../../../comm"
)

///------------------------------------------------------------------
type OrderCoitainerBase1Sk struct {
	*comm.IDOrderMap
	*OrdersList
}

func NewOrderCoitainerBase1Sk() *OrderCoitainerBase1Sk {
	o := new(OrderCoitainerBase1Sk)
	o.IDOrderMap = comm.NewIDOrderMap()
	o.OrdersList = NewOrdersList()
	return o
}

func (t *OrderCoitainerBase1Sk) Push(order *comm.Order) {
	if _, ok := t.IDOrderMap.Get(order.ID); ok {
		panic(fmt.Errorf("OrderCoitainerBase1Sk Push ID(%d) replication, check it!", order.ID))
	}
	t.IDOrderMap.Set(order.ID, order)

	if order.AorB == comm.TradeType_BID {
		t.OrdersList.bidOrders.Push(order)
	} else {
		t.OrdersList.askOrders.Push(order)
	}
}

func (t *OrderCoitainerBase1Sk) Pop(aorb comm.TradeType) (order *comm.Order) {
	if aorb == comm.TradeType_BID {
		order = t.OrdersList.bidOrders.Pop()
	} else {
		order = t.OrdersList.askOrders.Pop()
	}
	if order != nil {
		t.IDOrderMap.Remove(order.ID)
	}

	return order
}

func (t *OrderCoitainerBase1Sk) GetTop(aorb comm.TradeType) (order *comm.Order) {
	if aorb == comm.TradeType_BID {
		order = t.OrdersList.bidOrders.GetTop()
	} else {
		order = t.OrdersList.askOrders.GetTop()
	}

	return order
}

func (t *OrderCoitainerBase1Sk) Get(id int64) *comm.Order {
	if order, ok := t.IDOrderMap.Get(id); ok {
		return order
	} else {
		return nil
	}
}

func (t *OrderCoitainerBase1Sk) Dump() {
	fmt.Printf("===================Dump OrderCoitainerBase1Sk======================\n")
	t.IDOrderMap.Dump()
	t.OrdersList.Dump()
	fmt.Printf("===================================================================\n")
}

func (t *OrderCoitainerBase1Sk) BidSize() int64 {
	return int64(t.OrdersList.bidOrders.Len())
}

func (t *OrderCoitainerBase1Sk) AskSize() int64 {
	return int64(t.OrdersList.askOrders.Len())
}

func (t *OrderCoitainerBase1Sk) TheSize() int64 {
	return t.IDOrderMap.Len()
}

///------------------------------------------------------------------

///------------------------------------------------------------------
type OrdersList struct {
	bidOrders OrdersByScoreItf
	askOrders OrdersByScoreItf
}

func NewOrdersList() *OrdersList {
	o := new(OrdersList)
	// o.bidOrders = NewOrdersByScoreBase1dSk()
	// o.askOrders = NewOrdersByScoreBase1dSk()
	// o.bidOrders = NewOrdersByScoreBaseConSk()
	// o.askOrders = NewOrdersByScoreBaseConSk()
	o.bidOrders = NewOrdersByScoreBaseHM()
	o.askOrders = NewOrdersByScoreBaseHM()
	// o.bidOrders = NewOrdersByScoreBaseLazySk()
	// o.askOrders = NewOrdersByScoreBaseLazySk()

	return o
}

func (t *OrdersList) Dump() {
	fmt.Printf("======================Dump OrdersList==========================\n")
	fmt.Printf("bidOrders:\n")
	t.bidOrders.Dump()
	fmt.Printf("askOrders:\n")
	t.askOrders.Dump()
	fmt.Printf("===================================================================\n")
}

func (t *OrdersList) Pump() {
	fmt.Printf("======================Pump OrdersList==========================\n")
	fmt.Printf("bidOrders:\n")
	t.bidOrders.Pump()
	fmt.Printf("askOrders:\n")
	t.askOrders.Pump()
	fmt.Printf("===================================================================\n")
}
