// json_rpc
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"../config"
	. "../itf"
	jrpc2 "./jrpc2_"
)

// -------------------------------------------------------------------------------
// Sample:
// This struct is used for unmarshaling the method params
type AddParams struct {
	X *float64 `json:"x"`
	Y *float64 `json:"y"`
}

// Each params struct must implement the FromPositional method.
// This method will be passed an array of interfaces if positional parameters
// are passed in the rpc call
func (ap *AddParams) FromPositional(params []interface{}) error {
	if len(params) != 2 {
		return errors.New("exactly two integers are required")
	}

	x := params[0].(float64)
	y := params[1].(float64)
	ap.X = &x
	ap.Y = &y

	return nil
}

// Each method should match the prototype <fn(json.RawMessage) (inteface{}, *ErrorObject)>
func Add(params json.RawMessage) (interface{}, *jrpc2.ErrorObject) {
	p := new(AddParams)

	// ParseParams is a helper function that automatically invokes the FromPositional
	// method on the params instance if required
	if err := jrpc2.ParseParams(params, p); err != nil {
		return nil, err
	}

	if p.X == nil || p.Y == nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InvalidParamsCode,
			Message: jrpc2.InvalidParamsMsg,
			Data:    "exactly two integers are required",
		}
	}

	return *p.X + *p.Y, nil
}

// -------------------------------------------------------------------------------
// Add new coin
type AddCoinParams struct {
	Mark *string `json:"Mark"`
	ID   *int64  `json:"ID"`
}

func (t *AddCoinParams) FromPositional(params []interface{}) error {
	if len(params) != 2 {
		return errors.New("exactly AddCoinParams[Mark, ID] are required")
	}

	p0 := params[0].(string)
	p1 := params[1].(float64)
	t.Mark = &p0
	Ip1 := int64(p1)
	t.ID = &Ip1

	return nil
}

func AddCoin(params json.RawMessage) (interface{}, *jrpc2.ErrorObject) {
	p := new(AddCoinParams)

	if err := jrpc2.ParseParams(params, p); err != nil {
		return nil, err
	}

	if p.Mark == nil || p.ID == nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InvalidParamsCode,
			Message: jrpc2.InvalidParamsMsg,
			Data:    "exactly config.CoinInfo[Mark, ID] are required",
		}
	}

	info := config.CoinInfo{Mark: *p.Mark, ID: *p.ID}
	if err := config.AddCoinMap(info); err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "AddCoin to AddCoinMap fail! Error=%s\n", err.Error())
		return "Add coin to config.AddCoinMap fail.", nil
	}
	return "Add coin success.", nil
}

// -------------------------------------------------------------------------------
// Remove coin
type RemoveCoinParams struct {
	Mark *string `json:"Mark"`
}

func (t *RemoveCoinParams) FromPositional(params []interface{}) error {
	if len(params) != 1 {
		return errors.New("exactly RemoveCoinParams[Mark] are required")
	}

	p0 := params[0].(string)
	t.Mark = &p0

	return nil
}

func RemoveCoin(params json.RawMessage) (interface{}, *jrpc2.ErrorObject) {
	var (
		msg string = ""
		err error  = nil
	)
	p := new(RemoveCoinParams)

	if err := jrpc2.ParseParams(params, p); err != nil {
		return nil, err
	}

	if p.Mark == nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InvalidParamsCode,
			Message: jrpc2.InvalidParamsMsg,
			Data:    "exactly RemoveCoinParams[Mark] are required",
		}
	}

	//// operate:
	if msg, err = config.RemoveCoinMap(*p.Mark); err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "RemoveCoin to RemoveCoinMap fail! Error=%s\n", err.Error())
		return "Remove coin to config.RemoveCoinMap fail.", nil
	}
	return "Remove coin success with msg: " + msg, nil
}

// -------------------------------------------------------------------------------
// Add new market
type AddMarketParams struct {
	Sym             *config.Symbol `json:"Sym"`
	Market_MixHR    *bool          `json:"Market_MixHR"`
	Market_Human    *bool          `json:"Market_Human"`
	Market_Robot    *bool          `json:"Market_Robot"`
	RobotList       []int64        `json:"RobotList"`
	NoneFinanceList []int64        `json:"NoneFinanceList"`
}

func (t *AddMarketParams) FromPositional(params []interface{}) error {
	return nil
}

