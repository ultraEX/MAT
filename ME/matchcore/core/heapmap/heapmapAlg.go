package heapmap

import (
	"container/heap"
	"fmt"
	"sync"

	"../../../comm"
)

///------------------------------------------------------------------
type TradeContainer struct {
	comm.IDOrderMap
	TradePrices
}

func NewTradeContainer() *TradeContainer {
	o := new(TradeContainer)
	o.IDOrderMap = *comm.NewIDOrderMap()
	o.TradePrices = *NewTradePrices()
	return o
}

func (t *TradeContainer) Push(order *comm.Order) {
	if _, ok := t.IDOrderMap.Get(order.ID); ok {
		panic(fmt.Errorf("TradeContainer Push ID(%d) replication, reentry order!", order.ID))
	}
	t.IDOrderMap.Set(order.ID, order)

	if order.AorB == comm.TradeType_BID {
		t.TradePrices.BidPrices.Push(order)
	} else {
		t.TradePrices.AskPrices.Push(order)
	}
}

func (t *TradeContainer) Pop(aorb comm.TradeType) (order *comm.Order) {
	if aorb == comm.TradeType_BID {
		order = t.TradePrices.BidPrices.Pop()
	} else {
		order = t.TradePrices.AskPrices.Pop()
	}
	if order != nil {
		t.IDOrderMap.Remove(order.ID)
	}

	return order
}

func (t *TradeContainer) GetTop(aorb comm.TradeType) (order *comm.Order) {
	if aorb == comm.TradeType_BID {
		order = t.TradePrices.BidPrices.GetTop()
	} else {
		order = t.TradePrices.AskPrices.GetTop()
	}

	return order
}

func (t *TradeContainer) Get(id int64) *comm.Order {
	if order, ok := t.IDOrderMap.Get(id); ok {
		return order
	} else {
		return nil
	}
}

func (t *TradeContainer) Dump() {
	fmt.Printf("======================Dump TradeContainer==========================\n")
	t.IDOrderMap.Dump()
	t.TradePrices.Dump()
	fmt.Printf("===================================================================\n")
}

func (t *TradeContainer) BidSize() int64 {
	return t.TradePrices.BidPrices.Size()
}

func (t *TradeContainer) AskSize() int64 {
	return t.TradePrices.AskPrices.Size()
}

func (t *TradeContainer) TheSize() int64 {
	return t.IDOrderMap.Len()
}

///------------------------------------------------------------------
type AscPriceHeap struct {
	orders []float64
}

func (t *AscPriceHeap) Len() int {
	return len(t.orders)
}

func (t *AscPriceHeap) Less(i, j int) bool {
	return t.orders[i] > t.orders[j]
}

func (t *AscPriceHeap) Swap(i, j int) {
	tmp := t.orders[i]
	t.orders[i] = t.orders[j]
	t.orders[j] = tmp
}

func (t *AscPriceHeap) Push(x interface{}) {
	t.orders = append(t.orders, x.(float64))
}

func (t *AscPriceHeap) Pop() (ret interface{}) {
	l := len(t.orders)
	t.orders, ret = t.orders[:l-1], t.orders[l-1]
	return
}

func (t *AscPriceHeap) GetTop() (ret interface{}) {
	l := len(t.orders)
	if l == 0 {
		return nil
	} else {
		return t.orders[0]
	}
}

type BidPrices struct {
	bids AscPriceHeap
	PriceTimeMap

	ConMutex sync.Mutex
}

func NewBidPrices() *BidPrices {
	o := new(BidPrices)
	heap.Init(&o.bids)

	o.PriceTimeMap = *NewPriceTimeMap()
	return o
}

func (t *BidPrices) Push(order *comm.Order) {
	t.ConMutex.Lock()
	if orders, ok := t.PriceTimeMap.Get(order.Price); ok {
		orders.Push(order)
	} else {
		orders = NewOrdersByTime()
		orders.Push(order)
		t.PriceTimeMap.Set(order.Price, orders)
		heap.Push(&t.bids, order.Price)
	}
	t.ConMutex.Unlock()
}

func (t *BidPrices) Pop() (order *comm.Order) {
	t.ConMutex.Lock()
	defer t.ConMutex.Unlock()

	priceItf := t.bids.GetTop()
	if priceItf == nil {
		return nil
	}
	price := priceItf.(float64)

	if orders, ok := t.PriceTimeMap.Get(price); ok {
		order := orders.Pop()
		if orders.Len() == 0 {
			t.PriceTimeMap.Remove(price)
			heap.Pop(&t.bids)
		}
		return order
	} else {
		t.Dump()
		panic(fmt.Errorf("BidPrices Pop order with price(%.18f) fail, logic error!", price))
	}
}

