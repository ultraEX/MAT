package structmap

import (
	"fmt"

	"../../../comm"
	sk "../struct"
)

///------------------------------------------------------------------
type OrderCoitainerBase2Sk struct {
	comm.IDOrderMap
	TradePricesBaseSk
}

func NewOrderCoitainerBase2Sk() *OrderCoitainerBase2Sk {
	o := new(OrderCoitainerBase2Sk)
	o.IDOrderMap = *comm.NewIDOrderMap()
	o.TradePricesBaseSk = *NewTradePricesBaseSk()
	return o
}

func (t *OrderCoitainerBase2Sk) Push(order *comm.Order) {
	if _, ok := t.IDOrderMap.Get(order.ID); ok {
		panic(fmt.Errorf("OrderCoitainerBase2Sk Push ID(%d) replication, reentry order!", order.ID))
	}
	t.IDOrderMap.Set(order.ID, order)

	if order.AorB == comm.TradeType_BID {
		t.TradePricesBaseSk.BidPricesBaseSk.Push(order)
	} else {
		t.TradePricesBaseSk.AskPricesBaseSk.Push(order)
	}
}

func (t *OrderCoitainerBase2Sk) Pop(aorb comm.TradeType) (order *comm.Order) {
	if aorb == comm.TradeType_BID {
		order = t.TradePricesBaseSk.BidPricesBaseSk.Pop()
	} else {
		order = t.TradePricesBaseSk.AskPricesBaseSk.Pop()
	}
	if order != nil {
		t.IDOrderMap.Remove(order.ID)
	}

	return order
}

func (t *OrderCoitainerBase2Sk) GetTop(aorb comm.TradeType) (order *comm.Order) {
	if aorb == comm.TradeType_BID {
		order = t.TradePricesBaseSk.BidPricesBaseSk.GetTop()
	} else {
		order = t.TradePricesBaseSk.AskPricesBaseSk.GetTop()
	}

	return order
}

func (t *OrderCoitainerBase2Sk) Get(id int64) *comm.Order {
	if order, ok := t.IDOrderMap.Get(id); ok {
		return order
	} else {
		return nil
	}
}

func (t *OrderCoitainerBase2Sk) Dump() {
	fmt.Printf("======================Dump OrderCoitainerBase2Sk==========================\n")
	t.IDOrderMap.Dump()
	t.TradePricesBaseSk.Dump()
	fmt.Printf("===================================================================\n")
}

func (t *OrderCoitainerBase2Sk) Pump() {
	fmt.Printf("======================Pump OrderCoitainerBase2Sk==========================\n")
	fmt.Printf("===================================================================\n")
}

func (t *OrderCoitainerBase2Sk) BidSize() int64 {
	return int64(t.TradePricesBaseSk.BidPricesBaseSk.Size())
}

func (t *OrderCoitainerBase2Sk) AskSize() int64 {
	return int64(t.TradePricesBaseSk.AskPricesBaseSk.Size())
}

func (t *OrderCoitainerBase2Sk) TheSize() int64 {
	return t.IDOrderMap.Len()
}

///------------------------------------------------------------------
type AscPriceHeapBaseSk struct {
	*sk.SkipList
}

// func (t *AscPriceHeapBaseSk) Descending() bool {
// 	return true
// }

// func (t *AscPriceHeapBaseSk) Compare(lhs, rhs interface{}) bool {
// 	return lhs.(float64) > rhs.(float64)
// }

func NewAscPriceHeapBaseSk() *AscPriceHeapBaseSk {
	o := new(AscPriceHeapBaseSk)
	o.SkipList = sk.New(sk.Float64Descending)
	return o
}

type BidPricesBaseSk struct {
	*AscPriceHeapBaseSk
}

func NewBidPricesBaseSk() *BidPricesBaseSk {
	o := new(BidPricesBaseSk)
	o.AscPriceHeapBaseSk = NewAscPriceHeapBaseSk()
	return o
}

func (t *BidPricesBaseSk) Push(order *comm.Order) {
	t.Set(order)
}

func (t *BidPricesBaseSk) Pop() *comm.Order {
	elem := t.AscPriceHeapBaseSk.Front()

	if elem != nil {
		price := elem.Key().(float64)
		orders := elem.Value.(*OrdersByTimeBaseSk)
		order := orders.Pop()
		if orders.Len() <= 0 {
			t.AscPriceHeapBaseSk.Remove(price)
		}
		return order
	} else {
		return nil
	}
}

func (t *BidPricesBaseSk) GetTop() *comm.Order {
	elem := t.AscPriceHeapBaseSk.Front()

	if elem != nil {
		orders := elem.Value.(*OrdersByTimeBaseSk)
		return orders.GetTop()
	} else {
		return nil
	}
}

