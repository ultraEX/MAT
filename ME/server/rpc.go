// rpc
package server

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"../cmd"
	"../config"
	. "../itf"
)

const (
	MODULE_NAME string = "[ME RPC]: "
)

type Args struct {
	Command                                        string
	Param1, Param2, Param3, Param4, Param5, Param6 string
}

type Reply struct {
	Result string
}

type FaultyReply struct {
	IsFaulty bool
	Info     string
}

type PIDReply struct {
	PID int
}

type Handler struct {
}

func (t *Handler) CommandProc(args *Args, reply *Reply) error {
	switch args.Command {
	case "version":
		return t.version(args, reply)

	case "use":
		return t.use(args, reply)

	case "setlog":
		return t.setlog(args, reply)

	case "setlevel":
		return t.setlevel(args, reply)

	case "statics":
		return t.statics(args, reply)

	case "markets":
		return t.markets(args, reply)

	case "dump":
		return t.dump(args, reply)

	case "dumpcm":
		return t.dumpcm(args, reply)

	case "dumpch":
		return t.dumpch(args, reply)

	case "dumpticker":
		return t.dumpticker(args, reply)

	case "dumptrades":
		return t.dumptrades(args, reply)

	case "constructtickersfromhistorytrades":
		return t.constructtickersfromhistorytrades(args, reply)

	case "constructtickersfromhistorytradeswithfilter":
		return t.constructtickersfromhistorytradeswithfilter(args, reply)

	case "constructtickersfromhistorytradeswithuninitialized":
		return t.constructtickersfromhistorytradeswithuninitialized(args, reply)

	case "constructtickerfromhistorytrades":
		return t.constructtickerfromhistorytrades(args, reply)

	case "beatheart":
		return t.beatheart(args, reply)

	case "faulty":
		return t.faulty(args, reply)

	case "exitme":
		return t.exitme(args, reply)

	case "restartme":
		return t.restartme(args, reply)

	}
	return nil
}

func (t *Handler) GetFaulty(args *Args, reply *FaultyReply) error {
	switch args.Command {
	case "faulty":
		return t.getFaulty(args, reply)

	}
	return nil
}

func (t *Handler) GetPID(args *Args, reply *PIDReply) error {
	switch args.Command {
	case "getPID":
		reply.PID = os.Getpid()
		return nil
	}
	return nil
}

func (t *Handler) version(args *Args, reply *Reply) error {
	if args.Command == "version" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		strBuff := fmt.Sprintf(`
==============[Current ME Core Version Info]==============
Version: %s
==========================================================
`,
			VERSION_NO,
		)
		reply.Result += strBuff
	}
	return nil
}

func (t *Handler) use(args *Args, reply *Reply) error {
	if args.Command == "use" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		_, marketType, ok := cmd.Marks.MarketCheck(args.Param1, args.Param2)
		if ok {
			cmd.Use(args.Param1, marketType)
			reply.Result += fmt.Sprintf("Now using %s Match Engine for check.\n", args.Param1)
		} else {
			reply.Result += fmt.Sprintf("Invalid symbol or not a valid symbol in the ME.\n")
			reply.Result += cmd.Marks.Dump()
		}
	}
	return nil
}

func (t *Handler) setlog(args *Args, reply *Reply) error {
	if args.Command == "setlog" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		if args.Param1 == "true" {
			SetLog2File(true)
			reply.Result += "Setlog to file: true\n"
		} else if args.Param1 == "false" {
			reply.Result += "Setlog to file: false\n"
			SetLog2File(false)
		} else {
			SwitchLog()
			if LOG_TO_FILE {
				reply.Result += "Setlog to file: true\n"
			} else {
				reply.Result += "Setlog to file: false\n"
			}
		}
	}
	return nil
}

func (t *Handler) setlevel(args *Args, reply *Reply) error {
	if args.Command == "setlevel" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		if args.Param1 != "" {
			SetLevel(args.Param1)
		} else {
			SwitchLevel()
		}
		reply.Result += "Setlevel to " + LOG_LEVEL.String() + "\n"
	}
	return nil
}

