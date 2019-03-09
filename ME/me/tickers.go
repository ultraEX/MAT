// tickers
// me
package me

import (
	"container/list"
	"fmt"
	"runtime"
	"time"

	"../config"
	"../db/use_mysql"
	. "../itf"
)

const (
	MODULE_NAME_TICKERS string = "[Ticker Engine]: "

	TICKERS_ENGINE_DEPTH      int   = 1000
	OUTPUT_CHANNEL_SIZE       int64 = 68
	LATEST_TRADEL_SIZE        int64 = 100
	PERIOD_DS_DURATION_NANO   int64 = int64(time.Hour)
	LEVELORDER_DIRTY_DURATION int64 = 5 * int64(time.Second)
)

type TradePair struct {
	bidTrade *Trade
	askTrade *Trade
}

type getLevelFunc func(limit int64) ([]OrderLevel, error)
type LevelOrdersBuff struct {
	time    int64
	getFunc getLevelFunc

	///data [RESTFUL_MAX_ORDER_LEVELS]OrderLevel
	data []OrderLevel
}

func (t *LevelOrdersBuff) getLevelOrders(limit int64) ([]OrderLevel, error) {
	if limit == 0 {
		return t.data[0:0], nil
	}

	if time.Now().UnixNano()-t.time > LEVELORDER_DIRTY_DURATION || int64(len(t.data)) < limit {
		data, err := t.getFunc(limit)
		if err != nil {
			return nil, err
		} else {
			t.time = time.Now().UnixNano()
			t.data = data
			return data, nil
		}
	} else {
		return t.data[:limit], nil
	}

}

func CreateNewLevelOrdersBuff(f getLevelFunc) *LevelOrdersBuff {
	o := new(LevelOrdersBuff)
	o.time = 0
	o.getFunc = f
	o.data = make([]OrderLevel, 0)

	return o
}

type LevelOrders struct {
	askLevel *LevelOrdersBuff
	bidLevel *LevelOrdersBuff
}

func (t *TickerPool) initOrderLevels() {
	t.askLevel = CreateNewLevelOrdersBuff(t.getAskLevelsGroupByPrice)
	t.bidLevel = CreateNewLevelOrdersBuff(t.getBidLevelsGroupByPrice)
}

type TickerChannel struct {
	//	OutChannel chan *Trade
	OutChannel chan *TradePair
}

type TickersData struct {
	tickerRel_1min   TickerInfo
	ticker_1min      *list.List
	tickerRel_5min   TickerInfo
	ticker_5min      *list.List
	tickerRel_15min  TickerInfo
	ticker_15min     *list.List
	tickerRel_30min  TickerInfo
	ticker_30min     *list.List
	tickerRel_1hour  TickerInfo
	ticker_1hour     *list.List
	tickerRel_1day   TickerInfo
	ticker_1day      *list.List
	tickerRel_1week  TickerInfo
	ticker_1week     *list.List
	tickerRel_1month TickerInfo
	ticker_1month    *list.List
}

type LatestDSTime struct {
	time [TickerType_Num]int64
}

func (t *TickerPool) initLatestDSTime() {
	for _t := TickerType_1min; _t <= TickerType_1month; _t++ {
		tickList := t.getTickerList(_t)
		if tickList.Back() != nil {
			tickerInfo := tickList.Back().Value.(TickerInfo)
			t.LatestDSTime.time[_t] = tickerInfo.From
		} else {
			t.LatestDSTime.time[_t] = 0
		}
	}
}

/// single thread model
type LatestTrade struct {
	pos  int64
	size int64
	data [LATEST_TRADEL_SIZE]*Trade
}

func (t *LatestTrade) Init() {
	t.pos = 0
	t.size = 0
}

func (t *LatestTrade) AddTrade(trade *Trade) {
	t.data[t.pos] = trade

	t.size++
	if t.size >= LATEST_TRADEL_SIZE {
		t.size = LATEST_TRADEL_SIZE
	}
	t.pos++
	if t.pos >= LATEST_TRADEL_SIZE {
		t.pos = 0
	}
}

func (t *LatestTrade) GetTradesLimit(limit int64) ([]*Trade, error) {
	if limit > t.size {
		limit = t.size
	}

	var (
		pos    int64    = 0
		trades []*Trade = make([]*Trade, limit)
	)
	for i := int64(0); i < limit; i++ {
		pos = t.pos - i - 1
		if pos < 0 {
			pos += LATEST_TRADEL_SIZE
		}
		/// trades = append(trades, t.data[pos])
		trades[i] = t.data[pos]
	}

	return trades, nil
}

func (t *LatestTrade) Load(sym string) {
	fmt.Printf("[Ticker Engine]: Begin to Load LatestTrade to Ticker Module\n")
	TimeDot1 := time.Now().UnixNano()
	trades, _ := use_mysql.MEMySQLInstance().GetLatestTradeLimit(sym, LATEST_TRADEL_SIZE)

	size := len(trades)
	if size > 0 {
		for i := size - 1; i >= 0; i-- {
			t.AddTrade(trades[i])
		}
	}

	TimeDot2 := time.Now().UnixNano()
	fmt.Printf("[Ticker Engine]: "+
		"LatestTrade Load from DS complete: total num = %d, USE_TIME=%f (second)\n",
		len(trades),
		float64(TimeDot2-TimeDot1)/float64(1*time.Second),
	)
}