func (t *BidPricesBaseSk) Set(order *comm.Order) {
	elem := t.AscPriceHeapBaseSk.Get(order.Price)
	if elem != nil {
		elem.Value.(*OrdersByTimeBaseSk).Set(order)
	} else {
		orders := NewOrdersByTimeBaseSk()
		orders.Set(order)
		t.AscPriceHeapBaseSk.Set(order.Price, orders)
	}
}

func (t *BidPricesBaseSk) Get(price float64) *OrdersByTimeBaseSk {
	elem := t.AscPriceHeapBaseSk.Get(price)
	if elem != nil {
		return elem.Value.(*OrdersByTimeBaseSk)
	} else {
		return nil
	}
}

func (t *BidPricesBaseSk) Remove(price float64) *OrdersByTimeBaseSk {
	elem := t.AscPriceHeapBaseSk.Remove(price)
	if elem != nil {
		return elem.Value.(*OrdersByTimeBaseSk)
	} else {
		return nil
	}
}

func (t *BidPricesBaseSk) Len() int {
	return int(t.AscPriceHeapBaseSk.Len())
}

func (t *BidPricesBaseSk) Size() int {
	unboundIterator := t.AscPriceHeapBaseSk.Iterator()
	sum := 0
	for unboundIterator != nil {
		sum += unboundIterator.Value.(*OrdersByTimeBaseSk).Len()
		unboundIterator = unboundIterator.Next()
	}
	return sum
}

func (t *BidPricesBaseSk) Dump() {
	fmt.Printf("======================Dump BidPricesBaseSk==========================\n")
	fmt.Printf("bids len = %d\n", t.AscPriceHeapBaseSk.Len())
	unboundIterator := t.AscPriceHeapBaseSk.Iterator()
	count := 0
	for unboundIterator != nil {
		fmt.Printf("[%d] price = %.8f:\n", count, unboundIterator.Key().(float64))
		unboundIterator.Value.(*OrdersByTimeBaseSk).Dump()
		count++
		unboundIterator = unboundIterator.Next()
	}

	fmt.Printf("===================================================================\n")
}

///------------------------------------------------------------------
type DesPriceHeapBaseSk struct {
	*sk.SkipList
}

// func (t *DesPriceHeapBaseSk) Descending() bool {
// 	return false
// }

// func (t *DesPriceHeapBaseSk) Compare(lhs, rhs interface{}) bool {
// 	return lhs.(float64) > rhs.(float64)
// }

func NewDesPriceHeapBaseSk() *DesPriceHeapBaseSk {
	o := new(DesPriceHeapBaseSk)
	o.SkipList = sk.New(sk.Float64Ascending)
	return o
}

type AskPricesBaseSk struct {
	*DesPriceHeapBaseSk
}

func NewAskPricesBaseSk() *AskPricesBaseSk {
	o := new(AskPricesBaseSk)
	o.DesPriceHeapBaseSk = NewDesPriceHeapBaseSk()
	return o
}

func (t *AskPricesBaseSk) Push(order *comm.Order) {
	t.Set(order)
}

func (t *AskPricesBaseSk) Pop() *comm.Order {
	elem := t.DesPriceHeapBaseSk.Front()

	if elem != nil {
		price := elem.Key().(float64)
		orders := elem.Value.(*OrdersByTimeBaseSk)
		order := orders.Pop()
		if orders.Len() <= 0 {
			t.DesPriceHeapBaseSk.Remove(price)
		}
		return order
	} else {
		return nil
	}
}

func (t *AskPricesBaseSk) GetTop() *comm.Order {
	elem := t.DesPriceHeapBaseSk.Front()

	if elem != nil {
		orders := elem.Value.(*OrdersByTimeBaseSk)
		return orders.GetTop()
	} else {
		return nil
	}
}

func (t *AskPricesBaseSk) Set(order *comm.Order) {
	elem := t.DesPriceHeapBaseSk.Get(order.Price)
	if elem != nil {
		elem.Value.(*OrdersByTimeBaseSk).Set(order)
	} else {
		orders := NewOrdersByTimeBaseSk()
		orders.Set(order)
		t.DesPriceHeapBaseSk.Set(order.Price, orders)
	}
}

func (t *AskPricesBaseSk) Get(price float64) *OrdersByTimeBaseSk {
	elem := t.DesPriceHeapBaseSk.Get(price)
	if elem != nil {
		return elem.Value.(*OrdersByTimeBaseSk)
	} else {
		return nil
	}
}

func (t *AskPricesBaseSk) Remove(price float64) *OrdersByTimeBaseSk {
	elem := t.DesPriceHeapBaseSk.Remove(price)
	if elem != nil {
		return elem.Value.(*OrdersByTimeBaseSk)
	} else {
		return nil
	}
}

func (t *AskPricesBaseSk) Len() int {
	return int(t.DesPriceHeapBaseSk.Len())
}

