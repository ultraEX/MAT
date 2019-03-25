package zsetcluster

import (
	"fmt"

	"../../../comm"
	redis "../../../db/use_redis"
	"../../cluster"
)

///------------------------------------------------------------------

/// ME memory:
/// Redis cache:
type OrderContainer struct {
	bidOrders *cluster.OrderCache
	askOrders *cluster.OrderCache
}

func NewOrderContainer() *OrderContainer {
	o := new(OrderContainer)
	o.bidOrders = cluster.NewOrderCache(cluster.ORDER_CACHE_CONN_NAME_BID, redis.BID_ORDER_CONTAINER_KEY)
	o.askOrders = cluster.NewOrderCache(cluster.ORDER_CACHE_CONN_NAME_ASK, redis.ASK_ORDER_CONTAINER_KEY)
	return o
}

func (t *OrderContainer) Push(order *comm.Order) {

	if order.AorB == comm.TradeType_BID {
		t.bidOrders.AddOrderToCache(order)
		score := comm.BidScore(order.Price, order.Timestamp)
		index := redis.RedisDbInstance().ZSetAddInt64ForIndex(redis.BID_ORDER_CONTAINER_KEY, score, order.ID)
		t.bidOrders.AddOrderToCluster(order, index)
	} else if order.AorB == comm.TradeType_ASK {
		t.askOrders.AddOrderToCache(order)
		score := comm.AskScore(order.Price, order.Timestamp)
		index := redis.RedisDbInstance().ZSetAddInt64ForIndex(redis.ASK_ORDER_CONTAINER_KEY, score, order.ID)
		t.askOrders.AddOrderToCluster(order, index)
	} else {
		panic(fmt.Errorf("zsetcluster.OrderContainer.Push input illegal type = %d", order.AorB))
	}
}

func (t *OrderContainer) Pop(aorb comm.TradeType) (order *comm.Order) {
	var yes bool = false
	if aorb == comm.TradeType_BID {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.BID_ORDER_CONTAINER_KEY, 0); ok {
			redis.RedisDbInstance().ZSetRemoveInt64(redis.BID_ORDER_CONTAINER_KEY, id)

			if order, yes = t.bidOrders.Get(id); !yes {
				panic(fmt.Errorf("zsetcluster.OrderContainer.Pop id=%d order not exist.", id))
			}
			t.bidOrders.RemoveOrder(id)
			return order
		}
	} else if aorb == comm.TradeType_ASK {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.ASK_ORDER_CONTAINER_KEY, 0); ok {
			redis.RedisDbInstance().ZSetRemoveInt64(redis.ASK_ORDER_CONTAINER_KEY, id)

			if order, yes = t.askOrders.Get(id); !yes {
				panic(fmt.Errorf("zsetcluster.OrderContainer.Pop id=%d order not exist.", id))
			}
			t.askOrders.RemoveOrder(id)
			return order
		}
	} else {
		panic(fmt.Errorf("zsetcluster.OrderContainer.Pop input illegal type = %d", aorb))
	}

	return nil
}

func (t *OrderContainer) GetTop(aorb comm.TradeType) *comm.Order {
	if aorb == comm.TradeType_BID {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.BID_ORDER_CONTAINER_KEY, 0); ok {
			order, yes := t.bidOrders.Get(id)
			if !yes {
				panic(fmt.Errorf("zsetcluster.OrderContainer.GetTop id=%d order not exist  in cache.", id))
			}
			return order
		}
	} else if aorb == comm.TradeType_ASK {
		if id, ok := redis.RedisDbInstance().ZSetGetInt64(redis.ASK_ORDER_CONTAINER_KEY, 0); ok {
			order, yes := t.askOrders.Get(id)
			if !yes {
				panic(fmt.Errorf("zsetcluster.OrderContainer.GetTop id=%d order not exist in cache.", id))
			}
			return order
		}
	} else {
		panic(fmt.Errorf("zsetcluster.OrderContainer.GetTop input illegal type = %d", aorb))
	}

	return nil
}

func (t *OrderContainer) Get(id int64) *comm.Order {
	order := t.bidOrders.GetOrder(id)
	if order != nil {
		return order
	}

	order = t.askOrders.GetOrder(id)
	if order != nil {
		return order
	}

	return nil
}

func (t *OrderContainer) GetAll(aorb comm.TradeType) []int64 {
	if aorb == comm.TradeType_BID {
		return redis.RedisDbInstance().ZSetGetAll(redis.BID_ORDER_CONTAINER_KEY)
	} else if aorb == comm.TradeType_ASK {
		return redis.RedisDbInstance().ZSetGetAll(redis.ASK_ORDER_CONTAINER_KEY)
	} else {
		panic(fmt.Errorf("zsetcluster.OrderContainer.GetAll input illegal type = %d", aorb))
	}

}

func (t *OrderContainer) Dump() {
	fmt.Printf("======================Dump zsetcluster.OrderContainer==========================\n")
	fmt.Printf("----------------------Dump bid orders from redis zset ids----------------------\n")
	ids := t.GetAll(comm.TradeType_BID)
	t.bidOrders.Dump()
	var order *comm.Order
	for c, id := range ids {
		order = t.bidOrders.GetOrder(id)
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

	fmt.Printf("---------------------Dump ask orders from redis zset ids----------------------\n")
	ids = t.GetAll(comm.TradeType_ASK)
	t.askOrders.Dump()
	for c, id := range ids {
		order = t.askOrders.GetOrder(id)
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
	fmt.Printf("===============================================================================\n")
}

func (t *OrderContainer) BidSize() int64 {
	return redis.RedisDbInstance().ZSetGetSize(redis.BID_ORDER_CONTAINER_KEY)
}

func (t *OrderContainer) AskSize() int64 {
	return redis.RedisDbInstance().ZSetGetSize(redis.ASK_ORDER_CONTAINER_KEY)
}

func (t *OrderContainer) TheSize() int64 {
	return t.bidOrders.Len + t.askOrders.Len
}
