// adapter
package use_mysql

import (
	"fmt"
	"strconv"
	"strings"

	"../../config"
	. "../../itf"
)

//Type Adaptive---------------------------------------------------------------------------------------------------
func whoConvertTo(who string) int64 {
	iW, _ := strconv.ParseInt(who, 10, 64)
	return iW
}
func whoConvertFrom(who int64) string {
	return strconv.FormatInt(who, 10)
}

func idConvertTo(id int64) int32 {
	return int32(id)
}
func idConvertFrom(id int32) int64 {
	return int64(id)
}

func timeConvertTo(time int64) int32 {
	return int32(time)
}
func timeConvertFrom(time int64) int64 {
	return int64(time)
}

type CoinType int64

//const (
//	CONIABBRE_BTC  CoinType = 29
//	CONIABBRE_UBT  CoinType = 30
//	CONIABBRE_BCX  CoinType = 32
//	CONIABBRE_LTC  CoinType = 33
//	CONIABBRE_BCH  CoinType = 34
//	CONIABBRE_BTG  CoinType = 35
//	CONIABBRE_ETH  CoinType = 36
//	CONIABBRE_USDT CoinType = 37
//	CONIABBRE_NEO  CoinType = 38
//	CONIABBRE_UEX  CoinType = 39
//	CONIABBRE_EOS  CoinType = 40
//	CONIABBRE_XMR  CoinType = 41
//	CONIABBRE_DASH CoinType = 42
//	CONIABBRE_ZEC  CoinType = 43
//	CONIABBRE_ETC  CoinType = 44

//	CONIABBRE_NIL CoinType = -1 /// invalid coin
//)

//func (p CoinType) String() string {
//	switch p {
//	case CONIABBRE_BTC:
//		return "BTC"
//	case CONIABBRE_UBT:
//		return "UBT"
//	case CONIABBRE_BCX:
//		return "BCX"
//	case CONIABBRE_LTC:
//		return "LTC"
//	case CONIABBRE_BCH:
//		return "BCH"
//	case CONIABBRE_BTG:
//		return "BTG"
//	case CONIABBRE_ETH:
//		return "ETH"
//	case CONIABBRE_USDT:
//		return "USDT"
//	case CONIABBRE_NEO:
//		return "NEO"
//	case CONIABBRE_UEX:
//		return "UEX"
//	case CONIABBRE_EOS:
//		return "EOS"
//	case CONIABBRE_XMR:
//		return "XMR"
//	case CONIABBRE_DASH:
//		return "DASH"
//	case CONIABBRE_ZEC:
//		return "ZEC"
//	case CONIABBRE_ETC:
//		return "ETC"

//	}
//	return "NIL"
//}

//var coinMap map[string]CoinType = map[string]CoinType{
//	"BTC":  CONIABBRE_BTC,
//	"UBT":  CONIABBRE_UBT,
//	"BCX":  CONIABBRE_BCX,
//	"LTC":  CONIABBRE_LTC,
//	"BCH":  CONIABBRE_BCH,
//	"BTG":  CONIABBRE_BTG,
//	"ETH":  CONIABBRE_ETH,
//	"USDT": CONIABBRE_USDT,
//	"NEO":  CONIABBRE_NEO,
//	"UEX":  CONIABBRE_UEX,
//	"EOS":  CONIABBRE_EOS,
//	"XMR":  CONIABBRE_XMR,
//	"DASH": CONIABBRE_DASH,
//	"ZEC":  CONIABBRE_ZEC,
//	"ETC":  CONIABBRE_ETC,
//}

