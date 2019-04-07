package structmap

import (
	"container/heap"
	"fmt"
	"sync"

	"../../../comm"
)

///------------------------------------------------------------------

///------------------------------------------------------------------
type HMOrder struct {
	*comm.Order
	score string
}
type AscendScoreBaseHM struct {
	orders []*HMOrder
}

func NewAscendScoreBaseHM() *AscendScoreBaseHM {
	return new(AscendScoreBaseHM)
}

func (t *AscendScoreBaseHM) Len() int {
	return len(t.orders)
}

func (t *AscendScoreBaseHM) Less(i, j int) bool {
	return t.orders[i].score < t.orders[j].score
}

func (t *AscendScoreBaseHM) Swap(i, j int) {
	tmp := t.orders[i]
	t.orders[i] = t.orders[j]
	t.orders[j] = tmp
}

func (t *AscendScoreBaseHM) Push(x interface{}) {
	t.orders = append(t.orders, x.(*HMOrder))
}

func (t *AscendScoreBaseHM) Pop() (ret interface{}) {
	l := len(t.orders)
	t.orders, ret = t.orders[:l-1], t.orders[l-1]
	return
}

func (t *AscendScoreBaseHM) GetTop() (ret interface{}) {
	l := len(t.orders)
	if l == 0 {
		return nil
	} else {
		return t.orders[0]
	}
}

type OrdersByScoreBaseHM struct {
	*AscendScoreBaseHM
	ConMutex sync.Mutex
}

func NewOrdersByScoreBaseHM() *OrdersByScoreBaseHM {
	o := new(OrdersByScoreBaseHM)
	o.AscendScoreBaseHM = NewAscendScoreBaseHM()
	heap.Init(o.AscendScoreBaseHM)
	return o
}

func (t *OrdersByScoreBaseHM) Push(order *comm.Order) {
	t.Set(order)
}

func (t *OrdersByScoreBaseHM) Pop() *comm.Order {
	t.ConMutex.Lock()
	hmOrder := heap.Pop(t.AscendScoreBaseHM).(*HMOrder)
	t.ConMutex.Unlock()

	return hmOrder.Order
}

func (t *OrdersByScoreBaseHM) GetTop() *comm.Order {
	orderItf := t.AscendScoreBaseHM.GetTop()
	if orderItf != nil {
		hmOrder := orderItf.(*HMOrder)
		return hmOrder.Order
	} else {
		return nil
	}
}

func (t *OrdersByScoreBaseHM) Set(order *comm.Order) {
	var score string
	if order.AorB == comm.TradeType_BID {
		score = comm.BidKeyStr(order.Price, order.ID)
	}
	if order.AorB == comm.TradeType_ASK {
		score = comm.AskKeyStr(order.Price, order.ID)
	}
	hmOrder := HMOrder{Order: order, score: score}

	t.ConMutex.Lock()
	defer t.ConMutex.Unlock()

	heap.Push(t.AscendScoreBaseHM, &hmOrder)
}

func (t *OrdersByScoreBaseHM) Len() int {
	return t.AscendScoreBaseHM.Len()
}

func (t *OrdersByScoreBaseHM) Dump() {
	fmt.Printf("====================Dump OrdersByScoreBaseHM========================\n")
	fmt.Printf("order num = %d\n", t.Len())
	for c, hm := range t.AscendScoreBaseHM.orders {
		fmt.Printf("[%d] score = %s; order: id=%d; price=%.8f, time=%d, volume=%.8f, total=%.8f\n",
			c,
			hm.score,
			hm.Order.ID,
			hm.Order.Price,
			hm.Order.Timestamp,
			hm.Order.Volume,
			hm.Order.TotalVolume,
		)
	}

	fmt.Printf("===================================================================\n")
}

func (t *OrdersByScoreBaseHM) Pump() {
	fmt.Printf("====================Pump OrdersByScoreBaseHM========================\n")
	fmt.Printf("order num = %d\n", t.Len())

	c := 0
	for t.Len() > 0 {
		t.ConMutex.Lock()
		hm := heap.Pop(t.AscendScoreBaseHM).(*HMOrder)
		t.ConMutex.Unlock()

		fmt.Printf("[%d] score = %s; order: id=%d; price=%.8f, time=%d, volume=%.8f, total=%.8f\n",
			c,
			hm.score,
			hm.Order.ID,
			hm.Order.Price,
			hm.Order.Timestamp,
			hm.Order.Volume,
			hm.Order.TotalVolume,
		)
		c++
	}

	fmt.Printf("===================================================================\n")
}