func (t *Handler) statics(args *Args, reply *Reply) error {
	if args.Command == "statics" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
		if matchEng != nil {
			tp := matchEng.GetTradePool(cmd.MktType)
			if tp != nil {
				reply.Result = tp.Statics()
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}
		} else {
			reply.Result = cmd.MarketEngineNilWarningPrint()
		}
	}
	return nil
}

func (t *Handler) markets(args *Args, reply *Reply) error {
	if args.Command == "markets" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		reply.Result = cmd.Marks.TradeStaticsDump()
		reply.Result += cmd.PrintSysPortUsage()
	}
	return nil
}

func (t *Handler) dump(args *Args, reply *Reply) error {
	if args.Command == "dump" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		if args.Param1 == "true" {
			matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(cmd.MktType)
				if tp != nil {
					reply.Result = tp.DumpTradePool(true)
				} else {
					reply.Result = cmd.MarketEngineNilWarningPrint()
				}
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}

		} else if args.Param1 == "false" {
			matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(cmd.MktType)
				if tp != nil {
					reply.Result = tp.DumpTradePool(false)
				} else {
					reply.Result = cmd.MarketEngineNilWarningPrint()
				}
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}

		} else {
			matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
			if matchEng != nil {
				tp := matchEng.GetTradePool(cmd.MktType)
				if tp != nil {
					reply.Result = tp.DumpTradePool(false)
				} else {
					reply.Result = cmd.MarketEngineNilWarningPrint()
				}
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}

		}
	}
	return nil
}

func (t *Handler) dumpcm(args *Args, reply *Reply) error {
	if args.Command == "dumpcm" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
		if matchEng != nil {
			tp := matchEng.GetTradePool(cmd.MktType)
			if tp != nil {
				reply.Result = tp.DumpCM()
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}
		} else {
			reply.Result = cmd.MarketEngineNilWarningPrint()
		}

	}
	return nil
}

func (t *Handler) dumpch(args *Args, reply *Reply) error {
	if args.Command == "dumpch" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
		if matchEng != nil {
			tp := matchEng.GetTradePool(cmd.MktType)
			if tp != nil {
				reply.Result = tp.DumpChannel()
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}
		} else {
			reply.Result = cmd.MarketEngineNilWarningPrint()
		}

	}
	return nil
}

func (t *Handler) dumpticker(args *Args, reply *Reply) error {

	if args.Command == "dumpticker" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
		if matchEng != nil {
			tp := matchEng.GetTickersEngine()
			reply.Result = tp.DumpAllToBuff()
		} else {
			reply.Result = cmd.MarketEngineNilWarningPrint()
		}

	}
	return nil
}

func (t *Handler) dumptrades(args *Args, reply *Reply) error {

	if args.Command == "dumtrades" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
		if matchEng != nil {
			tp := matchEng.GetTickersEngine()
			reply.Result = tp.DumpLatestTradeBuff()
		} else {
			reply.Result = cmd.MarketEngineNilWarningPrint()
		}

	}
	return nil
}

func (t *Handler) constructtickersfromhistorytrades(args *Args, reply *Reply) error {

	if args.Command == "constructtickersfromhistorytrades" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		timeFrom, err := strconv.ParseInt(args.Param1, 10, 64)
		if err != nil || args.Param2 != config.RECONSTRUCT_TICKERS_PASSWORD {
			reply.Result = "constructtickersfromhistorytrades command cannot execute."
			return nil
		}
		go cmd.Marks.ConstructTickers(timeFrom)
		reply.Result = "constructtickersfromhistorytrades command deliver ok."
	}
	return nil
}

