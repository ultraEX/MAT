package structmap

import (
	"fmt"
	"sync"

	"../../../comm"
	sk "../struct"
)

///------------------------------------------------------------------
type LSkOrder struct {
	*comm.Order
	score string
}

///------------------------------------------------------------------
type AscendScoreBaseLazySk struct {
	*sk.LazySkipList
}

func NewAscendScoreBaseLazySk() *AscendScoreBaseLazySk {
	o := new(AscendScoreBaseLazySk)
	o.LazySkipList = sk.NewLazySkipList(func(v1, v2 interface{}) bool { return v1.(*LSkOrder).score < v2.(*LSkOrder).score }, func(v1, v2 interface{}) bool { return v1.(*LSkOrder).score == v2.(*LSkOrder).score })
	return o
}

type OrdersByScoreBaseLazySk struct {
	*AscendScoreBaseLazySk
	sync.Mutex
}

func NewOrdersByScoreBaseLazySk() *OrdersByScoreBaseLazySk {
	o := new(OrdersByScoreBaseLazySk)
	o.AscendScoreBaseLazySk = NewAscendScoreBaseLazySk()
	return o
}

func (t *OrdersByScoreBaseLazySk) Push(order *comm.Order) {
	t.Set(order)
}

func (t *OrdersByScoreBaseLazySk) Pop() *comm.Order {
retry:
	elem := t.AscendScoreBaseLazySk.Front()
	if elem != nil {
		lskOrder := elem.(*LSkOrder)
		// score := lskOrder.score
		order := lskOrder.Order
		if t.AscendScoreBaseLazySk.Remove(lskOrder) {
			return order
		} else {
			goto retry
		}
	} else {
		return nil
	}
}

func (t *OrdersByScoreBaseLazySk) GetTop() *comm.Order {
	elem := t.AscendScoreBaseLazySk.Front()

	if elem != nil {
		order := elem.(*LSkOrder).Order
		return order
	} else {
		return nil
	}
}

func (t *OrdersByScoreBaseLazySk) Set(order *comm.Order) {
	if order == nil {
		panic(fmt.Errorf("OrdersByScoreBaseLazySk.Set input nil order"))
	}

	var score string
	if order.AorB == comm.TradeType_BID {
		score = comm.BidKeyStr(order.Price, order.ID)
	}
	if order.AorB == comm.TradeType_ASK {
		score = comm.AskKeyStr(order.Price, order.ID)
	}

	var lskOrder LSkOrder = LSkOrder{Order: order, score: score}
	if !t.AscendScoreBaseLazySk.Add(&lskOrder) {
		panic(fmt.Errorf("OrdersByScoreBaseLazySk.Set input exist order, id = %d", order.ID))
	}
}

func (t *OrdersByScoreBaseLazySk) Len() int {
	return int(t.AscendScoreBaseLazySk.Len())
}

func (t *OrdersByScoreBaseLazySk) Dump() {
	fmt.Printf("====================Dump OrdersByScoreBaseLazySk========================\n")
	c := 0
	iterator := t.IteratorLSk()
	for {
		if v, ok := iterator.Next(); ok {
			fmt.Printf("\t[%d] id = %d, type = %s, price = %.8f, time = %d, volume = %.8f, tvolume = %.8f, status = %s\n",
				c,
				v.(*LSkOrder).ID,
				v.(*LSkOrder).AorB,
				v.(*LSkOrder).Price,
				v.(*LSkOrder).Timestamp,
				v.(*LSkOrder).Volume,
				v.(*LSkOrder).TotalVolume,
				v.(*LSkOrder).Status,
			)
			c++
		} else {
			break
		}
	}

	fmt.Printf("===================================================================\n")
}

func (t *OrdersByScoreBaseLazySk) Pump() {
	fmt.Printf("====================Pump OrdersByScoreBaseLazySk========================\n")

	fmt.Printf("===================================================================\n")
}
