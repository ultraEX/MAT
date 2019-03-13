// cli
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/rpc"
	"os/exec"
	"strings"
	"time"

	. "../ME/comm"
	"../ME/config"
)

var (
	SERVER_PORT string = config.GetMEConfig().ManagerRPCIPPort
)

type Args struct {
	Command                                        string
	Param1, Param2, Param3, Param4, Param5, Param6 string
}

type Reply struct {
	Result string
}

var commandList *HashSet

func initCommandList() {
	commandList = NewHashSet()
	commandList.Add("version")
	commandList.Add("use")
	commandList.Add("setlog")
	commandList.Add("setlevel")
	commandList.Add("statics")
	commandList.Add("markets")
	commandList.Add("dump")
	commandList.Add("dumpch")
	commandList.Add("constructtickersfromhistorytrades")
	commandList.Add("constructtickersfromhistorytradeswithfilter")
	commandList.Add("constructtickersfromhistorytradeswithuninitialized")
	commandList.Add("constructtickerfromhistorytrades")
	commandList.Add("dumpticker")
	commandList.Add("dumptrades")
	commandList.Add("beatheart")
	commandList.Add("faulty")
	commandList.Add("exitme")
	commandList.Add("restartme")
	commandList.Add("envclear")
}
func isValidCommand(cmd string) bool {
	return commandList.Contains(cmd)
}

func cmd(client *rpc.Client) (command string, reply Reply) {

	var (
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
		fmt.Printf("GetCommand: %s\n", command)

		/// Command filter:
		if !isValidCommand(command) {
			fmt.Printf("Command %s not a valid command\n", command)
			continue
		}

		/// Command local process:
		if command == "envclear" {
			strOut := Rememory()
			fmt.Print(strOut)
			DebugPrintf(COMMON_MODULE_NAME, LOG_LEVEL_FATAL, "%s", strOut)
			continue
		}

		/// Command remote process: Rpc call remote method and get action
		args := Args{Command: command, Param1: param1, Param2: param2, Param3: param3, Param4: param4, Param5: param5, Param6: param6}
		err := client.Call("Handler.CommandProc", args, &reply)
		if err != nil {
			log.Fatal("command execute fail:", err)
		}
		fmt.Printf("Get command result:\n%s\n", reply.Result)

		if command == "restartme" {
			break
		}
	}
	return command, reply
}

func cmd2(command string, reply Reply) string {
	switch command {
	case "restartme":
		restartME(reply.Result)
	}
	return command
}

func restartME(command string) {
	s := strings.Split(command, " ")
	if len(s) <= 0 {
		log.Fatal("restartME fail to parse command(%s) to execute...\n", command)
		return
	}

	var out bytes.Buffer
	cmd := exec.Command(s[0], s[1:]...)
	cmd.Stdout = &out
	wait := make(chan bool)
	defer close(wait)
	go func() {
		fmt.Printf("Wait 3 seconds to restart ME...\n")
		time.Sleep(3 * time.Second)
		fmt.Printf("Begin to restart %s with params: %s\n", s[0], strings.Join(s[1:], " "))
		//err := cmd.Run()
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		//		err = cmd.Wait()
		//		if err != nil {
		//			fmt.Printf("Command finished with error: %v", err)
		//		}
		fmt.Printf("Wait 2 seconds to reconnect Match Engine...\n")
		time.Sleep(2 * time.Second)
		fmt.Println(out.String())
		wait <- true
	}()
	_ = <-wait
}

func main() {
	fmt.Print(`
=====================Match Engine-Client=====================
Command format is: command [params...].
Command List:
	version
	use ETH/BTC
	setlog [true/false]
	statics
	markets
	dump [true/false]
	dumpch
	dumpticker
	dumptrades
	beatheart
	exitme
	restartme
	envclear
	constructtickersfromhistorytrades timefrom password
	constructtickersfromhistorytradeswithfilter timefrom filter timefrom
	constructtickerfromhistorytrades symbol timefrom password
	constructtickersfromhistorytradeswithuninitialized timefrom password
==============================================================	
`,
	)
	initCommandList()

reConnectME:
	fmt.Printf("Connect to ME TCP Server: Match Engine(port:%s)\n", SERVER_PORT)
	client, err := rpc.Dial("tcp", SERVER_PORT)
	if err != nil {
		log.Fatal("ME-Cli dial ME server fail:", err)
	}

	command := cmd2(cmd(client))
	if command == "restartme" {
		client.Close()
		goto reConnectME
	}
}
