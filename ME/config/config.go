// sym
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	. "../comm"
)

const (
	MODULE_NAME                  string = "[Config]: "
	RECONSTRUCT_TICKERS_PASSWORD string = "Iamsuretodoitpleaseaction!"
)

type MarketType int

const (
	MarketType_MixHR MarketType = 0
	MarketType_Human MarketType = 1
	MarketType_Robot MarketType = 2
	MarketType_Num   MarketType = 3
)

func (t MarketType) String() string {
	switch t {
	case MarketType_MixHR:
		return "MixHR Market"
	case MarketType_Human:
		return "Human Market"
	case MarketType_Robot:
		return "Robot Market"
	}
	return "Unset-Market"
}
func String2MarketType(str string) MarketType {
	switch str {
	case "mix":
		return MarketType_MixHR
	case "human":
		return MarketType_Human
	case "robot":
		return MarketType_Robot
	}
	return MarketType_Num
}

type Symbol struct {
	BaseCoin  string
	QuoteCoin string
}

func (t *Symbol) String() string {
	return t.BaseCoin + "/" + t.QuoteCoin
}

func Struct(str string) *Symbol {
	s := strings.Split(str, "/")
	if len(s) != 2 {
		return &Symbol{BaseCoin: "", QuoteCoin: ""}
	}

	return &Symbol{BaseCoin: s[0], QuoteCoin: s[1]}
}

type MarketInternal struct {
	Sym             Symbol  `json:"Sym"`
	Market_MixHR    bool    `json:"Market_MixHR"`
	Market_Human    bool    `json:"Market_Human"`
	Market_Robot    bool    `json:"Market_Robot"`
	RobotList       []int64 `json:"RobotList"`
	NoneFinanceList []int64 `json:"NoneFinanceList"`
}

type CoinInfo struct {
	Mark string `json:"Mark"`
	ID   int64  `json:"ID"`
}

type MarketConfig struct {
	Sym            Symbol
	Market_MixHR   bool
	Market_Human   bool
	Market_Robot   bool
	RobotSet       HashSet
	NoneFinanceSet HashSet
}

type MarketSlice struct {
	Markets []MarketInternal
}

type MEConfig struct {

	/// RPC param -------------------
	CommandRPCIPPort string `json:"CommandRPCIPPort"`
	ManagerRPCIPPort string `json:"ManagerRPCIPPort"`
	JSONRPCIPPort    string `json:"JSONRPCIPPort"`
	RESTfulIPPort    string `json:"RESTfulIPPort"`

	/// Mysql param -------------------
	MySQL_IP_ME   string `json:"MySQL_IP_ME"`
	MySQL_PORT_ME int64  `json:"MySQL_PORT_ME"`
	MySQL_User_ME string `json:"MySQL_User_ME"`
	MySQL_Pwd_ME  string `json:"MySQL_Pwd_ME"`

	MySQL_IP_TE   string `json:"MySQL_IP_TE"`
	MySQL_PORT_TE int64  `json:"MySQL_PORT_TE"`
	MySQL_User_TE string `json:"MySQL_User_TE"`
	MySQL_Pwd_TE  string `json:"MySQL_Pwd_TE"`

	/// InPool Mode param -------------
	InPoolMode     string `json:"InPoolMode"` /// block and unblock inputpool mode
	MatchAlgorithm string `json:"MatchAlgorithm"`

	/// CoinType Map -------------------
	CoinList []CoinInfo `json:"CoinList"`

	/// Match Engine -------------------
	Markets []MarketInternal `json:"Markets"`
}

func getConfigFileName() string {
	return WORK_DIR + "/" + CONFIG_JSON_FILE_NAME
}

func Load() (*MEConfig, error) {
	conf := getConfigFileName()
	fl, err := os.Open(conf)
	if err != nil {
		DebugPrintln(MODULE_NAME, LOG_LEVEL_FATAL, "Open ", conf, err)
		fmt.Println("Open ", conf, err)
		return nil, err
	}
	defer fl.Close()

	buf_len, _ := fl.Seek(0, os.SEEK_END)
	fl.Seek(0, os.SEEK_SET)

	buf := make([]byte, buf_len)
	_, err = fl.Read(buf)
	if err != nil {
		DebugPrintln(MODULE_NAME, LOG_LEVEL_FATAL, "Read ", conf, err)
		fmt.Println("Read ", conf, err)
		return nil, err
	}

	var meConfig MEConfig
	json.Unmarshal(buf, &meConfig)

	fmt.Printf("ME Config Loaded from %s:\n%+v\n", conf, meConfig)
	return &meConfig, nil
}

