package comm

type ExInterface interface {

	//---------------------------------------------------------------------------------------------------
	EnOrder(order *Order) error
	CancelOrder(id int64) error
	CancelTheOrder(order *Order) error
	CreateID() int64
	//---------------------------------------------------------------------------------------------------

}

type MxInterface interface {

	//---------------------------------------------------------------------------------------------------
	Start()
	Stop()
	Destroy()
	//---------------------------------------------------------------------------------------------------

}

type MTInterface interface {

	//---------------------------------------------------------------------------------------------------
	GetAskLevelOrders(limit int64) []*Order
	GetBidLevelOrders(limit int64) []*Order
	GetAskLevelsGroupByPrice(limit int64) []OrderLevel
	GetBidLevelsGroupByPrice(limit int64) []OrderLevel
	//---------------------------------------------------------------------------------------------------

}

type MonitorInterface interface {
	//---------------------------------------------------------------------------------------------------
	DumpTradePool(detail bool) string
	DumpTradePoolPrint(detail bool)
	DumpBeatHeart() string
	DumpChannel() string
	DumpChanlsMap()
	IsFaulty() bool
	RestartDebuginfo()
	ResetMatchCorePerform()
	Statics() string
	PrintHealth()
	Test(u string, p ...interface{})
	TradeCommand(u string, p ...interface{})

	GetTradeCompleteRate() float64
	GetAskPoolLen() int
	GetBidPoolLen() int
	GetPoolLen() int
	//---------------------------------------------------------------------------------------------------
}

type MatchCoreItf interface {
	//---------------------------------------------------------------------------------------------------
	ExInterface
	MonitorInterface
	MTInterface
	//---------------------------------------------------------------------------------------------------
}
