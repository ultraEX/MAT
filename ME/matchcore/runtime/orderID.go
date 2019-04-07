package runtime

import (
	"sync/atomic"
	"time"
)

type OrderID struct {
	ID_New   int64
	ID_Start int64
}

func NewOrderID(pre int) *OrderID {
	o := new(OrderID)

	o.ID_Start = int64(pre&0x3ff)<<53 | (time.Now().UnixNano() & (^0 >> 10))
	o.ID_New = o.ID_Start
	return o
}

func (t *OrderID) CreateID() int64 {
	id := atomic.AddInt64(&t.ID_New, 1)
	return id
}
