// go-sql-driver
package use_mysql

import (
	//	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	. "../../comm"
	"../../config"

	_ "github.com/go-sql-driver/mysql"
)

type MEMySQLDB struct {
	db *sql.DB
}

func (t *MEMySQLDB) Init() {
	t.db = newMySQL()
}

///==============DB pool to solve bad connection problem
//var MAX_POOL_SIZE=200
//var MySQLPool chan *sql.DB
//func getMySQL() *sql.DB {
//    if MySQLPool == nil {
//        MySQLPool = make(chan *sql.DB, MAX_POOL_SIZE)
//    }
//    if len(MySQLPool) == 0 {
//        go func() {
//            for i := 0; i < MAX_POOL_SIZE/2; i++ {
//		mysqlc, err := sql.Open("mymysql", "tcp:127.0.0.1:3306*xxxdb/root/123456789")
//                if err != nil {
//                    panic(err)
//                }
//                putMySQL(mysqlc)
//            }
//        }()
//    }
//    return <-MySQLPool
//}
//func putMySQL(conn *sql.DB) {
//    if MySQLPool == nil {
//        MySQLPool = make(chan *sql.DB, MAX_POOL_SIZE)
//    }

//    if len(MySQLPool) == MAX_POOL_SIZE {
//        conn.Close()
//        return
//    }
//    MySQLPool <- conn
//}
/////To use:
//dbx:=getMySQL()
//rows, err := dbx.Query(“select count(*)as tcount from xxx”)
//defer dbx.Close()
///==============DB pool to solve bad connection problem

