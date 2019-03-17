package comm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	MAX_FLOAT64 float64 = 1.797693134862315708145274237317043567981e+38
	MAX_INT64   int64   = int64(^uint64(0) >> 1)
)

////--------------------------------------------------------------------------------
const (
	COMMON_MODULE_NAME string = "[Common]: "
	VERSION_NO         string = "0.1.5"

	MECORE_MATCH_DURATION  time.Duration = 50 * time.Microsecond /// control match speed: value smaller, action faster, unit is milisecond
	MECORE_CANCEL_DURATION time.Duration = 50 * time.Microsecond /// control cancel speed: value smaller, action faster, unit is milisecond

	RESTFUL_MAX_ORDER_LEVELS int64 = 20
)

var (
	TEST_DEBUG_DISPLAY    bool     = false
	LOG_TO_FILE           bool     = false
	LOG_FILE_SIZE         int64    = 10 * 1000 * 1000
	LOG_LEVEL             LogLevel = LOG_LEVEL_FATAL
	LOG_FILE_NAME_DEBUG   string   = "debug.log"
	CONFIG_JSON_FILE_NAME string   = "MEconfig.json"
	WORK_DIR              string   = ""
	SCRIPT_DIR            string   = "script"
	SHELL_REMOMERY_FN     string   = "rememory.sh"
)

////--------------------------------------------------------------------------------

type LogLevel int64

const (
	LOG_LEVEL_DEBUG  LogLevel = 1
	LOG_LEVEL_TRACK  LogLevel = 2
	LOG_LEVEL_FATAL  LogLevel = 3
	LOG_LEVEL_ALWAYS LogLevel = 4
)

func (p LogLevel) String() string {
	switch p {
	case LOG_LEVEL_DEBUG:
		return "debug"
	case LOG_LEVEL_TRACK:
		return "track"
	case LOG_LEVEL_FATAL:
		return "fatal"
	}
	return "<UNSET>"
}

////--------------------------------------------------------------------------------

type Signal int

func (t Signal) String() string {
	switch t {
	case Signal_Quit:
		return "Quit"
	case Signal_Restart:
		return "Restart"
	}
	return "<Signal UNSET>"
}

const (
	Signal_Quit    Signal = -1
	Signal_Restart Signal = -2
)

var Wait chan Signal

func WaitingSignal() {
	Wait = make(chan Signal)
	signal := <-Wait
	fmt.Printf("Received %s signal to act\n", signal)
	switch signal {
	case Signal_Quit:
		break
	case Signal_Restart:
		break
	}
}

////--------------------------------------------------------------------------------
var (
	LogFile *LogToFile = nil
)

func SetLog2File(b bool) {
	if b {
		if LogFile == nil {
			LogFile = NewLogToFile(WORK_DIR + "/" + LOG_FILE_NAME_DEBUG)
			LogFile.Start()
		}
		LOG_TO_FILE = b
	}

	if !b {
		LOG_TO_FILE = b
		if LogFile != nil {
			LogFile.Stop()
			LogFile = nil
		}
	}
	fmt.Println("Set LOG_TO_FILE: ", LOG_TO_FILE, " complete.")
}
func SwitchLog() {
	SetLog2File(!LOG_TO_FILE)
}
func SetInfo(b bool) {
	TEST_DEBUG_DISPLAY = b
	fmt.Println("Set TEST_DEBUG_DISPLAY: ", TEST_DEBUG_DISPLAY, " complete.")
}
func SwitchInfo() {
	TEST_DEBUG_DISPLAY = !TEST_DEBUG_DISPLAY
	fmt.Println("Set TEST_DEBUG_DISPLAY: ", TEST_DEBUG_DISPLAY, " complete.")
}

func SetLevel(level string) {
	switch level {
	case LOG_LEVEL_DEBUG.String():
		LOG_LEVEL = LOG_LEVEL_DEBUG
	case LOG_LEVEL_TRACK.String():
		LOG_LEVEL = LOG_LEVEL_TRACK
	case LOG_LEVEL_FATAL.String():
		LOG_LEVEL = LOG_LEVEL_FATAL
	}
	fmt.Println("Set LOG_LEVEL: ", LOG_LEVEL, " complete.")
}
func SwitchLevel() {
	switch LOG_LEVEL {
	case LOG_LEVEL_DEBUG:
		LOG_LEVEL = LOG_LEVEL_TRACK
	case LOG_LEVEL_TRACK:
		LOG_LEVEL = LOG_LEVEL_FATAL
	case LOG_LEVEL_FATAL:
		LOG_LEVEL = LOG_LEVEL_DEBUG
	}
	fmt.Println("Set LOG_LEVEL: ", LOG_LEVEL, " complete.")
}

