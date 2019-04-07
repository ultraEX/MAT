package zset

import (
	"fmt"
	"sync"

	"../../../comm"
	redis "../../../db/use_redis"
)

///------------------------------------------------------------------
/// ME memory:
/// Redis cache:
type OrderContainer struct {
	comm.IDOrderMap

	ConMutex *sync.RWMutex
}

func NewOrderContainer() *OrderContainer {
	o := new(OrderContainer)
	o.IDOrderMap = *comm.NewIDOrderMap()
	o.ConMutex = new(sync.RWMutex)
	return o
}

func (t *OrderContainer) Push(order *comm.Order) {
	t.ConMutex.Lock()
	defer t.ConMutex.Unlock()

	t.IDOrderMap.Set(order.ID, order)

	if order.AorB == comm.TradeType_BID {
		score := comm.BidScore(order.Price, order.Timestamp)
		redis.RedisDbInstance().ZSetAddInt64(redis.BID_ORDER_CONTAINER_KEY, score, order.ID)
	} else if order.AorB == comm.TradeType_ASK {
		score := comm.AskScore(order.Price, order.Timestamp)
		redis.RedisDbInstance().ZSetAddInt64(redis.ASK_ORDER_CONTAINER_KEY, score, order.ID)
	} else {
		panic(fmt.Errorf("zset.OrderContainer.Push input illegal type = %d", order.AorB))
	}
}

func (t *OrderContainer) Pop(aorb comm.TradeType) (order *comm.Order) {
	t.ConMutex.Lock()
	defer t.ConMutex.Unlock()

	if aorb == comm.TradeType_BID {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.BID_ORDER_CONTAINER_KEY, 0); ok {
			redis.RedisDbInstance().ZSetRemoveInt64(redis.BID_ORDER_CONTAINER_KEY, id)
			order = t.Get(id)
			t.IDOrderMap.Remove(id)
			return order
		}
	} else if aorb == comm.TradeType_ASK {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.ASK_ORDER_CONTAINER_KEY, 0); ok {
			redis.RedisDbInstance().ZSetRemoveInt64(redis.ASK_ORDER_CONTAINER_KEY, id)
			order = t.Get(id)
			t.IDOrderMap.Remove(id)
			return order
		}
	} else {
		panic(fmt.Errorf("zset.OrderContainer.Pop input illegal type = %d", aorb))
	}

	return nil
}

func (t *OrderContainer) GetTop(aorb comm.TradeType) *comm.Order {
	t.ConMutex.RLock()
	defer t.ConMutex.RUnlock()

	if aorb == comm.TradeType_BID {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.BID_ORDER_CONTAINER_KEY, 0); ok {
			return t.Get(id)
		}
	} else if aorb == comm.TradeType_ASK {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.ASK_ORDER_CONTAINER_KEY, 0); ok {
			return t.Get(id)
		}
	} else {
		panic(fmt.Errorf("zset.OrderContainer.GetTop input illegal type = %d", aorb))
	}

	return nil
}

func (t *OrderContainer) Get(id int64) *comm.Order {
	order, ok := t.IDOrderMap.Get(id)
	if ok {
		return order
	} else {
		fmt.Printf("OrderContainer.Get(id=%d) nil\n", id)
		return nil
	}
}

func (t *OrderContainer) GetAll(aorb comm.TradeType) []int64 {
	t.ConMutex.RLock()
	defer t.ConMutex.RUnlock()

	if aorb == comm.TradeType_BID {
		return redis.RedisDbInstance().ZSetGetAll(redis.BID_ORDER_CONTAINER_KEY)
	} else if aorb == comm.TradeType_ASK {
		return redis.RedisDbInstance().ZSetGetAll(redis.ASK_ORDER_CONTAINER_KEY)
	} else {
		panic(fmt.Errorf("zset.OrderContainer.GetAll input illegal type = %d", aorb))
	}

}

func (t *OrderContainer) Dump() {
	fmt.Printf("======================Dump zset.OrderContainer==========================\n")
	t.IDOrderMap.Dump()
	ids := t.GetAll(comm.TradeType_BID)
	fmt.Printf("-------------------Dump bid orders from redis zset ids-------------------\n")
	var order *comm.Order
	for c, id := range ids {
		order = t.Get(id)
		if order != nil {
			fmt.Printf("[%d]: id = %d, price = %f, time = %d, volume = %f, totalvolume = %f\n",
				c,
				order.ID,
				order.Price,
				order.Timestamp,
				order.Volume,
				order.TotalVolume,
			)
		}
	}
	ids = t.GetAll(comm.TradeType_ASK)
	fmt.Printf("------------------Dump ask orders from redis zset ids-------------------\n")
	for c, id := range ids {
		order = t.Get(id)
		if order != nil {
			fmt.Printf("[%d]: id = %d, price = %f, time = %d, volume = %f, totalvolume = %f\n",
				c,
				order.ID,
				order.Price,
				order.Timestamp,
				order.Volume,
				order.TotalVolume,
			)
		}
	}
	fmt.Printf("===================================================================\n")
}

func (t *OrderContainer) Pump() {
	fmt.Printf("======================Pump zset.OrderContainer==========================\n")
	fmt.Printf("===================================================================\n")
}

func (t *OrderContainer) BidSize() int64 {
	return redis.RedisDbInstance().ZSetGetSize(redis.BID_ORDER_CONTAINER_KEY)
}

func (t *OrderContainer) AskSize() int64 {
	return redis.RedisDbInstance().ZSetGetSize(redis.ASK_ORDER_CONTAINER_KEY)
}

func (t *OrderContainer) TheSize() int64 {
	return t.IDOrderMap.Len()
}
