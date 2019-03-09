// MatchEngine.go
package me

import (
	"fmt"
	"math/rand"
	"time"

	"../config"
	//	"../doctor"
	. "../itf"
)

type MatchEngine struct {
	Symbol       string
	Config       config.MarketConfig
	tickerEngine *TickerPool
	TradePool    [config.MarketType_Num]*TradePool
}

func NewMatchEngine(symbol string, conf config.MarketConfig) *MatchEngine {
	o := new(MatchEngine)
	o.Symbol = symbol
	o.Config = conf

	///init Tickers Engine
	o.tickerEngine = NewTickerPool(o.Symbol, TICKERS_ENGINE_DEPTH)

	o.TradePool = [config.MarketType_Num]*TradePool{nil, nil, nil}

	if o.Config.Market_Human {
		o.TradePool[config.MarketType_Human] = NewTradePool(o.Symbol, config.MarketType_Human, &o.Config, o.tickerEngine)
		//go o.RunMatchEngine(config.MarketType_Human)
	}
	if o.Config.Market_Robot {
		o.TradePool[config.MarketType_Robot] = NewTradePool(o.Symbol, config.MarketType_Robot, &o.Config, o.tickerEngine)
		//go o.RunMatchEngine(config.MarketType_Robot)
	}
	if o.Config.Market_MixHR {
		o.TradePool[config.MarketType_MixHR] = NewTradePool(o.Symbol, config.MarketType_MixHR, &o.Config, o.tickerEngine)
		//go o.RunMatchEngine(config.MarketType_MixHR)
	}
	o.tickerEngine.setTradePool(o.TradePool)
	return o
}

func (t *MatchEngine) GetTradePool(maketType config.MarketType) *TradePool {
	if maketType > config.MarketType_Robot || maketType < config.MarketType_MixHR {
		panic(fmt.Errorf("MatchEngine GetTradePool input illegal."))
	}
	return t.TradePool[maketType]
}

func (t *MatchEngine) GetTickersEngine() *TickerPool {
	return t.tickerEngine
}

func (t *MatchEngine) GetSymbol() string {
	return t.Symbol
}

func (t *MatchEngine) GetConfig() config.MarketConfig {
	return t.Config
}

func (t *MatchEngine) test(tp *TradePool) {

	fmt.Println("pool data insert test:========================\n")
	volume := (10 + 10*float64(rand.Intn(10))/10)
	price := 1 + float64(rand.Intn(3))/10
	tmpBid := Order{time.Now().UnixNano(), "Hunter", TradeType_BID, t.Symbol, time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_SUBMIT, "localhost:IP"}
	tmpAsk := Order{time.Now().UnixNano(), "Hunter", TradeType_ASK, t.Symbol, time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_SUBMIT, "localhost:IP"}
	/// debug:
	TimeDot1 := time.Now().UnixNano()
	tp.add(&tmpBid)
	/// debug:
	TimeDot2 := time.Now().UnixNano()
	tp.add(&tmpAsk)
	/// debug:
	TimeDot3 := time.Now().UnixNano()
	fmt.Println("tradePool insert data test time log:========================\n",
		"test data scale = ", TEST_DATA_SCALE, "\n",
		"insert to bidPool = ", float64(TimeDot2-TimeDot1)/float64(1*time.Second), "s;\n",
		"insert to askPool = ", float64(TimeDot3-TimeDot2)/float64(1*time.Second), "s;\n")

	fmt.Println("pool data dump[after insert]:========================\n")
	tp.dump()

	//	fmt.Println("pool data popTops test:========================\n")
	//	bidOrder, numBid := tp.popTops(TradeType_BID)
	//	fmt.Println("tp.BID popTops: num=", numBid)
	//	for _, e := range bidOrder {
	//		fmt.Println("pop bidOrder.price = ", e.price)
	//	}
	//	askOrder, numAsk := tp.popTops(TradeType_ASK)
	//	fmt.Println("tp.ASK popTops: num=", numAsk)
	//	for _, e := range askOrder {
	//		fmt.Println("pop askOrder.price = ", e.price)
	//	}
	//	fmt.Println("pool data dump[after popTops]:========================\n")
	//	tp.dump()

	fmt.Println("pool data popTop test:========================\n")
	bidOrder, res := tp.popTop(TradeType_BID)
	fmt.Println("tp.BID popTop: res=", res, "; pop order.price=", bidOrder.Price, "; timestamp=", bidOrder.Timestamp)

	askOrder, res := tp.popTop(TradeType_ASK)
	fmt.Println("tp.ASK popTop: res=", res, "; pop order.price=", askOrder.Price, "; timestamp=", askOrder.Timestamp)

	fmt.Println("pool data dump[after popTop]:========================\n")
	tp.dump()
}

//func (t *MatchEngine) RunMatchEngine(marketType config.MarketType) {

//	//	t.test(tp)

//	fmt.Println("=====================================================================")
//	fmt.Printf("Start Match Engine %s...\n", t.Symbol)
//	go t.TradePool[marketType].match()
//	///go tp.Output()
//	///go t.TradePool[marketType].input()
//	go t.TradePool[marketType].cancel()
//	fmt.Printf("Start Match Engine %s complete.\n", t.Symbol)

//	t.TradePool[marketType].doctor.SetProgress(doctor.Progress_Working)
//}

func (t *MatchEngine) IsFaulty() bool {
	isHumanMarketFaulty := false
	if t.Config.Market_Human {
		isHumanMarketFaulty = t.TradePool[config.MarketType_Human].IsFaulty()
	}
	isRobotMarketFaulty := false
	if t.Config.Market_Robot {
		isRobotMarketFaulty = t.TradePool[config.MarketType_Robot].IsFaulty()
	}
	isMixHRMarketFaulty := false
	if t.Config.Market_MixHR {
		isMixHRMarketFaulty = t.TradePool[config.MarketType_MixHR].IsFaulty()
	}
	return isHumanMarketFaulty || isRobotMarketFaulty || isMixHRMarketFaulty
}