func DebugPrintln(module string, level LogLevel, a ...interface{}) (n int, err error) {
	if LOG_TO_FILE {
		if level >= LOG_LEVEL {
			dateTime := GetDateTime()
			ma := []interface{}{}
			ma = append(ma, dateTime)
			ma = append(ma, module)
			ma = append(ma, a...)
			n, err = LogFile.Println(ma...)
		}
	}
	if TEST_DEBUG_DISPLAY {
		if level >= LOG_LEVEL {
			dateTime := GetDateTime()
			ma := []interface{}{}
			ma = append(ma, dateTime)
			ma = append(ma, module)
			ma = append(ma, a...)
			n, err = fmt.Println(ma...)
		}
	}

	return n, err
}
func DebugPrintf(module string, level LogLevel, format string, a ...interface{}) (n int, err error) {
	if LOG_TO_FILE {
		if level >= LOG_LEVEL {
			dateTime := GetDateTime()
			n, err = LogFile.Printf(dateTime+module+format, a...)
		}
	}
	if TEST_DEBUG_DISPLAY {
		if level >= LOG_LEVEL {
			dateTime := GetDateTime()
			n, err = fmt.Printf(dateTime+module+format, a...)
		}
	}

	return n, err
}

type PINGPONG bool

const (
	PING PINGPONG = false
	PONG PINGPONG = true
)

func (t PINGPONG) String() string {
	switch t {
	case PING:
		return "PING"
	case PONG:
		return "PONG"
	}
	return "<UNSET>"
}

type LogToFile struct {
	f_debug *os.File

	isRun bool
	wait  chan bool
	path  string
}

func NewLogToFile(path string) *LogToFile {
	obj := new(LogToFile)
	f, err := CreateLogFile(path)
	if err != nil {
		panic(err)
	}
	obj.f_debug = f
	obj.path = path
	return obj
}

func (t *LogToFile) SwitchLogFile() {
	for t.isRun {
		if t.f_debug == nil {
			break
		}

		fLen, err := t.f_debug.Seek(0, os.SEEK_END)
		if err != nil {
			panic(err)
		}

		if fLen > LOG_FILE_SIZE {
			/// rename original one as path.old
			if t.isRun {
				t.Stop()
			}
			oldFileName := t.path + ".old"
			if _, err := os.Stat(oldFileName); err == nil {
				err = os.Remove(oldFileName)
				if err != nil {
					panic(err)
				}
			}
			err = os.Rename(t.path, oldFileName)
			if err != nil {
				panic(err)
			}

			/// create a new one for record
			f, err := CreateLogFile(t.path)
			if err != nil {
				panic(err)
			}

			t.f_debug = f
			t.Start()
			runtime.Goexit()
		}
		time.Sleep(3 * time.Second)
	}
}

func (t *LogToFile) Sync() {
	for t.isRun {
		time.Sleep(1 * time.Second)
		t.f_debug.Sync()
	}
	t.wait <- true
}

func (t *LogToFile) Start() {
	t.wait = make(chan bool)
	t.isRun = true
	go t.Sync()
	///go t.SwitchLogFile()
}

func (t *LogToFile) Stop() {
	t.isRun = false

	/// Wait sync end
tagRetry:
	select {
	case _ = <-t.wait:
		//_ = res
	case <-time.After(1 * time.Second):
		/// Test Fail
		fmt.Printf("Wait LogToFile Sync end waist one more 1s...\n")
		goto tagRetry
	}
	close(t.wait)

	t.f_debug.Close()
}

func (t *LogToFile) Printf(format string, a ...interface{}) (n int, err error) {

	buff := fmt.Sprintf(format, a...)
	n, err = t.f_debug.Write([]byte(buff))
	if err != nil {
		fmt.Println(err)
	}
	return n, err
}

func (t *LogToFile) Println(a ...interface{}) (n int, err error) {

	buff := fmt.Sprintln(a...)
	n, err = t.f_debug.Write([]byte(buff))
	if err != nil {
		fmt.Println(err)
	}
	return n, err
}

func CreateLogFile(path string) (*os.File, error) {

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Println("OpenFile ", path, err)
	}

	return f, err
}

