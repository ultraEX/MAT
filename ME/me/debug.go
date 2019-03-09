// debug
package me

import (
	"time"
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
	obj.addTimes = 0
	obj.quickAddTimes = 0

	obj.tradeOutputTimes = 0
	obj.tradeCompleteTimes = 0

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

type DebugInfo struct {
	askPoolDebugInfo DebugUnit
	bidPoolDebugInfo DebugUnit

	startTime int64
}

func NewDebugInfo() *DebugInfo {
	var obj = new(DebugInfo)
	obj.askPoolDebugInfo = *NewDebugUnit()
	obj.bidPoolDebugInfo = *NewDebugUnit()

	obj.startTime = time.Now().UnixNano()

	return obj
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
