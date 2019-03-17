// cmd
package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	. "../comm"
	"../config"
	"../db/use_mysql"
	"../db/use_redis"
	"../markets"
)

var Marks *markets.Markets = nil
var Symbol string = "ETH/BTC"
var MktType config.MarketType = config.MarketType_MixHR

func Use(s string, marketType config.MarketType) {
	Symbol = s
	MktType = marketType
}

func PrintSysPortUsage() string {
	strCmd := `netstat -n | awk '/^tcp/ {++state[$NF]} END {for(key in state) print key,"\t",state[key]}'`
	cmd := exec.Command("/bin/bash")
	cmd.Stdin = strings.NewReader(strCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("cmd.Run(%s) fail.\n", strCmd)
		return ""
	}
	strOut := fmt.Sprintf("System port usage:\n%s\n", out.String())
	fmt.Print(strOut)
	return strOut
}

func MarketEngineNilWarningPrint() string {
	strBuff := fmt.Sprintf("Warning: current using %s-%s not exist.\n", Symbol, MktType.String())
	fmt.Print(strBuff)
	return strBuff
}
func Command(m *markets.Markets) {
	if m == nil {
		panic("cmd.Command input a nil pointer of m = *markets.Markets")
	}
	fmt.Print("========================Matching Debug Back End Work Start========================\n")
	fmt.Print(`
  ==  Command	list below is support to debug the engine
  ==  dump:		list all orders. eg. dump; 
  ==  dumpch:	list outpool channel usage info. 
  ==  dumpcm:	list order ids map channellist info. 
  ==  dumpticker:	list tickers info of current symbol. 
  ==  dumptrades:	dump latest trades (about 100 items)
  ==  beatheart:	list all heart beat info in the me
  ==  statics:	get engine match work statics info
  ==  markets:	get engines' match work statics info(all markets)
  ==  faulty:	get markets faulty infomation, if one of more market fault, markets would be faulty
  ==  setInfo:	set debug info auto display. eg. setInfo true
  ==  setlevel: set debug info level
  ==  setdbpass: set db through pass out using to test core throughput
  ==  resetinfo: reset debug staics info from now on
  ==  setlog:	control log switch
  ==  setnecolog:	control neco log switch
  ---------------------------------------------------------------------------------
  ==  constructtickersfromhistorytrades:		construct tickers list(MM&DS) and current ticker(MM) from history trades(from some date<nano second as unit> as parmameter) for all markets symbols
  ==  constructtickersfromhistorytradeswithfilter:	construct tickers list(MM&DS) and current ticker(MM) from history trades(from some date<nano second as unit> as parmameter) for all markets symbols(filtered(use comma to seperate))	
  ==  constructtickersfromhistorytradeswithuninitialized:	construct tickers list(MM&DS) and current ticker(MM) from history trades(from some date<nano second as unit> as parmameter) for all markets symbols(uninitialized tickers)
  ==  constructtickerfromhistorytrades:		construct ticker list(MM&DS) and current ticker(MM) from history trades(from some date<nano second as unit> as parmameter) for special symbol
	** Please make sure you do want to do it and would risk the venture(Iamsuretodoitpleaseaction!)
  ---------------------------------------------------------------------------------
  ==  test:		test engine core function
  ==  trade:		control core engine work status
  ---------------------------------------------------------------------------------
  ==  redis:		test redis interface function
  ==  mysql:		test mysql interface function
  ---------------------------------------------------------------------------------
  ==  use:		to specify the market, eg. ETH/BTC(default), LTC/BTC... and mix(default), human, robot
  ---------------------------------------------------------------------------------
  ==  stop/exit:	to quit program
  ==  Version/version:	show Match Engine version info
  ---------------------------------------------------------------------------------

  Want to get more detail info please use command: help or manual.
`,
	)
	strSym, _ := m.GetDefaultMatchEngine()
	Symbol = strSym.String()
	fmt.Printf("Use default Match Engine as %s.\n", Symbol)
	fmt.Print("========================**********************************========================\n")

	Marks = m
	var (
		command                                        string
		param1, param2, param3, param4, param5, param6 string
	)
	for {
		param1 = ""
		param2 = ""
		param3 = ""
		param4 = ""
		param5 = ""
		param6 = ""
		fmt.Scanln(&command, &param1, &param2, &param3, &param4, &param5, &param6)

		/// Parse input
		/// setInfo command:========================================
		if command == "setinfo" {
			fmt.Printf("GetCommand: %s %s\n", command, param1)
			if param1 == "true" {
				SetInfo(true)
			} else if param1 == "false" {
				SetInfo(false)
			} else {
				SwitchInfo()
			}
		}

		/// setnecolog command:========================================
		if command == "setnecolog" {
			fmt.Printf("GetCommand: %s %s\n", command, param1)
			if param1 == "true" {
				SetNecoLog(true)
			} else if param1 == "false" {
				SetNecoLog(false)
			} else {
				SwitchNecoLog()
			}
		}

		/// setlog command:========================================
		if command == "setlog" {
			fmt.Printf("GetCommand: %s %s\n", command, param1)
			if param1 == "true" {
				SetLog2File(true)
			} else if param1 == "false" {
				SetLog2File(false)
			} else {
				SwitchLog()
			}
		}

		/// setlevel command:========================================
		if command == "setlevel" {
			fmt.Printf("GetCommand: %s %s\n", command, param1)
			if param1 != "" {
				SetLevel(param1)
			} else {
				SwitchLevel()
			}
		}

		/// setdbpass command:========================================
		if command == "setdbpass" {
			fmt.Printf("GetCommand: %s %s\n", command, param1)
			if param1 == "true" {
				SetDbThroughpass(true)
			} else {
				SetDbThroughpass(false)
			}

		}

		/// dump command:========================================
		if command == "dump" {
			fmt.Printf("GetCommand: %s %s\n", command, param1)
			if param1 == "true" {
				matchEng, err := Marks.GetMatchEngine(Symbol)
				if matchEng != nil {
					tp := matchEng.GetTradePool(MktType)
					if tp != nil {
						tp.DumpTradePoolPrint(true)
					} else {
						MarketEngineNilWarningPrint()
					}
				} else {
					fmt.Println(err)
				}
			} else if param1 == "false" {
				matchEng, err := Marks.GetMatchEngine(Symbol)
				if matchEng != nil {
					tp := matchEng.GetTradePool(MktType)
					if tp != nil {
						tp.DumpTradePoolPrint(false)
					} else {
						MarketEngineNilWarningPrint()
					}
				} else {
					fmt.Println(err)
				}
			} else {
				matchEng, err := Marks.GetMatchEngine(Symbol)
				if matchEng != nil {
					tp := matchEng.GetTradePool(MktType)
					if tp != nil {
						tp.DumpTradePoolPrint(false)
					} else {
						MarketEngineNilWarningPrint()
					}
				} else {
					fmt.Println(err)
				}
			}
		}

		/// dumpch command:========================================
		if command == "dumpch" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(MktType)
				if tp != nil {
					tp.DumpChannel()
				} else {
					MarketEngineNilWarningPrint()
				}
			} else {
				fmt.Println(err)
			}
		}

		/// dumpcm command:========================================
		if command == "dumpcm" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(MktType)
				if tp != nil {
					tp.DumpChanlsMap()
				} else {
					MarketEngineNilWarningPrint()
				}
			} else {
				fmt.Println(err)
			}
		}

		/// dumpticker command:========================================
		if command == "dumpticker" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTickersEngine()
				tp.DumpAllToPrint()
			} else {
				fmt.Println(err)
			}
		}

		/// dumtrades command:========================================
		if command == "dumptrades" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTickersEngine()
				tp.DumpLatestTradePrint()
			} else {
				fmt.Println(err)
			}
		}

		/// beatheart command:========================================
		if command == "beatheart" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(MktType)
				if tp != nil {
					tp.DumpBeatHeart()
				} else {
					MarketEngineNilWarningPrint()
				}
			} else {
				fmt.Println(err)
			}
		}

		/// Statics command:========================================
		if command == "statics" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(MktType)
				if tp != nil {
					tp.Statics()
				} else {
					MarketEngineNilWarningPrint()
				}
			} else {
				fmt.Println(err)
			}
		}

		/// Markets command:========================================
		if command == "markets" {
			fmt.Printf("GetCommand: %s\n", command)
			fmt.Print(Marks.TradeStaticsDump())

			PrintSysPortUsage()
		}

		/// faulty command:========================================
		if command == "faulty" {
			fmt.Printf("GetCommand: %s\n", command)
			fmt.Println(Marks.IsFaulty())
		}

		/// resetinfo command:========================================
		if command == "resetinfo" {
			fmt.Printf("GetCommand: %s\n", command)

			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(MktType)
				if tp != nil {
					tp.RestartDebuginfo()
					tp.ResetMatchCorePerform()
					fmt.Printf("Debuginfo Reset at time: %s\n", GetDateTime())
				} else {
					MarketEngineNilWarningPrint()
				}
			} else {
				fmt.Println(err)
			}
		}

		/// constructtickersfromhistorytrades command:========================================
		if command == "constructtickersfromhistorytrades" {
			fmt.Printf("GetCommand: %s\n", command)
			from, err := strconv.ParseInt(param1, 10, 64)
			if err == nil && param2 == config.RECONSTRUCT_TICKERS_PASSWORD {
				Marks.ConstructTickers(from)
			}
		}

		/// constructtickersfromhistorytradeswithfilter command:========================================
		if command == "constructtickersfromhistorytradeswithfilter" {
			fmt.Printf("GetCommand: %s\n", command)
			from, err := strconv.ParseInt(param1, 10, 64)
			if err == nil && param3 == config.RECONSTRUCT_TICKERS_PASSWORD {
				Marks.ConstructTickersWithFilter(from, param2)
			}
		}

		/// constructtickersfromhistorytradeswithuninitialized command:========================================
		if command == "constructtickersfromhistorytradeswithuninitialized" {
			fmt.Printf("GetCommand: %s\n", command)
			from, err := strconv.ParseInt(param1, 10, 64)
			if err == nil && param2 == config.RECONSTRUCT_TICKERS_PASSWORD {
				Marks.ConstructUnInitializedTickers(from)
			}
		}

		/// constructtickerfromhistorytrades command:========================================
		if command == "constructtickerfromhistorytrades" {
			fmt.Printf("GetCommand: %s\n", command)
			from, err := strconv.ParseInt(param2, 10, 64)
			if err == nil && param3 == config.RECONSTRUCT_TICKERS_PASSWORD {
				Marks.ConstructTicker(param1, from)
			}
		}

		/// Test command:========================================
		if command == "test" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(MktType)
				if tp != nil {
					tp.Test(param1, param2, param3, param4, param5, param6)
				} else {
					MarketEngineNilWarningPrint()
				}
			} else {
				fmt.Println(err)
			}
		}

		/// trade control:========================================
		if command == "trade" || command == "td" {
			fmt.Printf("GetCommand: %s\n", command)
			matchEng, err := Marks.GetMatchEngine(Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(MktType)
				if tp != nil {
					tp.TradeCommand(param1, param2, param3)
				} else {
					MarketEngineNilWarningPrint()
				}
			} else {
				fmt.Println(err)
			}
		}

		/// redis test
		if command == "redis" {
			fmt.Printf("GetCommand: %s\n", command)
			use_redis.RedisDbInstance().TestRedis(param1, param2, param3, param4, param5, param6)
		}

		/// mysql test
		if command == "mysql" {
			fmt.Printf("GetCommand: %s\n", command)
			use_mysql.TestMysql(param1, param2, param3, param4, param5, param6)
		}

		if command == "stop" || command == "exit" {
			fmt.Printf("GetCommand: %s\n", command)
			break
		}

		/// version command
		if command == "Version" || command == "version" {
			fmt.Printf("GetCommand: %s\n", command)
			fmt.Printf("Current ME Version: %s\n", VERSION_NO)
		}

		/// use command
		if command == "use" {
			fmt.Printf("GetCommand: %s\n", command)
			if param1 == "" {
				fmt.Printf("Current using symbol=%s, marketType=%s\n", Symbol, MktType.String())
				matchEng, err := Marks.GetMatchEngine(Symbol)
				if matchEng != nil {
					if matchEng.GetTradePool(MktType) == nil {
						MarketEngineNilWarningPrint()
					}

				} else {
					fmt.Println(err)
				}
				continue
			}
			_, mktType, ok := Marks.MarketCheck(param1, param2)
			if ok {
				Use(param1, mktType)
				fmt.Printf("Now using %s Match Engine for check.\n", param1)
			} else {
				fmt.Printf("Invalid symbol(sym:%s, marketType:%s) or not a valid symbol in the ME.\n", param1, param2)
				Marks.Dump()
			}
		}

		if command == "help" || command == "manual" {
			fmt.Printf("GetCommand: %s", command)
			fmt.Println(`
=============================================================================
[Control Type]
	Debug Info Switch:
	setinfo [true/false]
	setlog	[true/false]
	setnecolog [true/false]
	setdbpass [true/false]
	resetinfo
	-------------------------------------------------------------------------
	Engine Work Status:
	trade stop
	trade pause
	trade resume
	-------------------------------------------------------------------------
	Market select:
	use symbole
	-------------------------------------------------------------------------
	Markets statics:
	markets
	-------------------------------------------------------------------------
	Pool dump:
	dump true/false
	dumpch
	dumpcm
	dumpticker
	dumptrades
	faulty
	beatheart
	-------------------------------------------------------------------------
	Exit:
	stop
	exit
	version
=============================================================================	
	constructtickersfromhistorytrades timefrom password
	constructtickerfromhistorytrades symbol timefrom password
	constructtickersfromhistorytradeswithfilter timefrom filter password 
	constructtickersfromhistorytradeswithuninitialized timefrom password 
=============================================================================	
[Moduler Type]
	[Match Engine]
		Cancel Routine:
		test cancel user id symbol
		----------------------------------------------------------------------
		Statics Info:
		statics
		----------------------------------------------------------------------
	[Redis]
		Redis Order Function:
		redis order add
		redis order rm user id symbol
		redis order get user id symbol
		redis order all
		redis order ones user symbol
		----------------------------------------------------------------------
		Redis Trade Function:
		redis trade add
		redis trade rm user id symbol
		redis trade get user id symbol
		redis trade all
		redis trade ones user symbol
		----------------------------------------------------------------------
		redis zset add key score mem
		redis zset rm key mem
		redis zset get key index
		redis zset gets key start stop
		redis zset all key
		----------------------------------------------------------------------
	[Mysql]
		Mysql Order Function:
		mysql order add
		mysql order update
		mysql order rm user id symbol
		mysql order rm2
		mysql order get
		mysql order all
		mysql order ones
		----------------------------------------------------------------------
		Mysql Trade Function:
		mysql trade add
		mysql trade add2
		mysql trade rm user id symbol
		mysql trade get
		mysql trade all
		mysql trade ones
		----------------------------------------------------------------------
		Mysql Fund Function:
		mysql fund get user
		mysql fund freeze user aorb(1/2) price volume
		mysql fund unfreeze user aorb(1/2) price volume
		mysql fund settletx user aorb(1/2) amount
		mysql fund settlequick user aorb(1/2) amount
		mysql fund settle biduser askuser amount

		Mysql Finance Function:
		mysql finance get user id symbol
		mysql finance add2 bidOrderID askOrderID

		Mysql Tickers Function:
		mysql tickers init symbol
		mysql tickers add symbol tickType(1~8) 
		mysql tickers get symbol tickType(1~8) 
		mysql tickers getlimit symbol tickType(1~8) size
=============================================================================`, "\n")
		}
	}

	fmt.Print(`
========================Matching Debug Back End Work End=====================
=======   ME will exit, Bye Bye!!!
`)
}
