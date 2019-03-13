// testME
package use_mysql

import (
	"fmt"

	. "../../comm"
	"../../config"
)

//Order:--------------
func (t *MEMySQLDB) testMysql_AddOrder(order *Order) {
	if err, _ := t.AddOrder(order, nil); err != nil {
		panic(err)
	}

	fmt.Println("testMysql_AddOrder execute complete! Please check it use: SELECT * FROM ", TABLE_ORDER)
}

func (t *MEMySQLDB) testMysql_UpdateOrder(order *Order) {
	if err := t.UpdateOrder(order, nil); err != nil {
		panic(err)
	}

	fmt.Println("testMysql_UpdateOrder execute complete! Please check it use: SELECT * FROM ", TABLE_ORDER)
}

func (t *MEMySQLDB) testMysql_RmOrder(user string, id int64, symbol string) {
	if err := t.RmOrder(user, id, symbol, nil); err != nil {
		panic(err)
	}

	fmt.Println("testMysql_RmOrder execute complete! Please check it use: SELECT * FROM ", TABLE_ORDER, " WHERE orders_id=", id)
}

func (t *MEMySQLDB) testMysql_RmOrderCouple(bid *Order, ask *Order) {
	if err := t.RmOrderCouple(bid, ask, nil); err != nil {
		panic(err)
	}

	fmt.Println("testMysql_RmOrderCouple execute complete! Please check it use: SELECT * FROM ", TABLE_ORDER, " WHERE orders_id=", bid.ID, " or orders_id=", ask.ID)
}