func getCoinFromSymbol(symbol string) (int64, int64, error) {
	s := strings.Split(symbol, "/")
	if len(s) != 2 {
		return -1, -1, fmt.Errorf("getCoinFromSymbol input symbol %s cannot convert correctly!", symbol)
	}
	vB, okB := config.GetCoinMapInt()[s[0]]
	vQ, okQ := config.GetCoinMapInt()[s[1]]
	if okB && okQ {
		return vB, vQ, nil
	}
	return -1, -1, fmt.Errorf("getCoinFromSymbol is processing incorrect trade couple with symbol %s/%s", config.GetCoinMapMark()[vB], config.GetCoinMapMark()[vQ])
}
func getCoinString(coinID int64) (string, bool) {
	coinName, ok := config.GetCoinMapMark()[coinID]
	return coinName, ok
}
func symbolConvertTo(symbol string) (b int64, r int64) {
	s := strings.Split(symbol, "/")
	if len(s) != 2 {
		panic(fmt.Errorf("symbolConvertTo input symbol %s cannot convert correctly!", symbol))
	}
	vB, okB := config.GetCoinMapInt()[s[0]]
	vQ, okQ := config.GetCoinMapInt()[s[1]]
	if okB && okQ {
		return int64(vB), int64(vQ)
	}

	panic(fmt.Errorf("symbolConvertTo is processing incorrect trade couple with symbol %s", symbol))
}
func symbolConvertFrom(b int64, r int64) string {
	coinMapMark := config.GetCoinMapMark()
	vB, okB := coinMapMark[b]
	vR, okR := coinMapMark[r]
	if okB && okR {
		return vB + "/" + vR
	} else {
		fmt.Printf("symbolConvertFrom(b:%d, r:%d) fail.\n", b, r)
		return "Invalid"
	}
	//	vB := CoinType(b).String()
	//	vR := CoinType(r).String()
	//	symbol := vB + "/" + vR
	//	return symbol
}

func tradeTypeConvertTo(aorb TradeType) string {
	switch aorb {
	case TradeType_BID:
		return "buy"
	case TradeType_ASK:
		return "sell"
	}
	return "NULL"
}
func tradeTypeConvertFrom(aorb string) TradeType {
	switch aorb {
	case "buy":
		return TradeType_BID
	case "sell":
		return TradeType_ASK
	}
	return TradeType_UNSET
}

func fundStausConvertTo(fs FundStatus) int64 {
	switch fs {
	case FundStatus_OKK:
		return 0
	case FundStatus_ABN:
		return 1
	}
	return -1
}
func fundStausConvertFrom(fs int64) FundStatus {
	switch fs {
	case 0:
		return FundStatus_OKK
	case 1:
		return FundStatus_ABN
	}
	return FundStatus_UNSET
}

func tradeStatusConvertTo(s TradeStatus) int64 {
	switch s {
	case ORDER_SUBMIT:
		return 0
	case ORDER_FILLED:
		return 2
	case ORDER_PARTIAL_FILLED:
		return 1
	case ORDER_PARTIAL_CANCEL:
		return -1
	case ORDER_CANCELED:
		return -1
	case ORDER_CANCELING:
		return -1
	}
	return 3 /// 3 not define at uex order table
}
func tradeStatusConvertFrom(s int64) TradeStatus {
	switch s {
	case 0:
		return ORDER_SUBMIT
	case 1:
		return ORDER_PARTIAL_FILLED
	case 2:
		return ORDER_FILLED
	case -1:
		return ORDER_CANCELED
	}
	return ORDER_UNKNOWN /// 3 not define at uex order table
}

func ioOConvertTo(io InOrOut) int64 {
	switch io {
	case InOrOut_Earn:
		return 1
	case InOrOut_Pay:
		return 2
	}
	return 0
}
func ioOConvertFrom(io int64) InOrOut {
	switch io {
	case 1:
		return InOrOut_Earn
	case 2:
		return InOrOut_Pay
	}
	return InOrOut_Unknown
}

func fTypeConvertTo(fType FinanceType) int64 {
	switch fType {
	case FinanceType_Profit:
		return 10
	case FinanceType_Rebate:
		return 9
	case FinanceType_Encharge:
		return 6
	case FinanceType_TradeFee:
		return 11
	}
	return 0
}
func fTypeConvertFrom(fType int64) FinanceType {
	switch fType {
	case 6:
		return FinanceType_Encharge
	case 9:
		return FinanceType_Rebate
	case 10:
		return FinanceType_Profit
	case 11:
		return FinanceType_TradeFee
	}
	return FinanceType_Unknown
}

func getInOrOutFromFType(f FinanceType) InOrOut {
	switch f {
	case FinanceType_Profit:
		return InOrOut_Earn
	case FinanceType_Rebate:
		return InOrOut_Earn
	case FinanceType_Encharge:
		return InOrOut_Earn
	case FinanceType_TradeFee:
		return InOrOut_Pay
	}
	return InOrOut_Unknown
}
