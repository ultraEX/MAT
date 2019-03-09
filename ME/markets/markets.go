// markets
package markets

import (
	"fmt"
	"strings"
	"time"

	"../config"
	. "../itf"
	"../me"
)

const (
	MODULE_NAME string = "[Markets]: "
)

var (
	marketsRef *Markets = nil
)

type Markets struct {
	meMap      map[config.Symbol]*me.MatchEngine
	MARKET_NUM int64
}

func CreateMarkets() *Markets {
	if marketsRef != nil {
		return marketsRef
	}

	o := new(Markets)
	o.meMap = make(map[config.Symbol]*me.MatchEngine)
	o.MARKET_NUM = 0
	err := o.Setup()
	if err != nil {
		panic(err)
	}

	marketsRef = o
	return o
}

func GetMarkets() *Markets {
	if marketsRef == nil {
		panic(fmt.Errorf("markets.CreateMarkets should be invoke first.\n"))
	}
	return marketsRef
}

func (t *Markets) GetDefaultMatchEngine() (config.Symbol, *me.MatchEngine) {
	for s, m := range t.meMap {
		if m != nil {
			return s, m
		}
	}
	return config.Symbol{}, nil
}

func (t *Markets) GetMatchEngine(symbol string) (*me.MatchEngine, error) {
	m, ok := t.meMap[*config.Struct(symbol)]
	if ok {
		return m, nil
	} else {
		return nil, fmt.Errorf("GetMatchEngine fail. Illegal symbol(%s) input.", symbol)
	}
}

func (t *Markets) GetMatchEngineExistence(symbol string) bool {
	sym := config.Struct(symbol)
	if sym == nil {
		return false
	}

	_, ok := t.meMap[*sym]
	return ok
}

func (t *Markets) MarketCheck(symbol string, marketType string) (*config.Symbol, config.MarketType, bool) {
	mktType := config.String2MarketType(marketType)
	if mktType > config.MarketType_Robot || mktType < config.MarketType_MixHR {
		return &config.Symbol{}, config.MarketType_Num, false
	}

	var ok bool = false
	var m *me.MatchEngine = nil
	sym := config.Struct(symbol)
	if sym != nil {
		if m, ok = t.meMap[*config.Struct(symbol)]; ok {
			if m.TradePool[mktType] == nil {
				ok = false
			}
		}
	}

	return sym, mktType, ok
}

func (t *Markets) MarketTypeCheck(symbol string, marketType config.MarketType) bool {
	if marketType > config.MarketType_Robot || marketType < config.MarketType_MixHR {
		return false
	}

	var ok bool = false
	var m *me.MatchEngine = nil
	sym := config.Struct(symbol)
	if sym != nil {
		m, ok = t.meMap[*config.Struct(symbol)]
		if m.TradePool[marketType] == nil {
			ok = false
		}
	}

	return ok
}

func (t *Markets) Dump() string {
	strBuff := fmt.Sprintf("=========[Current ME Support Market List]=========\n")
	//	count := 0
	for sym, m := range t.meMap {
		strBuff = fmt.Sprintf(strBuff + sym.String() + ": ")
		strBuff += config.GetMarket(m.Config.Sym).Dump()
	}
	strBuff = fmt.Sprintf(strBuff + "\n=================================================\n")

	fmt.Print(strBuff)
	return strBuff
}

func (t *Markets) Setup() error {
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "Markets Setup Begin...\n")

	s := config.GetSymbols()
	for _, elem := range s {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "<+++>Start to setup %s Market...\n", elem.String())
		t.add1Market(elem)
		//		t.meMap[elem] = me.NewMatchEngine(elem.String(), config.GetMarket(elem))
		/// go t.meMap[elem].Run()
	}

	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "Markets Setup Complete.\n")
	return nil
}

func (t *Markets) add1Market(sym config.Symbol) error {
	t.meMap[sym] = me.NewMatchEngine(sym.String(), config.GetMarket(sym))
	return nil
}

func (t *Markets) Add1Market(sym config.Symbol) error {
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "Add1Market Begin...\n")

	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "<+++>Start to setup %s Market...\n", sym.String())
	t.add1Market(sym)

	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "Add1Market Complete.\n")
	return nil
}

func (t *Markets) IsFaulty() bool {
	isFaulty := false
	for _, me := range t.meMap {
		if me.IsFaulty() {
			isFaulty = true
			break
		}
	}
	return isFaulty
}

func (t *Markets) GetMarketNum() int64 {
	if t.MARKET_NUM == 0 {
		marketNum := int64(0)
		for _, market := range t.meMap {
			if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
				marketNum++
			}
			if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
				marketNum++
			}
			if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
				marketNum++
			}
		}
		t.MARKET_NUM = marketNum
	}

	return t.MARKET_NUM
}

