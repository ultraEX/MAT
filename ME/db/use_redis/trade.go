package use_redis

import (
	"fmt"
	"strconv"

	. "../../comm"
	"github.com/gomodule/redigo/redis"
)

//---------------------------------------------------------------------------------------------------

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
func (t *RedisDb) AddTrade(trade *Trade) error {
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

	err := t.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.Send("SADD", tradeSetKey(trade.Symbol), tradeHashKey(trade.Who, trade.ID))
	if err != nil {
		panic(err)
	}
	err = t.Send("HMSET", redis.Args{}.Add(tradeHashKey(trade.Who, trade.ID)).AddFlat(m)...)
	if err != nil {
		panic(err)
	}
	_, err = t.Do("EXEC")
	if err != nil {
		panic(err)
	}

	return nil
}

func (t *RedisDb) RmTrade(user string, id int64, symbol string) error {
	err := t.Send("MULTI")
	if err != nil {
		panic(err)
	}
	err = t.Send("DEL", tradeHashKey(user, id))
	if err != nil {
		panic(err)
	}
	err = t.Send("SREM", tradeSetKey(symbol), tradeHashKey(user, id))
	if err != nil {
		panic(err)
	}
	_, err = t.Do("EXEC")
	if err != nil {
		panic(err)
	}
	return nil
}

func (t *RedisDb) GetTrade(user string, id int64, symbol string) (*Trade, error) {
	isMember, err := redis.Bool(t.Do("SISMEMBER", tradeSetKey(symbol), tradeHashKey(user, id)))
	if err != nil {
		panic(err)
	}
	if !isMember {
		return nil, fmt.Errorf("the trade id(%d) not exist int the set(%s)", id, tradeSetKey(symbol))
	}

	var ro rTrade
	v, err := redis.Values(t.Do("HGETALL", tradeHashKey(user, id)))
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

func (t *RedisDb) getTradeDetail(trades []interface{}) (to []*Trade, err error) {
	var ro rTrade
	for _, elem := range trades {
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

func (t *RedisDb) GetAllTrade(symbol string) (to []*Trade, err error) {
	trades, err := redis.Values(t.Do("SMEMBERS", tradeSetKey(symbol)))
	if err != nil {
		panic(err)
	}

	return t.getTradeDetail(trades)
}

func (t *RedisDb) GetOnesTrade(user string, symbol string) (to []*Trade, err error) {
	trades, err := redis.Values(t.Do("SSCAN", tradeSetKey(symbol), 0, "MATCH", tradeHashKeyByUser(user), "COUNT", 100))
	if err != nil {
		panic(err)
	}

	return t.getTradeDetail(trades[1].([]interface{}))
}
