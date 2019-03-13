package runtime

import (
	"time"

	"../../comm"
)

const (
	MODULE_NAME string = "[Match Core]: "
)

type DebugUnit struct {
	addTimes      int64
	quickAddTimes int64

	tradeOutputTimes   int64
	tradeCompleteTimes int64
}

func NewDebugUnit() *DebugUnit {
	var obj = new(DebugUnit)
	obj.Reset()

	return obj
}

func (t *DebugUnit) AddTimesInc() {
	t.addTimes++
}
func (t *DebugUnit) QuickAddTimesInc() {
	t.quickAddTimes++
}
func (t *DebugUnit) TradeOutputTimesInc() {
	t.tradeOutputTimes++
}
func (t *DebugUnit) TradeCompleteTimesInc() {
	t.tradeCompleteTimes++
}
func (t *DebugUnit) Reset() {
	t.addTimes = 0
	t.quickAddTimes = 0

	t.tradeOutputTimes = 0
	t.tradeCompleteTimes = 0
}

type MatchCorePerform struct {
	max   int64
	min   int64
	ave   int64
	sum   int64
	count int64
}

func (t *MatchCorePerform) Record(v int64) {
	t.count++
	t.sum += v
	if v > t.max {
		t.max = v
	}
	if v < t.min {
		t.min = v
	}
	t.ave = t.sum / t.count
}

func (t *MatchCorePerform) Reset() {
	t.max, t.min, t.ave, t.sum, t.count = 0, comm.MAX_INT64, 0, 0, 0
}

func (t *MatchCorePerform) GetPerform() (max, min, ave float64) {
	return float64(t.max) / float64(1*time.Second), float64(t.min) / float64(1*time.Second), float64(t.ave) / float64(1*time.Second)
}

type DebugInfo struct {
	askPoolDebugInfo DebugUnit
	bidPoolDebugInfo DebugUnit
	MatchCorePerform

	startTime int64
}

func NewDebugInfo() *DebugInfo {
	var obj = new(DebugInfo)
	obj.askPoolDebugInfo = *NewDebugUnit()
	obj.bidPoolDebugInfo = *NewDebugUnit()

	obj.startTime = time.Now().UnixNano()
	obj.MatchCorePerform = MatchCorePerform{0, comm.MAX_INT64, 0, 0, 0}

	return obj
}

func (t *DebugInfo) DebugInfo_RestartDebuginfo() {
	t.askPoolDebugInfo.Reset()
	t.bidPoolDebugInfo.Reset()
	t.startTime = time.Now().UnixNano()
}

func (t *DebugInfo) DebugInfo_AskEnOrderNormalAdd() {
	t.askPoolDebugInfo.AddTimesInc()
}

func (t *DebugInfo) DebugInfo_AskEnOrderQuickAdd() {
	t.askPoolDebugInfo.QuickAddTimesInc()
}

func (t *DebugInfo) DebugInfo_BidEnOrderNormalAdd() {
	t.bidPoolDebugInfo.AddTimesInc()
}

func (t *DebugInfo) DebugInfo_BidEnOrderQuickAdd() {
	t.bidPoolDebugInfo.QuickAddTimesInc()
}

func (t *DebugInfo) DebugInfo_AskTradeOutputAdd() {
	t.askPoolDebugInfo.TradeOutputTimesInc()
}

func (t *DebugInfo) DebugInfo_AskTradeCompleteAdd() {
	t.askPoolDebugInfo.TradeCompleteTimesInc()
}

func (t *DebugInfo) DebugInfo_BidTradeOutputAdd() {
	t.bidPoolDebugInfo.TradeOutputTimesInc()
}

func (t *DebugInfo) DebugInfo_BidTradeCompleteAdd() {
	t.bidPoolDebugInfo.TradeCompleteTimesInc()
}

func (t *DebugInfo) DebugInfo_GetEnOrders() int64 {
	return t.askPoolDebugInfo.addTimes +
		t.askPoolDebugInfo.quickAddTimes +
		t.bidPoolDebugInfo.addTimes +
		t.bidPoolDebugInfo.quickAddTimes
}

func (t *DebugInfo) DebugInfo_GetUserAskEnOrders() int64 {
	return t.askPoolDebugInfo.addTimes
}

func (t *DebugInfo) DebugInfo_GetUserBidEnOrders() int64 {
	return t.bidPoolDebugInfo.addTimes
}

func (t *DebugInfo) DebugInfo_GetUserEnOrders() int64 {
	return t.bidPoolDebugInfo.addTimes + t.askPoolDebugInfo.addTimes
}

func (t *DebugInfo) DebugInfo_GetAskEnOrders() int64 {
	return t.askPoolDebugInfo.addTimes + t.askPoolDebugInfo.quickAddTimes
}

func (t *DebugInfo) DebugInfo_GetBidEnOrders() int64 {
	return t.bidPoolDebugInfo.addTimes + t.bidPoolDebugInfo.quickAddTimes
}

func (t *DebugInfo) DebugInfo_GetTradeOutputs() int64 {
	return t.bidPoolDebugInfo.tradeOutputTimes + t.askPoolDebugInfo.tradeOutputTimes
}

func (t *DebugInfo) DebugInfo_GetTradeCompletes() int64 {
	return t.bidPoolDebugInfo.tradeCompleteTimes + t.askPoolDebugInfo.tradeCompleteTimes
}

func (t *DebugInfo) DebugInfo_GetTradeCompleteRate() float64 {
	return float64(t.DebugInfo_GetTradeCompletes()) / float64((time.Now().UnixNano()-t.startTime)/int64(time.Second))
}

func (t *DebugInfo) DebugInfo_GetTradeOutputRate() float64 {
	return float64(t.DebugInfo_GetTradeOutputs()) / float64((time.Now().UnixNano()-t.startTime)/int64(time.Second))
}

func (t *DebugInfo) DebugInfo_GetUserEnOrderRate() float64 {
	return float64(t.DebugInfo_GetUserEnOrders()) / float64((time.Now().UnixNano()-t.startTime)/int64(time.Second))
}

func (t *DebugInfo) DebugInfo_RecordCorePerform(v int64) {
	t.MatchCorePerform.Record(v)
}

func (t *DebugInfo) DebugInfo_GetCorePerform() (max, min, ave float64) {
	return t.MatchCorePerform.GetPerform()
}

func (t *DebugInfo) DebugInfo_ResetMatchCorePerform() {
	t.MatchCorePerform.Reset()
}
