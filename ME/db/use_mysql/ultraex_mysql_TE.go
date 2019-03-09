// ultraex_tickers_mysql
// go-sql-driver
package use_mysql

import (
	//	"context"
	"database/sql"
	"fmt"
	"log"

	"strings"
	"sync"
	"time"

	. "../../itf"

	_ "github.com/go-sql-driver/mysql"
)

type TEMySQLDB struct {
	db *sql.DB
}

func (t *TEMySQLDB) Init() {
	t.db = newTEMySQL()
}

func newTEMySQL() *sql.DB {
	db, err := sql.Open("mysql",
		DB_USER_TE+":"+DB_PWD_TE+"@tcp("+DB_IP_TE+":"+DB_PORT_TE+")/"+TE_DB_NAME)
	if err != nil {
		log.Fatal(err)
	}

	// Open doesn't open a connection. Validate DSN data:
	db.SetMaxIdleConns(MAX_IDLE_CONNS)
	db.SetMaxOpenConns(MAX_OPEN_CONNS)
	db.SetConnMaxLifetime(MAX_CONN_LIFE_TIME)
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

var mySQLDBTEObj *TEMySQLDB
var onceTE sync.Once

func TEMySQLInstance() *TEMySQLDB {

	onceTE.Do(func() {
		mySQLDBTEObj = new(TEMySQLDB)
		mySQLDBTEObj.Init()
	})

	return mySQLDBTEObj
}

//Tickers: Get and Add and Update itf---------------------------------------------------------------------------------------------------
func getTickersTableName(sym string) string {
	return TABLE_TICKERS + "_" + strings.Replace(sym, "/", "_", -1)
}

func (t *TEMySQLDB) InitTickersTable(sym string) error {
	var (
		err error
		res sql.Result
	)

	cmd := `CREATE TABLE IF NOT EXISTS %s (
	  id bigint(32) NOT NULL AUTO_INCREMENT,
	  period_type int(10) NOT NULL COMMENT 'period type: 1(1),5(2),15(3),30(4),60(5)min and 1day(6) 1week(7) or 1month(8)',
	  timefrom bigint(20) NOT NULL DEFAULT '0' COMMENT 'ticker time from(use unix second unit with 10 bit)',
	  timeend bigint(20) NOT NULL DEFAULT '0' COMMENT 'next ticker from time(use unix second unit with 10 bit)',
	  openprice decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: open price in this period',
	  closeprice decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: close price in this period',
	  lowprice decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: lowest price in this period',
	  highprice decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: highest price in this period',
	  volume decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: trade volume in this period',
	  amount decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT 'ticker: trade amount in this period',
	  PRIMARY KEY (id),
	  UNIQUE KEY pt (period_type,timefrom)
	)ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=0`
	sql := fmt.Sprintf(cmd, getTickersTableName(sym))
	res, err = t.db.Exec(sql)
	checkErr(err)

	MySQLOpeResaultLog(res, "InitTickersTable")

	return nil
}

func (t *TEMySQLDB) GetTickers(sym string, _type TickerType) ([]*TickerInfo, error) {

	var (
		id       int64
		tickType int64

		From       int64
		End        int64
		OpenPrice  float64
		ClosePrice float64
		LowPrice   float64
		HightPrice float64
		Volume     float64
		Amount     float64

		tickers []*TickerInfo
	)

	sql := fmt.Sprintf("SELECT * FROM %s WHERE period_type=?", getTickersTableName(sym))
	rows, err := t.db.Query(sql, _type)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&id, &tickType, &From, &End, &OpenPrice, &ClosePrice, &LowPrice, &HightPrice, &Volume, &Amount)
		checkErr(err)

		o := &TickerInfo{
			From:       From * int64(time.Second),
			End:        End * int64(time.Second),
			OpenPrice:  OpenPrice,
			ClosePrice: ClosePrice,
			LowPrice:   LowPrice,
			HightPrice: HightPrice,
			Volume:     Volume,
			Amount:     Amount,
		}
		tickers = append(tickers, o)
	}

	return tickers, nil
}

func (t *TEMySQLDB) GetTickersLimit(sym string, _type TickerType, _size int) ([]*TickerInfo, error) {

	var (
		id       int64
		tickType int64

		From       int64
		End        int64
		OpenPrice  float64
		ClosePrice float64
		LowPrice   float64
		HightPrice float64
		Volume     float64
		Amount     float64

		tickers []*TickerInfo
	)

	sql := fmt.Sprintf("SELECT * FROM %s WHERE period_type=? ORDER BY timefrom DESC LIMIT ?", getTickersTableName(sym))
	rows, err := t.db.Query(sql, _type, _size)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&id, &tickType, &From, &End, &OpenPrice, &ClosePrice, &LowPrice, &HightPrice, &Volume, &Amount)
		checkErr(err)

		o := &TickerInfo{
			From:       From * int64(time.Second),
			End:        End * int64(time.Second),
			OpenPrice:  OpenPrice,
			ClosePrice: ClosePrice,
			LowPrice:   LowPrice,
			HightPrice: HightPrice,
			Volume:     Volume,
			Amount:     Amount,
		}
		tickers = append(tickers, o)
	}

	return tickers, nil
}

func (t *TEMySQLDB) AddTicker(sym string, _type TickerType, ticker *TickerInfo) error {
	var (
		err error
		res sql.Result
	)

	cmd := `INSERT INTO %s
			(period_type, timefrom, timeend, openprice, closeprice, lowprice, highprice, volume, amount) VALUES
			(?,?,?,?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE
			timeend=VALUES(timeend),
			openprice=VALUES(openprice),
			closeprice=VALUES(closeprice),
			lowprice=VALUES(lowprice),
			highprice=VALUES(highprice),
			volume=VALUES(volume),
			amount=VALUES(amount)`
	sql := fmt.Sprintf(cmd, getTickersTableName(sym))
	res, err = t.db.Exec(sql,
		_type,
		ticker.From/int64(time.Second),
		ticker.End/int64(time.Second),
		ticker.OpenPrice,
		ticker.ClosePrice,
		ticker.LowPrice,
		ticker.HightPrice,
		ticker.Volume,
		ticker.Amount,
	)
	checkErr(err)

	MySQLOpeResaultLog(res, "AddTicker")
	return nil
}