func newMySQL() *sql.DB {
	db, err := sql.Open("mysql",
		//"root:root@tcp(127.0.0.1:3306)/ubtdb")
		DB_USER_ME+":"+DB_PWD_ME+"@tcp("+DB_IP_ME+":"+DB_PORT_ME+")/"+ME_DB_NAME)
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

var mySQLDBObj *MEMySQLDB
var onceME sync.Once

func MEMySQLInstance() *MEMySQLDB {

	onceME.Do(func() {
		mySQLDBObj = new(MEMySQLDB)
		mySQLDBObj.Init()
	})

	return mySQLDBObj
}

//IdbLongTerm::Order---------------------------------------------------------------------------------------------------
func (t *MEMySQLDB) AddOrder(order *Order, tx *sql.Tx) (error, ErrorCode) {
	var (
		err error
		res sql.Result
	)

	/// order record at order table
	orderID, _ := strconv.ParseInt(order.Who, 10, 64)
	isRobot := 0
	robotSet := config.GetRobotSet(order.Symbol)
	if robotSet.Contains(orderID) {
		isRobot = 1
	} else {
		isRobot = 0
	}
	cB, cR := symbolConvertTo(order.Symbol)
	cmd := `INSERT INTO %s 
			(orders_id, member_id, currency_id, currency_trade_id, price, num, trade_num, fee, type, add_time, trade_time, status, is_robot) VALUES 
			(?,?,?,?,?,?,?,?,?,?,?,?,?)`
	sql := fmt.Sprintf(cmd, TABLE_ORDER)
	if tx == nil {
		res, err = t.db.Exec(sql,
			order.ID,
			whoConvertTo(order.Who),

			cB,
			cR,

			order.EnOrderPrice,
			order.TotalVolume,
			order.Volume,
			order.Fee,

			tradeTypeConvertTo(order.AorB),

			order.Timestamp/int64(time.Second),
			time.Now().Unix(),
			tradeStatusConvertTo(order.Status),
			isRobot,
		)
	} else {
		res, err = tx.Exec(sql,
			order.ID,
			whoConvertTo(order.Who),

			cB,
			cR,

			order.EnOrderPrice,
			order.TotalVolume,
			order.Volume,
			order.Fee,

			tradeTypeConvertTo(order.AorB),

			order.Timestamp/int64(time.Second),
			time.Now().Unix(),
			tradeStatusConvertTo(order.Status),
			isRobot,
		)
	}
	if err != nil {
		///Error 1062: Duplicate entry '1538117113568831117' for key 'PRIMARY'
		isDup, _ := regexp.Match("Error 1062: Duplicate entry .* for key 'PRIMARY'", []byte(err.Error()))
		if isDup {
			fmt.Printf("AddOrder Met Duplicate entry for PRIMARY, will reproduce a new order ID.\n")
			return err, ErrorCode_DupPrimateKey
		} else {
			panic(err)
		}
	}

	MySQLOpeResaultLog(res, "AddOrder")
	return nil, ErrorCode_OK
}

func (t *MEMySQLDB) UpdateOrder(order *Order, tx *sql.Tx) error {
	var (
		err error
		res sql.Result
	)

	/// order record at order table
	baseCoinID, quoteCoinID := symbolConvertTo(order.Symbol)
	cmd := `UPDATE %s SET trade_num = trade_num + ?, trade_time = ?, status = ? WHERE orders_id=? AND currency_id=? AND currency_trade_id=?`
	sql := fmt.Sprintf(cmd, TABLE_ORDER)
	if tx == nil {
		res, err = t.db.Exec(sql,
			order.Volume,
			time.Now().Unix(),
			tradeStatusConvertTo(order.Status),
			order.ID,
			baseCoinID,
			quoteCoinID,
		)
	} else {
		res, err = tx.Exec(sql,
			order.Volume,
			time.Now().Unix(),
			tradeStatusConvertTo(order.Status),
			order.ID,
			baseCoinID,
			quoteCoinID,
		)
	}
	//	cmd := `INSERT INTO %s
	//			(orders_id, member_id, currency_id, currency_trade_id, price, num, trade_num, fee, type, add_time, trade_time, status, is_robot) VALUES
	//			(?,?,?,?,?,?,?,?,?,?,?,?,?)
	//			ON DUPLICATE KEY UPDATE
	//			price=VALUES(price),
	//			num=VALUES(num),
	//			trade_num=trade_num+VALUES(trade_num),
	//			fee=VALUES(fee),
	//			type=VALUES(type),
	//			add_time=VALUES(add_time),
	//			trade_time=VALUES(trade_time),
	//			status=VALUES(status),
	//			is_robot=VALUES(is_robot)`
	//	sql := fmt.Sprintf(cmd, TABLE_ORDER)
	//	if tx == nil {
	//		res, err = t.db.Exec(sql,
	//			order.ID,
	//			whoConvertTo(order.Who),

	//			cB,
	//			cR,

	//			order.Price,
	//			order.TotalVolume,
	//			order.Volume,
	//			order.Fee,

	//			tradeTypeConvertTo(order.AorB),

	//			order.Timestamp/int64(time.Second),
	//			time.Now().Unix(),
	//			tradeStatusConvertTo(order.Status),
	//			0)
	//	} else {
	//		res, err = tx.Exec(sql,
	//			order.ID,
	//			whoConvertTo(order.Who),

	//			cB,
	//			cR,

	//			order.Price,
	//			order.TotalVolume,
	//			order.Volume,
	//			order.Fee,

	//			tradeTypeConvertTo(order.AorB),

	//			order.Timestamp/int64(time.Second),
	//			time.Now().Unix(),
	//			tradeStatusConvertTo(order.Status),
	//			0)
	//	}
	checkErr(err)

	MySQLOpeResaultLog(res, "UpdateOrder")
	return nil
}

func (t *MEMySQLDB) RmOrder(user string, id int64, symbol string, tx *sql.Tx) error {
	var (
		err error
		res sql.Result
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("DELETE FROM %s WHERE orders_id=? AND currency_id=? AND currency_trade_id=?", TABLE_ORDER)
	if tx == nil {
		res, err = t.db.Exec(sql, id, baseCoinID, quoteCoinID)
	} else {
		res, err = tx.Exec(sql, id, baseCoinID, quoteCoinID)
	}
	checkErr(err)

	MySQLOpeResaultLog(res, "RmOrder")
	return nil
}

func (t *MEMySQLDB) RmOrderCouple(bid *Order, ask *Order, tx *sql.Tx) error {
	var (
		err error
		res sql.Result
	)

	sql := fmt.Sprintf("DELETE FROM %s WHERE orders_id=? OR orders_id=?", TABLE_ORDER)
	if tx == nil {
		res, err = t.db.Exec(sql, bid.ID, ask.ID)
	} else {
		res, err = tx.Exec(sql, bid.ID, ask.ID)
	}
	checkErr(err)

	MySQLOpeResaultLog(res, "RmOrderCouple")
	return nil
}

func (t *MEMySQLDB) GetOrder(user string, id int64, symbol string, tx *sql.Tx) (*Order, error) {
	var (
		orders_id         int64
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64 /// total volume
		trade_num         float64 /// had trade volume
		fee               float64
		aorb              string
		add_time          int64
		trade_time        int64
		status            int64
		is_robot          int64

		err error
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	if tx == nil {
		sql := fmt.Sprintf("SELECT * FROM %s WHERE orders_id=? AND currency_id=? AND currency_trade_id=?", TABLE_ORDER)
		err = t.db.QueryRow(sql, id, baseCoinID, quoteCoinID).Scan(&orders_id, &member_id, &currency_id, &currency_trade_id, &price, &num, &trade_num, &fee, &aorb, &add_time, &trade_time, &status, &is_robot)
	} else {
		sql := fmt.Sprintf("SELECT * FROM %s WHERE orders_id=? AND currency_id=? AND currency_trade_id=?", TABLE_ORDER)
		err = tx.QueryRow(sql, id, baseCoinID, quoteCoinID).Scan(&orders_id, &member_id, &currency_id, &currency_trade_id, &price, &num, &trade_num, &fee, &aorb, &add_time, &trade_time, &status, &is_robot)
	}
	if err != nil {
		return nil, err
	}

	if num < trade_num {
		fmt.Printf("=========Illegal Order detail=========\n\tOrder: Type(%s), User(%s), ID(%d), Status(%s), Price(%f), TotalVolume(%f), TradeVolume(%f)\n",
			tradeTypeConvertFrom(aorb), whoConvertFrom(member_id), orders_id, tradeStatusConvertFrom(status), price, num, trade_num)
		panic(fmt.Errorf("GetOrder Met illegal Order with trade_num bigger than total order num."))
	}

	return &Order{
		ID:           orders_id,
		Who:          whoConvertFrom(member_id),
		AorB:         tradeTypeConvertFrom(aorb),
		Symbol:       symbolConvertFrom(currency_id, currency_trade_id),
		Timestamp:    add_time * int64(time.Second),
		EnOrderPrice: price,
		Price:        price,
		Volume:       num - trade_num,
		TotalVolume:  num,
		Fee:          fee,
		Status:       tradeStatusConvertFrom(status),
	}, nil
}

/// use at initHistoryOrder
func (t *MEMySQLDB) GetAllOrder(symbol string) (to []*Order, err error) {
	var (
		orders_id         int64
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		trade_num         float64
		fee               float64
		aorb              string
		add_time          int64
		trade_time        int64
		status            int64
		is_robot          int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("SELECT * FROM %s  WHERE currency_id=? AND currency_trade_id=?", TABLE_ORDER)
	rows, err := t.db.Query(sql, baseCoinID, quoteCoinID)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&orders_id, &member_id, &currency_id, &currency_trade_id, &price, &num, &trade_num, &fee, &aorb, &add_time, &trade_time, &status, &is_robot)
		checkErr(err)

		if num < trade_num {
			fmt.Printf("=========Illegal Order detail=========\n\tOrder: Type(%s), User(%s), ID(%d), Status(%s), Price(%f), TotalVolume(%f), TradeVolume(%f)\n",
				tradeTypeConvertFrom(aorb), whoConvertFrom(member_id), orders_id, tradeStatusConvertFrom(status), price, num, trade_num)
			panic(fmt.Errorf("GetAllOrder Met illegal Order with trade_num bigger than total order num."))
		}

		o := &Order{
			ID:           orders_id,
			Who:          whoConvertFrom(member_id),
			AorB:         tradeTypeConvertFrom(aorb),
			Symbol:       symbolConvertFrom(currency_id, currency_trade_id),
			Timestamp:    add_time * int64(time.Second),
			EnOrderPrice: price,
			Price:        price,
			Volume:       num - trade_num,
			TotalVolume:  num,
			Fee:          fee,
			Status:       tradeStatusConvertFrom(status),
		}
		to = append(to, o)
	}

	return to, nil
}

/// use at initHistoryOrder
func (t *MEMySQLDB) GetAllHumanOrder(symbol string, robots []interface{}) (to []*Order, err error) {
	var (
		orders_id         int64
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		trade_num         float64
		fee               float64
		aorb              string
		add_time          int64
		trade_time        int64
		status            int64
		is_robot          int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	human := fmt.Sprintf("member_id != %d", robots[0])
	for _, robot := range robots[1:] {
		human += fmt.Sprintf(" AND member_id != %d", robot.(int64))
	}
	/// debug for del:
	fmt.Printf(human + "\n")

	sql := fmt.Sprintf("SELECT * FROM %s  WHERE currency_id=? AND currency_trade_id=? AND %s", TABLE_ORDER, human)
	rows, err := t.db.Query(sql, baseCoinID, quoteCoinID)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&orders_id, &member_id, &currency_id, &currency_trade_id, &price, &num, &trade_num, &fee, &aorb, &add_time, &trade_time, &status, &is_robot)
		checkErr(err)

		if num < trade_num {
			fmt.Printf("=========Illegal Order detail=========\n\tOrder: Type(%s), User(%s), ID(%d), Status(%s), Price(%f), TotalVolume(%f), TradeVolume(%f)\n",
				tradeTypeConvertFrom(aorb), whoConvertFrom(member_id), orders_id, tradeStatusConvertFrom(status), price, num, trade_num)
			panic(fmt.Errorf("GetAllHumanOrder Met illegal Order with trade_num bigger than total order num."))
		}

		o := &Order{
			ID:           orders_id,
			Who:          whoConvertFrom(member_id),
			AorB:         tradeTypeConvertFrom(aorb),
			Symbol:       symbolConvertFrom(currency_id, currency_trade_id),
			Timestamp:    add_time * int64(time.Second),
			EnOrderPrice: price,
			Price:        price,
			Volume:       num - trade_num,
			TotalVolume:  num,
			Fee:          fee,
			Status:       tradeStatusConvertFrom(status),
		}
		to = append(to, o)
	}

	return to, nil
}

/// use at initHistoryOrder
func (t *MEMySQLDB) GetAllRobotOrder(symbol string, robots []interface{}) (to []*Order, err error) {
	var (
		orders_id         int64
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		trade_num         float64
		fee               float64
		aorb              string
		add_time          int64
		trade_time        int64
		status            int64
		is_robot          int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	human := "( "
	human += fmt.Sprintf("member_id = %d", robots[0])
	for _, robot := range robots[1:] {
		human += fmt.Sprintf(" OR member_id = %d", robot.(int64))
	}
	human += " )"
	/// debug for del:
	fmt.Printf(human + "\n")

	sql := fmt.Sprintf("SELECT * FROM %s  WHERE currency_id=? AND currency_trade_id=? AND %s", TABLE_ORDER, human)
	rows, err := t.db.Query(sql, baseCoinID, quoteCoinID)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&orders_id, &member_id, &currency_id, &currency_trade_id, &price, &num, &trade_num, &fee, &aorb, &add_time, &trade_time, &status, &is_robot)
		checkErr(err)

		if num < trade_num {
			fmt.Printf("=========Illegal Order detail=========\n\tOrder: Type(%s), User(%s), ID(%d), Status(%s), Price(%f), TotalVolume(%f), TradeVolume(%f)\n",
				tradeTypeConvertFrom(aorb), whoConvertFrom(member_id), orders_id, tradeStatusConvertFrom(status), price, num, trade_num)
			panic(fmt.Errorf("GetAllRobotOrder Met illegal Order with trade_num bigger than total order num."))
		}

		o := &Order{
			ID:           orders_id,
			Who:          whoConvertFrom(member_id),
			AorB:         tradeTypeConvertFrom(aorb),
			Symbol:       symbolConvertFrom(currency_id, currency_trade_id),
			Timestamp:    add_time * int64(time.Second),
			EnOrderPrice: price,
			Price:        price,
			Volume:       num - trade_num,
			TotalVolume:  num,
			Fee:          fee,
			Status:       tradeStatusConvertFrom(status),
		}
		to = append(to, o)
	}

	return to, nil
}

