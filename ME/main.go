// MatchEngine.go
package main

import (
	"flag"
	"fmt"
	"os"

	. "./comm"
	cmd "./command"
	"./markets"
	"./server"
	"./server/thrift_rpc"

	"github.com/VividCortex/godaemon"
)

const (
	MODULE_NAME string = "[Main]: "
)

////--------------------------------------------------------------------------------
func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\n")
	}
	isDaemon := flag.Bool("daemon", false, "Run use daemon mode, work at back bench")
	flag.Parse()
	if *isDaemon {
		attr := godaemon.DaemonAttr{}
		attr.ProgramName = WORK_DIR + "/" + GetExeName()
		godaemon.MakeDaemon(&attr)
	}

	markets := markets.CreateMarkets()
	fmt.Println("Start Thrift RPC Server...")
	go thrift_rpc.RPCMain(markets)

	fmt.Println("Start Command Server...")
	go cmd.Command(markets)

	fmt.Println("Start RPC Server...")
	server.ServerMain(markets)

	WaitingSignal()
}