func (t *AskPricesBaseSk) Size() int {
	unboundIterator := t.DesPriceHeapBaseSk.Iterator()

	sum := 0
	for unboundIterator != nil {
		sum += unboundIterator.Value.(*OrdersByTimeBaseSk).Len()
		unboundIterator = unboundIterator.Next()
	}
	return sum
}

func (t *AskPricesBaseSk) Dump() {
	fmt.Printf("======================Dump AskPricesBaseSk==========================\n")
	fmt.Printf("bids len = %d\n", t.DesPriceHeapBaseSk.Len())
	unboundIterator := t.DesPriceHeapBaseSk.Iterator()
	if unboundIterator == nil {
		return
	}

	count := 0
	for unboundIterator != nil {
		fmt.Printf("[%d] price = %.8f:\n", count, unboundIterator.Key().(float64))
		unboundIterator.Value.(*OrdersByTimeBaseSk).Dump()
		count++
		unboundIterator = unboundIterator.Next()
	}

	fmt.Printf("===================================================================\n")
}

///------------------------------------------------------------------
type TradePricesBaseSk struct {
	*BidPricesBaseSk
	*AskPricesBaseSk
}

func NewTradePricesBaseSk() *TradePricesBaseSk {
	o := new(TradePricesBaseSk)
	o.BidPricesBaseSk = NewBidPricesBaseSk()
	o.AskPricesBaseSk = NewAskPricesBaseSk()
	return o
}

func (t *TradePricesBaseSk) Dump() {
	fmt.Printf("======================Dump TradePricesBaseSk==========================\n")
	fmt.Printf("BidPricesBaseSk:\n")
	t.BidPricesBaseSk.Dump()
	fmt.Printf("AskPricesBaseSk:\n")
	t.AskPricesBaseSk.Dump()
	fmt.Printf("===================================================================\n")
}

///------------------------------------------------------------------
type DesTimeHeapBaseSk struct {
	*sk.SkipList
}

// func (t *DesTimeHeapBaseSk) Descending() bool {
// 	return false
// }

// func (t *DesTimeHeapBaseSk) Compare(lhs, rhs interface{}) bool {
// 	return lhs.(int64) > rhs.(int64)
// }

func NewDesTimeHeapBaseSk() *DesTimeHeapBaseSk {
	o := new(DesTimeHeapBaseSk)
	o.SkipList = sk.New(sk.Int64Ascending)
	return o
}

type OrdersByTimeBaseSk struct {
	*DesTimeHeapBaseSk
}

func NewOrdersByTimeBaseSk() *OrdersByTimeBaseSk {
	o := new(OrdersByTimeBaseSk)
	o.DesTimeHeapBaseSk = NewDesTimeHeapBaseSk()
	return o
}

func (t *OrdersByTimeBaseSk) Push(order *comm.Order) {
	t.Set(order)
}

func (t *OrdersByTimeBaseSk) Pop() *comm.Order {
	elem := t.DesTimeHeapBaseSk.Front()

	if elem != nil {
		timestamp := elem.Key().(int64)
		order := elem.Value.(*comm.Order)
		t.DesTimeHeapBaseSk.Remove(timestamp)
		return order
	} else {
		return nil
	}
}

func (t *OrdersByTimeBaseSk) GetTop() *comm.Order {
	elem := t.DesTimeHeapBaseSk.Front()

	if elem != nil {
		order := elem.Value.(*comm.Order)
		return order
	} else {
		return nil
	}
}

func (t *OrdersByTimeBaseSk) Set(order *comm.Order) {
	t.DesTimeHeapBaseSk.Set(order.Timestamp, order)
}

func (t *OrdersByTimeBaseSk) Get(timestamp int64) *comm.Order {
	elem := t.DesTimeHeapBaseSk.Get(timestamp)
	if elem != nil {
		return elem.Value.(*comm.Order)
	} else {
		return nil
	}
}

func (t *OrdersByTimeBaseSk) Remove(timestamp int64) *comm.Order {
	elem := t.DesTimeHeapBaseSk.Remove(timestamp)
	if elem != nil {
		return elem.Value.(*comm.Order)
	} else {
		return nil
	}
}

func (t *OrdersByTimeBaseSk) Len() int {
	return int(t.DesTimeHeapBaseSk.Len())
}

func (t *OrdersByTimeBaseSk) Dump() {
	fmt.Printf("====================Dump OrdersByTimeBaseSk========================\n")
	c := 0
	iterator := t.Iterator()
	for iterator != nil {
		fmt.Printf("\t[%d] time = %d, price = %.8f, volume = %.8f, tvolume = %.8f;\n",
			c,
			iterator.Key().(int64),
			iterator.Value.(*comm.Order).Price,
			iterator.Value.(*comm.Order).Volume,
			iterator.Value.(*comm.Order).TotalVolume,
		)
		iterator = iterator.Next()
		c++
	}

	fmt.Printf("===================================================================\n")
}

func (t *OrdersByTimeBaseSk) Pump() {

}