/// use at CancelRobotOverTimeOrder
func (t *MEMySQLDB) GetOnesOverTimeOrder(symbol string, users []interface{}, ot int64) (to []*Order, err error) {
	var (
		orders_id         int64
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		trade_num         float64
		fee               float64
		aorb              string
		add_time          int64
		trade_time        int64
		status            int64
		is_robot          int64
	)

	userConditions := "( "
	userConditions += fmt.Sprintf("member_id = %d", users[0])
	for _, user := range users[1:] {
		userConditions += fmt.Sprintf(" OR member_id = %d", user.(int64))
	}
	userConditions += " )"

	overTimeCondition := "( "
	conditionTime := time.Now().Unix() - ot
	overTimeCondition += fmt.Sprintf("add_time < %d", conditionTime)
	overTimeCondition += " )"

	/// debug for del:
	///fmt.Printf(userConditions + " AND " + overTimeCondition + "\n")

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("SELECT * FROM %s  WHERE currency_id=? AND currency_trade_id=? AND %s AND %s", TABLE_ORDER, userConditions, overTimeCondition)
	rows, err := t.db.Query(sql, baseCoinID, quoteCoinID)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&orders_id, &member_id, &currency_id, &currency_trade_id, &price, &num, &trade_num, &fee, &aorb, &add_time, &trade_time, &status, &is_robot)
		checkErr(err)

		if num < trade_num {
			fmt.Printf("=========Illegal Order detail=========\n\tOrder: Type(%s), User(%s), ID(%d), Status(%s), Price(%f), TotalVolume(%f), TradeVolume(%f)\n",
				tradeTypeConvertFrom(aorb), whoConvertFrom(member_id), orders_id, tradeStatusConvertFrom(status), price, num, trade_num)
			panic(fmt.Errorf("GetOnesOverTimeOrder Met illegal Order with trade_num bigger than total order num."))
		}

		o := &Order{
			ID:           orders_id,
			Who:          whoConvertFrom(member_id),
			AorB:         tradeTypeConvertFrom(aorb),
			Symbol:       symbolConvertFrom(currency_id, currency_trade_id),
			Timestamp:    add_time * int64(time.Second),
			EnOrderPrice: price,
			Price:        price,
			Volume:       num - trade_num,
			TotalVolume:  num,
			Fee:          fee,
			Status:       tradeStatusConvertFrom(status),
		}
		to = append(to, o)
	}

	return to, nil
}

func (t *MEMySQLDB) GetOnesOrder(user string, symbol string) (to []*Order, err error) {
	var (
		orders_id         int64
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		trade_num         float64
		fee               float64
		aorb              string
		add_time          int64
		trade_time        int64
		status            int64
		is_robot          int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE member_id=? AND currency_id=? AND currency_trade_id=?", TABLE_ORDER)
	rows, err := t.db.Query(sql, user, baseCoinID, quoteCoinID)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&orders_id, &member_id, &currency_id, &currency_trade_id, &price, &num, &trade_num, &fee, &aorb, &add_time, &trade_time, &status, &is_robot)
		checkErr(err)

		if num < trade_num {
			fmt.Printf("=========Illegal Order detail=========\n\tOrder: Type(%s), User(%s), ID(%d), Status(%s), Price(%f), TotalVolume(%f), TradeVolume(%f)\n",
				tradeTypeConvertFrom(aorb), whoConvertFrom(member_id), orders_id, tradeStatusConvertFrom(status), price, num, trade_num)
			panic(fmt.Errorf("GetOnesOrder Met illegal Order with trade_num bigger than total order num."))
		}

		o := &Order{
			ID:           orders_id,
			Who:          whoConvertFrom(member_id),
			AorB:         tradeTypeConvertFrom(aorb),
			Symbol:       symbolConvertFrom(currency_id, currency_trade_id),
			Timestamp:    add_time * int64(time.Second),
			EnOrderPrice: price,
			Price:        price,
			Volume:       num - trade_num,
			TotalVolume:  num,
			Fee:          fee,
			Status:       tradeStatusConvertFrom(status),
		}
		to = append(to, o)
	}

	return to, nil
}

//IdbLongTerm::Trade---------------------------------------------------------------------------------------------------
/// trade record at trade table
func (t *MEMySQLDB) AddTrade(trade *Trade, tx *sql.Tx) error {
	var (
		cB, cR int64 = 0, 0
		err    error
		res    sql.Result
	)

	cB, cR = symbolConvertTo(trade.Symbol)

	cmd := `INSERT INTO %s
			(trade_id, trade_no, member_id, currency_id, currency_trade_id, price, num, money, fee, type, add_time, status) VALUES
			(?,?,?,?,?,?,?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE
			price=VALUES(price),
			num=num+VALUES(num),
			money=money+VALUES(money),
			fee=fee+VALUES(fee),
			type=VALUES(type),
			add_time=VALUES(add_time),
			status=VALUES(status)`
	sql := fmt.Sprintf(cmd, TABLE_TRADE)
	if tx == nil {
		res, err = t.db.Exec(sql,
			trade.ID,
			strconv.FormatInt(trade.ID, 10),
			whoConvertTo(trade.Who),

			cB,
			cR,

			trade.Price,
			trade.Volume,
			trade.Price*trade.Volume,

			trade.FeeCost,

			tradeTypeConvertTo(trade.AorB),

			trade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(trade.Status))
	} else {
		res, err = tx.Exec(sql,
			trade.ID,
			strconv.FormatInt(trade.ID, 10),
			whoConvertTo(trade.Who),

			cB,
			cR,

			trade.Price,
			trade.Volume,
			trade.Price*trade.Volume,

			trade.FeeCost,

			tradeTypeConvertTo(trade.AorB),

			trade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(trade.Status))
	}
	checkErr(err)

	MySQLOpeResaultLog(res, "AddTrade")
	return nil
}