func (t *LatestTrade) DumpToPrint() {
	var (
		c     int
		trade *Trade
	)
	trades, _ := t.GetTradesLimit(LATEST_TRADEL_SIZE)

	fmt.Printf("==================[LatestTrade Data Dump]=====================\n")
	for c, trade = range trades {
		fmt.Printf("[trade %d]Sym: %s, Time: %d, Type: %s, Price: %f, Volume: %f\n",
			c,
			trade.Symbol,
			trade.TradeTime,
			trade.AorB,
			trade.Price,
			trade.Volume,
		)
	}
	fmt.Printf("Total: %d\n", c+1)
	fmt.Printf("==============================================================\n")
}

func (t *LatestTrade) DumpToBuff(buff *string) *string {
	var (
		c     int
		trade *Trade
	)
	trades, _ := t.GetTradesLimit(LATEST_TRADEL_SIZE)

	*buff += fmt.Sprintf("==================[LatestTrade Data Dump]=====================\n")
	for c, trade = range trades {
		*buff += fmt.Sprintf("[trade %d]Sym: %s, Time: %d, Type: %s, Price: %f, Volume: %f\n",
			c,
			trade.Symbol,
			trade.TradeTime,
			trade.AorB,
			trade.Price,
			trade.Volume,
		)
	}
	*buff += fmt.Sprintf("Total: %d\n", c+1)
	*buff += fmt.Sprintf("==============================================================\n")
	return buff
}

type TickerPool struct {
	sym  string
	size int

	newestPrice float64
	spread      float64
	location    *time.Location

	TickersData
	LatestTrade
	LatestDSTime

	LevelOrders

	TickerChannel

	tpRef [config.MarketType_Num]*TradePool
}

func NewTickerPool(sym string, _size int) *TickerPool {
	t := new(TickerPool)

	t.init(sym, _size)
	t.tpRef = [config.MarketType_Num]*TradePool{nil, nil, nil}

	go t.outPutProcess()

	return t
}

func (t *TickerPool) getLocalTimeUnixNano() int64 {
	return time.Now().In(t.location).UnixNano()
}

func (t *TickerPool) setTradePool(p [config.MarketType_Num]*TradePool) {
	t.tpRef = p
}

func (t *TickerPool) getTradePool() *TradePool {
	if t.tpRef[config.MarketType_MixHR] != nil {
		return t.tpRef[config.MarketType_MixHR]
	}
	if t.tpRef[config.MarketType_Human] != nil {
		return t.tpRef[config.MarketType_Human]
	}
	return nil
}

func (t *TickerPool) init(sym string, _size int) error {
	t.sym = sym
	t.size = _size
	t.location, _ = time.LoadLocation("Local")
	t.newestPrice = 0
	tickerInfoInit := TickerInfo{
		From:       t.getLocalTimeUnixNano(),
		End:        t.getLocalTimeUnixNano(),
		OpenPrice:  0,
		ClosePrice: 0,
		LowPrice:   0,
		HightPrice: 0,
		Volume:     0,
		Amount:     0,
	}

	t.tickerRel_1min = tickerInfoInit
	t.ticker_1min = list.New()
	t.ticker_1min.Init()
	t.tickerRel_5min = tickerInfoInit
	t.ticker_5min = list.New()
	t.ticker_5min.Init()
	t.tickerRel_15min = tickerInfoInit
	t.ticker_15min = list.New()
	t.ticker_15min.Init()
	t.tickerRel_30min = tickerInfoInit
	t.ticker_30min = list.New()
	t.ticker_30min.Init()
	t.tickerRel_1hour = tickerInfoInit
	t.ticker_1hour = list.New()
	t.ticker_1hour.Init()
	t.tickerRel_1day = tickerInfoInit
	t.ticker_1day = list.New()
	t.ticker_1day.Init()
	t.tickerRel_1week = tickerInfoInit
	t.ticker_1week = list.New()
	t.ticker_1week.Init()
	t.tickerRel_1month = tickerInfoInit
	t.ticker_1month = list.New()
	t.ticker_1month.Init()

	/// init ds table
	err := use_mysql.TEMySQLInstance().InitTickersTable(t.sym)
	if err != nil {
		panic(err)
	}

	/// init ouput channel
	t.OutChannel = make(chan *TradePair, OUTPUT_CHANNEL_SIZE)

	/// history data init
	t.Load()

	/// latest trade init
	t.LatestTrade.Init()
	t.LatestTrade.Load(t.sym)

	/// latest DS time init
	t.initLatestDSTime()

	/// level orders
	t.initOrderLevels()

	return nil
}

func (t *TickerPool) AddTickerByType(_type TickerType, _info TickerInfo) (*list.Element, *list.List) {

	tickList := t.getTickerList(_type)

	var elem *list.Element = nil
	appendingElem := tickList.Back()
	if appendingElem != nil && t.isInSpan(_type, appendingElem.Value.(TickerInfo).From, _info.From) {
		tickList.Remove(appendingElem)
		elem = tickList.PushBack(_info)
	} else {
		elem = tickList.PushBack(_info)
		if tickList.Len() > t.size {
			tickList.Remove(tickList.Front())
		}
	}

	return elem, tickList
}

