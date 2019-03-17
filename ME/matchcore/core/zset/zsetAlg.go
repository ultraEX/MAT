package zset

import (
	"fmt"

	"../../../comm"
	redis "../../../db/use_redis"
)

///------------------------------------------------------------------
const (
	PRICE_MULTI_FACTOR float64 = 10000000000000000000
	PRICE_MAX_VALUE    float64 = 10000000000000000000
	TIME_DIV_FACTOR    float64 = 10000000000000000000

	BID_ORDER_CONTAINER_KEY string = "bid_orders_container"
	ASK_ORDER_CONTAINER_KEY string = "ask_orders_container"
)

/// ME memory:
/// Redis cache:
type OrderContainer struct {
	comm.IDOrderMap
}

func NewOrderContainer() *OrderContainer {
	o := new(OrderContainer)
	o.IDOrderMap = *comm.NewIDOrderMap()
	return o
}

func (t *OrderContainer) Push(order *comm.Order) {

	t.IDOrderMap.Set(order.ID, order)

	if order.AorB == comm.TradeType_BID {
		score := (PRICE_MAX_VALUE - order.Price*PRICE_MULTI_FACTOR) + float64(order.Timestamp)
		redis.RedisDbInstance().ZSetAddInt64(BID_ORDER_CONTAINER_KEY, score, order.ID)
	} else if order.AorB == comm.TradeType_ASK {
		score := order.Price*PRICE_MULTI_FACTOR + float64(order.Timestamp)
		redis.RedisDbInstance().ZSetAddInt64(ASK_ORDER_CONTAINER_KEY, score, order.ID)
	} else {
		panic(fmt.Errorf("OrderContainer.Push input illegal type = %d", order.AorB))
	}
}

func (t *OrderContainer) Pop(aorb comm.TradeType) (order *comm.Order) {
	if aorb == comm.TradeType_BID {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(BID_ORDER_CONTAINER_KEY, 0); ok {
			redis.RedisDbInstance().ZSetRemoveInt64(BID_ORDER_CONTAINER_KEY, id)
			order = t.Get(id)
			t.IDOrderMap.Remove(id)
			return order
		}
	} else if aorb == comm.TradeType_ASK {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(ASK_ORDER_CONTAINER_KEY, 0); ok {
			redis.RedisDbInstance().ZSetRemoveInt64(ASK_ORDER_CONTAINER_KEY, id)
			order = t.Get(id)
			t.IDOrderMap.Remove(id)
			return order
		}
	} else {
		panic(fmt.Errorf("OrderContainer.Pop input illegal type = %d", aorb))
	}

	return nil
}

func (t *OrderContainer) GetTop(aorb comm.TradeType) *comm.Order {
	if aorb == comm.TradeType_BID {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(BID_ORDER_CONTAINER_KEY, 0); ok {
			return t.Get(id)
		}
	} else if aorb == comm.TradeType_ASK {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(ASK_ORDER_CONTAINER_KEY, 0); ok {
			return t.Get(id)
		}
	} else {
		panic(fmt.Errorf("OrderContainer.GetTop input illegal type = %d", aorb))
	}

	return nil
}

func (t *OrderContainer) Get(id int64) *comm.Order {
	order, ok := t.IDOrderMap.Get(id)
	if ok {
		return order
	} else {
		return nil
	}
}

func (t *OrderContainer) GetAll(aorb comm.TradeType) []int64 {
	if aorb == comm.TradeType_BID {
		return redis.RedisDbInstance().ZSetGetAll(BID_ORDER_CONTAINER_KEY)
	} else if aorb == comm.TradeType_ASK {
		return redis.RedisDbInstance().ZSetGetAll(ASK_ORDER_CONTAINER_KEY)
	} else {
		panic(fmt.Errorf("OrderContainer.GetAll input illegal type = %d", aorb))
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

func (t *OrderContainer) BidSize() int64 {
	return redis.RedisDbInstance().ZSetGetSize(BID_ORDER_CONTAINER_KEY)
}

func (t *OrderContainer) AskSize() int64 {
	return redis.RedisDbInstance().ZSetGetSize(ASK_ORDER_CONTAINER_KEY)
}

func (t *OrderContainer) TheSize() int64 {
	return t.IDOrderMap.Len()
}