/// seperate trade record item
func (t *MEMySQLDB) InsertTrade(trade *Trade, tx *sql.Tx) error {
	var (
		cB, cR int64 = 0, 0
		err    error
		res    sql.Result
	)

	cB, cR = symbolConvertTo(trade.Symbol)

	cmd := `INSERT INTO %s
			(trade_no, member_id, currency_id, currency_trade_id, price, num, money, fee, type, add_time, status) VALUES
			(?,?,?,?,?,?,?,?,?,?,?)`
	sql := fmt.Sprintf(cmd, TABLE_TRADE)
	if tx == nil {
		res, err = t.db.Exec(sql,
			strconv.FormatInt(trade.ID, 10),
			whoConvertTo(trade.Who),

			cB,
			cR,

			trade.Price,
			trade.Volume,
			trade.Price*trade.Volume,

			trade.FeeCost,

			tradeTypeConvertTo(trade.AorB),

			trade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(trade.Status))
	} else {
		res, err = tx.Exec(sql,
			strconv.FormatInt(trade.ID, 10),
			whoConvertTo(trade.Who),

			cB,
			cR,

			trade.Price,
			trade.Volume,
			trade.Price*trade.Volume,

			trade.FeeCost,

			tradeTypeConvertTo(trade.AorB),

			trade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(trade.Status))
	}
	checkErr(err)

	MySQLOpeResaultLog(res, "InsertTrade")
	return nil
}

/// trade couple(composed of bid and ask trades) record at trade table
func (t *MEMySQLDB) AddTradeCouple(bidTrade *Trade, askTrade *Trade, tx *sql.Tx) error {
	var (
		err error
		res sql.Result
	)

	if bidTrade.Symbol != askTrade.Symbol {
		return fmt.Errorf("AddTradeCouple trade couple not corresponding, cause: bidTrade.Symbol != askTrade.Symbol")
	}
	cB, cQ, err := getCoinFromSymbol(bidTrade.Symbol)
	if err != nil {
		return fmt.Errorf("AddTradeCouple fail! Cause: bid with illegal bid symbol(%s)", bidTrade.Symbol)
	}

	cmd := `INSERT INTO %s
			(trade_id, trade_no, member_id, currency_id, currency_trade_id, price, num, money, fee, type, add_time, status) VALUES
			(?,?,?,?,?,?,?,?,?,?,?,?),
			(?,?,?,?,?,?,?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE
			price=VALUES(price),
			num=num+VALUES(num),
			money=money+VALUES(money),
			fee=fee+VALUES(fee),
			type=VALUES(type),
			add_time=VALUES(add_time),
			status=VALUES(status)`
	sql := fmt.Sprintf(cmd, TABLE_TRADE)
	if tx == nil {
		res, err = t.db.Exec(sql,
			bidTrade.ID,
			strconv.FormatInt(bidTrade.ID, 10),
			whoConvertTo(bidTrade.Who),
			int64(cB),
			int64(cQ),
			bidTrade.Price,
			bidTrade.Volume,
			bidTrade.Price*bidTrade.Volume,
			bidTrade.FeeCost,
			tradeTypeConvertTo(bidTrade.AorB),
			bidTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(bidTrade.Status),

			askTrade.ID,
			strconv.FormatInt(askTrade.ID, 10),
			whoConvertTo(askTrade.Who),
			int64(cB),
			int64(cQ),
			askTrade.Price,
			askTrade.Volume,
			askTrade.Price*askTrade.Volume,
			askTrade.FeeCost,
			tradeTypeConvertTo(askTrade.AorB),
			askTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(askTrade.Status),
		)
	} else {
		res, err = tx.Exec(sql,
			bidTrade.ID,
			strconv.FormatInt(bidTrade.ID, 10),
			whoConvertTo(bidTrade.Who),
			int64(cB),
			int64(cQ),
			bidTrade.Price,
			bidTrade.Volume,
			bidTrade.Price*bidTrade.Volume,
			bidTrade.FeeCost,
			tradeTypeConvertTo(bidTrade.AorB),
			bidTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(bidTrade.Status),

			askTrade.ID,
			strconv.FormatInt(askTrade.ID, 10),
			whoConvertTo(askTrade.Who),
			int64(cB),
			int64(cQ),
			askTrade.Price,
			askTrade.Volume,
			askTrade.Price*askTrade.Volume,
			askTrade.FeeCost,
			tradeTypeConvertTo(askTrade.AorB),
			askTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(askTrade.Status),
		)
	}
	checkErr(err)

	MySQLOpeResaultLog(res, "AddTradeCouple")
	return nil
}

/// trade couple(composed of bid and ask trades) record at trade table seperately.
func (t *MEMySQLDB) InsertTradeCouple(bidTrade *Trade, askTrade *Trade, tx *sql.Tx) error {
	var (
		err error
		res sql.Result
	)

	if bidTrade.Symbol != askTrade.Symbol {
		return fmt.Errorf("AddTradeCouple trade couple not corresponding, cause: bidTrade.Symbol != askTrade.Symbol")
	}
	cB, cQ, err := getCoinFromSymbol(bidTrade.Symbol)
	if err != nil {
		return fmt.Errorf("AddTradeCouple fail! Cause: bid with illegal bid symbol(%s)", bidTrade.Symbol)
	}

	cmd := `INSERT INTO %s
			(trade_no, member_id, currency_id, currency_trade_id, price, num, money, fee, type, add_time, status) VALUES
			(?,?,?,?,?,?,?,?,?,?,?),
			(?,?,?,?,?,?,?,?,?,?,?)`
	sql := fmt.Sprintf(cmd, TABLE_TRADE)
	if tx == nil {
		res, err = t.db.Exec(sql,
			strconv.FormatInt(bidTrade.ID, 10),
			whoConvertTo(bidTrade.Who),
			int64(cB),
			int64(cQ),
			bidTrade.Price,
			bidTrade.Volume,
			bidTrade.Price*bidTrade.Volume,
			bidTrade.FeeCost,
			tradeTypeConvertTo(bidTrade.AorB),
			bidTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(bidTrade.Status),

			strconv.FormatInt(askTrade.ID, 10),
			whoConvertTo(askTrade.Who),
			int64(cB),
			int64(cQ),
			askTrade.Price,
			askTrade.Volume,
			askTrade.Price*askTrade.Volume,
			askTrade.FeeCost,
			tradeTypeConvertTo(askTrade.AorB),
			askTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(askTrade.Status),
		)
	} else {
		res, err = tx.Exec(sql,
			strconv.FormatInt(bidTrade.ID, 10),
			whoConvertTo(bidTrade.Who),
			int64(cB),
			int64(cQ),
			bidTrade.Price,
			bidTrade.Volume,
			bidTrade.Price*bidTrade.Volume,
			bidTrade.FeeCost,
			tradeTypeConvertTo(bidTrade.AorB),
			bidTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(bidTrade.Status),

			strconv.FormatInt(askTrade.ID, 10),
			whoConvertTo(askTrade.Who),
			int64(cB),
			int64(cQ),
			askTrade.Price,
			askTrade.Volume,
			askTrade.Price*askTrade.Volume,
			askTrade.FeeCost,
			tradeTypeConvertTo(askTrade.AorB),
			askTrade.TradeTime/int64(time.Second),
			tradeStatusConvertTo(askTrade.Status),
		)
	}
	checkErr(err)

	MySQLOpeResaultLog(res, "InsertTradeCouple")
	return nil
}