type DebugLock struct {
	lock *sync.RWMutex
}

func NewDebugLock(tag string) *DebugLock {
	///DebugPrintln(COMMON_MODULE_NAME, "DebugLock new: ", tag)
	obj := new(DebugLock)
	obj.lock = new(sync.RWMutex)
	return obj
}
func (t *DebugLock) Lock(tag string) {
	///DebugPrintln(COMMON_MODULE_NAME, "DebugLock Lock: ", tag)
	t.lock.Lock()
}
func (t *DebugLock) Unlock(tag string) {
	///DebugPrintln(COMMON_MODULE_NAME, "DebugLock Unlock: ", tag)
	t.lock.Unlock()
}
func (t *DebugLock) RLock(tag string) {
	///DebugPrintln(COMMON_MODULE_NAME, "DebugLock RLock: ", tag)
	t.lock.RLock()
}
func (t *DebugLock) RUnlock(tag string) {
	///DebugPrintln(COMMON_MODULE_NAME, "DebugLock RUnlock: ", tag)
	t.lock.RUnlock()
}

var db_through_pass bool = false

func SetDbThroughpass(is bool) {
	db_through_pass = is
}
func GetDbThroughpass() bool {
	return db_through_pass
}

type TradeType int64

const (
	TradeType_BID   TradeType = 1
	TradeType_ASK   TradeType = 2
	TradeType_UNSET TradeType = -1
)

func (p TradeType) String() string {
	switch p {
	case TradeType_BID:
		return "BID"
	case TradeType_ASK:
		return "ASK"
	}
	return "<UNSET>"
}

type TradeStatus int64

const (
	ORDER_SUBMIT         TradeStatus = 1 // 已挂单
	ORDER_FILLED         TradeStatus = 2 // 已成交
	ORDER_PARTIAL_FILLED TradeStatus = 3 // 已部分成交
	ORDER_PARTIAL_CANCEL TradeStatus = 4 // 已部分成交后取消订单
	ORDER_CANCELED       TradeStatus = 5 // 已取消订单
	ORDER_CANCELING      TradeStatus = 6 // 取消订单ing
	ORDER_UNKNOWN        TradeStatus = 7 // 状态未知
)

func (p TradeStatus) String() string {
	switch p {
	case ORDER_SUBMIT:
		return "ORDER_SUBMIT"
	case ORDER_FILLED:
		return "ORDER_FILLED"
	case ORDER_PARTIAL_FILLED:
		return "ORDER_PARTIAL_FILLED"
	case ORDER_PARTIAL_CANCEL:
		return "ORDER_PARTIAL_CANCEL"
	case ORDER_CANCELED:
		return "ORDER_CANCELED"
	case ORDER_CANCELING:
		return "ORDER_CANCELING"
	}
	return "<UNSET>"
}

type Order struct {
	ID           int64
	Who          string
	AorB         TradeType
	Symbol       string
	Timestamp    int64
	EnOrderPrice float64
	Price        float64
	Volume       float64 /// this turn to trade
	TotalVolume  float64 /// total to trade
	Fee          float64
	Status       TradeStatus
	IPAddr       string
}

type Trade struct {
	Order
	Amount    float64 /// ones get
	TradeTime int64
	FeeCost   float64
}

type OrderLevel struct {
	Price       float64
	Volume      float64
	TotalVolume float64
}

type FundStatus int64

const (
	FundStatus_OKK   FundStatus = 1 // fund ok = normal
	FundStatus_ABN   FundStatus = 2 // fund abn = abnormal
	FundStatus_UNSET FundStatus = -1
)

func (p FundStatus) String() string {
	switch p {
	case FundStatus_OKK:
		return "FundStatus Normal"
	case FundStatus_ABN:
		return "FundStatus Abnormal"
	}
	return "<UNSET>"
}

type Fund struct {
	User string

	AvailableMoney map[string]float64
	FreezedMoney   map[string]float64

	TotalMoney map[string]float64    /// for check uniform
	Status     map[string]FundStatus /// fund status
}

type Finance struct {
	Trade

	FType   FinanceType
	FAmount float64

	UserIP string
}

type InOrOut int64

const (
	InOrOut_Earn    InOrOut = 1 // Income of coin
	InOrOut_Pay     InOrOut = 2 // Payment of coin
	InOrOut_Unknown InOrOut = 0 // Unknown
)

func (t InOrOut) String() string {
	switch t {
	case InOrOut_Earn:
		return "Earn coin"
	case InOrOut_Pay:
		return "Pay coin"
	}
	return "<InOrOut UNSET>"
}

