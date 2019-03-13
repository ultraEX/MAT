// mysql_common
package use_mysql

import (
	"database/sql"
	"log"
	"time"

	. "../../comm"
)

const (
	MAX_IDLE_CONNS     int           = 100
	MAX_OPEN_CONNS     int           = 100
	MAX_CONN_LIFE_TIME time.Duration = 10 * time.Second
)

type ErrorCode int64

const (
	ErrorCode_OK   ErrorCode = 1
	ErrorCode_Fail ErrorCode = 0

	ErrorCode_IllSymbol       ErrorCode = -1
	ErrorCode_RecordLocked    ErrorCode = -2
	ErrorCode_FundNoEnough    ErrorCode = -3
	ErrorCode_IllUnfreezeFund ErrorCode = -4
	ErrorCode_DupPrimateKey   ErrorCode = -5
	ErrorCode_NoRecord        ErrorCode = -6
)

func (p ErrorCode) String() string {
	switch p {
	case ErrorCode_RecordLocked:
		return "item locked"
	case ErrorCode_FundNoEnough:
		return "fund no enough"
	}
	return "<UNSET>"
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func MySQLOpeResaultLog(res sql.Result, tag string) {
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	DebugPrintf(MODULE_NAME, LOG_LEVEL_TRACK, "%s Use_mysql driver, LastInsertID = %d, affected = %d\n", tag, lastId, rowCnt)
}