/// Index need: should add TABLE_TRADE trade_id unique index
func (t *MEMySQLDB) RmTrade(user string, id int64, symbol string) error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE trade_id=?", TABLE_TRADE)
	res, err := t.db.Exec(sql, id)
	checkErr(err)

	MySQLOpeResaultLog(res, "RmTrade")
	return nil
}

func (t *MEMySQLDB) GetTrade(user string, id int64, symbol string) (*Trade, error) {
	var (
		trade_id          int64
		trade_no          string
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		money             float64
		fee               float64
		aorb              string
		add_time          int64
		status            int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE trade_id=? AND currency_id=? AND currency_trade_id=?", TABLE_TRADE)
	err := t.db.QueryRow(sql, id, baseCoinID, quoteCoinID).Scan(&trade_id, &trade_no, &member_id, &currency_id, &currency_trade_id, &price, &num, &money, &fee, &aorb, &add_time, &status)
	if err != nil {
		fmt.Printf("GetTrade QueryRow fail.\n")
		fmt.Print(err)
		return nil, err
	}

	return &Trade{Order{trade_id, whoConvertFrom(member_id), tradeTypeConvertFrom(aorb), symbolConvertFrom(currency_id, currency_trade_id), 0, 0, price, num, num, 0, tradeStatusConvertFrom(status), ""},
		0, add_time * int64(time.Second), fee}, nil
}

func (t *MEMySQLDB) GetAllTrade(symbol string) (to []*Trade, err error) {
	var (
		trade_id          int64
		trade_no          string
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		money             float64
		fee               float64
		aorb              string
		add_time          int64
		status            int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE currency_id=? AND currency_trade_id=?", TABLE_TRADE)
	rows, err := t.db.Query(sql, baseCoinID, quoteCoinID)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&trade_id, &trade_no, &member_id, &currency_id, &currency_trade_id, &price, &num, &money, &fee, &aorb, &add_time, &status)
		checkErr(err)

		o := &Trade{Order{trade_id, whoConvertFrom(member_id), tradeTypeConvertFrom(aorb), symbolConvertFrom(currency_id, currency_trade_id), 0, 0, price, num, num, 0, tradeStatusConvertFrom(status), ""},
			0, add_time * int64(time.Second), fee}
		to = append(to, o)
	}
	return to, nil
}

func (t *MEMySQLDB) GetOnesTrade(user string, symbol string) (to []*Trade, err error) {
	var (
		trade_id          int64
		trade_no          string
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		money             float64
		fee               float64
		aorb              string
		add_time          int64
		status            int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE member_id=? AND currency_id=? AND currency_trade_id=?", TABLE_TRADE)
	rows, err := t.db.Query(sql, user, baseCoinID, quoteCoinID)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&trade_id, &trade_no, &member_id, &currency_id, &currency_trade_id, &price, &num, &money, &fee, &aorb, &add_time, &status)
		checkErr(err)

		o := &Trade{Order{trade_id, whoConvertFrom(member_id), tradeTypeConvertFrom(aorb), symbolConvertFrom(currency_id, currency_trade_id), 0, 0, price, num, num, 0, tradeStatusConvertFrom(status), ""},
			0, add_time * int64(time.Second), fee}
		to = append(to, o)
	}

	return to, nil
}

func (t *MEMySQLDB) GetRelTradeForTickers(symbol string, startTime int64) (to []*Trade, err error) {
	var (
		trade_id          int64
		trade_no          string
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		money             float64
		fee               float64
		aorb              string
		add_time          int64
		status            int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	///sql := fmt.Sprintf("SELECT * FROM %s WHERE currency_id=? AND currency_trade_id=? AND add_time>=? AND type=? ORDER BY add_time ASC", TABLE_TRADE)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE currency_id=? AND currency_trade_id=? AND add_time>? AND type=? ORDER BY add_time ASC", TABLE_TRADE)
	rows, err := t.db.Query(sql, baseCoinID, quoteCoinID, startTime/int64(time.Second), "sell")
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&trade_id, &trade_no, &member_id, &currency_id, &currency_trade_id, &price, &num, &money, &fee, &aorb, &add_time, &status)
		checkErr(err)

		o := &Trade{Order{trade_id, whoConvertFrom(member_id), tradeTypeConvertFrom(aorb), symbolConvertFrom(currency_id, currency_trade_id), 0, 0, price, num, num, 0, tradeStatusConvertFrom(status), ""},
			money, add_time * int64(time.Second), fee}
		to = append(to, o)
	}

	return to, nil
}

func (t *MEMySQLDB) GetLatestTradeLimit(symbol string, limit int64) (to []*Trade, err error) {
	var (
		trade_id          int64
		trade_no          string
		member_id         int64
		currency_id       int64
		currency_trade_id int64
		price             float64
		num               float64
		money             float64
		fee               float64
		aorb              string
		add_time          int64
		status            int64
	)

	baseCoinID, quoteCoinID := symbolConvertTo(symbol)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE currency_id=? AND currency_trade_id=? ORDER BY add_time DESC LIMIT ?", TABLE_TRADE)
	rows, err := t.db.Query(sql, baseCoinID, quoteCoinID, limit)
	defer rows.Close()
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&trade_id, &trade_no, &member_id, &currency_id, &currency_trade_id, &price, &num, &money, &fee, &aorb, &add_time, &status)
		checkErr(err)

		o := &Trade{Order{trade_id, whoConvertFrom(member_id), tradeTypeConvertFrom(aorb), symbolConvertFrom(currency_id, currency_trade_id), 0, 0, price, num, num, 0, tradeStatusConvertFrom(status), ""},
			money, add_time * int64(time.Second), fee}
		to = append(to, o)
	}

	return to, nil
}

//IdbLongTerm::Fund---------------------------------------------------------------------------------------------------
func (t *MEMySQLDB) GetFund(user string) (*Fund, error) {
	var (
		cu_id        int64
		member_id    int64
		currency_id  int64
		num          float64
		forzen_num   float64
		status       int64
		chongzhi_url string
	)
	/// Index need: should add member_id index for query quickly
	sql := fmt.Sprintf("SELECT * FROM %s WHERE member_id=?", TABLE_MONEY)
	rows, err := t.db.Query(sql, whoConvertTo(user))
	defer rows.Close()
	checkErr(err)

	coinMapMark := config.GetCoinMapMark()
	var fund Fund = Fund{User: user, AvailableMoney: make(map[string]float64), FreezedMoney: make(map[string]float64), TotalMoney: make(map[string]float64), Status: make(map[string]FundStatus)}
	for rows.Next() {
		err = rows.Scan(&cu_id, &member_id, &currency_id, &num, &forzen_num, &status, &chongzhi_url)
		checkErr(err)
		//coin := CoinType(currency_id).String()
		coin, ok := coinMapMark[currency_id]
		if ok {
			fund.AvailableMoney[coin] = num
			fund.FreezedMoney[coin] = forzen_num
			fund.TotalMoney[coin] = num + forzen_num
			fund.Status[coin] = fundStausConvertFrom(status)
		} else {
			fmt.Printf("Warn: GetFund Met coin(%d) does not config in MEconfig.json\n", currency_id)
		}
	}

	return &fund, nil
}