func (t *TickerPool) getTickerList(_type TickerType) *list.List {
	var tickers *list.List
	switch _type {
	case TickerType_1min:
		tickers = t.ticker_1min
	case TickerType_5min:
		tickers = t.ticker_5min
	case TickerType_15min:
		tickers = t.ticker_15min
	case TickerType_30min:
		tickers = t.ticker_30min
	case TickerType_1hour:
		tickers = t.ticker_1hour
	case TickerType_1day:
		tickers = t.ticker_1day
	case TickerType_1week:
		tickers = t.ticker_1week
	case TickerType_1month:
		tickers = t.ticker_1month
	default:
		panic(fmt.Errorf("getTickerList error input _type=%s", _type.String()))
	}
	return tickers
}

func (t *TickerPool) getTickerInfo(_type TickerType) *TickerInfo {
	var tickerInfo *TickerInfo
	switch _type {
	case TickerType_1min:
		tickerInfo = &t.tickerRel_1min
	case TickerType_5min:
		tickerInfo = &t.tickerRel_5min
	case TickerType_15min:
		tickerInfo = &t.tickerRel_15min
	case TickerType_30min:
		tickerInfo = &t.tickerRel_30min
	case TickerType_1hour:
		tickerInfo = &t.tickerRel_1hour
	case TickerType_1day:
		tickerInfo = &t.tickerRel_1day
	case TickerType_1week:
		tickerInfo = &t.tickerRel_1week
	case TickerType_1month:
		tickerInfo = &t.tickerRel_1month
	default:
		panic(fmt.Errorf("getTickerInfo error input _type=%s", _type.String()))
	}
	return tickerInfo
}

func (t *TickerPool) getTickerTimeSpan(_type TickerType) int64 {
	var span time.Duration
	switch _type {
	case TickerType_1min:
		span = time.Minute
	case TickerType_5min:
		span = 5 * time.Minute
	case TickerType_15min:
		span = 15 * time.Minute
	case TickerType_30min:
		span = 30 * time.Minute
	case TickerType_1hour:
		span = time.Hour
	case TickerType_1day:
		span = 24 * time.Hour
	case TickerType_1week:
		span = 7 * 24 * time.Hour
	case TickerType_1month: /// Have to special process
		span = 30 * 24 * time.Hour
	default:
		panic(fmt.Errorf("getTickerTimeSpan error input _type=%s", _type.String()))
	}
	return int64(span)
}

func (t *TickerPool) getTruncateTime(_type TickerType, _time int64) int64 {
	var (
		outTime int64 = 0
	)

	switch _type {
	case TickerType_1min:
		outTime = _time / int64(time.Minute) * int64(time.Minute)
	case TickerType_5min:
		outTime = _time / int64(5*time.Minute) * int64(5*time.Minute)
	case TickerType_15min:
		outTime = _time / int64(15*time.Minute) * int64(15*time.Minute)
	case TickerType_30min:
		outTime = _time / int64(30*time.Minute) * int64(30*time.Minute)
	case TickerType_1hour:
		outTime = _time / int64(time.Hour) * int64(time.Hour)
	case TickerType_1day:
		outTime = _time / int64(time.Hour) * int64(time.Hour)
	case TickerType_1week:
		outTime = _time / int64(time.Hour) * int64(time.Hour)
	case TickerType_1month:
		outTime = _time / int64(time.Hour) * int64(time.Hour)
	default:
		panic(fmt.Errorf("getTruncateTime error input _type=%s", _type.String()))
	}

	return outTime
}

func (t *TickerPool) dumpToBuff(_type TickerType, buff *string) *string {

	tickers := t.getTickerList(_type)

	*buff += fmt.Sprintf("==================[TickerPool Dump<%s>]=====================\n", _type.String())
	e := tickers.Front()
	count := 0
	for elem := e; elem != nil; elem = elem.Next() {
		info := elem.Value.(TickerInfo)
		info.DumpTickerInfoToBuff(buff)

		count++
	}
	*buff += fmt.Sprintf("[TickerPool]: Total(%d)\n", count)

	return buff
}

func (t *TickerPool) dumpToPrint(_type TickerType) {
	tickers := t.getTickerList(_type)

	fmt.Printf("==================[TickerPool Dump<%s>]=====================\n", _type.String())
	e := tickers.Front()
	count := 0
	for elem := e; elem != nil; elem = elem.Next() {
		info := elem.Value.(TickerInfo)
		info.DumpTickerInfoToPrint()
		count++
	}
	fmt.Printf("[TickerPool]: Total(%d)\n", count)
}