func (t *MEMySQLDB) testMysql_GetOrder(user string, id int64, symbol string) {
	order, err := t.GetOrder(user, id, symbol, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("GetOrder Return:\nid: ", order.ID,
		"; who: ", order.Who,
		"; AorB: ", order.AorB,
		"; Symbol: ", order.Symbol,
		"; Timestamp: ", order.Timestamp,
		"; price: ", order.Price,
		"; volume: ", order.Volume,
		"; TotalVolume: ", order.TotalVolume,
		"; Fee: ", order.Fee,
		"; Status: ", order.Status,

		"\n")
}

func (t *MEMySQLDB) testMysql_GetAllOrder(symbol string) {
	orders, err := t.GetAllOrder(symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range orders {
		fmt.Println("GetAllOrder Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,

			"\n")
	}
}

func (t *MEMySQLDB) testMysql_GetOnesOrder(user string, symbol string) {
	orders, err := t.GetOnesOrder(user, symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range orders {
		fmt.Println("GetOnesOrder Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,

			"\n")
	}
}

func (t *MEMySQLDB) testMysql_GetOnesOverTimeOrder(symbol string, users []interface{}, ot int64) {
	orders, err := t.GetOnesOverTimeOrder(symbol, users, ot)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range orders {
		fmt.Println("GetOnesOverTimeOrder Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,

			"\n")
	}
}

//Trade:--------------
func (t *MEMySQLDB) testMysql_AddTrade(trade *Trade) {
	if err := t.AddTrade(trade, nil); err != nil {
		panic(err)
	}

	fmt.Println("testMysql_AddTrade execute complete! Please check it use: SELECT * FROM ", TABLE_TRADE)
}

func (t *MEMySQLDB) testMysql_AddTradeCouple(bid *Trade, ask *Trade) {
	if err := t.AddTradeCouple(bid, ask, nil); err != nil {
		panic(err)
	}

	fmt.Println("testMysql_AddTradeCouple execute complete! Please check it use: SELECT * FROM ", TABLE_TRADE)
}

func (t *MEMySQLDB) testMysql_RmTrade(user string, id int64, symbol string) {
	if err := t.RmTrade(user, id, symbol); err != nil {
		panic(err)
	}

	fmt.Println("testMysql_RmTrade execute complete! Please check it use: SELECT * FROM ", TABLE_TRADE, " WHERE orders_id=", id)
}

func (t *MEMySQLDB) testMysql_GetTrade(user string, id int64, symbol string) {
	trade, err := t.GetTrade(user, id, symbol)
	if err != nil {
		panic(err)
	}

	fmt.Println("GetTrade Return:\nid: ", trade.ID,
		"; who: ", trade.Who,
		"; AorB: ", trade.AorB,
		"; Symbol: ", trade.Symbol,
		"; Timestamp: ", trade.Timestamp,
		"; price: ", trade.Price,
		"; volume: ", trade.Volume,
		"; TotalVolume: ", trade.TotalVolume,
		"; Fee: ", trade.Fee,
		"; Status: ", trade.Status,
		"; Amount: ", trade.Amount,
		"; TradeTime", trade.TradeTime,
		"; FeeCost", trade.FeeCost,

		"\n")
}

func (t *MEMySQLDB) testMysql_GetAllTrade(symbol string) {
	trades, err := t.GetAllTrade(symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range trades {
		fmt.Println("GetAllTrade Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,
			"; Amount: ", elem.Amount,
			"; TradeTime", elem.TradeTime,
			"; FeeCost", elem.FeeCost,

			"\n")
	}
}

func (t *MEMySQLDB) testMysql_GetOnesTrade(user string, symbol string) {
	trades, err := t.GetOnesTrade(user, symbol)
	if err != nil {
		panic(err)
	}

	/// debug:
	for _, elem := range trades {
		fmt.Println("GetOnesTrade Return:\nid: ", elem.ID,
			"; who: ", elem.Who,
			"; AorB: ", elem.AorB,
			"; Symbol: ", elem.Symbol,
			"; Timestamp: ", elem.Timestamp,
			"; price: ", elem.Price,
			"; volume: ", elem.Volume,
			"; TotalVolume: ", elem.TotalVolume,
			"; Fee: ", elem.Fee,
			"; Status: ", elem.Status,
			"; Amount: ", elem.Amount,
			"; TradeTime", elem.TradeTime,
			"; FeeCost", elem.FeeCost,

			"\n")
	}
}

/// Fund
func (t *MEMySQLDB) testMysql_GetFund(user string) {
	fmt.Println("=================================\n")

	fund, err := t.GetFund(user)
	if err != nil {
		panic(err)
	}

	/// get output
	coinMapMark := config.GetCoinNames()
	fmt.Println("GetFund Return:\nUser: ", user)
	for _, c := range coinMapMark {
		fmt.Printf("\nCoin Mark==%s: ", c)
		if v, ok := fund.AvailableMoney[c]; ok {
			fmt.Printf("	Available: %f; ", v)
		}
		if v, ok := fund.FreezedMoney[c]; ok {
			fmt.Printf("	Frozen: %f; ", v)
		}
		if v, ok := fund.TotalMoney[c]; ok {
			fmt.Printf("	Total:  %f; ", v)
		}
		if v, ok := fund.Status[c]; ok {
			fmt.Printf("	Status:  %s; ", v.String())
		}
	}

	fmt.Println("\n=================================\n")
}

func (t *MEMySQLDB) testMysql_FreezeFund(order *Order) {
	fmt.Println("=================================\n")

	fmt.Println("FreezeFund test start with order:\n")
	t.testMysql_GetFund(order.Who)

	fmt.Println("ID: ", order.ID,
		"; Who: ", order.Who,
		"; AorB: ", order.AorB,
		"; Symbol: ", order.Symbol,
		"; Timestamp: ", order.Timestamp,
		"; price: ", order.Price,
		"; volume: ", order.Volume,
		"; TotalVolume: ", order.TotalVolume,
		"; Fee: ", order.Fee,

		"\n")

	err, _ := t.FreezeFund(order, nil)
	if err != nil {
		panic(err)
	}

	/// get output
	fmt.Println("FreezeFund Complete, GetFUnd to check:\n")
	t.testMysql_GetFund(order.Who)

	fmt.Println("=================================\n")
}

func (t *MEMySQLDB) testMysql_UnfreezeFund(order *Order) {
	fmt.Println("=================================\n")

	fmt.Println("UnfreezeFund test start with order:\n")
	fmt.Println("ID: ", order.ID,
		"; Who: ", order.Who,
		"; AorB: ", order.AorB,
		"; Symbol: ", order.Symbol,
		"; Timestamp: ", order.Timestamp,
		"; price: ", order.Price,
		"; volume: ", order.Volume,
		"; TotalVolume: ", order.TotalVolume,
		"; Fee: ", order.Fee,

		"\n")
	t.testMysql_GetFund(order.Who)

	err, _ := t.UnfreezeFund(order, nil)
	if err != nil {
		panic(err)
	}

	/// get output
	fmt.Println("UnfreezeFund Complete, GetFUnd to check:\n")
	t.testMysql_GetFund(order.Who)

	fmt.Println("=================================\n")
}

func (t *MEMySQLDB) testMysql_SettleAccount(trade *Trade) {
	fmt.Println("=================================\n")
	t.testMysql_GetFund(trade.Who)

	fmt.Println("SettleAccount test start with trade:\n")
	fmt.Println("ID: ", trade.ID,
		"; Who: ", trade.Who,
		"; AorB: ", trade.AorB,
		"; Symbol: ", trade.Symbol,
		"; Timestamp: ", trade.Timestamp,
		"; price: ", trade.Price,
		"; volume: ", trade.Volume,
		"; TotalVolume: ", trade.TotalVolume,
		"; Fee: ", trade.Fee,
		"; Amount: ", trade.Amount,
		"; TradeTime", trade.TradeTime,
		"; Status: ", trade.Status,

		"\n")

	err := t.SettleAccount(trade)
	if err != nil {
		panic(err)
	}

	/// get output
	fmt.Println("SettleAccount Complete, GetFUnd to check:\n")
	t.testMysql_GetFund(trade.Who)

	fmt.Println("=================================\n")
}

func (t *MEMySQLDB) testMysql_SettleAccountQuick(trade *Trade) {
	fmt.Println("=================================\n")
	t.testMysql_GetFund(trade.Who)

	fmt.Println("SettleAccount test start with trade:\n")
	fmt.Println("ID: ", trade.ID,
		"; Who: ", trade.Who,
		"; AorB: ", trade.AorB,
		"; Symbol: ", trade.Symbol,
		"; Timestamp: ", trade.Timestamp,
		"; price: ", trade.Price,
		"; volume: ", trade.Volume,
		"; TotalVolume: ", trade.TotalVolume,
		"; Fee: ", trade.Fee,
		"; Amount: ", trade.Amount,
		"; TradeTime", trade.TradeTime,
		"; Status: ", trade.Status,

		"\n")

	err := t.SettleAccountQuick(trade)
	if err != nil {
		panic(err)
	}

	/// get output
	fmt.Println("SettleAccount Complete, GetFUnd to check:\n")
	t.testMysql_GetFund(trade.Who)

	fmt.Println("=================================\n")
}

func (t *MEMySQLDB) testMysql_Settle(bid *Trade, ask *Trade) {
	fmt.Println("=================================\n")
	t.testMysql_GetFund(bid.Who)
	t.testMysql_GetFund(ask.Who)

	fmt.Println("Settle test start with bid trade and ask trade resault:\n")
	fmt.Println("ID: ", bid.ID,
		"; Who: ", bid.Who,
		"; AorB: ", bid.AorB,
		"; Symbol: ", bid.Symbol,
		"; Timestamp: ", bid.Timestamp,
		"; price: ", bid.Price,
		"; volume: ", bid.Volume,
		"; TotalVolume: ", bid.TotalVolume,
		"; Fee: ", bid.Fee,
		"; Amount: ", bid.Amount,
		"; TradeTime", bid.TradeTime,
		"; Status: ", bid.Status,

		"\n")
	fmt.Println("ID: ", ask.ID,
		"; Who: ", ask.Who,
		"; AorB: ", ask.AorB,
		"; Symbol: ", ask.Symbol,
		"; Timestamp: ", ask.Timestamp,
		"; price: ", ask.Price,
		"; volume: ", ask.Volume,
		"; TotalVolume: ", ask.TotalVolume,
		"; Fee: ", ask.Fee,
		"; Amount: ", ask.Amount,
		"; TradeTime", ask.TradeTime,
		"; Status: ", ask.Status,

		"\n")

	err, _ := t.Settle(bid, ask, nil)
	if err != nil {
		panic(err)
	}

	/// get output
	fmt.Println("Settle test Complete, GetFund to check:\n")
	t.testMysql_GetFund(bid.Who)
	t.testMysql_GetFund(ask.Who)

	fmt.Println("=================================\n")
}

/// Finance
func (t *MEMySQLDB) testMysql_GetTradeFinance(user string, id int64, symbol string) {
	finance, err := t.GetTradeFinance(user, id, symbol)
	if err != nil {
		fmt.Printf("testMysql_GetTradeFinance GetTradeFinance fail, no return result.\n")
		return
	}

	fmt.Println("GetTradeFinance Return:\nid: ", finance.ID,
		"; who: ", finance.Who,
		"; AorB: ", finance.AorB,
		"; Symbol: ", finance.Symbol,
		"; Timestamp: ", finance.Timestamp,
		"; price: ", finance.Price,
		"; volume: ", finance.Volume,
		"; TotalVolume: ", finance.TotalVolume,
		"; Fee: ", finance.Fee,
		"; Status: ", finance.Status,
		"; Amount: ", finance.Amount,
		"; TradeTime", finance.TradeTime,
		"; FeeCost", finance.FeeCost,

		"; FType", finance.FType,
		"; FAmount", finance.FAmount,
		"; UserIP", finance.UserIP,

		"\n")
}

func (t *MEMySQLDB) testMysql_AddTradeFinanceCouple(bidF *Finance, askF *Finance) {
	fmt.Println("=================================\n")

	fmt.Println("AddTradeFinanceCouple test start with bid trade and ask trade resault:\n")
	fmt.Println("ID: ", bidF.ID,
		"; Who: ", bidF.Who,
		"; AorB: ", bidF.AorB,
		"; Symbol: ", bidF.Symbol,
		"; Timestamp: ", bidF.Timestamp,
		"; price: ", bidF.Price,
		"; volume: ", bidF.Volume,
		"; TotalVolume: ", bidF.TotalVolume,
		"; Fee: ", bidF.Fee,
		"; Amount: ", bidF.Amount,
		"; TradeTime", bidF.TradeTime,
		"; Status: ", bidF.Status,

		"; FType", bidF.FType,
		"; FAmount", bidF.FAmount,
		"; UserIP", bidF.UserIP,

		"\n")
	fmt.Println("ID: ", askF.ID,
		"; Who: ", askF.Who,
		"; AorB: ", askF.AorB,
		"; Symbol: ", askF.Symbol,
		"; Timestamp: ", askF.Timestamp,
		"; price: ", askF.Price,
		"; volume: ", askF.Volume,
		"; TotalVolume: ", askF.TotalVolume,
		"; Fee: ", askF.Fee,
		"; Amount: ", askF.Amount,
		"; TradeTime", askF.TradeTime,
		"; Status: ", askF.Status,

		"; FType", askF.FType,
		"; FAmount", askF.FAmount,
		"; UserIP", askF.UserIP,

		"\n")

	err := t.AddTradeFinanceCouple(bidF, askF, nil)
	if err != nil {
		panic(err)
	}

	/// get output
	fmt.Println("AddTradeFinanceCouple test Complete, GetTradeFinance to check:\n")
	t.testMysql_GetTradeFinance(bidF.Who, bidF.ID, bidF.Symbol)
	t.testMysql_GetTradeFinance(askF.Who, askF.ID, askF.Symbol)

	fmt.Println("=================================\n")
}