func (t *BidPrices) GetTop() (order *comm.Order) {
	priceItf := t.bids.GetTop()
	if priceItf == nil {
		return nil
	}

	price := priceItf.(float64)
	if orders, ok := t.PriceTimeMap.Get(price); ok {
		order := orders.GetTop()
		return order
	} else {
		return nil
	}
}

func (t *BidPrices) Len() int {
	return t.bids.Len()
}

func (t *BidPrices) Size() int64 {
	return t.PriceTimeMap.Size()
}

func (t *BidPrices) Dump() {
	fmt.Printf("======================Dump BidPrices==========================\n")
	fmt.Printf("bids len = %d\n", t.bids.Len())
	for c, price := range t.bids.orders {
		fmt.Printf("[%d] price = %.8f;\n", c, price)
	}
	t.PriceTimeMap.Dump()
	fmt.Printf("===================================================================\n")
}

///------------------------------------------------------------------
type DesPriceHeap struct {
	orders []float64
}

func (t *DesPriceHeap) Len() int {
	return len(t.orders)
}

func (t *DesPriceHeap) Less(i, j int) bool {
	return t.orders[i] < t.orders[j]
}

func (t *DesPriceHeap) Swap(i, j int) {
	tmp := t.orders[i]
	t.orders[i] = t.orders[j]
	t.orders[j] = tmp
}

func (t *DesPriceHeap) Push(x interface{}) {
	t.orders = append(t.orders, x.(float64))
}

func (t *DesPriceHeap) Pop() (ret interface{}) {
	l := len(t.orders)
	t.orders, ret = t.orders[:l-1], t.orders[l-1]
	return
}

func (t *DesPriceHeap) GetTop() (ret interface{}) {
	l := len(t.orders)
	if l == 0 {
		return nil
	} else {
		return t.orders[0]
	}
}

type AskPrices struct {
	asks DesPriceHeap
	PriceTimeMap

	ConMutex sync.Mutex
}

func NewAskPrices() *AskPrices {
	o := new(AskPrices)
	heap.Init(&o.asks)

	o.PriceTimeMap = *NewPriceTimeMap()
	return o
}

func (t *AskPrices) Push(order *comm.Order) {
	t.ConMutex.Lock()
	if orders, ok := t.PriceTimeMap.Get(order.Price); ok {
		orders.Push(order)
	} else {
		orders = NewOrdersByTime()
		orders.Push(order)
		t.PriceTimeMap.Set(order.Price, orders)
		heap.Push(&t.asks, order.Price)
	}
	t.ConMutex.Unlock()
}

func (t *AskPrices) Pop() (order *comm.Order) {
	t.ConMutex.Lock()
	defer t.ConMutex.Unlock()

	priceItf := t.asks.GetTop()
	if priceItf == nil {
		return nil
	}
	price := priceItf.(float64)

	if orders, ok := t.PriceTimeMap.Get(price); ok {
		order = orders.Pop()
		if orders.Len() == 0 {
			t.PriceTimeMap.Remove(price)
			heap.Pop(&t.asks)
		}
		return order
	} else {
		t.Dump()
		panic(fmt.Errorf("AskPrices Pop order with price(%.18f) fail, logic error!", price))
	}
}

func (t *AskPrices) GetTop() (order *comm.Order) {
	priceItf := t.asks.GetTop()
	if priceItf == nil {
		return nil
	}

	price := priceItf.(float64)
	if orders, ok := t.PriceTimeMap.Get(price); ok {
		order := orders.GetTop()
		return order
	} else {
		return nil
	}
}

func (t *AskPrices) Len() int {
	return t.asks.Len()
}

func (t *AskPrices) Size() int64 {
	return t.PriceTimeMap.Size()
}

func (t *AskPrices) Dump() {
	fmt.Printf("======================Dump AskPrices==========================\n")
	fmt.Printf("asks len = %d\n", t.asks.Len())
	for c, price := range t.asks.orders {
		fmt.Printf("[%d] price = %.8f;\n", c, price)
	}
	t.PriceTimeMap.Dump()
	fmt.Printf("===================================================================\n")
}

///------------------------------------------------------------------
type TradePrices struct {
	BidPrices
	AskPrices
}

func NewTradePrices() *TradePrices {
	o := new(TradePrices)
	o.BidPrices = *NewBidPrices()
	o.AskPrices = *NewAskPrices()
	return o
}

