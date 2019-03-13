package comm

type IdbShortTerm interface {

	//---------------------------------------------------------------------------------------------------
	AddOrder(order *Order) error
	RmOrder(user string, id int64, symbol string) error
	GetOrder(user string, id int64, symbol string) (*Order, error)
	GetAllOrder(symbol string) (so []*Order, err error)
	GetOnesOrder(user string, symbol string) (so []*Order, err error)

	//---------------------------------------------------------------------------------------------------
	AddTrade(trade *Trade) error
	RmTrade(user string, id int64, symbol string) error
	GetTrade(user string, id int64, symbol string) (*Trade, error)
	GetAllTrade(symbol string) (so []*Trade, err error)
	GetOnesTrade(user string, symbol string) (so []*Trade, err error)

	//---------------------------------------------------------------------------------------------------

}

type IdbLongTerm interface {

	//---------------------------------------------------------------------------------------------------
	AddTrade(trade *Trade) error
	RmTrade(user string, id int64, symbol string) error
	GetTrade(user string, id int64, symbol string) (*Trade, error)
	GetAllTrade(symbol string) (so []*Trade, err error)
	GetOnesTrade(user string, symbol string) (so []*Trade, err error)

	//---------------------------------------------------------------------------------------------------
	GetFund(user string) (*Fund, error)
	FreezeFund(order *Order) error
	SettleAccount(trade *Trade) error
}