func (t *TickerPool) DumpAllToPrint() {
	fmt.Printf("==================[DumpTickers Info]=====================\n")
	fmt.Printf("Symbol(%s)\n", t.sym)
	fmt.Printf("Size(%d)\n", t.size)
	//	fmt.Printf("candleStartTime(%d)\n", t.tickerStartTime)
	fmt.Printf("NewestPrice(%f)\n", t.newestPrice)

	t.tickerRel_1min.DumpTickerInfoToPrint()
	t.tickerRel_5min.DumpTickerInfoToPrint()
	t.tickerRel_15min.DumpTickerInfoToPrint()
	t.tickerRel_30min.DumpTickerInfoToPrint()
	t.tickerRel_1hour.DumpTickerInfoToPrint()
	t.tickerRel_1day.DumpTickerInfoToPrint()
	t.tickerRel_1week.DumpTickerInfoToPrint()
	t.tickerRel_1month.DumpTickerInfoToPrint()

	t.dumpToPrint(TickerType_1min)
	t.dumpToPrint(TickerType_5min)
	t.dumpToPrint(TickerType_15min)
	t.dumpToPrint(TickerType_30min)
	t.dumpToPrint(TickerType_1hour)

	t.dumpToPrint(TickerType_1day)
	t.dumpToPrint(TickerType_1week)
	t.dumpToPrint(TickerType_1month)

}

func (t *TickerPool) DumpAllToBuff() string {
	var strBuff string = ""
	strBuff += fmt.Sprintf("==================[DumpTickers Info]=====================\n")
	strBuff += fmt.Sprintf("Symbol(%s)\n", t.sym)
	strBuff += fmt.Sprintf("Size(%d)\n", t.size)
	strBuff += fmt.Sprintf("NewestPrice(%f)\n", t.newestPrice)

	t.tickerRel_1min.DumpTickerInfoToBuff(&strBuff)
	t.tickerRel_5min.DumpTickerInfoToBuff(&strBuff)
	t.tickerRel_15min.DumpTickerInfoToBuff(&strBuff)
	t.tickerRel_30min.DumpTickerInfoToBuff(&strBuff)
	t.tickerRel_1hour.DumpTickerInfoToBuff(&strBuff)
	t.tickerRel_1day.DumpTickerInfoToBuff(&strBuff)
	t.tickerRel_1week.DumpTickerInfoToBuff(&strBuff)
	t.tickerRel_1month.DumpTickerInfoToBuff(&strBuff)

	t.dumpToBuff(TickerType_1min, &strBuff)
	t.dumpToBuff(TickerType_5min, &strBuff)
	t.dumpToBuff(TickerType_15min, &strBuff)
	t.dumpToBuff(TickerType_30min, &strBuff)
	t.dumpToBuff(TickerType_1hour, &strBuff)

	t.dumpToBuff(TickerType_1day, &strBuff)
	t.dumpToBuff(TickerType_1week, &strBuff)
	t.dumpToBuff(TickerType_1month, &strBuff)

	return strBuff
}

func (t *TickerPool) DumpLatestTradePrint() {
	t.LatestTrade.DumpToPrint()
}

func (t *TickerPool) DumpLatestTradeBuff() string {
	var strBuff string = ""

	return *t.LatestTrade.DumpToBuff(&strBuff)
}

///func (t *TickerPool) isInSpan(_type TickerType, timeRef time.Time, now time.Time) bool {
func (t *TickerPool) isInSpan(_type TickerType, timeRef int64, now int64) bool {
	timeNow := time.Unix(0, now).In(t.location)
	timeRel := time.Unix(0, timeRef).In(t.location)

	year, month, day := timeNow.Date()
	_, week := timeNow.ISOWeek() /*Weekday()*/
	hour := timeNow.Hour()
	min := timeNow.Minute()
	yearRel, monthRel, dayRel := timeRel.Date()
	_, weekRel := timeRel.ISOWeek() /*Weekday()*/
	hourRel := timeRel.Hour()
	minRel := timeRel.Minute()

	switch _type {
	case TickerType_1min:
		return year == yearRel && day == dayRel && hour == hourRel && min == minRel
	case TickerType_5min:
		return year == yearRel && day == dayRel && hour == hourRel && min/5 == minRel/5
	case TickerType_15min:
		return year == yearRel && day == dayRel && hour == hourRel && min/15 == minRel/15
	case TickerType_30min:
		return year == yearRel && day == dayRel && hour == hourRel && min/30 == minRel/30
	case TickerType_1hour:
		return year == yearRel && day == dayRel && hour == hourRel
	case TickerType_1day:
		return year == yearRel && day == dayRel
	case TickerType_1week:
		return year == yearRel && week == weekRel
	case TickerType_1month:
		return year == yearRel && month == monthRel
	default:
		panic(fmt.Errorf("isInSpan error input _type=%s", _type.String()))
	}
}

func (t *TickerPool) updateTicker(trade *Trade) {
	t.newestPrice = trade.Price

	for _t := TickerType_1min; _t <= TickerType_1month; _t++ {
		trade.TradeTime = t.getLocalTimeUnixNano()
		t.updateTickerByType(_t, trade)
	}
}

