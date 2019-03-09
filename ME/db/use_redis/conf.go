// utils
package use_redis

import (
	"strconv"
)

func orderSetKey(symbol string) string {
	return "os_" + symbol
}

func orderHashKey(user string, id int64) string {
	return "oh_" + user + strconv.FormatInt(id, 10)
}

func orderHashKeyByID(id int64) string {
	return "oh_" + "*" + strconv.FormatInt(id, 10)
}

func orderHashKeyByUser(user string) string {
	return "oh_" + user + "*"
}

func tradeSetKey(symbol string) string {
	return "ts_" + symbol
}

func tradeHashKey(user string, id int64) string {
	return "th_" + user + strconv.FormatInt(id, 10)
}

func tradeHashKeyByID(id int64) string {
	return "th_" + "*" + strconv.FormatInt(id, 10)
}

func tradeHashKeyByUser(user string) string {
	return "th_" + user + "*"
}