/// config store
func Save() error {

	jstr, err := json.Marshal(meConfig)
	if err != nil {
		panic(fmt.Errorf("Construct meConfig fail."))
	}

	var Jout bytes.Buffer
	err = json.Indent(&Jout, jstr, "", "\t")
	if err != nil {
		panic(fmt.Errorf("Format MEconfig.json fail."))
	}

	DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "Begin to update config file.\n")
	conf := getConfigFileName()
	fl, err := os.OpenFile(conf, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "OpenFile %s to save fail. error: %s\n", conf, err.Error())
		fmt.Println("OpenFile ", conf, err)
		return err
	}
	defer fl.Close()

	_, err = fl.Write(Jout.Bytes())
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "Save config to write fail. file: %s, error: %s\n", conf, err.Error())
		fmt.Println(err)
		return err
	}

	DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "Update config file success.\n")
	fmt.Printf("Save MEConfig To Update to %s:\n%+v\n", conf, meConfig)
	return nil
}

var (
	meConfig *MEConfig
	//	marketMap       map[Symbol]MarketInternal
	marketConfigMap map[Symbol]MarketConfig

	coinMapInt  map[string]int64
	coinMapMark map[int64]string
)

func init() {
	fmt.Printf("Start to load config from %s\n", CONFIG_JSON_FILE_NAME)

	var err error
	meConfig, err = Load()
	if err != nil {
		panic(err)
	}

	/// Map to marketConfigMap
	//	marketMap = make(map[Symbol]MarketInternal)
	marketConfigMap = make(map[Symbol]MarketConfig)
	for _, market := range meConfig.Markets {
		addMarketConfigMap(market)
	}

	/// Map to coinMap
	coinMapInt = make(map[string]int64)
	coinMapMark = make(map[int64]string)
	for _, coin := range meConfig.CoinList {
		coinMapInt[coin.Mark] = coin.ID
		coinMapMark[coin.ID] = coin.Mark
	}

	/// Config validity check
	CheckMarketConfig()

	/// Config 3part Modules Param
}

func GetMEConfig() *MEConfig {
	return meConfig
}

func GetCoinMapInt() map[string]int64 {
	return coinMapInt
}

func GetCoinMapMark() map[int64]string {
	return coinMapMark
}

func GetCoinList() []CoinInfo {
	return meConfig.CoinList
}

func RemoveDuplicateCoin(info CoinInfo, list []CoinInfo) ([]CoinInfo, error) {
	if list == nil {
		return list, fmt.Errorf("RemoveDuplicateCoin input error.")
	}

	for i := 0; i < len(list); {
		if info.Mark == list[i].Mark {
			list = append(list[:i], list[i+1:]...)
		} else {
			i++
		}
	}

	return list, nil
}

func AddCoinMap(info CoinInfo) error {

	/// remove duplication
	id, ok := coinMapInt[info.Mark]
	if ok {
		if id == info.ID {
			return fmt.Errorf("Coin had existed.")
		} else {
			meConfig.CoinList, _ = RemoveDuplicateCoin(info, meConfig.CoinList)
			//fmt.Println(meConfig.CoinList)
		}
	}

	/// running time config
	meConfig.CoinList = append(meConfig.CoinList, info)
	coinMapInt[info.Mark] = info.ID
	coinMapMark[info.ID] = info.Mark

	if err := Save(); err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "AddCoinMap to save config fail. coin: %s:%d, error: %s\n", info.Mark, info.ID, err.Error())
		fmt.Println(err)
		return err
	}

	return nil
}

func RemoveCoinMap(mark string) (string, error) {

	/// remove duplication
	id, ok := coinMapInt[mark]
	if ok {
		meConfig.CoinList, _ = RemoveDuplicateCoin(CoinInfo{Mark: mark, ID: id}, meConfig.CoinList)
		fmt.Println(meConfig.CoinList)

		delete(coinMapInt, mark)
		delete(coinMapMark, id)

		if err := Save(); err != nil {
			DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "RemoveCoinMap to save config fail. coin: %s:%d, error: %s\n", mark, id, err.Error())
			fmt.Println(err)
			return "RemoveCoinMap to save config fail", err
		}
		return "RemoveCoinMap success", nil
	} else {
		return "The coin removing not exist", nil
	}
}

func RemoveDuplicateMarket(info MarketInternal, list []MarketInternal) ([]MarketInternal, error) {
	if list == nil {
		return list, fmt.Errorf("RemoveDuplicateMarket input error.")
	}

	for i := 0; i < len(list); {
		if info.Sym == list[i].Sym {
			list = append(list[:i], list[i+1:]...)
		} else {
			i++
		}
	}

	return list, nil
}