/// use transaction for fund operate -> frozen fund
func (t *MEMySQLDB) FreezeFund(order *Order, tx *sql.Tx) (error, ErrorCode) {
	var (
		cu_id        int64
		member_id    int64
		currency_id  int64
		num          float64
		forzen_num   float64
		status       int64
		chongzhi_url string

		err      error
		res      sql.Result
		bLocalTx bool = false
	)

	if tx == nil {
		bLocalTx = true
		tx, err = t.db.Begin()
		if err != nil {
			log.Fatalln(err)
		}
		defer func(tx *sql.Tx) {
			err := tx.Rollback()
			if err != sql.ErrTxDone && err != nil {
				log.Fatalln(err)
			}
		}(tx)
	}

	cB, cQ, err := getCoinFromSymbol(order.Symbol)
	if err != nil {
		return fmt.Errorf("FreezeFund fail! Cause: order with illegal trade symbol(%s)", order.Symbol), ErrorCode_IllSymbol
	}

	///  Index need: should add member_id index and currency_id index for update quickly, 创建多列唯一索引
	switch order.AorB {
	case TradeType_ASK:
		statement := fmt.Sprintf("SELECT * FROM %s WHERE member_id=? AND currency_id=? FOR UPDATE", TABLE_MONEY)
		//		newCtx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		//		err = tx.QueryRowContext(newCtx, statement, whoConvertTo(order.Who), int64(cB)).Scan(&cu_id, &member_id, &currency_id, &num, &forzen_num, &status, &chongzhi_url)
		err = tx.QueryRow(statement, whoConvertTo(order.Who), int64(cB)).Scan(&cu_id, &member_id, &currency_id, &num, &forzen_num, &status, &chongzhi_url)
		if err != nil {
			isDeadLock, _ := regexp.Match("Error 1213: Deadlock found when trying to get lock;", []byte(err.Error()))
			if isDeadLock {
				DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "FreezeFund met deadlock.\n")
				return err, ErrorCode_RecordLocked
			} else {
				return err, ErrorCode_Fail
			}
		}

		//if num < order.TotalVolume {
		if float32(num) < float32(order.TotalVolume) {
			coinB, _ := getCoinString(cB)
			return fmt.Errorf("FreezeFund fail! Cause: available coin(%s) volume(%f) for sell less than sell order volume(%f)", coinB, num, order.TotalVolume), ErrorCode_FundNoEnough
		}

		statement = fmt.Sprintf("UPDATE %s SET num=num-?, forzen_num=forzen_num+?  WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, order.TotalVolume, order.TotalVolume, whoConvertTo(order.Who), int64(cB))
		checkErr(err)

		///debug:
		DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "[Fund info] ASK: %d, FreezeFund.....................................CB:num=num-%f, forzen_num=forzen_num+%f\n", order.ID, order.TotalVolume, order.TotalVolume)

	case TradeType_BID:
		statement := fmt.Sprintf("SELECT * FROM %s WHERE member_id=? AND currency_id=? FOR UPDATE", TABLE_MONEY)
		//		newCtx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		//		err = tx.QueryRowContext(newCtx, statement, whoConvertTo(order.Who), int64(cQ)).Scan(&cu_id, &member_id, &currency_id, &num, &forzen_num, &status, &chongzhi_url)
		err = tx.QueryRow(statement, whoConvertTo(order.Who), int64(cQ)).Scan(&cu_id, &member_id, &currency_id, &num, &forzen_num, &status, &chongzhi_url)
		if err != nil {
			isDeadLock, _ := regexp.Match("Error 1213: Deadlock found when trying to get lock;", []byte(err.Error()))
			if isDeadLock {
				DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "FreezeFund met deadlock.\n")
				return err, ErrorCode_RecordLocked
			} else {
				return err, ErrorCode_Fail
			}
		}

		bidCost := order.TotalVolume * order.EnOrderPrice
		//if num < bidCost {
		if float32(num) < float32(bidCost) {
			return fmt.Errorf("FreezeFund fail! Cause: available coin(%s) volume(%f) for buy less than buy order bidCost(%f)", cQ, num, bidCost), ErrorCode_FundNoEnough
		}

		statement = fmt.Sprintf("UPDATE %s SET num=num-?, forzen_num=forzen_num+?  WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, bidCost, bidCost, whoConvertTo(order.Who), int64(cQ))
		checkErr(err)

		///debug:
		DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "[Fund info] BID: %d, FreezeFund.....................................CQ:num=num-%f, forzen_num=forzen_num+%f\n", order.ID, bidCost, bidCost)

	default:
		panic("FreezeFund met illegal order input!")
	}

	if bLocalTx {
		if err := tx.Commit(); err != nil {
			log.Fatalln(err)
		}
	}

	MySQLOpeResaultLog(res, "FreezeFund")
	return nil, ErrorCode_OK
}

/// use transaction for fund operate -> frozen fund
func (t *MEMySQLDB) UnfreezeFund(order *Order, tx *sql.Tx) (error, ErrorCode) {
	var (
		cu_id        int64
		member_id    int64
		currency_id  int64
		num          float64
		forzen_num   float64
		status       int64
		chongzhi_url string

		err      error
		res      sql.Result
		bLocalTx bool = false
	)

	if tx == nil {
		bLocalTx = true
		tx, err = t.db.Begin()
		if err != nil {
			log.Fatalln(err)
		}
		defer func(tx *sql.Tx) {
			err := tx.Rollback()
			if err != sql.ErrTxDone && err != nil {
				log.Fatalln(err)
			}
		}(tx)
	}

	cB, cQ, err := getCoinFromSymbol(order.Symbol)
	if err != nil {
		return fmt.Errorf("UnfreezeFund fail! Cause: order with illegal trade symbol(%s)", order.Symbol), ErrorCode_IllSymbol
	}

	///  Index need: should add member_id index and currency_id index for update quickly, 创建多列唯一索引
	switch order.AorB {
	case TradeType_ASK:

		statement := fmt.Sprintf("SELECT * FROM %s WHERE member_id=? AND currency_id=? FOR UPDATE", TABLE_MONEY)
		err := tx.QueryRow(statement, whoConvertTo(order.Who), int64(cB)).Scan(&cu_id, &member_id, &currency_id, &num, &forzen_num, &status, &chongzhi_url)
		if err != nil {
			isDeadLock, _ := regexp.Match("Error 1213: Deadlock found when trying to get lock;", []byte(err.Error()))
			if isDeadLock {
				DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "UnfreezeFund met deadlock.\n")
				return err, ErrorCode_RecordLocked
			} else {
				return err, ErrorCode_Fail
			}
		}

		//if forzen_num < order.Volume {
		if float32(forzen_num) < float32(order.Volume) {
			coinB, _ := getCoinString(cB)
			return fmt.Errorf("UnfreezeFund fail! Cause: To unfrozen coin(%s) volume(%f) for sell less than sell order volume(%f)", coinB, forzen_num, order.Volume), ErrorCode_IllUnfreezeFund
		}

		statement = fmt.Sprintf("UPDATE %s SET num=num+?, forzen_num=forzen_num-?  WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, order.Volume, order.Volume, whoConvertTo(order.Who), int64(cB))
		checkErr(err)

		///debug:
		DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "[Fund info] ASK: %d, UnfreezeFund.....................................CB:num=num+%f, forzen_num=forzen_num-%f\n", order.ID, order.Volume, order.Volume)

	case TradeType_BID:

		statement := fmt.Sprintf("SELECT * FROM %s WHERE member_id=? AND currency_id=? FOR UPDATE", TABLE_MONEY)
		err := tx.QueryRow(statement, whoConvertTo(order.Who), int64(cQ)).Scan(&cu_id, &member_id, &currency_id, &num, &forzen_num, &status, &chongzhi_url)
		if err != nil {
			isDeadLock, _ := regexp.Match("Error 1213: Deadlock found when trying to get lock;", []byte(err.Error()))
			if isDeadLock {
				DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "UnfreezeFund met deadlock.\n")
				return err, ErrorCode_RecordLocked
			} else {
				return err, ErrorCode_Fail
			}
		}

		bidCost := order.Volume * order.EnOrderPrice
		//if forzen_num < bidCost {
		if float32(forzen_num) < float32(bidCost) {
			coinQ, _ := getCoinString(cQ)
			return fmt.Errorf("UnfreezeFund fail! Cause: To unfrozen coin(%s) volume(%f) for buy less than buy order bidCost(%f)", coinQ, forzen_num, bidCost), ErrorCode_IllUnfreezeFund
		}

		statement = fmt.Sprintf("UPDATE %s SET num=num+?, forzen_num=forzen_num-?  WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, bidCost, bidCost, whoConvertTo(order.Who), int64(cQ))
		checkErr(err)

		///debug:
		DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "[Fund info] BID: %d, UnfreezeFund.....................................CQ:num=num+%f, forzen_num=forzen_num-%f\n", order.ID, bidCost, bidCost)

	default:
		panic("UnfreezeFund met illegal order input!")
	}

	if bLocalTx {
		if err := tx.Commit(); err != nil {
			log.Fatalln(err)
		}
	}

	MySQLOpeResaultLog(res, "UnfreezeFund")
	return nil, ErrorCode_OK
}