func (t *Markets) GetTotalTradeCompleteRate() float64 {
	tradeCPR := float64(0)
	for _, market := range t.meMap {
		if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
			tradeCPR += market.TradePool[config.MarketType_MixHR].GetTradeCompleteRate()
		}
		if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
			tradeCPR += market.TradePool[config.MarketType_Human].GetTradeCompleteRate()
		}
		if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
			tradeCPR += market.TradePool[config.MarketType_Robot].GetTradeCompleteRate()
		}
	}
	return tradeCPR
}

func (t *Markets) GetAskPoolLen() int {
	num := 0
	for _, market := range t.meMap {
		if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
			num += market.TradePool[config.MarketType_MixHR].GetAskPoolLen()
		}
		if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
			num += market.TradePool[config.MarketType_Human].GetAskPoolLen()
		}
		if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
			num += market.TradePool[config.MarketType_Robot].GetAskPoolLen()
		}
	}
	return num
}

func (t *Markets) GetBidPoolLen() int {
	num := 0
	for _, market := range t.meMap {
		if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
			num += market.TradePool[config.MarketType_MixHR].GetBidPoolLen()
		}
		if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
			num += market.TradePool[config.MarketType_Human].GetBidPoolLen()
		}
		if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
			num += market.TradePool[config.MarketType_Robot].GetBidPoolLen()
		}
	}
	return num
}

func (t *Markets) GetPoolLen() int {
	num := 0
	for _, market := range t.meMap {
		if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
			num += market.TradePool[config.MarketType_MixHR].GetPoolLen()
		}
		if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
			num += market.TradePool[config.MarketType_Human].GetPoolLen()
		}
		if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
			num += market.TradePool[config.MarketType_Robot].GetPoolLen()
		}
	}
	return num
}

func (t *Markets) TradeStaticsDump() string {
	strBuff := fmt.Sprintf("===============[TradeStatics Trade Info]==============\n")
	strBuff += fmt.Sprintf("Market num: %d\n", t.GetMarketNum())
	strBuff += fmt.Sprintf("Market askpool len: %d\n", t.GetAskPoolLen())
	strBuff += fmt.Sprintf("Market bidpool len: %d\n", t.GetBidPoolLen())
	strBuff += fmt.Sprintf("Market pool len: %d\n", t.GetPoolLen())
	strBuff += fmt.Sprintf("Markets trade complete rate: %f\n", t.GetTotalTradeCompleteRate())
	strBuff += fmt.Sprintf("=======================================================\n")
	return strBuff
}

func (t *Markets) ConstructTicker(sym string, from int64) {
	market, _ := t.GetMatchEngine(sym)
	if market == nil {
		fmt.Printf("ConstructTicker illegal symbol input(%s)\n", sym)
		return
	}

	fmt.Printf("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	fmt.Printf("Begin to construct symbol(%s)'s tickers...\n", sym)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Begin to construct symbol(%s)'s tickers...\n", sym)

	if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
		market.TradePool[config.MarketType_MixHR].TradeCommand("stop")
	}
	if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
		market.TradePool[config.MarketType_Human].TradeCommand("stop")
	}
	if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
		market.TradePool[config.MarketType_Robot].TradeCommand("stop")
	}

	market.GetTickersEngine().ConstructTickersFromHistoryTrades(from)

	if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
		market.TradePool[config.MarketType_MixHR].TradeCommand("resume")
	}
	if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
		market.TradePool[config.MarketType_Human].TradeCommand("resume")
	}
	if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
		market.TradePool[config.MarketType_Robot].TradeCommand("resume")
	}

	fmt.Printf("To construct symbol(%s)'s tickers complete.\n", market.Symbol)
	fmt.Printf("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "To construct symbol(%s)'s tickers complete.\n", market.Symbol)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
}