func (t *TickerPool) updateTickerByPeriod(symbol string, _t TickerType, tickInfo *TickerInfo, tradeTime int64, duration int64) {
	if _t <= TickerType_1hour || _t >= TickerType_Num {
		return
	}

	/// inner ticker process:
	if tradeTime-t.LatestDSTime.time[_t] > duration {
		/// to do: record to DS
		tickInfo.From = t.getTruncateTime(_t, tickInfo.From)
		tickInfo.End = tradeTime
		use_mysql.TEMySQLInstance().UpdateTicker(symbol, _t, tickInfo)
		t.LatestDSTime.time[_t] = tradeTime
	}

	DebugPrintf(MODULE_NAME_TICKERS, LOG_LEVEL_TRACK,
		`updateTickerByPeriod in DS: Symbol: %s[%s], From: %d, End: %d, OpenPrice: %f, ClosePrice: %f, LowPrice: %f, HightPrice: %f, Volume: %f	, Amount: %f	
`,
		symbol,
		_t.String(),
		tickInfo.From,
		tickInfo.End,
		tickInfo.OpenPrice,
		tickInfo.ClosePrice,
		tickInfo.LowPrice,
		tickInfo.HightPrice,
		tickInfo.Volume,
		tickInfo.Amount,
	)
}

func (t *TickerPool) updateTickerByType(_t TickerType, trade *Trade) {
	t.newestPrice = trade.Price
	tickInfo := t.getTickerInfo(_t)

	/// inner ticker process:
	/// if t.isInSpan(_t, year, month, day, hour, min, yearRel, monthRel, dayRel, hourRel, minRel) {
	if t.isInSpan(_t, tickInfo.From, trade.TradeTime) {
		if tickInfo.HightPrice < t.newestPrice {
			tickInfo.HightPrice = t.newestPrice
		}
		if tickInfo.LowPrice > t.newestPrice || tickInfo.LowPrice == 0 {
			tickInfo.LowPrice = t.newestPrice
		}

		tickInfo.ClosePrice = t.newestPrice
		tickInfo.Volume += trade.Volume
		tickInfo.Amount += trade.Amount
		if tickInfo.OpenPrice == 0 {
			tickInfo.OpenPrice = t.newestPrice
		}
		tickInfo.End = trade.TradeTime

		DebugPrintf(MODULE_NAME_TICKERS, LOG_LEVEL_TRACK,
			`updateTickerUsingHistrade in memory: Symbol: %s[%s], From: %d, End: %d, OpenPrice: %f, ClosePrice: %f, LowPrice: %f, HightPrice: %f	, Volume: %f	, Amount: %f	
`,
			trade.Symbol,
			_t.String(),
			tickInfo.From,
			tickInfo.End,
			tickInfo.OpenPrice,
			tickInfo.ClosePrice,
			tickInfo.LowPrice,
			tickInfo.HightPrice,
			tickInfo.Volume,
			tickInfo.Amount,
		)
		t.updateTickerByPeriod(trade.Symbol, _t, tickInfo, trade.TradeTime, PERIOD_DS_DURATION_NANO)
	} else {
		/// span ticker process:
		tickerInfo := *tickInfo

		/// update tickerRel to new minute ticker
		tickInfo.From = trade.TradeTime
		tickInfo.End = trade.TradeTime
		tickInfo.OpenPrice = t.newestPrice
		tickInfo.ClosePrice = t.newestPrice
		tickInfo.LowPrice = t.newestPrice
		tickInfo.HightPrice = t.newestPrice
		tickInfo.Volume = trade.Volume
		tickInfo.Amount = trade.Amount

		/// record ticker to Mem&DS
		/// Record to memory
		tickerInfo.From = t.getTruncateTime(_t, tickerInfo.From)
		///tickerInfo.End = trade.TradeTime
		t.AddTickerByType(_t, tickerInfo)

		/// to do: record to DS
		use_mysql.TEMySQLInstance().UpdateTicker(trade.Symbol, _t, &tickerInfo)

		DebugPrintf(MODULE_NAME_TICKERS, LOG_LEVEL_TRACK,
			`updateTickerUsingHistrade in DS: Symbol: %s[%s], From: %d, End: %d, OpenPrice: %f, ClosePrice: %f, LowPrice: %f, HightPrice: %f	, Volume: %f	, Amount: %f	
`,
			trade.Symbol,
			_t.String(),
			tickInfo.From,
			tickInfo.End,
			tickInfo.OpenPrice,
			tickInfo.ClosePrice,
			tickInfo.LowPrice,
			tickInfo.HightPrice,
			tickInfo.Volume,
			tickInfo.Amount,
		)
	}
}

func (t *TickerPool) UpdateTicker(trade *TradePair) {
	t.OutChannel <- trade
}

func (t *TickerPool) GetTicker(_type TickerType) ([]*TickerInfo, error) {
	var (
		tickers []*TickerInfo
		c       int64 = 0
	)

	tickInfo := t.getTickerInfo(_type)
	tickList := t.getTickerList(_type)

	tickers = make([]*TickerInfo, tickList.Len()+1)

	curTicker := *tickInfo
	curTicker.From = curTicker.From / int64(time.Second)
	curTicker.End = curTicker.End / int64(time.Second)
	tickers[0] = &curTicker

	e := tickList.Back()
	if e != nil {
		v := e.Value.(TickerInfo)
		if t.isInSpan(_type, tickInfo.From, v.From) {
			e = e.Prev()
		}
	}

	c = 1
	for ; e != nil; e = e.Prev() {
		v := e.Value.(TickerInfo)
		v.From = v.From / int64(time.Second)
		v.End = v.End / int64(time.Second)
		tickers[c] = &v
		c++
	}

	return tickers, nil
}

