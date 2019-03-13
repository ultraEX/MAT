// redisgo
package use_redis

import (
	"fmt"
	"math/rand"

	. "../../comm"

	//	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type RediGODB struct {
	client redis.Conn
}

func (t *RediGODB) Init() {
	t.client = newRediGO()
}

func newRediGO() redis.Conn {
	client, err := redis.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println("connect redis error :", err)
		panic(err)
	}

	return client
}

var rediGOClientObj *RediGODB
var once sync.Once

func ReidsGOInstance() *RediGODB {

	once.Do(func() {
		rediGOClientObj = new(RediGODB)
		rediGOClientObj.Init()
	})

	return rediGOClientObj
}

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
func (t *RediGODB) AddOrder(order *Order) error {
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

	err := t.client.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.client.Send("SADD", orderSetKey(order.Symbol), orderHashKey(order.Who, order.ID))
	if err != nil {
		panic(err)
	}
	err = t.client.Send("HMSET", redis.Args{}.Add(orderHashKey(order.Who, order.ID)).AddFlat(m)...)
	if err != nil {
		panic(err)
	}
	_, err = t.client.Do("EXEC")
	if err != nil {
		panic(err)
	}

	/// redis set use name with order's symbol field like: ETH/BTC, Value is the order id.
	//	_, err = t.client.Do("SADD", orderSetKey(order.Symbol), order.ID)
	//	if err != nil {
	//		panic(err)
	//	}
	//	if _, err := t.client.Do("HMSET", redis.Args{}.Add(orderHashKey(order.ID)).AddFlat(m)...); err != nil {
	//		panic(err)
	//	}

	return nil
}

func (t *RediGODB) RmOrder(user string, id int64, symbol string) error {
	err := t.client.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.client.Send("DEL", orderHashKey(user, id))
	if err != nil {
		panic(err)
	}
	err = t.client.Send("SREM", orderSetKey(symbol), orderHashKey(user, id))
	if err != nil {
		panic(err)
	}
	_, err = t.client.Do("EXEC")
	if err != nil {
		panic(err)
	}
	//	/// delete order Hash detail
	//	_, err := t.client.Do("DEL", orderHashKey(id))
	//	if err != nil {
	//		panic(err)
	//	}
	//	/// delete the symbol order Set item
	//	_, err = t.client.Do("SREM", orderSetKey(symbol), id)
	//	if err != nil {
	//		panic(err)
	//	}

	return nil
}

func (t *RediGODB) GetOrder(user string, id int64, symbol string) (*Order, error) {
	isMember, err := redis.Bool(t.client.Do("SISMEMBER", orderSetKey(symbol), orderHashKey(user, id)))
	if err != nil {
		panic(err)
	}
	if !isMember {
		return nil, fmt.Errorf("the order id(%d) not exist int the set(%s)", id, orderSetKey(symbol))
	}

	var ro rOrder
	v, err := redis.Values(t.client.Do("HGETALL", orderHashKey(user, id)))
	if err != nil {
		panic(err)
	}
	if err := redis.ScanStruct(v, &ro); err != nil {
		panic(err)
	}
	fmt.Printf("Get order: %+v\n", ro)

	order := new(Order)
	order.ID, err = strconv.ParseInt(ro.ID, 10, 64)
	if err != nil {
		panic(err)
	}
	order.Who = ro.Who
	iv, err := strconv.ParseInt(ro.AorB, 10, 64)
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

	return order, nil
}