func (t *Markets) ConstructTickerByTickerType(sym string, tickerType TickerType, from int64) {
	market, _ := t.GetMatchEngine(sym)
	if market == nil {
		fmt.Printf("ConstructTicker illegal symbol input(%s)\n", sym)
		return
	}

	fmt.Printf("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	fmt.Printf("Begin to construct symbol(%s)'s tickers by type(%s)...\n", sym, tickerType)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Begin to construct symbol(%s)'s tickers by type(%s)...\n", sym, tickerType)

	market.GetTickersEngine().ConstructTickersFromHistoryTradesByTickerType(tickerType, from)

	fmt.Printf("To construct symbol(%s)'s tickers by type(%s) complete\n", sym, tickerType)
	fmt.Printf("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "To construct symbol(%s)'s tickers by type(%s) complete\n", sym, tickerType)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
}

func (t *Markets) ConstructUnInitializedTicker(sym string, from int64) {
	market, _ := t.GetMatchEngine(sym)
	if market == nil {
		fmt.Printf("ConstructTicker illegal symbol input(%s)\n", sym)
		return
	}

	fmt.Printf("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	fmt.Printf("Begin to construct unInitialized symbol(%s)'s tickers...\n", sym)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Begin to construct unInitialized symbol(%s)'s tickers...\n", sym)

	if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
		market.TradePool[config.MarketType_MixHR].TradeCommand("stop")
	}
	if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
		market.TradePool[config.MarketType_Human].TradeCommand("stop")
	}
	if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
		market.TradePool[config.MarketType_Robot].TradeCommand("stop")
	}

	for _type := TickerType_1min; _type <= TickerType_1month; _type++ {
		if !market.GetTickersEngine().IsInitialized(_type) {
			t.ConstructTickerByTickerType(sym, _type, from)
		}
	}

	if market.Config.Market_MixHR && market.TradePool[config.MarketType_MixHR] != nil {
		market.TradePool[config.MarketType_MixHR].TradeCommand("resume")
	}
	if market.Config.Market_Human && market.TradePool[config.MarketType_Human] != nil {
		market.TradePool[config.MarketType_Human].TradeCommand("resume")
	}
	if market.Config.Market_Robot && market.TradePool[config.MarketType_Robot] != nil {
		market.TradePool[config.MarketType_Robot].TradeCommand("resume")
	}

	fmt.Printf("To construct unInitialized symbol(%s)'s tickers complete.\n", market.Symbol)
	fmt.Printf("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "To construct unInitialized symbol(%s)'s tickers complete.\n", market.Symbol)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
}

func (t *Markets) ConstructTickers(from int64) {
	loc, _ := time.LoadLocation("Local")
	fmt.Printf("=======================Begin To construct tickers=========================\n")
	fmt.Printf("Begin time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "=======================Begin To construct tickers=========================\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Begin time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))

	c := 0
	for _, market := range t.meMap {

		t.ConstructTicker(market.Symbol, from)

		c++
		fmt.Printf("Total Progress:<%d/%d>\n", c, len(t.meMap))
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Total Progress:<%d/%d>\n", c, len(t.meMap))
	}

	fmt.Printf("========================To construct tickers complete.========================\n")
	fmt.Printf("End time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "========================To construct tickers complete.========================\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "End time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
}

func (t *Markets) ConstructTickersWithFilter(from int64, filter string) {
	excludedSymbols := strings.Split(filter, ",")
	isFiltered := false
	loc, _ := time.LoadLocation("Local")

	fmt.Printf("=======================Begin To construct tickers with filter=========================\n")
	fmt.Printf("Begin time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "=======================Begin To construct tickers  with filter=========================\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Begin time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))

	c := 0
	for _, market := range t.meMap {
		isFiltered = false
		for _, exSym := range excludedSymbols {
			if exSym == market.Symbol {
				isFiltered = true
				fmt.Printf("Symbol(%s) is filtered, jump to next symbol.\n", market.Symbol)
				DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Symbol(%s) is filtered, jump to next symbol.\n", market.Symbol)
				break
			}
		}
		if isFiltered {
			continue
		}

		t.ConstructTicker(market.Symbol, from)

		c++
		fmt.Printf("Total Progress:<%d/%d>\n", c, len(t.meMap))
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Total Progress:<%d/%d>\n", c, len(t.meMap))
	}

	fmt.Printf("========================To construct tickers  with filter complete.========================\n")
	fmt.Printf("End time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "========================To construct tickers  with filter complete.========================\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "End time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
}

func (t *Markets) ConstructUnInitializedTickers(from int64) {
	loc, _ := time.LoadLocation("Local")

	fmt.Printf("=======================Begin To Construct UnInitialized Tickers=========================\n")
	fmt.Printf("Begin time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "=======================Begin To Construct UnInitialized Tickers=========================\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Begin time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))

	c := 0
	for _, market := range t.meMap {

		t.ConstructUnInitializedTicker(market.Symbol, from)

		c++
		fmt.Printf("Total Progress:<%d/%d>\n", c, len(t.meMap))
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Total Progress:<%d/%d>\n", c, len(t.meMap))
	}

	fmt.Printf("========================To Construct UnInitialized Tickers complete.========================\n")
	fmt.Printf("End time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "========================To Construct UnInitialized Tickers complete.========================\n")
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "End time: %s\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"))
}