func AddMarket(params json.RawMessage) (interface{}, *jrpc2.ErrorObject) {
	p := new(AddMarketParams)

	if err := jrpc2.ParseParams(params, p); err != nil {
		return nil, err
	}

	if p.Sym == nil || p.Market_MixHR == nil || p.Market_Human == nil || p.Market_Robot == nil || p.RobotList == nil || p.NoneFinanceList == nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InvalidParamsCode,
			Message: jrpc2.InvalidParamsMsg,
			Data:    "exactly AddMarketParams{Sym, Market_MixHR, Market_Human, Market_Robot, RobotList, NoneFinanceList} are required",
		}
	}

	market := config.MarketInternal{
		Sym:             *p.Sym,
		Market_MixHR:    *p.Market_MixHR,
		Market_Human:    *p.Market_Human,
		Market_Robot:    *p.Market_Robot,
		RobotList:       p.RobotList,
		NoneFinanceList: p.NoneFinanceList,
	}

	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
			go restartME()
		}
	}()

	err, up := config.AddMarketMap(market)
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "AddMarket to AddMarketMap fail! Error=%s\n", err.Error())
		return "Add market to config.AddMarketMap fail.", nil
	}

	/// active market
	if !up {
		if err = marketsRef.Add1Market(market.Sym); err != nil {
			go restartME()
			return "AddMarket Something wrong to active new market. ME would restart to restore from errors.", nil
		}

		/// if no error, save conifg
		if err := config.Save(); err != nil {
			DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "AddMarket to save config fail. Error: %s\n", err.Error())
			fmt.Println(err)
			go restartME()
			return "AddMarket Something wrong to act new market and save config file. ME would restart to restore from errors.", nil
		}
	} else {
		/// save conifg
		if err := config.Save(); err != nil {
			DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "AddMarket to save config fail. Error: %s\n", err.Error())
			fmt.Println(err)
			go restartME()
			return "AddMarket Something wrong to save config file. ME would restart to restore from errors.", nil
		}

		/// to restart me
		go restartME()
	}

	return "Add market success.", nil
}

// -------------------------------------------------------------------------------
// Remove new market
type RemoveMarketParams struct {
	Sym *config.Symbol `json:"Sym"`
}

func (t *RemoveMarketParams) FromPositional(params []interface{}) error {
	return nil
}

func RemoveMarket(params json.RawMessage) (interface{}, *jrpc2.ErrorObject) {
	p := new(RemoveMarketParams)

	if err := jrpc2.ParseParams(params, p); err != nil {
		return nil, err
	}

	if p.Sym == nil {
		return nil, &jrpc2.ErrorObject{
			Code:    jrpc2.InvalidParamsCode,
			Message: jrpc2.InvalidParamsMsg,
			Data:    "exactly RemoveMarketParams{Sym[BaseCoin, QuoteCoin]} are required",
		}
	}

	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
			go restartME()
		}
	}()

	/// to operate
	err, up := config.RemoveMarketMap(*p.Sym)
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "RemoveMarket to RemoveMarketMap fail! Error=%s\n", err.Error())
		return "RemoveMarket to RemoveMarketMap fail.", nil
	}

	/// active market
	if !up {
		return "Params wrong, nothing was affected.", nil
	} else {
		/// update conifg
		if err := config.Save(); err != nil {
			DebugPrintf(MODULE_NAME, LOG_LEVEL_FATAL, "RemoveMarket to save config fail. Error: %s\n", err.Error())
			fmt.Println(err)
			go restartME()
			return "RemoveMarket Something wrong to save config file. ME would restart to restore from errors.", nil
		}

		/// to restart me
		go restartME()
	}

	return "Remove Market success.", nil
}

func restartME() {
	/// to restart me
	time.Sleep(1 * time.Second)
	Wait <- Signal_Restart
}

func Jrpc2Main() {
	// create a new server instance
	s := jrpc2.NewServer(config.GetMEConfig().JSONRPCIPPort, "/api/v1/rpc", nil)

	// register the add method
	s.Register("add", jrpc2.Method{Method: Add})
	s.Register("addcoin", jrpc2.Method{Method: AddCoin})
	s.Register("addmarket", jrpc2.Method{Method: AddMarket})
	s.Register("removecoin", jrpc2.Method{Method: RemoveCoin})
	s.Register("removemarket", jrpc2.Method{Method: RemoveMarket})

	// register the subtract method to proxy another rpc server
	// s.Register("add", jrpc2.Method{Url: "http://localhost:9999/api/v1/rpc"})

	// start the server instance
	s.Start()
}
