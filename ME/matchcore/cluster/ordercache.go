package cluster

import (
	"fmt"
	"sync/atomic"
	"time"

	"../../comm"
	redis "../../db/use_redis"
)

///------------------------------------------------------------------
const (
	CACHE_SCAN_DURATION        = 1 * time.Second
	ORDER_STRUCT_EVALUATE_SIZE = 300
	ORDER_MEMORY_CACHE_SIZE    = 1000000000 //// should be configed occording ram size
	ORDER_MEMORY_CACHE_MAX_NUM = ORDER_MEMORY_CACHE_SIZE / ORDER_STRUCT_EVALUATE_SIZE
	ORDER_MEMORY_CACHE_MIN_NUM = (ORDER_MEMORY_CACHE_SIZE / ORDER_STRUCT_EVALUATE_SIZE) >> 1
	ORDER_MEMORY_CACHE_FIT_NUM = ORDER_MEMORY_CACHE_MIN_NUM + (ORDER_MEMORY_CACHE_MAX_NUM-ORDER_MEMORY_CACHE_MIN_NUM)>>1

	ORDER_CACHE_CONN_NAME_BID string = "cachesyncconn_bid"
	ORDER_CACHE_CONN_NAME_ASK string = "cachesyncconn_ask"
	MODULE_NAME_ORDERCACHE    string = "[MatchCore.Cluster.OrderCache]: "
)

type WorkStatus int

const (
	WorkStatus_Idle      = 0
	WorkStatus_ToCluster = 1
	WorkStatus_ToCache   = 2
)

type OrderCache struct {
	comm.IDOrderMap
	WorkStatus

	Len      int64
	ConnName string
	ZsetName string

	toCluter chan bool
	toCache  chan bool
}

func NewOrderCache(connName string, zsetName string) *OrderCache {
	o := new(OrderCache)
	o.IDOrderMap = *comm.NewIDOrderMap()
	o.WorkStatus = WorkStatus_Idle
	o.Len = 0
	o.ConnName = connName
	o.ZsetName = zsetName
	o.toCluter = make(chan bool)
	o.toCache = make(chan bool)

	go func() {
		for {
			select {
			case <-o.toCluter:
				if o.IsFull() && o.WorkStatus == WorkStatus_Idle {
					o.WorkStatus = WorkStatus_ToCluster
				}
			case <-o.toCache:
				if o.IsLack() && o.WorkStatus == WorkStatus_Idle {
					o.WorkStatus = WorkStatus_ToCache
				}
			}
		}
	}()
	go func() {
		for {
			comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_DEBUG, "Cache<->Cluster task running: [Len=%d, WorkStatus=%d]\n", o.Len, o.WorkStatus)
			if o.WorkStatus == WorkStatus_ToCluster {
				comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_TRACK, "Just to CacheToCluster: [Len=%d, WorkStatus=%d]\n", o.Len, o.WorkStatus)
				o.CacheToCluster()
			} else if o.WorkStatus == WorkStatus_ToCache {
				comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_TRACK, "Just to ClusterToCache: [Len=%d, WorkStatus=%d]\n", o.Len, o.WorkStatus)
				o.ClusterToCache()
			}
			time.Sleep(CACHE_SCAN_DURATION)
		}
	}()
	return o
}

func (t *OrderCache) AddOrderToCluster(order *comm.Order, index int64) {
	if index >= ORDER_MEMORY_CACHE_MAX_NUM {
		redis.RedisDbInstance().AddCacheOrderByNamedConn(t.ConnName, order)
		t.toCache <- true
		comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_DEBUG, "AddOrderToCluster , id = %d [Len=%d, WorkStatus=%d]\n", order.ID, t.Len, t.WorkStatus)
	}

	return
}

func (t *OrderCache) AddOrderToCache(order *comm.Order) {
	t.IDOrderMap.Set(order.ID, order)
	atomic.AddInt64(&t.Len, 1)
	t.toCluter <- true
	return
}

func (t *OrderCache) RemoveOrderFromCache(id int64) {
	_, ok := t.IDOrderMap.Get(id)
	if ok {
		t.IDOrderMap.Remove(id)
		atomic.AddInt64(&t.Len, -1)
		t.toCache <- true
	}

	return
}

func (t *OrderCache) GetOrder(id int64) *comm.Order {
	order, ok := t.IDOrderMap.Get(id)
	if !ok {
		order, _ = redis.RedisDbInstance().GetCacheOrder(id)
		comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_TRACK, "Order get from cluster, id = %d [Len=%d, WorkStatus=%d]\n", id, t.Len, t.WorkStatus)
	}

	return order
}

func (t *OrderCache) RemoveOrder(id int64) {
	_, ok := t.IDOrderMap.Get(id)
	if ok {
		t.IDOrderMap.Remove(id)
		atomic.AddInt64(&t.Len, -1)
		t.toCache <- true
		comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_DEBUG, "RemoveOrder from cache, id = %d [Len=%d, WorkStatus=%d]\n", id, t.Len, t.WorkStatus)
	} else {
		redis.RedisDbInstance().RmCacheOrder(id)
		comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_DEBUG, "RemoveOrder from cluster, id = %d [Len=%d, WorkStatus=%d]\n", id, t.Len, t.WorkStatus)
	}

	return
}

func (t *OrderCache) IsLack() bool {
	return t.Len < ORDER_MEMORY_CACHE_MIN_NUM
}

func (t *OrderCache) IsFull() bool {
	return t.Len >= ORDER_MEMORY_CACHE_MAX_NUM
}

func (t *OrderCache) CacheToCluster() {

	ids := redis.RedisDbInstance().ZSetGetRangeInt64sByNamedConn(t.ConnName, t.ZsetName, int64(ORDER_MEMORY_CACHE_FIT_NUM-1), int64(ORDER_MEMORY_CACHE_MAX_NUM-1))
	for _, id := range ids {
		if order, ok := t.IDOrderMap.Get(id); ok {
			redis.RedisDbInstance().AddCacheOrderByNamedConn(t.ConnName, order)
		}
		t.IDOrderMap.Remove(id)
		l := atomic.AddInt64(&t.Len, -1)
		if l <= int64(ORDER_MEMORY_CACHE_FIT_NUM) {
			break
		}
		comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_DEBUG, "CacheToCluster, id = %d [Len=%d, WorkStatus=%d]\n", id, t.Len, t.WorkStatus)
	}

	if t.IsFull() {
		panic(fmt.Errorf("Product fster than consume from cache!!!"))
	}

	t.WorkStatus = WorkStatus_Idle
}

func (t *OrderCache) ClusterToCache() {

	ids := redis.RedisDbInstance().ZSetGetRangeInt64sByNamedConn(t.ConnName, t.ZsetName, int64(ORDER_MEMORY_CACHE_MIN_NUM-1), int64(ORDER_MEMORY_CACHE_FIT_NUM-1))
	for _, id := range ids {
		order, _ := redis.RedisDbInstance().GetCacheOrderByNamedConn(t.ConnName, id)
		if order != nil {
			redis.RedisDbInstance().RmCacheOrderByNamedConn(t.ConnName, id)
			t.IDOrderMap.Set(id, order)
			l := atomic.AddInt64(&t.Len, 1)
			if l >= int64(ORDER_MEMORY_CACHE_FIT_NUM) {
				break
			}
			comm.DebugPrintf(MODULE_NAME_ORDERCACHE, comm.LOG_LEVEL_DEBUG, "ClusterToCache, id = %d [Len=%d, WorkStatus=%d]\n", id, t.Len, t.WorkStatus)
		}
	}

	t.WorkStatus = WorkStatus_Idle
}