/// settle account use at the end of match output to consume frozen fund and get equivalent
func (t *MEMySQLDB) SettleAccount(trade *Trade) error {
	cB, cQ, err := getCoinFromSymbol(trade.Symbol)
	if err != nil {
		return fmt.Errorf("SettleAccount fail! Cause: trade with illegal trade symbol(%s)", trade.Symbol)
	}

	tx, err := t.db.Begin()
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != sql.ErrTxDone && err != nil {
			log.Fatalln(err)
		}
	}(tx)

	var (
		res sql.Result
	)
	switch trade.AorB {
	case TradeType_ASK:
		statement := fmt.Sprintf("UPDATE %s SET forzen_num=forzen_num-? WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, trade.Volume, whoConvertTo(trade.Who), int64(cB))
		checkErr(err)
		statement = fmt.Sprintf("UPDATE %s SET num=num+? WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, trade.Amount, whoConvertTo(trade.Who), int64(cQ))
		checkErr(err)

	case TradeType_BID:
		bidCost := trade.Volume * trade.EnOrderPrice
		repayDiffFromFreezeFund := (trade.EnOrderPrice - trade.Price) * trade.Volume
		statement := fmt.Sprintf("UPDATE %s SET forzen_num=forzen_num-?, num=num+? WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, bidCost, repayDiffFromFreezeFund, whoConvertTo(trade.Who), int64(cQ))
		checkErr(err)
		statement = fmt.Sprintf("UPDATE %s SET num=num+? WHERE member_id=? AND currency_id=?", TABLE_MONEY)
		res, err = tx.Exec(statement, trade.Amount, whoConvertTo(trade.Who), int64(cB))
		checkErr(err)

	default:
		panic("SettleAccount met illegal trade input!")
	}

	if err := tx.Commit(); err != nil {
		log.Fatalln(err)
	}

	MySQLOpeResaultLog(res, "SettleAccount")
	return nil
}

/// use one sql statement to complete 2 record update, not need use transaction
func (t *MEMySQLDB) SettleAccountQuick(trade *Trade) error {
	cB, cQ, err := getCoinFromSymbol(trade.Symbol)
	if err != nil {
		return fmt.Errorf("SettleAccountQuick fail! Cause: trade with illegal trade symbol(%s)", trade.Symbol)
	}

	var (
		res sql.Result
	)
	switch trade.AorB {
	case TradeType_ASK:
		cmd := `UPDATE %s SET forzen_num = CASE currency_id
		  WHEN ? THEN forzen_num-?
		  WHEN ? THEN forzen_num
		  END,
		  num = CASE currency_id
		  WHEN ? THEN num
		  WHEN ? THEN num+?
		  END
		WHERE currency_id IN(?,?) AND member_id=?`
		//		cmd := "UPDATE %s SET forzen_num = CASE currency_id WHEN ? THEN forzen_num-? WHEN ? THEN forzen_num END, num = CASE currency_id WHEN ? THEN num WHEN ? THEN num+? END WHERE currency_id IN(?,?) AND member_id=?"
		statement := fmt.Sprintf(cmd, TABLE_MONEY)
		res, err = t.db.Exec(statement, int64(cB), trade.Volume, int64(cQ), int64(cB), int64(cQ), trade.Amount, int64(cB), int64(cQ), whoConvertTo(trade.Who))
		checkErr(err)

	case TradeType_BID:
		bidCost := trade.Volume * trade.EnOrderPrice
		repayDiffFromFreezeFund := (trade.EnOrderPrice - trade.Price) * trade.Volume
		cmd := `UPDATE %s SET forzen_num = CASE currency_id
		  WHEN ? THEN forzen_num-?
		  WHEN ? THEN forzen_num
		  END,
		  num = CASE currency_id
		  WHEN ? THEN num+?
		  WHEN ? THEN num+?
		  END
		WHERE currency_id IN(?,?) AND member_id=?`
		//		cmd := "UPDATE %s SET forzen_num = CASE currency_id WHEN ? THEN forzen_num-? WHEN ? THEN forzen_num END, num = CASE currency_id WHEN ? THEN num WHEN ? THEN num+? END WHERE currency_id IN(?, ?) AND member_id=?"
		statement := fmt.Sprintf(cmd, TABLE_MONEY)
		res, err = t.db.Exec(statement, int64(cQ), bidCost, int64(cB), int64(cQ), repayDiffFromFreezeFund, int64(cB), trade.Amount, int64(cB), int64(cQ), whoConvertTo(trade.Who))
		checkErr(err)

	default:
		panic("SettleAccountQuick met illegal trade input!")
	}

	MySQLOpeResaultLog(res, "SettleAccountQuick")
	return nil
}

/// Settle method is used to settle the trade resault, that is bid and ask part must sync operate their account or both not.
/// notice: must setup unique multi index use command: alter table ubt_currency_user add unique index uc(member_id,currency_id)
func (t *MEMySQLDB) Settle(bid *Trade, ask *Trade, tx *sql.Tx) (error, ErrorCode) {
	var (
		err error
		res sql.Result
	)

	if bid.Symbol != ask.Symbol {
		return fmt.Errorf("Settle couple not corresponding, cause: bid.Symbol != ask.Symbol"), ErrorCode_IllSymbol
	}
	cB, cQ, err := getCoinFromSymbol(bid.Symbol)
	if err != nil {
		return fmt.Errorf("Settle fail! Cause: bid with illegal bid symbol(%s)", bid.Symbol), ErrorCode_IllSymbol
	}

	/// To lock
	if tx != nil {
		statement := fmt.Sprintf("SELECT * FROM %s WHERE member_id=? AND currency_id=? OR member_id=? AND currency_id=? OR member_id=? AND currency_id=? OR member_id=? AND currency_id=? FOR UPDATE", TABLE_MONEY)
		_, err := tx.Exec(statement,
			whoConvertTo(bid.Who), int64(cB),
			whoConvertTo(bid.Who), int64(cQ),
			whoConvertTo(ask.Who), int64(cB),
			whoConvertTo(ask.Who), int64(cQ),
		)
		if err != nil {
			///Error 1213: Deadlock found when trying to get lock;
			isDeadLock, _ := regexp.Match("Error 1213: Deadlock found when trying to get lock;", []byte(err.Error()))
			if isDeadLock {
				DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Record lock occur(return from mysql:Error 1213: Deadlock found when trying to get lock) when Settle to select for update fund.\n")
				return err, ErrorCode_RecordLocked
			} else {
				return err, ErrorCode_Fail
			}
		}
	}

	/// bid part should pay
	bidCost := bid.Volume * bid.EnOrderPrice
	/// when bid: it freeze fund use bid.TotalVolume * bid.EnOrderPrice, but trade use bid.price,
	/// when settle, freeze partial had been rately consumed with bid.Volume * bid.EnOrderPrice,
	/// so the difference should be payed to the bidder
	repayDiffFromFreezeFund := (bid.EnOrderPrice - bid.Price) * bid.Volume
	cmd := `INSERT INTO %s (member_id, currency_id, num, forzen_num, status, chongzhi_url) VALUES 
			(?, ?, 0, ?, ?, ""), 
			(?, ?, ?, 0, ?, ""), 
			(?, ?, ?, ?, ?, ""), 
			(?, ?, ?, 0, ?, "")
			ON DUPLICATE KEY UPDATE forzen_num=forzen_num-VALUES(forzen_num), num=num+VALUES(num)`
	statement := fmt.Sprintf(cmd, TABLE_MONEY)
	if tx == nil {
		res, err = t.db.Exec(statement,
			whoConvertTo(ask.Who), int64(cB), ask.Volume, fundStausConvertTo(FundStatus_ABN),
			whoConvertTo(ask.Who), int64(cQ), ask.Amount, fundStausConvertTo(FundStatus_ABN),

			whoConvertTo(bid.Who), int64(cQ), repayDiffFromFreezeFund, bidCost, fundStausConvertTo(FundStatus_ABN),
			whoConvertTo(bid.Who), int64(cB), bid.Amount, fundStausConvertTo(FundStatus_ABN))
	} else {
		res, err = tx.Exec(statement,
			whoConvertTo(ask.Who), int64(cB), ask.Volume, fundStausConvertTo(FundStatus_ABN),
			whoConvertTo(ask.Who), int64(cQ), ask.Amount, fundStausConvertTo(FundStatus_ABN),

			whoConvertTo(bid.Who), int64(cQ), repayDiffFromFreezeFund, bidCost, fundStausConvertTo(FundStatus_ABN),
			whoConvertTo(bid.Who), int64(cB), bid.Amount, fundStausConvertTo(FundStatus_ABN))
	}
	if err != nil {
		///Error 1213: Deadlock found when trying to get lock;
		isDeadLock, _ := regexp.Match("Error 1213: Deadlock found when trying to get lock;", []byte(err.Error()))
		if isDeadLock {
			DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Record lock occur(return from mysql:Error 1213: Deadlock found when trying to get lock) when Settle to INSERT(update) fund.\n")
			return err, ErrorCode_RecordLocked
		} else {
			return err, ErrorCode_Fail
		}
	}

	///debug:
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "[Fund info] ASK: %d, Settle.....................................CB:forzen_num-%f, CQ:num=num+%f\n", ask.ID, ask.Volume, ask.Amount)
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "[Fund info] BID: %d, Settle.....................................CB:num=num+%f, CQ:forzen_num-%f\n", bid.ID, bid.Amount, bidCost)

	MySQLOpeResaultLog(res, "Settle")
	return nil, ErrorCode_OK
}

