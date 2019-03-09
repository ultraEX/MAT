// servers
package server

import (
	"fmt"

	"../markets"
)

// -------------------------------------------------------------------------------
var (
	marketsRef *markets.Markets = nil
)

func ServerMain(marketsInfo *markets.Markets) {
	if marketsInfo == nil {
		panic(fmt.Errorf("ServerMain cannot initialize as the input markets not been initialized first."))
	}
	marketsRef = marketsInfo

	fmt.Println("Start Cli-RPC Server...")
	go CliRpcMain()

	fmt.Println("Start HTTP Json RPC Server...")
	go Jrpc2Main()

	fmt.Println("Start HTTP RESTful Server...")
	go RESTfulMain()
}