func (t *TradePrices) Dump() {
	fmt.Printf("======================Dump TradePrices==========================\n")
	fmt.Printf("BidPrices:\n")
	t.BidPrices.Dump()
	fmt.Printf("AskPrices:\n")
	t.AskPrices.Dump()
	fmt.Printf("===================================================================\n")
}

///------------------------------------------------------------------
type DesTimeHeap struct {
	orders []*comm.Order
}

func (t *DesTimeHeap) Len() int {
	return len(t.orders)
}

func (t *DesTimeHeap) Less(i, j int) bool {
	return t.orders[i].Price > t.orders[j].Price
}

func (t *DesTimeHeap) Swap(i, j int) {
	tmp := t.orders[i]
	t.orders[i] = t.orders[j]
	t.orders[j] = tmp
}

func (t *DesTimeHeap) Push(x interface{}) {
	t.orders = append(t.orders, x.(*comm.Order))
}

func (t *DesTimeHeap) Pop() (ret interface{}) {
	l := len(t.orders)
	t.orders, ret = t.orders[:l-1], t.orders[l-1]
	return
}

func (t *DesTimeHeap) GetTop() (ret interface{}) {
	l := len(t.orders)
	if l == 0 {
		return nil
	} else {
		return t.orders[0]
	}
}

type OrdersByTime struct {
	orders DesTimeHeap

	ConMutex sync.Mutex
}

func NewOrdersByTime() *OrdersByTime {
	o := new(OrdersByTime)
	heap.Init(&o.orders)
	return o
}

func (t *OrdersByTime) Push(order *comm.Order) {
	t.ConMutex.Lock()
	heap.Push(&t.orders, order)
	t.ConMutex.Unlock()
}

func (t *OrdersByTime) Pop() (order *comm.Order) {
	t.ConMutex.Lock()
	order = heap.Pop(&t.orders).(*comm.Order)
	t.ConMutex.Unlock()
	return
}

func (t *OrdersByTime) Len() int64 {
	return int64(t.orders.Len())
}

func (t *OrdersByTime) GetTop() *comm.Order {
	t.ConMutex.Lock()
	defer t.ConMutex.Unlock()
	orderItf := t.orders.GetTop()
	if orderItf != nil {
		return orderItf.(*comm.Order)
	} else {
		return nil
	}
}

///------------------------------------------------------------------
type PriceTimeMap struct {
	// m map[float64]*OrdersByTime
	m sync.Map
}

func NewPriceTimeMap() *PriceTimeMap {
	o := new(PriceTimeMap)
	// o.m = make(map[float64]*OrdersByTime)
	return o
}

func (t *PriceTimeMap) Set(price float64, orders *OrdersByTime) {
	// t.m[price] = orders
	t.m.Store(price, orders)
}

func (t *PriceTimeMap) Get(price float64) (orders *OrdersByTime, ok bool) {
	// orders, ok = t.m[price]
	if ordersItf, ok := t.m.Load(price); ok {
		return ordersItf.(*OrdersByTime), ok
	} else {
		return nil, ok
	}
}

func (t *PriceTimeMap) Remove(price float64) {
	// delete(t.m, price)
	t.m.Delete(price)
}

func (t *PriceTimeMap) Len() int64 {
	// return len(t.m)
	return comm.LenOfSyncMap(&t.m)
}

func (t *PriceTimeMap) Size() int64 {
	var size int64 = 0
	// for _, orders := range t.m {
	// 	size += orders.Len()
	// }
	t.m.Range(func(price, orders interface{}) bool {
		size += orders.(*OrdersByTime).Len()
		return true
	})
	return size
}

func (t *PriceTimeMap) Dump() {
	fmt.Printf("======================Dump PriceTimeMap==========================\n")
	fmt.Printf("PriceTimeMap len = %d\n", t.Len())
	// for price, orders := range t.m {
	// 	fmt.Printf("price = %f, orders = %v;\n", price, orders)
	// }
	t.m.Range(func(price, ordersItf interface{}) bool {
		orders := ordersItf.(*OrdersByTime)
		fmt.Printf("price = %f, orders(len=%d):\n", price, len(orders.orders.orders))
		for count, order := range orders.orders.orders {
			fmt.Printf("[%d]: price = %.8f, time = %d, volume = %.8f, totalvolume = %.8f;\n",
				count,
				order.Price,
				order.Timestamp,
				order.Volume,
				order.TotalVolume,
			)
		}
		return true
	})
	fmt.Printf("===================================================================\n")
}