func (t *RediGODB) getOrderDetail(orders []interface{}) (so []*Order, err error) {
	var ro rOrder
	for _, elem := range orders {
		id, err := redis.String(elem, nil)
		if err != nil {
			panic(err)
		}
		v, err := redis.Values(t.client.Do("HGETALL", id))
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
//fmt.Println("reply id enum: ", id)
//fmt.Printf("%+v\n", ro)
func (t *RediGODB) GetAllOrder(symbol string) (so []*Order, err error) {
	//orders, err := redis.Values(t.client.Do("SSCAN", orderSetKey(symbol), 0, "MATCH", "*"))
	orders, err := redis.Values(t.client.Do("SMEMBERS", orderSetKey(symbol)))
	if err != nil {
		panic(err)
	}

	return t.getOrderDetail(orders)
}

func (t *RediGODB) GetOnesOrder(user string, symbol string) (so []*Order, err error) {
	orders, err := redis.Values(t.client.Do("SSCAN", orderSetKey(symbol), 0, "MATCH", orderHashKeyByUser(user), "COUNT", 100))
	if err != nil {
		panic(err)
	}

	return t.getOrderDetail(orders[1].([]interface{}))
}

//---------------------------------------------------------------------------------------------------
//func (t *RediGODB) AddTrade(trade *Trade) error {
//	return fmt.Errorf("use_redis: unexpected type=%T for String", t)
//}

//func (t *RediGODB) RmTrade(id int64) error {
//	return fmt.Errorf("use_redis: unexpected type=%T for String", t)
//}

//func (t *RediGODB) GetTrade(id int64) (*Trade, error) {
//	return nil, fmt.Errorf("use_redis: unexpected type=%T for String", t)
//}

type rTrade struct {
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
	Amount      string `redis:"Amount"`
	TradeTime   string `redis:"TradeTime"`
	FeeCost     string `redis:"FeeCost"`
}

//---------------------------------------------------------------------------------------------------
/// Use Redis Set to store complete process trade list
/// Use Redis Hash to store trade detail info
func (t *RediGODB) AddTrade(trade *Trade) error {
	/// use Hash to store trade detail
	m := map[string]string{
		"ID":          strconv.FormatInt(trade.ID, 10),
		"Who":         trade.Who,
		"AorB":        strconv.FormatInt(int64(trade.AorB), 10),
		"Symbol":      trade.Symbol,
		"Timestamp":   strconv.FormatInt(trade.Timestamp, 10),
		"Price":       strconv.FormatFloat(trade.Price, 'f', -1, 64),
		"Volume":      strconv.FormatFloat(trade.Volume, 'f', -1, 64),
		"TotalVolume": strconv.FormatFloat(trade.TotalVolume, 'f', -1, 64),
		"Fee":         strconv.FormatFloat(trade.Fee, 'f', -1, 64),
		"Status":      strconv.FormatInt(int64(trade.Status), 10),
		"Amount":      strconv.FormatFloat(trade.Amount, 'f', -1, 64),
		"TradeTime":   strconv.FormatInt(trade.TradeTime, 10),
		"FeeCost":     strconv.FormatFloat(trade.FeeCost, 'f', -1, 64),
	}

	err := t.client.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.client.Send("SADD", tradeSetKey(trade.Symbol), tradeHashKey(trade.Who, trade.ID))
	if err != nil {
		panic(err)
	}
	err = t.client.Send("HMSET", redis.Args{}.Add(tradeHashKey(trade.Who, trade.ID)).AddFlat(m)...)
	if err != nil {
		panic(err)
	}
	_, err = t.client.Do("EXEC")
	if err != nil {
		panic(err)
	}

	return nil
}

func (t *RediGODB) RmTrade(user string, id int64, symbol string) error {
	err := t.client.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.client.Send("DEL", tradeHashKey(user, id))
	if err != nil {
		panic(err)
	}
	err = t.client.Send("SREM", tradeSetKey(symbol), tradeHashKey(user, id))
	if err != nil {
		panic(err)
	}
	_, err = t.client.Do("EXEC")
	if err != nil {
		panic(err)
	}
	return nil
}

func (t *RediGODB) GetTrade(user string, id int64, symbol string) (*Trade, error) {
	isMember, err := redis.Bool(t.client.Do("SISMEMBER", tradeSetKey(symbol), tradeHashKey(user, id)))
	if err != nil {
		panic(err)
	}
	if !isMember {
		return nil, fmt.Errorf("the trade id(%d) not exist int the set(%s)", id, tradeSetKey(symbol))
	}

	var ro rTrade
	v, err := redis.Values(t.client.Do("HGETALL", tradeHashKey(user, id)))
	if err != nil {
		panic(err)
	}
	if err := redis.ScanStruct(v, &ro); err != nil {
		panic(err)
	}
	fmt.Printf("Get trade: %+v\n", ro)

	trade := new(Trade)
	trade.ID, err = strconv.ParseInt(ro.ID, 10, 64)
	if err != nil {
		panic(err)
	}
	trade.Who = ro.Who
	iv, err := strconv.ParseInt(ro.AorB, 10, 64)
	if err != nil {
		panic(err)
	}
	trade.AorB = TradeType(iv)
	trade.Symbol = ro.Symbol
	if err != nil {
		panic(err)
	}
	trade.Timestamp, err = strconv.ParseInt(ro.Timestamp, 10, 64)
	if err != nil {
		panic(err)
	}
	trade.Price, err = strconv.ParseFloat(ro.Price, 64)
	if err != nil {
		panic(err)
	}
	trade.Volume, err = strconv.ParseFloat(ro.Volume, 64)
	if err != nil {
		panic(err)
	}

	trade.TotalVolume, err = strconv.ParseFloat(ro.TotalVolume, 64)
	if err != nil {
		panic(err)
	}
	trade.Fee, err = strconv.ParseFloat(ro.Fee, 64)
	if err != nil {
		panic(err)
	}

	iv, err = strconv.ParseInt(ro.Status, 10, 64)
	if err != nil {
		panic(err)
	}
	trade.Status = TradeStatus(iv)

	if err != nil {
		panic(err)
	}
	trade.Amount, err = strconv.ParseFloat(ro.Amount, 64)
	if err != nil {
		panic(err)
	}
	trade.TradeTime, err = strconv.ParseInt(ro.TradeTime, 10, 64)
	if err != nil {
		panic(err)
	}
	trade.FeeCost, err = strconv.ParseFloat(ro.FeeCost, 64)
	if err != nil {
		panic(err)
	}
	return trade, nil
}

func (t *RediGODB) getTradeDetail(trades []interface{}) (to []*Trade, err error) {
	var ro rTrade
	for _, elem := range trades {
		id, err := redis.String(elem, nil)
		if err != nil {
			panic(err)
		}
		v, err := redis.Values(t.client.Do("HGETALL", id))
		if err != nil {
			panic(err)
		}

		if err := redis.ScanStruct(v, &ro); err != nil {
			panic(err)
		}

		o := new(Trade)
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
		o.Amount, _ = strconv.ParseFloat(ro.Amount, 64)
		o.TradeTime, _ = strconv.ParseInt(ro.TradeTime, 10, 64)
		o.FeeCost, _ = strconv.ParseFloat(ro.FeeCost, 64)

		to = append(to, o)
	}

	return to, nil
}

func (t *RediGODB) GetAllTrade(symbol string) (to []*Trade, err error) {
	trades, err := redis.Values(t.client.Do("SMEMBERS", tradeSetKey(symbol)))
	if err != nil {
		panic(err)
	}

	return t.getTradeDetail(trades)
}

func (t *RediGODB) GetOnesTrade(user string, symbol string) (to []*Trade, err error) {
	trades, err := redis.Values(t.client.Do("SSCAN", tradeSetKey(symbol), 0, "MATCH", tradeHashKeyByUser(user), "COUNT", 100))
	if err != nil {
		panic(err)
	}

	return t.getTradeDetail(trades[1].([]interface{}))
}

//Test:---------------------------------------------------------------------------------------------------
func (t *RediGODB) TestRedis(u string, o string, p ...interface{}) {
	switch u {
	case "order":
		price := 1 + float64(rand.Intn(3))/10
		volume := (10 + 10*float64(rand.Intn(10))/10)
		order := Order{time.Now().UnixNano(), "Hunter", TradeType_ASK, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_SUBMIT, "localhost:IP"}

		switch o {
		case "add":
			t.testRedis_AddOrder(&order)
		case "rm":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			t.testRedis_RmOrder(user, id, symbol)
		case "get":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			t.testRedis_GetOrder(user, id, symbol)
		case "all":
			symbol, _ := p[0].(string)
			t.testRedis_GetAllOrder(symbol)
		case "ones":
			user, _ := p[0].(string)
			symbol, _ := p[1].(string)
			t.testRedis_GetOnesOrder(user, symbol)
		}

	case "trade":
		price := 1 + float64(rand.Intn(3))/10
		volume := (10 + 10*float64(rand.Intn(10))/10)
		trade := Trade{Order{time.Now().UnixNano(), "Hunter", TradeType_ASK, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_FILLED, "localhost:IP"},
			price * volume, time.Now().UnixNano(), volume * price * 0.001}

		switch o {
		case "add":
			t.testRedis_AddTrade(&trade)
		case "rm":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			t.testRedis_RmTrade(user, id, symbol)
		case "get":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			t.testRedis_GetTrade(user, id, symbol)
		case "all":
			symbol, _ := p[0].(string)
			t.testRedis_GetAllTrade(symbol)
		case "ones":
			user, _ := p[0].(string)
			symbol, _ := p[1].(string)
			t.testRedis_GetOnesTrade(user, symbol)
		}
	default:

	}
}

//Order:--------------
func (t *RediGODB) testRedis_AddOrder(order *Order) {
	if err := t.AddOrder(order); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_AddOrder execute complete! Please check it use: KEYS * ")
}

func (t *RediGODB) testRedis_RmOrder(user string, id int64, symbol string) {
	if err := t.RmOrder(user, id, symbol); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_RmOrder execute complete! Please check it use: GET ", orderHashKey(user, id), "\nAnd use: SISMEMBER ", orderSetKey(symbol), " ", orderHashKey(user, id))
}

func (t *RediGODB) testRedis_GetOrder(user string, id int64, symbol string) {
	order, err := t.GetOrder(user, id, symbol)
	if err != nil {
		panic(err)
	}

	fmt.Println("GetOrder Return:\nid: ", order.ID,
		"; who: ", order.Who,
		"; AorB: ", order.AorB,
		"; Symbol: ", order.Symbol,
		"; Timestamp: ", order.Timestamp,
		"; price: ", order.Price,
		"; volume: ", order.Volume,
		"; TotalVolume: ", order.TotalVolume,
		"; Fee: ", order.Fee,
		"; Status: ", order.Status,

		"\n")
}

func (t *RediGODB) testRedis_GetAllOrder(symbol string) {
	orders, err := t.GetAllOrder(symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range orders {
		fmt.Println("GetAllOrder Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,

			"\n")
	}
}

func (t *RediGODB) testRedis_GetOnesOrder(user string, symbol string) {
	orders, err := t.GetOnesOrder(user, symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range orders {
		fmt.Println("GetOnesOrder Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,

			"\n")
	}
}

//Trade:--------------
func (t *RediGODB) testRedis_AddTrade(trade *Trade) {
	if err := t.AddTrade(trade); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_AddTrade execute complete! Please check it use: KEYS * ")
}

func (t *RediGODB) testRedis_RmTrade(user string, id int64, symbol string) {
	if err := t.RmTrade(user, id, symbol); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_RmTrade execute complete! Please check it use: GET ", tradeHashKey(user, id), "\nAnd use: SISMEMBER ", tradeSetKey(symbol), " ", tradeHashKey(user, id))
}

func (t *RediGODB) testRedis_GetTrade(user string, id int64, symbol string) {
	trade, err := t.GetTrade(user, id, symbol)
	if err != nil {
		panic(err)
	}

	fmt.Println("GetTrade Return:\nid: ", trade.ID,
		"; who: ", trade.Who,
		"; AorB: ", trade.AorB,
		"; Symbol: ", trade.Symbol,
		"; Timestamp: ", trade.Timestamp,
		"; price: ", trade.Price,
		"; volume: ", trade.Volume,
		"; TotalVolume: ", trade.TotalVolume,
		"; Fee: ", trade.Fee,
		"; Status: ", trade.Status,
		"; Amount: ", trade.Amount,
		"; TradeTime", trade.TradeTime,

		"\n")
}

func (t *RediGODB) testRedis_GetAllTrade(symbol string) {
	trades, err := t.GetAllTrade(symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range trades {
		fmt.Println("GetAllTrade Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,
			"; Amount: ", elem.Amount,
			"; TradeTime", elem.TradeTime,

			"\n")
	}
}

func (t *RediGODB) testRedis_GetOnesTrade(user string, symbol string) {
	trades, err := t.GetOnesTrade(user, symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range trades {
		fmt.Println("GetOnesTrade Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,
			"; Amount: ", elem.Amount,
			"; TradeTime", elem.TradeTime,

			"\n")
	}
}
