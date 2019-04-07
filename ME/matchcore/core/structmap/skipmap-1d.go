package structmap

import (
	"fmt"
	"sync"

	"../../../comm"
	sk "../struct"
)

///------------------------------------------------------------------
///------------------------------------------------------------------
type AscendScoreBase1dSk struct {
	*sk.SkipList
}

func NewAscendScoreBase1dSk() *AscendScoreBase1dSk {
	o := new(AscendScoreBase1dSk)
	o.SkipList = sk.New(sk.Float64Ascending)
	return o
}

type OrdersByScoreBase1dSk struct {
	*AscendScoreBase1dSk
	sync.Mutex
}

func NewOrdersByScoreBase1dSk() *OrdersByScoreBase1dSk {
	o := new(OrdersByScoreBase1dSk)
	o.AscendScoreBase1dSk = NewAscendScoreBase1dSk()
	return o
}

func (t *OrdersByScoreBase1dSk) Push(order *comm.Order) {
	t.Set(order)
}

func (t *OrdersByScoreBase1dSk) Pop() *comm.Order {
	t.Lock()
	defer t.Unlock()
	elem := t.AscendScoreBase1dSk.Front()

	if elem != nil {
		score := elem.Key().(float64)
		order := elem.Value.(*comm.Order)
		// t.AscendScoreBase1dSk.Remove(score)
		t.AscendScoreBase1dSk.RemoveCon(score)
		return order
	} else {
		return nil
	}
}

func (t *OrdersByScoreBase1dSk) GetTop() *comm.Order {
	elem := t.AscendScoreBase1dSk.Front()

	if elem != nil {
		order := elem.Value.(*comm.Order)
		return order
	} else {
		return nil
	}
}

func (t *OrdersByScoreBase1dSk) Set(order *comm.Order) {
	if order == nil {
		panic(fmt.Errorf("OrdersByScoreBase1dSk.Set input nil order"))
	}

	var score float64
	if order.AorB == comm.TradeType_BID {
		score = comm.BidKey(order.Price, order.Timestamp, order.ID)
	}
	if order.AorB == comm.TradeType_ASK {
		score = comm.AskKey(order.Price, order.Timestamp, order.ID)
	}
	// t.AscendScoreBase1dSk.Set(score, order)
	t.Lock()
	defer t.Unlock()
	t.AscendScoreBase1dSk.SetCon(score, order)
}

func (t *OrdersByScoreBase1dSk) Len() int {
	return int(t.AscendScoreBase1dSk.Len())
}

func (t *OrdersByScoreBase1dSk) Dump() {
	fmt.Printf("====================Dump OrdersByScoreBase1dSk========================\n")
	c := 0
	iterator := t.Iterator()
	for iterator != nil {
		fmt.Printf("\t[%d] id = %d, type = %s, price = %.8f, time = %d, volume = %.8f, tvolume = %.8f, status = %s\n",
			c,
			iterator.Value.(*comm.Order).ID,
			iterator.Value.(*comm.Order).AorB,
			iterator.Value.(*comm.Order).Price,
			iterator.Value.(*comm.Order).Timestamp,
			iterator.Value.(*comm.Order).Volume,
			iterator.Value.(*comm.Order).TotalVolume,
			iterator.Value.(*comm.Order).Status,
		)
		iterator = iterator.Next()
		c++
	}

	fmt.Printf("===================================================================\n")
}

func (t *OrdersByScoreBase1dSk) Pump() {
	fmt.Printf("====================Pump OrdersByScoreBase1dSk========================\n")

	fmt.Printf("===================================================================\n")
}