func (t *Handler) constructtickersfromhistorytradeswithfilter(args *Args, reply *Reply) error {

	if args.Command == "constructtickersfromhistorytradeswithfilter" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		timeFrom, err := strconv.ParseInt(args.Param1, 10, 64)
		if err != nil || args.Param3 != config.RECONSTRUCT_TICKERS_PASSWORD {
			reply.Result = "constructtickersfromhistorytradeswithfilter command cannot execute."
			return nil
		}
		go cmd.Marks.ConstructTickersWithFilter(timeFrom, args.Param2)
		reply.Result = "constructtickersfromhistorytradeswithfilter command deliver ok."
	}
	return nil
}

func (t *Handler) constructtickersfromhistorytradeswithuninitialized(args *Args, reply *Reply) error {

	if args.Command == "constructtickersfromhistorytradeswithuninitialized" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		timeFrom, err := strconv.ParseInt(args.Param1, 10, 64)
		if err != nil || args.Param2 != config.RECONSTRUCT_TICKERS_PASSWORD {
			reply.Result = "constructtickersfromhistorytradeswithuninitialized command cannot execute."
			return nil
		}
		go cmd.Marks.ConstructUnInitializedTickers(timeFrom)
		reply.Result = "constructtickersfromhistorytradeswithuninitialized command deliver ok."
	}
	return nil
}

func (t *Handler) constructtickerfromhistorytrades(args *Args, reply *Reply) error {

	if args.Command == "constructtickerfromhistorytrades" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		timeFrom, err := strconv.ParseInt(args.Param2, 10, 64)
		if err != nil || args.Param3 != config.RECONSTRUCT_TICKERS_PASSWORD {
			reply.Result = "constructtickerfromhistorytrades command cannot execute."
			return nil
		}
		go cmd.Marks.ConstructTicker(args.Param1, timeFrom)
		reply.Result = "constructtickerfromhistorytrades command deliver ok."
	}
	return nil
}

func (t *Handler) beatheart(args *Args, reply *Reply) error {
	if args.Command == "beatheart" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
		if matchEng != nil {
			tp := matchEng.GetTradePool(cmd.MktType)
			if tp != nil {
				reply.Result = tp.DumpBeatHeart()
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}
		} else {
			reply.Result = cmd.MarketEngineNilWarningPrint()
		}

	}
	return nil
}

func (t *Handler) faulty(args *Args, reply *Reply) error {
	if args.Command == "faulty" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		matchEng, _ := cmd.Marks.GetMatchEngine(cmd.Symbol)
		if matchEng != nil {
			tp := matchEng.GetTradePool(cmd.MktType)
			if tp != nil {
				reply.Result = strconv.FormatBool(tp.IsFaulty())
			} else {
				reply.Result = cmd.MarketEngineNilWarningPrint()
			}
		} else {
			reply.Result = cmd.MarketEngineNilWarningPrint()
		}

	}
	return nil
}

func (t *Handler) getFaulty(args *Args, reply *FaultyReply) error {
	if args.Command == "faulty" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		reply.IsFaulty = cmd.Marks.IsFaulty()
		reply.Info = "success"

	}
	return nil
}

func (t *Handler) exitme(args *Args, reply *Reply) error {
	if args.Command == "exitme" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		go func() {
			time.Sleep(100 * time.Millisecond)
			Wait <- Signal_Quit
		}()
		reply.Result += "exit command process complete, ME have closed.\n"
	}
	return nil
}

func (t *Handler) restartme(args *Args, reply *Reply) error {
	if args.Command == "restartme" {
		fmt.Printf("Get client command: %s [%s, %s, %s, %s, %s, %s]\n",
			args.Command, args.Param1, args.Param2, args.Param3, args.Param4, args.Param5, args.Param6)

		reply.Result += WORK_DIR + "/" + GetExeName() + " "
		for _, osArg := range os.Args[1:] {
			reply.Result += (osArg + " ")
		}
		go func() {
			time.Sleep(1 * time.Second)
			Wait <- Signal_Restart
		}()
	}
	return nil
}

func CliRpcMain() {

	handler := new(Handler)
	rpc.Register(handler)

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.GetMEConfig().ManagerRPCIPPort) ///Default:"localhost:1937"
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
