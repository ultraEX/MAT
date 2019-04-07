// ultraex_atom
package use_mysql

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	. "../../comm"
	"../../config"

	_ "github.com/go-sql-driver/mysql"
)

///---------------------------------------------------------------------------------------------------
/// EnOrder = FundFreeze + AddOrder
func (t *MEMySQLDB) EnOrder(order *Order) (error, ErrorCode) {
	if !GetDbThroughpass() {
		return t.enOrder(order)
	} else {
		return nil, ErrorCode_OK
	}
}

///---------------------------------------------------------------------------------------------------
/// CancelOrder = UnfreezeFund + RmOrder
func (t *MEMySQLDB) CancelOrder(order *Order) (error, ErrorCode) {
	if !GetDbThroughpass() {
		return t.cancelOrder(order)
	} else {
		return nil, ErrorCode_OK
	}
}

/// UpdateTrade = Update trade couple = (BID+ASK)*(Trade update + Order update + Fund update)
func (t *MEMySQLDB) UpdateTrade(bidTrade *Trade, askTrade *Trade) (error, ErrorCode) {
	if !GetDbThroughpass() {
		return t.updateTrade(bidTrade, askTrade)
	} else {
		return nil, ErrorCode_OK
	}
}

/// UpdateTicker with parameters
func (t *TEMySQLDB) UpdateTicker(sym string, _type TickerType, ticker *TickerInfo) (error, ErrorCode) {
	if !GetDbThroughpass() {
		return t.updateTicker(sym, _type, ticker)
	} else {
		return nil, ErrorCode_OK
	}
}

/// EnOrder = FundFreeze + AddOrder
func (t *MEMySQLDB) enOrder(order *Order) (error, ErrorCode) {
	var (
		tx      *sql.Tx
		err     error
		errCode ErrorCode
	)
	/// debug:
	TimeDot1 := time.Now().UnixNano()

retry:
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

	//// FreezeFund:
	orderID, _ := strconv.ParseInt(order.Who, 10, 64)
	noneFinanceSet := config.GetNoneFinanceSet(order.Symbol)
	if !noneFinanceSet.Contains(orderID) {
		err, errCode = t.FreezeFund(order, tx)
		if err != nil {
			DebugPrintln(MODULE_NAME, LOG_LEVEL_ALWAYS, err)
			if errCode == ErrorCode_RecordLocked {
				err := tx.Rollback()
				if err != sql.ErrTxDone && err != nil {
					log.Fatalln(err)
				}
				time.Sleep(MECORE_MATCH_DURATION)
				goto retry
			} else {
				return err, errCode
			}
		}
	}

	//// AddOrder:
	order.Volume = 0
	err, errCode = t.AddOrder(order, tx)
	if err != nil {
		return err, errCode
	}

	if err := tx.Commit(); err != nil {
		log.Fatalln(err)
		return fmt.Errorf("EnOrder tx commit fail!"), ErrorCode_Fail
	}

	/// debug:
	TimeDot2 := time.Now().UnixNano()
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "EnOrder== EnOrder(User(%s), ID(%d), Price(%f), Volume(%f)) to Engine complete. ****USE_TIME: %f.\n",
		order.Who, order.ID, order.Price, order.TotalVolume, float64(TimeDot2-TimeDot1)/float64(1*time.Second))
	return nil, ErrorCode_OK
}

///---------------------------------------------------------------------------------------------------
/// CancelOrder = UnfreezeFund + RmOrder
func (t *MEMySQLDB) cancelOrder(order *Order) (error, ErrorCode) {
	var (
		tx      *sql.Tx
		err     error
		errCode ErrorCode
	)

	/// debug:
	TimeDot1 := time.Now().UnixNano()

retry:
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

	if order.Volume > order.TotalVolume {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "[Illegal Order detail] Order: ",
			"Type(%s), User(%s), ID(%d), Status(%s), Price(%f), TotalVolume(%f), TradeVolume(%f)\n",
			order.AorB, order.Who, order.ID, order.Price, order.TotalVolume, order.Volume)
		panic(fmt.Errorf("CancelOrder Met illegal Order with Volume bigger than TotalVolume."))
	}

	//// UnfreezeFund:
	orderID, _ := strconv.ParseInt(order.Who, 10, 64)
	noneFinanceSet := config.GetNoneFinanceSet(order.Symbol)
	if !noneFinanceSet.Contains(orderID) {
		err, errCode = t.UnfreezeFund(order, tx)
		if err != nil {
			DebugPrintln(MODULE_NAME, LOG_LEVEL_ALWAYS, err)
			if errCode == ErrorCode_RecordLocked {
				err := tx.Rollback()
				if err != sql.ErrTxDone && err != nil {
					log.Fatalln(err)
				}
				time.Sleep(MECORE_MATCH_DURATION)
				goto retry
			} else {
				return err, errCode
			}
		}
	}

	//// RmOrder:
	err = t.RmOrder(order.Who, order.ID, order.Symbol, tx)
	if err != nil {
		panic(err)
	}

	//// AddTrade (ORDER_PARTIAL_CANCEL or ORDER_CANCELED)
	/// notice: trade.Amount and trade.FeeCost not exactly because the amount info not recorded in every sub trade, would be fixed !
	/// in this mechanize: only output a roughly report to trade table
	trade := Trade{*order, 0, 0, 0}
	switch order.AorB {
	case TradeType_BID:
		trade.Amount = trade.Volume
	case TradeType_ASK:
		trade.Amount = trade.Price * (trade.Volume)
	default:
		panic(fmt.Errorf("CancelOrder input order tradetype error!"))
	}
	trade.TradeTime = time.Now().UnixNano()
	trade.FeeCost = trade.Amount * trade.Fee
	if trade.Volume == trade.TotalVolume {
		trade.Status = ORDER_CANCELED
	} else {
		trade.Status = ORDER_PARTIAL_CANCEL
	}
	if trade.Volume != 0 {
		err = t.InsertTrade(&trade, tx)
		if err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalln(err)
		return fmt.Errorf("CancelOrder tx commit fail!"), ErrorCode_Fail
	}

	/// debug:
	TimeDot2 := time.Now().UnixNano()
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK,
		"CancelOrder== CancelOrder( User(%s), ID(%d), Price(%f), Volume(%f)) from Duration Storage complete. ****USE_TIME: %f\n",
		order.Who, order.ID, order.Price, order.TotalVolume, float64(TimeDot2-TimeDot1)/float64(1*time.Second))
	return nil, ErrorCode_OK
}

