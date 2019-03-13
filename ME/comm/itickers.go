package comm

import (
	"fmt"
)

/// store 1,5,15,30min, and 1hour, 1week, 1month tickers
type TickerType int64

const (
	TickerType_Invalid TickerType = 0
	TickerType_1min    TickerType = 1
	TickerType_5min    TickerType = 2
	TickerType_15min   TickerType = 3
	TickerType_30min   TickerType = 4
	TickerType_1hour   TickerType = 5
	TickerType_1day    TickerType = 6
	TickerType_1week   TickerType = 7
	TickerType_1month  TickerType = 8

	TickerType_Num TickerType = 9
)

func (t TickerType) String() string {
	switch t {
	case TickerType_1min:
		return "Ticker_1min"
	case TickerType_5min:
		return "Ticker_5min"
	case TickerType_15min:
		return "Ticker_15min"
	case TickerType_30min:
		return "Ticker_30min"
	case TickerType_1hour:
		return "Ticker_1hour"
	case TickerType_1day:
		return "Ticker_1day"
	case TickerType_1week:
		return "Ticker_1week"
	case TickerType_1month:
		return "Ticker_1month"

	}
	return "<TickerType_UNSET>"
}

/// ticker info
type TickerInfo struct {
	From       int64
	End        int64
	OpenPrice  float64
	ClosePrice float64
	LowPrice   float64
	HightPrice float64
	Volume     float64
	Amount     float64
}

func (t *TickerInfo) DumpTickerInfoToPrint() {
	fmt.Printf(
		"TickerInfo: from:%d, end:%d, open:%f, close:%f, low:%f, high:%f, volume:%f, amount:%f\n",
		t.From,
		t.End,
		t.OpenPrice,
		t.ClosePrice,
		t.LowPrice,
		t.HightPrice,
		t.Volume,
		t.Amount,
	)
}

func (t *TickerInfo) DumpTickerInfoToBuff(buff *string) *string {
	*buff += fmt.Sprintf(
		"TickerInfo: from:%d, end:%d, open:%f, close:%f, low:%f, high:%f, volume:%f, amount:%f\n",
		t.From,
		t.End,
		t.OpenPrice,
		t.ClosePrice,
		t.LowPrice,
		t.HightPrice,
		t.Volume,
		t.Amount,
	)
	return buff
}

/// ticker pack unit
type TickerUnit struct {
	Symbol string
	Type   TickerType
	*TickerInfo
}

type QuoteInfo struct {
	Ask1stPrice      float64
	Bid1stPrice      float64
	LatestTradePrice float64
	Spread           float64
	DayOpenPrice     float64
	PreDayClosePrice float64
	DayHighPrice     float64
	DayLowPrice      float64

	ChangePrice     float64
	ChangePriceRate float64

	DayVolume          float64
	DayAmount          float64
	ChangePriceRate_7D float64
}

func (t *QuoteInfo) DumpQuoteInfoToPrint() {
	fmt.Printf(
		"QuoteInfo: LatestTradePrice:%f, Ask1stPrice:%f, Bid1stPrice:%f, Spread:%f, DayOpenPrice:%f, PreDayClosePrice:%f, DayHighPrice:%f, DayLowPrice:%f, DayVolume:%f, ChangePrice:%f, ChangePriceRate:%f\n",
		t.LatestTradePrice,
		t.Ask1stPrice,
		t.Bid1stPrice,
		t.Spread,
		t.DayOpenPrice,
		t.PreDayClosePrice,
		t.DayHighPrice,
		t.DayLowPrice,
		t.DayVolume,
		t.ChangePrice,
		t.ChangePriceRate,
	)
}

func (t *QuoteInfo) DumpQuoteInfoToBuff(buff *string) *string {
	*buff += fmt.Sprintf(
		"QuoteInfo: LatestTradePrice:%f, Ask1stPrice:%f, Bid1stPrice:%f, Spread:%f, DayOpenPrice:%f, PreDayClosePrice:%f, DayHighPrice:%f, DayLowPrice:%f, DayVolume:%f, ChangePrice:%f, ChangePriceRate:%f\n",
		t.LatestTradePrice,
		t.Ask1stPrice,
		t.Bid1stPrice,
		t.Spread,
		t.DayOpenPrice,
		t.PreDayClosePrice,
		t.DayHighPrice,
		t.DayLowPrice,
		t.DayVolume,
		t.ChangePrice,
		t.ChangePriceRate,
	)
	return buff
}