func AddMarketMap(market MarketInternal) (error, bool) {
	/// add new market to running time markets
	err, up := addMarketConfigMap(market)
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "AddMarketMap to addMarketConfigMap fail. Error: %s\n", err.Error())
		fmt.Println(err)
		return err, false
	}

	/// remove duplication
	meConfig.Markets, _ = RemoveDuplicateMarket(market, meConfig.Markets)

	/// add market item to meConfig
	meConfig.Markets = append(meConfig.Markets, market)

	return nil, up
}

func addMarketConfigMap(market MarketInternal) (error, bool) {
	var needUpdate bool = false
	_, ok := marketConfigMap[market.Sym]
	if ok {
		needUpdate = true
	}

	robotSet := NewHashSet()
	for _, robot := range market.RobotList {
		robotSet.Add(robot)
	}
	noneFinanceSet := NewHashSet()
	for _, robot := range market.NoneFinanceList {
		noneFinanceSet.Add(robot)
	}
	marketConfig := MarketConfig{
		Sym:            market.Sym,
		Market_MixHR:   market.Market_MixHR,
		Market_Human:   market.Market_Human,
		Market_Robot:   market.Market_Robot,
		RobotSet:       *robotSet,
		NoneFinanceSet: *noneFinanceSet,
	}
	marketConfigMap[market.Sym] = marketConfig
	return nil, needUpdate
}

func RemoveMarketMap(sym Symbol) (error, bool) {
	/// remove market from running time markets
	var needUpdate bool = false

	_, ok := marketConfigMap[sym]
	if ok {
		needUpdate = true

		/// remove from meConfig.Markets
		for i := 0; i < len(meConfig.Markets); {
			if sym == meConfig.Markets[i].Sym {
				meConfig.Markets = append(meConfig.Markets[:i], meConfig.Markets[i+1:]...)
			} else {
				i++
			}
		}
	}

	return nil, needUpdate
}

func GetCoinNames() []string {
	names := []string{}
	for _, coin := range meConfig.CoinList {
		names = append(names, coin.Mark)
	}
	return names
}

func GetSymbols() []Symbol {
	syms := []Symbol{}
	for _, market := range meConfig.Markets {
		syms = append(syms, market.Sym)
	}

	return syms
}

func GetMarket(symbol Symbol) MarketConfig {
	return marketConfigMap[symbol]
}

func GetNoneFinanceSet(symbol string) *HashSet {
	marketConfig, ok := marketConfigMap[*Struct(symbol)]
	if !ok {
		fmt.Printf("GetNoneFinanceSet(%s) fail.\n", symbol)
		return nil
	}
	return &marketConfig.NoneFinanceSet
}

//func GetRobotList(symbol string) ([]int64, error) {
//	marketConfig, ok := marketMap[*Struct(symbol)]
//	if !ok {
//		return nil, fmt.Errorf("GetRobotList(%s) fail.\n", symbol)
//	}
//	return marketConfig.RobotList, nil
//}
func GetRobotSet(symbol string) *HashSet {
	marketConfig, ok := marketConfigMap[*Struct(symbol)]
	if !ok {
		fmt.Printf("GetRobotSet(%s) fail.\n", symbol)
		return nil
	}
	return &marketConfig.RobotSet
}

func CheckMarketConfig() {
	for _, market := range meConfig.Markets {
		if (market.Market_MixHR && market.Market_Human) ||
			(market.Market_MixHR && market.Market_Robot) {
			panic(fmt.Errorf("MixHR market exist does not permit solo market exist"))
		}
	}
}

func (mktConf MarketConfig) Dump() string {
	strBuff := fmt.Sprintf("Symbol: %s \t", mktConf.Sym.String())
	strBuff += fmt.Sprintf("Market_MixHR: %t: \t", mktConf.Market_MixHR)
	strBuff += fmt.Sprintf("Market_Human: %t: \t", mktConf.Market_Human)
	strBuff += fmt.Sprintf("Market_Robot: %t: \t", mktConf.Market_Robot)
	strBot := ""
	for _, robot := range mktConf.RobotSet.Elements() {
		strBot += (strconv.FormatInt(robot.(int64), 10) + ",")
	}
	strBuff += fmt.Sprintf("Robot Set: %s\t", strBot)
	strNoneFinance := ""
	for _, noneFinance := range mktConf.NoneFinanceSet.Elements() {
		strNoneFinance += (strconv.FormatInt(noneFinance.(int64), 10) + ",")
	}
	strBuff += fmt.Sprintf("NoneFinanceRobot Set: %s\n", strBot)

	return strBuff
}