/// UpdateTrade = Update trade couple = (BID+ASK)*(Trade update + Order update + Fund update)
func (t *MEMySQLDB) updateTrade(bidTrade *Trade, askTrade *Trade) (error, ErrorCode) {
	var (
		tx      *sql.Tx
		err     error
		errCode ErrorCode
	)

	/// debug:
	TimeDot1 := time.Now().UnixNano()

retry:
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

	/// UpdateUserFund:
	err, errCode = t.Settle(bidTrade, askTrade, tx)
	if err != nil {
		DebugPrintln(MODULE_NAME, LOG_LEVEL_ALWAYS, err)
		if errCode == ErrorCode_RecordLocked {
			err := tx.Rollback()
			if err != sql.ErrTxDone && err != nil {
				log.Fatalln(err)
			}
			time.Sleep(MECORE_MATCH_DURATION)
			goto retry
		} else {
			return err, errCode
		}
	}

	/// ask + bid all filled -> couple record trade and remove from order table
	/// ask or bid filled -> eighter to record trade and remove from order table
	if bidTrade.Status == ORDER_FILLED && askTrade.Status == ORDER_FILLED {
		err = t.InsertTradeCouple(bidTrade, askTrade, tx)
		if err != nil {
			panic(err)
		}

		err = t.UpdateOrder(&bidTrade.Order, tx)
		if err != nil {
			panic(err)
		}
		err = t.UpdateOrder(&askTrade.Order, tx)
		if err != nil {
			panic(err)
		}
	} else {
		err = t.InsertTrade(bidTrade, tx)
		if err != nil {
			panic(err)
		}
		err = t.UpdateOrder(&bidTrade.Order, tx)
		if err != nil {
			panic(err)
		}
		err = t.UpdateOrder(&askTrade.Order, tx)
		if err != nil {
			panic(err)
		}
		err = t.InsertTrade(askTrade, tx)
		if err != nil {
			panic(err)
		}
	}

	/// To record Finance info to Mysql database: Finance table
	bidID, _ := strconv.ParseInt(bidTrade.Who, 10, 64)
	askID, _ := strconv.ParseInt(askTrade.Who, 10, 64)
	noneFinanceSet := config.GetNoneFinanceSet(bidTrade.Symbol)
	if !noneFinanceSet.Contains(bidID) || !noneFinanceSet.Contains(askID) {
		bidFinance := Finance{*bidTrade, FinanceType_TradeFee, bidTrade.FeeCost, bidTrade.IPAddr}
		askFinance := Finance{*askTrade, FinanceType_TradeFee, askTrade.FeeCost, askTrade.IPAddr}
		err = t.AddTradeFinanceCouple(&bidFinance, &askFinance, tx)
		if err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalln(err)
		return fmt.Errorf("UpdateTrade tx commit fail!"), ErrorCode_Fail
	}

	/// debug:
	TimeDot2 := time.Now().UnixNano()
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK,
		"UpdateTrade== Engine UpdateTrade Couple Trade Complete. ****USE_TIME: %f.\n\tBid: User(%s), ID(%d), Price(%f), Volume/TotalVolume(%f/%f)\n\tAsk: User(%s), ID(%d), Price(%f), Volume/TotalVolume(%f/%f).\n",
		float64(TimeDot2-TimeDot1)/float64(1*time.Second),
		bidTrade.Who, bidTrade.ID, bidTrade.Price, bidTrade.Volume, bidTrade.TotalVolume, askTrade.Who, askTrade.ID, askTrade.Price, askTrade.Volume, askTrade.TotalVolume)
	return nil, ErrorCode_OK
}

/// UpdateTicker with parameters
func (t *TEMySQLDB) updateTicker(sym string, _type TickerType, ticker *TickerInfo) (error, ErrorCode) {
	err := t.AddTicker(sym, _type, ticker)
	if err != nil {
		panic(err)
	}
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK,
		"UpdateTicker== Tickers Engine UpdateTicker with ticker info:\tSym: %s, TickerType: %s, From: %s,  OpenPrice: %f, ClosePrice: %f, LowPrice: %f, HightPrice: %f, Volume: %f, Amount: %f\n",
		sym,
		_type.String(),
		time.Unix(0, ticker.From).Format("2006-01-02T15:04:05Z07:00"),
		ticker.OpenPrice,
		ticker.ClosePrice,
		ticker.LowPrice,
		ticker.HightPrice,
		ticker.Volume,
		ticker.Amount,
	)

	return nil, ErrorCode_OK
}
