package core

import (
	"fmt"
	"time"

	"../../config"
)

func (t *MEXCore) DumpTradePool(detail bool) string {
	return "heap map algorithm monitor itf"
}

func (t *MEXCore) DumpTradePoolPrint(detail bool) {
	fmt.Printf("==================[%s-%s Dump MEXCore Trade container]=====================\n", t.Symbol, t.MarketType.String())
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	fmt.Printf("Date Time: %s\n", time.Now().In(loc).Format(formate))

	t.OrderContainerItf.Dump()

	fmt.Printf("=============================================================\n")
}

func (t *MEXCore) DumpBeatHeart() string {
	return "to implement!"
}

func (t *MEXCore) DumpChannel() string {
	return "to implement!"
}

func (t *MEXCore) DumpChanlsMap() {
	fmt.Printf("==================[%s-%s Channel Map Infoo]=====================\n", t.Symbol, t.MarketType.String())
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	fmt.Printf("Date Time: %s\n", time.Now().In(loc).Format(formate))
	t.MultiChans_Out.Dump()
	fmt.Printf("=======================================================\n")
}

func (t *MEXCore) IsFaulty() bool {
	/// to do

	return false
}

///------------------------------------------------------------------
func (t *MEXCore) RestartDebuginfo() {
	t.DebugInfo_RestartDebuginfo()
}

func (t *MEXCore) ResetMatchCorePerform() {
	t.DebugInfo_ResetMatchCorePerform()
}

func (t *MEXCore) Statics() string {
	fmt.Printf("===============[Market %s-%s Trade Info]==============\n", t.Symbol, t.MarketType.String())
	fmt.Printf("===================(User Input Order)====================\n")
	fmt.Printf("ASK ORDERS		: %d\n", t.DebugInfo_GetUserAskEnOrders())
	fmt.Printf("BID ORDERS		: %d\n", t.DebugInfo_GetUserBidEnOrders())
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("====================(Output+Complete)====================\n")
	fmt.Printf("TRADE OUTPUTS	: %d\n", t.DebugInfo_GetTradeOutputs())
	fmt.Printf("TRADE COMPLETES	: %d\n", t.DebugInfo_GetTradeCompletes())
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("Ask Pool Scale	:	%d\n", t.OrderContainerItf.AskSize())
	fmt.Printf("Bid Pool Scale	:	%d\n", t.OrderContainerItf.BidSize())
	fmt.Printf("Newest Price	:	%f\n", t.latestPrice)
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("=====================[Trade Statics]=====================\n")
	fmt.Printf("Trade Complete Rate	: %f\n", t.DebugInfo_GetTradeCompleteRate())
	fmt.Printf("Trade Output Rate	: %f\n", t.DebugInfo_GetTradeOutputRate())
	fmt.Printf("Trade UserInput Rate	: %f\n", t.DebugInfo_GetUserEnOrderRate())
	fmt.Printf("----------------------------------------------------------\n")
	max, min, ave := t.DebugInfo_GetCorePerform()
	fmt.Printf("Match core performance(second/round):\n\tmin=%.9f, max=%.9f, ave=%.9f\n", min, max, ave)
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("InChannel Pool Work Mode:	%s\n", config.GetMEConfig().InPoolMode)
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("MultiChans_In Pool Size	:	%d\n", t.MultiChans_In.Len())
	fmt.Printf("MultiChans_In Buff Size	:	%d\n", t.MultiChans_In.ChanCap())
	t.MultiChans_In.Summary()
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("----------------------------------------------------------\n")
	fmt.Printf("MultiChans_Out Pool Size	:	%d\n", t.MultiChans_Out.Len())
	fmt.Printf("MultiChans_Out Buff Size	:	%d\n", t.MultiChans_Out.ChanCap())
	t.MultiChans_Out.Summary()
	fmt.Printf("----------------------------------------------------------\n")
	IDs, CHs, chnums := t.MultiChans_Out.GetChanUseStatus()
	fmt.Printf("MultiChansOut Chans Usage status: %d; %d; %d\n", IDs, CHs, chnums)

	fmt.Printf("=======================================================\n")

	return "heap map algorithm monitor itf"
}

func (t *MEXCore) PrintHealth() {
	// to do
}

func (t *MEXCore) Test(u string, p ...interface{}) {
	// to do
}

func (t *MEXCore) TradeCommand(u string, p ...interface{}) {
	// to do
}

func (t *MEXCore) GetTradeCompleteRate() float64 {
	return t.DebugInfo_GetTradeCompleteRate()
}

func (t *MEXCore) GetAskPoolLen() int {
	return int(t.OrderContainerItf.AskSize())
}

func (t *MEXCore) GetBidPoolLen() int {
	return int(t.OrderContainerItf.BidSize())
}

func (t *MEXCore) GetPoolLen() int {
	return int(t.OrderContainerItf.TheSize())
}