func (t *TickerPool) GetTickerLimit(_type TickerType, limit int64) ([]*TickerInfo, error) {
	var (
		tickers []*TickerInfo
		c       int64 = 0
	)

	tickInfo := t.getTickerInfo(_type)
	tickList := t.getTickerList(_type)

	if limit > int64(t.size) {
		limit = int64(t.size)
	}
	limit = MinINT64(limit, int64(tickList.Len()+1))
	if limit == 0 {
		return tickers, nil
	}

	tickers = make([]*TickerInfo, limit)

	curTicker := *tickInfo
	curTicker.From = curTicker.From / int64(time.Second)
	curTicker.End = curTicker.End / int64(time.Second)
	tickers[0] = &curTicker

	e := tickList.Back()
	if e != nil {
		v := e.Value.(TickerInfo)
		if t.isInSpan(_type, tickInfo.From, v.From) {
			e = e.Prev()
		}
	}

	c = 1
	for ; e != nil; e = e.Prev() {
		if c >= limit {
			break
		}

		v := e.Value.(TickerInfo)
		v.From = v.From / int64(time.Second)
		v.End = v.End / int64(time.Second)
		/// tickers = append(tickers, &v)
		tickers[c] = &v
		c++
	}

	return tickers, nil
}

func (t *TickerPool) GetTickerDuration(_type TickerType, from int64, to int64) ([]*TickerInfo, error) {
	var (
		tickers    []*TickerInfo
		backTicker *TickerInfo = nil
	)

	if from > to {
		return nil, fmt.Errorf("GetTickerDuration input time scale error.")
	}

	tickInfo := t.getTickerInfo(_type)
	tickList := t.getTickerList(_type)

	e := tickList.Front()
	for elem := e; elem != nil; elem = elem.Next() {
		v := elem.Value.(TickerInfo)
		if v.From >= from && v.From <= to {
			tickers = append(tickers, &v)
			backTicker = &v
		} else if v.From > to {
			break
		}
	}
	if tickInfo.From >= from && tickInfo.From <= to {
		if backTicker != nil && t.isInSpan(_type, tickInfo.From, backTicker.From) {
			tickers = tickers[0 : len(tickers)-1]
		}
		tickers = append(tickers, tickInfo)
	}

	return tickers, nil
}

func (t *TickerPool) GetQuote() (*QuoteInfo, error) {
	var (
		ch, chp, prev_close_price float64
		ch_1Week, chp_1Week       float64
		ask1st, bid1st            float64 = float64(0), float64(0)
		askOrders, bidOrders      []*Order
	)

	tp := t.getTradePool()
	if tp == nil {
		return nil, fmt.Errorf("GetQuote not ready!")
	}
	askOrders = tp.GetAskLevelOrders(1)
	bidOrders = tp.GetBidLevelOrders(1)
	if len(askOrders) != 0 {
		ask1st = askOrders[0].Price
	}
	if len(bidOrders) != 0 {
		bid1st = bidOrders[0].Price
	}

	tickInfo_1Day := t.getTickerInfo(TickerType_1day)
	tickInfo_1Week := t.getTickerInfo(TickerType_1week)
	ch = t.newestPrice - tickInfo_1Day.OpenPrice
	ch_1Week = t.newestPrice - tickInfo_1Week.OpenPrice
	if tickInfo_1Day.OpenPrice == 0 || tickInfo_1Week.OpenPrice == 0 {
		return nil, fmt.Errorf("GetQuote not ready to work!")
	} else {
		chp = ch / tickInfo_1Day.OpenPrice
		chp_1Week = ch_1Week / tickInfo_1Week.OpenPrice
	}

	tickList := t.getTickerList(TickerType_1day)
	if tickList.Back() == nil {
		prev_close_price = tickInfo_1Day.OpenPrice
	} else {
		prev_close_price = tickList.Back().Value.(TickerInfo).ClosePrice
	}

	return &QuoteInfo{
		Ask1stPrice:      ask1st,
		Bid1stPrice:      bid1st,
		LatestTradePrice: t.newestPrice,
		Spread:           t.spread,
		DayOpenPrice:     tickInfo_1Day.OpenPrice,
		PreDayClosePrice: prev_close_price,
		DayHighPrice:     tickInfo_1Day.HightPrice,
		DayLowPrice:      tickInfo_1Day.LowPrice,
		ChangePrice:      ch,
		ChangePriceRate:  chp,

		DayVolume:          tickInfo_1Day.Volume,
		DayAmount:          tickInfo_1Day.Amount,
		ChangePriceRate_7D: chp_1Week,
	}, nil
}

func (t *TickerPool) GetLatestTradeLimit(limit int64) ([]*Trade, error) {
	return t.LatestTrade.GetTradesLimit(limit)
}

