// test
package use_mysql

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	. "../../itf"
)

//Test:---------------------------------------------------------------------------------------------------
var countOV int64 = 0

func TestMysql(u string, o string, p ...interface{}) {
	switch u {
	case "order":
		price := 1 + float64(rand.Intn(3))/10
		volume := (10 + 10*float64(rand.Intn(10))/10)
		bidOrder := Order{54321, "Hunter", TradeType_BID, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_SUBMIT, "localhost:IP"}
		askOrder := Order{12345, "Hunter", TradeType_ASK, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_SUBMIT, "localhost:IP"}
		switch o {
		case "add":
			if countOV%2 == 0 {
				bidOrder.Volume = 0
				MEMySQLInstance().testMysql_AddOrder(&bidOrder)
			} else {
				askOrder.Volume = 0
				MEMySQLInstance().testMysql_AddOrder(&askOrder)
			}
			countOV++
		case "update":
			if countOV%2 == 0 {
				bidOrder.Volume = bidOrder.TotalVolume / 10
				MEMySQLInstance().testMysql_UpdateOrder(&bidOrder)
			} else {
				askOrder.Volume = askOrder.TotalVolume / 10
				MEMySQLInstance().testMysql_UpdateOrder(&askOrder)
			}
			countOV++
		case "rm":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			MEMySQLInstance().testMysql_RmOrder(user, id, symbol)
		case "rm2":
			MEMySQLInstance().testMysql_RmOrderCouple(&bidOrder, &askOrder)

		case "get":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			MEMySQLInstance().testMysql_GetOrder(user, id, symbol)
		case "all":
			symbol, _ := p[0].(string)
			MEMySQLInstance().testMysql_GetAllOrder(symbol)
		case "ones":
			user, _ := p[0].(string)
			symbol, _ := p[1].(string)
			MEMySQLInstance().testMysql_GetOnesOrder(user, symbol)
		case "overtime":
			symbol, _ := p[0].(string)
			user, _ := p[1].(string)
			userID, _ := strconv.ParseInt(user, 10, 64)
			MEMySQLInstance().testMysql_GetOnesOverTimeOrder(symbol, []interface{}{userID, int64(5)}, 100)
		}
	case "trade":
		price := 1 + float64(rand.Intn(3))/10
		volume := (10 + 10*float64(rand.Intn(10))/10)
		askTrade := Trade{Order{time.Now().UnixNano(), "Hunter", TradeType_ASK, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_FILLED, "localhost:IP"},
			price * volume, time.Now().UnixNano(), price * volume * 0.001}
		bidTrade := Trade{Order{time.Now().UnixNano(), "Hunter", TradeType_BID, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_FILLED, "localhost:IP"},
			price * volume, time.Now().UnixNano(), volume * 0.001}
		switch o {
		case "add":
			if countOV%2 == 0 {
				MEMySQLInstance().testMysql_AddTrade(&bidTrade)
			} else {
				MEMySQLInstance().testMysql_AddTrade(&askTrade)
			}
			countOV++
		case "add2":
			askTrade.ID = 12345
			bidTrade.ID = 54321
			MEMySQLInstance().testMysql_AddTradeCouple(&bidTrade, &askTrade)
		case "rm":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			MEMySQLInstance().testMysql_RmTrade(user, id, symbol)
		case "get":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			MEMySQLInstance().testMysql_GetTrade(user, id, symbol)
		case "all":
			symbol, _ := p[0].(string)
			MEMySQLInstance().testMysql_GetAllTrade(symbol)
		case "ones":
			user, _ := p[0].(string)
			symbol, _ := p[1].(string)
			MEMySQLInstance().testMysql_GetOnesTrade(user, symbol)
		}

	case "fund":
		price := 1 + float64(rand.Intn(3))/10
		volume := (10 + 10*float64(rand.Intn(10))/10)
		order := Order{time.Now().UnixNano(), "Hunter", TradeType_ASK, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_SUBMIT, "localhost:IP"}
		trade := Trade{order, price * volume, time.Now().UnixNano(), 0}

		switch o {
		case "get":
			user, _ := p[0].(string)
			MEMySQLInstance().testMysql_GetFund(user)
		case "freeze":
			order.Who, _ = p[0].(string)
			aorb, _ := p[1].(string)
			if aorb == "buy" {
				order.AorB = TradeType_BID
			} else {
				order.AorB = TradeType_ASK
			}
			order.EnOrderPrice, _ = strconv.ParseFloat(p[2].(string), 64)
			order.Price = order.EnOrderPrice
			order.Volume, _ = strconv.ParseFloat(p[3].(string), 64)
			MEMySQLInstance().testMysql_FreezeFund(&order)
		case "unfreeze":
			order.Who, _ = p[0].(string)
			aorb, _ := p[1].(string)
			if aorb == "buy" {
				order.AorB = TradeType_BID
			} else {
				order.AorB = TradeType_ASK
			}
			order.EnOrderPrice, _ = strconv.ParseFloat(p[2].(string), 64)
			order.Price = order.EnOrderPrice
			order.Volume, _ = strconv.ParseFloat(p[3].(string), 64)
			MEMySQLInstance().testMysql_UnfreezeFund(&order)
		case "settletx":
			user, _ := p[0].(string)
			trade.Who = user
			aorb, _ := p[1].(string)
			if aorb == "buy" {
				trade.AorB = TradeType_BID
				trade.FeeCost = trade.Volume * trade.Fee
			} else {
				trade.AorB = TradeType_ASK
				trade.FeeCost = trade.Volume * trade.Price * trade.Fee
			}
			amount, _ := strconv.ParseFloat(p[2].(string), 64)
			trade.Amount = amount
			MEMySQLInstance().testMysql_SettleAccount(&trade)
		case "settlequick":
			user, _ := p[0].(string)
			trade.Who = user
			aorb, _ := p[1].(string)
			if aorb == "buy" {
				trade.AorB = TradeType_BID
				trade.FeeCost = trade.Volume * trade.Fee
			} else {
				trade.AorB = TradeType_ASK
				trade.FeeCost = trade.Volume * trade.Price * trade.Fee
			}
			amount, _ := strconv.ParseFloat(p[2].(string), 64)
			trade.Amount = amount
			MEMySQLInstance().testMysql_SettleAccountQuick(&trade)
		case "settle":
			bidUser, _ := p[0].(string)
			askUser, _ := p[1].(string)
			amount, _ := strconv.ParseFloat(p[2].(string), 64)
			order.AorB = TradeType_BID
			order.Who = bidUser
			bid := Trade{order, amount, time.Now().UnixNano(), order.Volume * order.Fee}
			order.AorB = TradeType_ASK
			order.Who = askUser
			ask := Trade{order, amount, time.Now().UnixNano(), order.Volume * order.Price * order.Fee}
			MEMySQLInstance().testMysql_Settle(&bid, &ask)
		}

	case "finance":
		price := 1 + float64(rand.Intn(3))/10
		volume := (10 + 10*float64(rand.Intn(10))/10)
		bidTrade := Trade{Order{time.Now().UnixNano(), "Hunter", TradeType_BID, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_FILLED, "localhost:IP"},
			price * volume, time.Now().UnixNano(), volume * 0.001}
		askTrade := Trade{Order{time.Now().UnixNano(), "Hunter", TradeType_ASK, "ETH/BTC", time.Now().UnixNano(), price, price, volume, volume, 0.001, ORDER_FILLED, "localhost:IP"},
			price * volume, time.Now().UnixNano(), price * volume * 0.001}
		bifF := Finance{bidTrade, FinanceType_TradeFee, bidTrade.FeeCost, "bid localhost testip"}
		askF := Finance{askTrade, FinanceType_TradeFee, askTrade.FeeCost, "ask localhost testip"}

		switch o {
		case "add":
			fmt.Printf("Interface not support for now...\n")
		case "add2":
			bidID, _ := strconv.ParseInt(p[0].(string), 10, 64)
			askID, _ := strconv.ParseInt(p[1].(string), 10, 64)
			bifF.ID = bidID
			askF.ID = askID
			MEMySQLInstance().testMysql_AddTradeFinanceCouple(&bifF, &askF)
		case "rm":
			fmt.Printf("Interface not support for now...\n")
		case "get":
			user, _ := p[0].(string)
			id, _ := strconv.ParseInt(p[1].(string), 10, 64)
			symbol, _ := p[2].(string)
			MEMySQLInstance().testMysql_GetTradeFinance(user, id, symbol)
		case "all":
			fmt.Printf("Interface not support for now...\n")
		case "ones":
			fmt.Printf("Interface not support for now...\n")
		}

	case "tickers":
		switch o {
		case "init":
			symbol, _ := p[0].(string)
			TEMySQLInstance().testMysql_InitializeTickersTable(symbol)
		case "add":
			symbol, _ := p[0].(string)
			tickType, _ := strconv.ParseInt(p[1].(string), 10, 64)
			price := 1 + float64(rand.Intn(3))/10
			volume := (10 + 10*float64(rand.Intn(10))/10)
			loc, _ := time.LoadLocation("Local")
			ticker := TickerInfo{
				From:       time.Now().In(loc).UnixNano(),
				OpenPrice:  price,
				ClosePrice: price + 10*float64(rand.Intn(10))/10,
				LowPrice:   1,
				HightPrice: 2,
				Volume:     volume,
				Amount:     volume * 1.5,
			}
			TEMySQLInstance().testMysql_AddTicker(symbol, tickType, &ticker)
		case "get":
			symbol, _ := p[0].(string)
			tickType, _ := strconv.ParseInt(p[1].(string), 10, 64)
			TEMySQLInstance().testMysql_GetTickers(symbol, tickType)
		case "getlimit":
			symbol, _ := p[0].(string)
			tickType, _ := strconv.ParseInt(p[1].(string), 10, 64)
			size, _ := strconv.ParseInt(p[2].(string), 10, 64)
			TEMySQLInstance().testMysql_GetTickersLimit(symbol, tickType, int(size))
		}

	default:
	}
}
