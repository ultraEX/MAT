package use_redis

import (
	"fmt"
	"strconv"

	. "../../comm"
	"github.com/gomodule/redigo/redis"
)

type rOrder struct {
	ID          string `redis:"ID"`
	Who         string `redis:"Who"`
	AorB        string `redis:"AorB"`
	Symbol      string `redis:"Symbol"`
	Timestamp   string `redis:"Timestamp"`
	Price       string `redis:"Price"`
	Volume      string `redis:"Volume"`
	TotalVolume string `redis:"TotalVolume"`
	Fee         string `redis:"Fee"`
	Status      string `redis:"Status"`
}

//---------------------------------------------------------------------------------------------------
/// Use Redis Set to store wait process order list
/// Use Redis Hash to store order detail info
func (t *RedisDb) AddOrderToSet(order *Order) error {
	/// use Hash to store order detail
	m := map[string]string{
		"ID":          strconv.FormatInt(order.ID, 10),
		"Who":         order.Who,
		"AorB":        strconv.FormatInt(int64(order.AorB), 10),
		"Symbol":      order.Symbol,
		"Timestamp":   strconv.FormatInt(order.Timestamp, 10),
		"Price":       strconv.FormatFloat(order.Price, 'f', -1, 64),
		"Volume":      strconv.FormatFloat(order.Volume, 'f', -1, 64),
		"TotalVolume": strconv.FormatFloat(order.TotalVolume, 'f', -1, 64),
		"Fee":         strconv.FormatFloat(order.Fee, 'f', -1, 64),
		"Status":      strconv.FormatInt(int64(order.Status), 10),
	}

	err := t.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.Send("SADD", orderSetKey(order.Symbol), orderHashKey(order.Who, order.ID))
	if err != nil {
		panic(err)
	}
	err = t.Send("HMSET", redis.Args{}.Add(orderHashKey(order.Who, order.ID)).AddFlat(m)...)
	if err != nil {
		panic(err)
	}
	_, err = t.Do("EXEC")
	if err != nil {
		panic(err)
	}

	return nil
}

func (t *RedisDb) AddCacheOrderByNamedConn(name string, order *Order) error {
	/// use Hash to store order detail
	m := map[string]string{
		"ID":          strconv.FormatInt(order.ID, 10),
		"Who":         order.Who,
		"AorB":        strconv.FormatInt(int64(order.AorB), 10),
		"Symbol":      order.Symbol,
		"Timestamp":   strconv.FormatInt(order.Timestamp, 10),
		"Price":       strconv.FormatFloat(order.Price, 'f', -1, 64),
		"Volume":      strconv.FormatFloat(order.Volume, 'f', -1, 64),
		"TotalVolume": strconv.FormatFloat(order.TotalVolume, 'f', -1, 64),
		"Fee":         strconv.FormatFloat(order.Fee, 'f', -1, 64),
		"Status":      strconv.FormatInt(int64(order.Status), 10),
	}

	_, err := t.GetNamedLongConn(name).Do("HMSET", redis.Args{}.Add(cacheOrderHashKey(order.ID)).AddFlat(m)...)
	if err != nil {
		panic(err)
	}

	return nil
}

func (t *RedisDb) RmOrderFromSet(user string, id int64, symbol string) error {
	err := t.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.Send("DEL", orderHashKey(user, id))
	if err != nil {
		panic(err)
	}
	err = t.Send("SREM", orderSetKey(symbol), orderHashKey(user, id))
	if err != nil {
		panic(err)
	}
	_, err = t.Do("EXEC")
	if err != nil {
		panic(err)
	}

	return nil
}

func (t *RedisDb) RmCacheOrder(id int64) error {
	_, err := t.Do("DEL", cacheOrderHashKey(id))
	if err != nil {
		panic(err)
	}
	return nil
}

func (t *RedisDb) RmCacheOrderByNamedConn(name string, id int64) error {
	_, err := t.GetNamedLongConn(name).Do("DEL", cacheOrderHashKey(id))
	if err != nil {
		panic(err)
	}

	return nil
}

