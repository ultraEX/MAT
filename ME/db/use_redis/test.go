package use_redis

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"../../comm"
)

//Test:---------------------------------------------------------------------------------------------------
func (t *RedisDb) TestRedis(u string, o string, p ...interface{}) {
	switch u {
	case "order":
		price := 1 + float64(rand.Intn(3))/10
		volume := (10 + 10*float64(rand.Intn(10))/10)
		order := comm.Order{
			Timestamp:    time.Now().UnixNano(),
			Who:          "Hunter",
			AorB:         comm.TradeType_ASK,
			Symbol:       "ETH/BTC",
			Price:        price,
			EnOrderPrice: price,
			Volume:       volume,
			TotalVolume:  volume,
			Fee:          0.001,
			Status:       comm.ORDER_SUBMIT,
			IPAddr:       "localhost:IP",
		}

		switch o {
		case "add":
			t.testRedis_AddOrderToSet(&order)
		case "rm":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			t.testRedis_RmOrderFromSet(user, id, symbol)
		case "get":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			t.testRedis_GetOrderFromSet(user, id, symbol)
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
		trade := comm.Trade{
			Order: comm.Order{
				Timestamp:    time.Now().UnixNano(),
				Who:          "Hunter",
				AorB:         comm.TradeType_ASK,
				Symbol:       "ETH/BTC",
				Price:        price,
				EnOrderPrice: price,
				Volume:       volume,
				TotalVolume:  volume,
				Fee:          0.001,
				Status:       comm.ORDER_FILLED,
				IPAddr:       "localhost:IP",
			},
			Amount:    price * volume,
			TradeTime: time.Now().UnixNano(),
			FeeCost:   volume * price * 0.001,
		}

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

	case "zset":
		switch o {
		case "add":
			key, _ := p[0].(string)
			score, _ := strconv.ParseFloat(p[1].(string), 64)
			mem, _ := strconv.ParseInt(p[2].(string), 10, 64)
			t.testRedis_ZSetAddInt64(key, score, mem)
		case "addforindex":
			key, _ := p[0].(string)
			score, _ := strconv.ParseFloat(p[1].(string), 64)
			mem, _ := strconv.ParseInt(p[2].(string), 10, 64)
			t.testRedis_ZSetAddInt64ForIndex(key, score, mem)
		case "rm":
			key, _ := p[0].(string)
			mem, _ := strconv.ParseInt(p[1].(string), 10, 64)
			t.testRedis_ZSetRemoveInt64(key, mem)

		case "get":
			key, _ := p[0].(string)
			index, _ := strconv.ParseInt(p[1].(string), 10, 64)
			t.testRedis_ZSetGetInt64(key, index)

		case "gets":
			key, _ := p[0].(string)
			start, _ := strconv.ParseInt(p[1].(string), 10, 64)
			stop, _ := strconv.ParseInt(p[2].(string), 10, 64)
			t.testRedis_ZSetGetRangeInt64s(key, start, stop)
		case "all":
			key, _ := p[0].(string)
			t.testRedis_ZSetGetAll(key)
		}
	default:

	}
}

//Order:--------------
func (t *RedisDb) testRedis_AddOrderToSet(order *comm.Order) {
	if err := t.AddOrderToSet(order); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_AddOrderToSet execute complete! Please check it use: KEYS * ")
}

func (t *RedisDb) testRedis_RmOrderFromSet(user string, id int64, symbol string) {
	if err := t.RmOrderFromSet(user, id, symbol); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_RmOrderFromSet execute complete! Please check it use: GET ", orderHashKey(user, id), "\nAnd use: SISMEMBER ", orderSetKey(symbol), " ", orderHashKey(user, id))
}

func (t *RedisDb) testRedis_GetOrderFromSet(user string, id int64, symbol string) {
	order, err := t.GetOrderFromSet(user, id, symbol)
	if err != nil {
		panic(err)
	}

	fmt.Println("testRedis_GetOrderFromSet Return:\nid: ", order.ID,
		"; who: ", order.Who,
		"; AorB: ", order.AorB,
		"; Symbol: ", order.Symbol,
		"; Timestamp: ", order.Timestamp,
		"; price: ", order.Price,
		"; volume: ", order.Volume,
		"; TotalVolume: ", order.TotalVolume,
		"; Fee: ", order.Fee,
		"; Status: ", order.Status,
	)
}

func (t *RedisDb) testRedis_GetAllOrder(symbol string) {
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
		)
	}
}

func (t *RedisDb) testRedis_GetOnesOrder(user string, symbol string) {
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
		)
	}
}

//Trade:--------------
func (t *RedisDb) testRedis_AddTrade(trade *comm.Trade) {
	if err := t.AddTrade(trade); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_AddTrade execute complete! Please check it use: KEYS * ")
}

func (t *RedisDb) testRedis_RmTrade(user string, id int64, symbol string) {
	if err := t.RmTrade(user, id, symbol); err != nil {
		panic(err)
	}

	fmt.Println("testRedis_RmTrade execute complete! Please check it use: GET ", tradeHashKey(user, id), "\nAnd use: SISMEMBER ", tradeSetKey(symbol), " ", tradeHashKey(user, id))
}

func (t *RedisDb) testRedis_GetTrade(user string, id int64, symbol string) {
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
	)
}

func (t *RedisDb) testRedis_GetAllTrade(symbol string) {
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
		)
	}
}

func (t *RedisDb) testRedis_GetOnesTrade(user string, symbol string) {
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
		)
	}
}

//Zset:--------------
func (t *RedisDb) testRedis_ZSetAddInt64(key string, score float64, mem int64) {
	t.ZSetAddInt64(key, score, mem)
	fmt.Printf("testRedis_ZSetAddInt64: key = %s, score = %f, mem = %d\n", key, score, mem)
}

func (t *RedisDb) testRedis_ZSetAddInt64ForIndex(key string, score float64, mem int64) {
	index := t.ZSetAddInt64ForIndex(key, score, mem)
	fmt.Printf("testRedis_ZSetAddInt64ForIndex: key = %s, score = %f, mem = %d\n", key, score, mem)
	fmt.Printf("Get index = %d\n", index)
}

func (t *RedisDb) testRedis_ZSetRemoveInt64(key string, mem int64) {
	t.ZSetRemoveInt64(key, mem)
	fmt.Printf("testRedis_ZSetRemoveInt64: key = %s, mem = %d\n", key, mem)
}

func (t *RedisDb) testRedis_ZSetGetRangeInt64s(key string, start int64, stop int64) {
	vs := t.ZSetGetRangeInt64s(key, start, stop)
	fmt.Printf("testRedis_ZSetGetRangeInt64s: key = %s, start = %d, stop = %d, get values:\n", key, start, stop)
	for c, v := range vs {
		fmt.Printf("[%d]: %d\n", c, v)
	}
}

func (t *RedisDb) testRedis_ZSetGetInt64(key string, index int64) {
	if v, ok := t.ZSetGetInt64(key, index); ok {
		fmt.Printf("testRedis_ZSetGetInt64: key = %s, index = %d, get value: %d\n", key, index, v)
	} else {
		fmt.Printf("testRedis_ZSetGetInt64: key = %s, index = %d, no value got.\n", key, index)
	}
}

func (t *RedisDb) testRedis_ZSetGetAll(key string) {
	vs := t.ZSetGetAll(key)
	fmt.Printf("testRedis_ZSetGetAll: key = %s, get values:\n", key)
	for c, v := range vs {
		fmt.Printf("[%d]: %d\n", c, v)
	}
}