//IdbLongTerm::Finance---------------------------------------------------------------------------------------------------
/// Finance record at finance table
func (t *MEMySQLDB) GetTradeFinance(user string, id int64, symbol string) (*Finance, error) {
	var (
		finance_id  int64
		member_id   int64
		type_       int64
		content     string
		money_type  int64
		money       float64
		add_time    int64
		currency_id int64
		ip          string
	)

	sql := fmt.Sprintf("SELECT * FROM %s WHERE finance_id=?", TABLE_FINANCE)
	err := t.db.QueryRow(sql, id).Scan(&finance_id, &member_id, &type_, &content, &money_type, &money, &add_time, &currency_id, &ip)
	if err != nil {
		fmt.Printf("GetTradeFinance query(id:%d,user:%s,symbol:%s) fail!\n", id, user, symbol)
		fmt.Print(err)
		return nil, err
	}

	trade, err := t.GetTrade(user, id, symbol)
	if err != nil {
		fmt.Printf("GetTradeFinance GetTrade(id:%d,user:%s,symbol:%s) fail!\n", id, user, symbol)
		return nil, err
	}
	return &Finance{*trade, fTypeConvertFrom(type_), money, ip}, nil
}

func (t *MEMySQLDB) AddTradeFinanceCouple(bidF *Finance, askF *Finance, tx *sql.Tx) error {
	var (
		err error
		res sql.Result
	)
	if bidF.Symbol != askF.Symbol {
		return fmt.Errorf("AddTradeFinanceCouple trade couple not corresponding, cause: bidF.Symbol != askF.Symbol")
	}
	cB, cQ, err := getCoinFromSymbol(bidF.Symbol)
	if err != nil {
		return fmt.Errorf("AddTradeFinanceCouple fail! Cause: bidF with illegal bid symbol(%s)", bidF.Symbol)
	}

	cmd := `INSERT INTO %s
			(finance_id, member_id, type, content, money_type, money, add_time, currency_id, ip) VALUES
			(?,?,?,?,?,?,?,?,?),
			(?,?,?,?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE
			money=money+VALUES(money),
			add_time=VALUES(add_time)`
	sql := fmt.Sprintf(cmd, TABLE_FINANCE)
	if tx == nil {
		res, err = t.db.Exec(sql,
			bidF.ID,
			whoConvertTo(bidF.Who),
			fTypeConvertTo(bidF.FType),
			///"Trade fee cost by order(id="+strconv.FormatInt(bidF.ID, 10)+")",
			"交易手续费",
			ioOConvertTo(getInOrOutFromFType(bidF.FType)),
			bidF.FAmount,
			time.Now().Unix(),
			int64(cB),
			bidF.UserIP,

			askF.ID,
			whoConvertTo(askF.Who),
			fTypeConvertTo(askF.FType),
			///"Trade fee cost by order(id="+strconv.FormatInt(bidF.ID, 10)+")",
			"交易手续费",
			ioOConvertTo(getInOrOutFromFType(askF.FType)),
			askF.FAmount,
			time.Now().Unix(),
			int64(cQ),
			askF.UserIP,
		)
	} else {
		res, err = tx.Exec(sql,
			bidF.ID,
			whoConvertTo(bidF.Who),
			fTypeConvertTo(bidF.FType),
			///"Trade fee cost by order(id="+strconv.FormatInt(bidF.ID, 10)+")",
			"交易手续费",
			ioOConvertTo(getInOrOutFromFType(bidF.FType)),
			bidF.FAmount,
			time.Now().Unix(),
			int64(cB),
			bidF.UserIP,

			askF.ID,
			whoConvertTo(askF.Who),
			fTypeConvertTo(askF.FType),
			///"Trade fee cost by order(id="+strconv.FormatInt(bidF.ID, 10)+")",
			"交易手续费",
			ioOConvertTo(getInOrOutFromFType(askF.FType)),
			askF.FAmount,
			time.Now().Unix(),
			int64(cQ),
			askF.UserIP,
		)
	}
	checkErr(err)

	MySQLOpeResaultLog(res, "AddTradeFinanceCouple")
	return nil
}