func parseOrder(v []interface{}) *Order {
	var ro rOrder
	var err error
	var iv int64
	var order Order

	if err := redis.ScanStruct(v, &ro); err != nil {
		panic(err)
	}
	fmt.Printf("Get order: %+v\n", ro)

	order.ID, err = strconv.ParseInt(ro.ID, 10, 64)
	if err != nil {
		panic(err)
	}
	order.Who = ro.Who
	iv, err = strconv.ParseInt(ro.AorB, 10, 64)
	if err != nil {
		panic(err)
	}
	order.AorB = TradeType(iv)
	order.Symbol = ro.Symbol
	if err != nil {
		panic(err)
	}
	order.Timestamp, err = strconv.ParseInt(ro.Timestamp, 10, 64)
	if err != nil {
		panic(err)
	}
	order.Price, err = strconv.ParseFloat(ro.Price, 64)
	if err != nil {
		panic(err)
	}
	order.Volume, err = strconv.ParseFloat(ro.Volume, 64)
	if err != nil {
		panic(err)
	}

	order.TotalVolume, err = strconv.ParseFloat(ro.TotalVolume, 64)
	if err != nil {
		panic(err)
	}
	order.Fee, err = strconv.ParseFloat(ro.Fee, 64)
	if err != nil {
		panic(err)
	}

	iv, err = strconv.ParseInt(ro.Status, 10, 64)
	if err != nil {
		panic(err)
	}
	order.Status = TradeStatus(iv)

	return &order
}

func (t *RedisDb) GetOrderFromSet(user string, id int64, symbol string) (*Order, error) {
	isMember, err := redis.Bool(t.Do("SISMEMBER", orderSetKey(symbol), orderHashKey(user, id)))
	if err != nil {
		panic(err)
	}
	if !isMember {
		return nil, fmt.Errorf("the order id(%d) not exist int the set(%s)", id, orderSetKey(symbol))
	}

	v, err := redis.Values(t.Do("HGETALL", orderHashKey(user, id)))
	if err != nil {
		panic(err)
	}
	order := parseOrder(v)

	return order, nil
}

func (t *RedisDb) GetCacheOrder(id int64) (*Order, error) {

	v, err := redis.Values(t.Do("HGETALL", cacheOrderHashKey(id)))
	if err != nil {
		panic(err)
	}
	order := parseOrder(v)

	return order, nil
}

func (t *RedisDb) GetCacheOrderByNamedConn(name string, id int64) (*Order, error) {
	v, err := redis.Values(t.GetNamedLongConn(name).Do("HGETALL", cacheOrderHashKey(id)))
	if err != nil {
		panic(err)
	}
	order := parseOrder(v)

	return order, nil
}

func (t *RedisDb) getOrderDetail(orders []interface{}) (so []*Order, err error) {
	var ro rOrder
	for _, elem := range orders {
		id, err := redis.String(elem, nil)
		if err != nil {
			panic(err)
		}
		v, err := redis.Values(t.Do("HGETALL", id))
		if err != nil {
			panic(err)
		}

		if err := redis.ScanStruct(v, &ro); err != nil {
			panic(err)
		}

		o := new(Order)
		o.ID, _ = strconv.ParseInt(ro.ID, 10, 64)
		o.Who = ro.Who
		iv, _ := strconv.ParseInt(ro.AorB, 10, 64)
		o.AorB = TradeType(iv)
		o.Symbol = ro.Symbol
		o.Timestamp, _ = strconv.ParseInt(ro.Timestamp, 10, 64)
		o.Price, _ = strconv.ParseFloat(ro.Price, 64)
		o.Volume, _ = strconv.ParseFloat(ro.Volume, 64)
		o.TotalVolume, _ = strconv.ParseFloat(ro.TotalVolume, 64)
		o.Fee, _ = strconv.ParseFloat(ro.Fee, 64)
		iv, _ = strconv.ParseInt(ro.Status, 10, 64)
		o.Status = TradeStatus(iv)

		so = append(so, o)
	}

	return so, nil
}

/// debug
func (t *RedisDb) GetAllOrder(symbol string) (so []*Order, err error) {
	//orders, err := redis.Values(t.Do("SSCAN", orderSetKey(symbol), 0, "MATCH", "*"))
	orders, err := redis.Values(t.Do("SMEMBERS", orderSetKey(symbol)))
	if err != nil {
		panic(err)
	}

	return t.getOrderDetail(orders)
}

func (t *RedisDb) GetOnesOrder(user string, symbol string) (so []*Order, err error) {
	orders, err := redis.Values(t.Do("SSCAN", orderSetKey(symbol), 0, "MATCH", orderHashKeyByUser(user), "COUNT", 100))
	if err != nil {
		panic(err)
	}

	return t.getOrderDetail(orders[1].([]interface{}))
}
