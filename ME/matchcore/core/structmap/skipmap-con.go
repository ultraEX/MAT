package structmap

import (
	"fmt"
	"unsafe"

	"../../../comm"
	sk "../struct"
)

///------------------------------------------------------------------

///------------------------------------------------------------------
type AscendScoreBaseConSk struct {
	*sk.Skiplist
}

func NewAscendScoreBaseConSk() *AscendScoreBaseConSk {
	o := new(AscendScoreBaseConSk)
	o.Skiplist = sk.NewSkiplist(32)
	return o
}

type OrdersByScoreBaseConSk struct {
	*AscendScoreBaseConSk
}

func NewOrdersByScoreBaseConSk() *OrdersByScoreBaseConSk {
	o := new(OrdersByScoreBaseConSk)
	o.AscendScoreBaseConSk = NewAscendScoreBaseConSk()
	return o
}

func (t *OrdersByScoreBaseConSk) Push(order *comm.Order) {
	t.Set(order)
}

func (t *OrdersByScoreBaseConSk) Pop() *comm.Order {
	it, _ := sk.NewIterator(t.Skiplist, nil, nil)
	if it.Next() {
		key, _ := it.NextNode()
		value, _ := t.AscendScoreBaseConSk.Remove(key)
		return (*comm.Order)(value)
	} else {
		return nil
	}
}

func (t *OrdersByScoreBaseConSk) GetTop() *comm.Order {
	it, _ := sk.NewIterator(t.Skiplist, nil, nil)
	if it.Next() {
		_, value := it.NextNode()
		return (*comm.Order)(value)
	} else {
		return nil
	}
}

func (t *OrdersByScoreBaseConSk) Set(order *comm.Order) {
	if order == nil {
		panic(fmt.Errorf("OrdersByScoreBaseConSk.Set input nil order"))
	}

	var score string
	if order.AorB == comm.TradeType_BID {
		score = comm.BidKeyStr(order.Price, order.ID)
	}
	if order.AorB == comm.TradeType_ASK {
		score = comm.AskKeyStr(order.Price, order.ID)
	}

	t.AscendScoreBaseConSk.Put([]byte(score), unsafe.Pointer(order))
}

// func (t *OrdersByScoreBaseConSk) Get(score float64) *comm.Order {
// 	key := []byte(strconv.FormatFloat(score, 'f', -1, 64))
// 	elem, _ := t.AscendScoreBaseConSk.Get(key)
// 	if elem != nil {
// 		return (*comm.Order)(elem)
// 	} else {
// 		return nil
// 	}
// }

// func (t *OrdersByScoreBaseConSk) Remove(score float64) *comm.Order {
// 	key := []byte(strconv.FormatFloat(score, 'f', -1, 64))
// 	elem, _ := t.AscendScoreBaseConSk.Remove(key)
// 	if elem != nil {
// 		return (*comm.Order)(elem)
// 	} else {
// 		return nil
// 	}
// }

func (t *OrdersByScoreBaseConSk) Len() int {
	return int(t.AscendScoreBaseConSk.Count())
}

func (t *OrdersByScoreBaseConSk) Dump() {
	fmt.Printf("====================Dump OrdersByScoreBaseConSk========================\n")
	c := 0
	iterator, _ := sk.NewIterator(t.Skiplist, nil, nil)

	for iterator.Next() {
		key, value := iterator.NextNode()
		fmt.Printf("\t[%d]key=%s: id = %d, type = %s, price = %.8f, time = %d, volume = %.8f, tvolume = %.8f, status = %s\n",
			c,
			key,
			(*comm.Order)(value).ID,
			(*comm.Order)(value).AorB,
			(*comm.Order)(value).Price,
			(*comm.Order)(value).Timestamp,
			(*comm.Order)(value).Volume,
			(*comm.Order)(value).TotalVolume,
			(*comm.Order)(value).Status,
		)
		c++
	}

	fmt.Printf("===================================================================\n")
}