func (t *TickerPool) GetAskLevelOrders(limit int64) ([]*Order, error) {
	tp := t.getTradePool()
	if tp == nil {
		return nil, fmt.Errorf("GetAskLevelOrders not ready!")
	}
	if limit > RESTFUL_MAX_ORDER_LEVELS {
		limit = RESTFUL_MAX_ORDER_LEVELS
	}
	return tp.GetAskLevelOrders(limit), nil
}

func (t *TickerPool) GetBidLevelOrders(limit int64) ([]*Order, error) {
	tp := t.getTradePool()
	if tp == nil {
		return nil, fmt.Errorf("GetBidLevelOrders not ready!")
	}
	if limit > RESTFUL_MAX_ORDER_LEVELS {
		limit = RESTFUL_MAX_ORDER_LEVELS
	}
	return tp.GetBidLevelOrders(limit), nil
}

func (t *TickerPool) getAskLevelsGroupByPrice(limit int64) ([]OrderLevel, error) {
	tp := t.getTradePool()
	if tp == nil {
		return nil, fmt.Errorf("GetAskLevelsGroupByPrice not ready!")
	}
	if limit > RESTFUL_MAX_ORDER_LEVELS {
		limit = RESTFUL_MAX_ORDER_LEVELS
	}
	return tp.GetAskLevelsGroupByPrice(limit), nil
}

func (t *TickerPool) getBidLevelsGroupByPrice(limit int64) ([]OrderLevel, error) {
	tp := t.getTradePool()
	if tp == nil {
		return nil, fmt.Errorf("GetBidLevelsGroupByPrice not ready!")
	}
	if limit > RESTFUL_MAX_ORDER_LEVELS {
		limit = RESTFUL_MAX_ORDER_LEVELS
	}
	return tp.GetBidLevelsGroupByPrice(limit), nil
}

func (t *TickerPool) GetAskLevelsGroupByPrice(limit int64) ([]OrderLevel, error) {
	///return t.getAskLevelsGroupByPrice(limit)
	return t.LevelOrders.askLevel.getLevelOrders(limit)
}

func (t *TickerPool) GetBidLevelsGroupByPrice(limit int64) ([]OrderLevel, error) {
	///return t.getBidLevelsGroupByPrice(limit)
	return t.LevelOrders.bidLevel.getLevelOrders(limit)
}

func (t *TickerPool) Load() {
	t.LoadHistoryTickers()

	t.ReconstructRelTickers()
}

func (t *TickerPool) LoadHistoryTickers() error {
	var (
		tickers []*TickerInfo
	)
	loc, _ := time.LoadLocation("Local")

	for _t := TickerType_1min; _t <= TickerType_1month; _t++ {
		tickers, _ = use_mysql.TEMySQLInstance().GetTickersLimit(t.sym, _t, t.size)
		size := len(tickers)
		for i := size - 1; i >= 0; i-- {
			if i == size-1 {
				fmt.Printf("Load ticker from time: %s to ", time.Unix(0, tickers[i].From).In(loc).Format("2006-01-02T15:04:05Z07:00"))
			}

			t.AddTickerByType(_t, *tickers[i])

			if i == 0 {
				fmt.Printf("%s, total items = %d\n", time.Unix(0, tickers[i].From).In(loc).Format("2006-01-02T15:04:05Z07:00"), size)
			}
		}
		fmt.Printf("[Ticker Engine]: Starting.LoadHistoryTickers."+
			`LoadHistoryTickers load Symbol(%s)-%s tickers complete, %d items loaded.
`,
			t.sym,
			_t.String(),
			size,
		)
	}

	/// to construct rel tickers
	for _t := TickerType_1min; _t <= TickerType_1month; _t++ {
		tickInfo := t.getTickerInfo(_t)
		tickList := t.getTickerList(_t)
		if tickList.Back() != nil {
			lastTicker := tickList.Back().Value.(TickerInfo)
			tickInfo.From = lastTicker.End
			tickInfo.End = lastTicker.End
			tickInfo.OpenPrice = lastTicker.ClosePrice
			tickInfo.ClosePrice = lastTicker.ClosePrice
			tickInfo.LowPrice = lastTicker.ClosePrice
			tickInfo.HightPrice = lastTicker.ClosePrice
			tickInfo.Volume = 0
			tickInfo.Amount = 0
		}
		fmt.Printf("[Ticker Engine]: Starting to construct rel tickers."+
			`LoadHistoryTickers to construct rel tickInfo Symbol(%s)-%s complete.
`,
			t.sym,
			_t.String(),
		)
	}

	return nil
}