/// 分红; 交易手续费; 充值; 返利
type FinanceType int64

const (
	FinanceType_Profit   FinanceType = 1
	FinanceType_Rebate   FinanceType = 2
	FinanceType_Encharge FinanceType = 3
	FinanceType_TradeFee FinanceType = 4

	FinanceType_Unknown FinanceType = 5
)

func (t FinanceType) String() string {
	switch t {
	case FinanceType_Profit:
		return "Profit"
	case FinanceType_Rebate:
		return "Rebate"
	case FinanceType_Encharge:
		return "Encharge"
	case FinanceType_TradeFee:
		return "TradeFee"
	}
	return "<FinanceType UNSET>"
}

func init() {

	WORK_DIR, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	if LOG_TO_FILE {
		LogFile = NewLogToFile(WORK_DIR + "/" + LOG_FILE_NAME_DEBUG)
		LogFile.Start()
	}

	//// Out print start info...
	DebugPrintf(GetDateTime(), LOG_LEVEL_FATAL, ": Match Engine-%s(Version: %s) begin to start...\n", GetExeName(), VERSION_NO)
	fmt.Printf(GetDateTime()+": Match Engine-%s(Version: %s) begin to start...\n", GetExeName(), VERSION_NO)
	///-----------------------------------------------------------------------
	DebugPrintf(COMMON_MODULE_NAME, LOG_LEVEL_FATAL, "Current work dir: %s\n", WORK_DIR)
	DebugPrintf(COMMON_MODULE_NAME, LOG_LEVEL_FATAL, "Current program name: %s\n", os.Args[0])
	///-----------------------------------------------------------------------
	fmt.Printf("Current work dir: %s\n", WORK_DIR)
	fmt.Printf("Current program name: %s\n", os.Args[0])
}

func GetWorkDir() string {
	if WORK_DIR != "" {
		return WORK_DIR
	} else {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return dir
	}
}

func GetExeName() string {
	p := strings.LastIndexAny(os.Args[0], "/")
	if p == -1 {
		p = 0
	}
	return os.Args[0][p+1:]
}

func GetDateTime() string {
	formate := "2006-01-02T15:04:05Z07:00"
	loc, _ := time.LoadLocation("Local")
	return time.Now().In(loc).Format(formate)
}

func MinINT64(a int64, b int64) int64 {
	if a > b {
		return b
	} else {
		return a
	}
}

func CommandExecute(strCmd string) (string, error) {

	cmd := exec.Command("/bin/bash")
	cmd.Stdin = strings.NewReader(strCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("cmd.Run(%s) fail.\n", strCmd)
		return "", fmt.Errorf("cmd.Run(%s) fail.\n", strCmd)
	}

	return out.String(), nil
}

func getRememoryShellFile() string {
	return WORK_DIR + "/" + SCRIPT_DIR + "/" + SHELL_REMOMERY_FN
}
func Rememory() string {
	var strOut string = ""
	scriptFile := getRememoryShellFile()

	fl, err := os.Open(scriptFile)
	if err != nil {
		strOut = fmt.Sprintf("Rememory fail, no script file(%s) found.\n", scriptFile)

		return strOut
	}
	defer fl.Close()
	buf_len, _ := fl.Seek(0, os.SEEK_END)
	fl.Seek(0, os.SEEK_SET)
	buf := make([]byte, buf_len)
	_, err = fl.Read(buf)

	strOut, _ = CommandExecute(string(buf))
	strOut = fmt.Sprintf("Rememory Result:\n%s\n", strOut)

	return strOut
}

////--------------------------------------------------------------------------------
/// NECO Logger Switch
var NECO_LOG_SWITCH bool = false

func SetNecoLog(onOff bool) {
	NECO_LOG_SWITCH = onOff
	fmt.Println("Set NECO_LOG_SWITCH: ", NECO_LOG_SWITCH, " complete.")
}
func SwitchNecoLog() {
	NECO_LOG_SWITCH = !NECO_LOG_SWITCH
	fmt.Println("Set NECO_LOG_SWITCH: ", NECO_LOG_SWITCH, " complete.")
}
func GetNecoLog() bool {
	return NECO_LOG_SWITCH
}

func LenOfSyncMap(m *sync.Map) int64 {
	var c int64 = 0
	m.Range(func(k, v interface{}) bool {
		c++
		return true
	})
	return c
}
