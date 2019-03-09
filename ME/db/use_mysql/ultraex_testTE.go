// testTE
package use_mysql

import (
	"fmt"
	"time"

	. "../../itf"
)

/// tickers
func (t *TEMySQLDB) testMysql_InitializeTickersTable(symbol string) {
	fmt.Printf("===============testMysql InitializeTickersTable==================\n")

	err := t.InitTickersTable(symbol)
	if err != nil {
		panic(err)
	}

	fmt.Printf("=================================:testMysql_InitializeTickersTable Complete.\n")
	fmt.Printf("Please check if the tickers table is created.\n")
}

func (t *TEMySQLDB) testMysql_AddTicker(symbol string, tickType int64, ticker *TickerInfo) {
	fmt.Printf("===============testMysql AddTicker==================\n")

	err := t.AddTicker(symbol, TickerType(tickType), ticker)
	if err != nil {
		panic(err)
	}

	fmt.Printf("=================================:testMysql_AddTicker Complete.\n")
	fmt.Printf("Please check the result in the ticker db.\n")
}

func (t *TEMySQLDB) testMysql_GetTickers(symbol string, tickType int64) {
	fmt.Printf("===============testMysql GetTickers==================\n")

	tickers, err := t.GetTickers(symbol, TickerType(tickType))
	if err != nil {
		panic(err)
	}

	for count, elem := range tickers {
		from := time.Unix(0, elem.From).Format("2006-01-02T15:04:05Z07:00")
		fmt.Printf("Tickers[%d]: From: %s,  OpenPrice: %f, ClosePrice: %f, LowPrice: %f, HightPrice: %f, Volume: %f, Amount: %f\n",
			count,
			from,
			elem.OpenPrice,
			elem.ClosePrice,
			elem.LowPrice,
			elem.HightPrice,
			elem.Volume,
			elem.Amount,
		)
	}

	fmt.Printf("=================================:testMysql_GetTickers Complete.\n")
}

func (t *TEMySQLDB) testMysql_GetTickersLimit(symbol string, tickType int64, _size int) {
	fmt.Printf("===============testMysql GetTickersLimit==================\n")

	tickers, err := t.GetTickersLimit(symbol, TickerType(tickType), _size)
	if err != nil {
		panic(err)
	}

	for count, elem := range tickers {
		from := time.Unix(0, elem.From).Format("2006-01-02T15:04:05Z07:00")
		fmt.Printf("Tickers[%d]: From: %s,  OpenPrice: %f, ClosePrice: %f, LowPrice: %f, HightPrice: %f, Volume: %f, Amount: %f\n",
			count,
			from,
			elem.OpenPrice,
			elem.ClosePrice,
			elem.LowPrice,
			elem.HightPrice,
			elem.Volume,
			elem.Amount,
		)
	}

	fmt.Printf("=================================:testMysql_GetTickersLimit Complete.\n")
}