func (t *TickerPool) ReconstructRelTickers() error {
	var (
		count    int    = 0
		elem     *Trade = nil
		trades   []*Trade
		_type    TickerType  = TickerType_1min
		tickList *list.List  = nil
		tickInfo *TickerInfo = nil
	)
	loc, _ := time.LoadLocation("Local")

	for _type = TickerType_1min; _type <= TickerType_1month; _type++ {
		tickInfo = t.getTickerInfo(_type)
		tickList = t.getTickerList(_type)
		if tickList.Back() != nil {
			TimeDot1 := time.Now().UnixNano()
			fmt.Printf("[%s]Begin to load trades from ds from time(%d)\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), tickList.Back().Value.(TickerInfo).End)
			trades, _ = use_mysql.MEMySQLInstance().GetRelTradeForTickers(t.sym, tickList.Back().Value.(TickerInfo).End)
			fmt.Printf("[%s]Complete to load trades from ds, %d items loaded.\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), len(trades))
			*tickInfo = tickList.Back().Value.(TickerInfo)
			for count, elem = range trades {
				t.updateTickerByType(_type, elem)
			}
			fmt.Printf("[%s]Complete to load trades to memory using updateTickerByType(%s), %d items loaded.\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), _type, count)
			TimeDot2 := time.Now().UnixNano()
			fmt.Printf("[Ticker Engine]: Starting to ReconstructRelTickers."+
				`ReconstructRelTickers Symbol(%s) tickers in recent %s (from time: %d)complete, %d items trades used, USE_TIME= %f(second).
`,
				t.sym,
				_type.String(),
				tickList.Back().Value.(TickerInfo).End,
				count,
				float64(TimeDot2-TimeDot1)/float64(1*time.Second),
			)
		}
	}

	return nil
}

func (t *TickerPool) IsInitialized(_t TickerType) bool {
	tickList := t.getTickerList(_t)
	if tickList.Len() == 0 || tickList.Back() == nil {
		return false
	} else {
		return true
	}
}

func (t *TickerPool) ConstructTickersFromHistoryTradesByTickerType(_type TickerType, from int64) error {
	var (
		count  int    = 0
		elem   *Trade = nil
		trades []*Trade

		tickerInfoInit TickerInfo = TickerInfo{
			From:       0,
			End:        0,
			OpenPrice:  0,
			ClosePrice: 0,
			LowPrice:   0,
			HightPrice: 0,
			Volume:     0,
			Amount:     0,
		}
	)

	loc, _ := time.LoadLocation("Local")
	tickInfo := t.getTickerInfo(_type)
	*tickInfo = tickerInfoInit
	tickList := t.getTickerList(_type)
	tickList.Init()

	TimeDot1 := time.Now().UnixNano()
	fmt.Printf("[%s]Begin to load trades from ds from time(%d)\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), from)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "[%s]Begin to load trades from ds from time(%d)\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), from)
	trades, _ = use_mysql.MEMySQLInstance().GetRelTradeForTickers(t.sym, from)
	fmt.Printf("[%s]Complete to load trades from ds, %d items loaded.\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), len(trades))
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "[%s]Complete to load trades from ds, %d items loaded.\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), len(trades))

	for count, elem = range trades {
		t.updateTickerByType(_type, elem)
	}

	fmt.Printf("[%s]Complete to load trades to memory using updateTickerByType(%s), %d items loaded.\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), _type, count)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "[%s]Complete to load trades to memory using updateTickerByType(%s), %d items loaded.\n", time.Now().In(loc).Format("2006-01-02T15:04:05Z07:00"), _type, count)

	TimeDot2 := time.Now().UnixNano()
	fmt.Printf("[Ticker Engine]: "+
		`ConstructTickersFromHistoryTrades Symbol(%s) tickers in recent %s (from time: %d)complete, %d items trades used, USE_TIME= %f(second).
`,
		t.sym,
		_type.String(),
		from,
		count,
		float64(TimeDot2-TimeDot1)/float64(1*time.Second),
	)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "[Ticker Engine]: "+
		`ConstructTickersFromHistoryTrades Symbol(%s) tickers in recent %s (from time: %d)complete, %d items trades used, USE_TIME= %f(second).
`,
		t.sym,
		_type.String(),
		from,
		count,
		float64(TimeDot2-TimeDot1)/float64(1*time.Second),
	)

	return nil
}

func (t *TickerPool) ConstructTickersFromHistoryTrades(from int64) error {
	var (
		_type TickerType = TickerType_1min
	)

	for _type = TickerType_1min; _type <= TickerType_1month; _type++ {
		t.ConstructTickersFromHistoryTradesByTickerType(_type, from)
	}

	return nil
}

func (t *TickerPool) outPutProcess() {
	var (
		v  *TradePair = nil
		ok bool       = false
	)

	for {
		v, ok = <-t.OutChannel
		if ok {
			t.updateTicker(v.askTrade)

			tradeTime := time.Now().UnixNano()
			v.bidTrade.TradeTime = tradeTime
			v.askTrade.TradeTime = tradeTime

			t.LatestTrade.AddTrade(v.bidTrade)
			t.LatestTrade.AddTrade(v.askTrade)

			DebugPrintf(MODULE_NAME_TICKERS, LOG_LEVEL_TRACK,
				`------------------------------------>>>>>>>>>>[%s]Ticker Output with Trade:
	ID: %d, Price: %f, Volume: %f, Amount: %f, TradeTime: %d	
`,
				v.askTrade.Symbol,
				v.askTrade.ID,
				v.askTrade.Price,
				v.askTrade.Volume,
				v.askTrade.Amount,
				v.askTrade.TradeTime,
			)

		} else {
			panic(fmt.Errorf("TickerPool OutChannel outPutProcess exception occur!"))
		}
		runtime.Gosched()
	}
}
