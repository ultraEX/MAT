// cli
package main

import (
	"bytes"
	"fmt"

	//"log"
	"flag"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"

	"../ME/config"
	. "../ME/itf"
	"github.com/VividCortex/godaemon"
)

const (
	MODULE_NAME string = "[StartME]: "
)

var (
	SERVER_PORT string = config.GetMEConfig().ManagerRPCIPPort
)

const (
	RECONNECT_RETRY_MAXTIMES   int           = 20
	RECONNECT_WAIT_TIME_SECOND time.Duration = 30

	START_ERROR_RETRY_TIME_SECOND  int = 60
	CYCLECHECK_INTERVALTIME_SECOND int = 60
	CHECKFAULT_ACTION_TIMES        int = 3
)

var (
	EXE_NAME     string   = ""
	DAEMON_NAME  string   = "ME"
	DAEMON_PARAM []string = []string{"-daemon"}

	DAEMON_PID      int  = -1
	CMD_EXCUTE_BUSY bool = false
)

////--------------------------------------------------------------------------------
type StartMeSignal int

const (
	StartMeSignal_DaemonQuit   StartMeSignal = -1 //// Daemon had exit
	StartMeSignal_Reconncect   StartMeSignal = -2 //// Cannot connect the daemon
	StartMeSignal_DaemonFaulty StartMeSignal = -3 //// Daemon report faulty
)

func (t StartMeSignal) String() string {
	switch t {
	case StartMeSignal_DaemonQuit:
		return "Daemon Quit"

	case StartMeSignal_Reconncect:
		return "Reconnect Daemon"

	case StartMeSignal_DaemonFaulty:
		return "Daemon In Faulty"

	}
	return "<StartMeSignal-UNSET>"
}

var waitStartMeSignal chan StartMeSignal

////--------------------------------------------------------------------------------

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

func envInit() {
	WORK_DIR = GetWorkDir()
	EXE_NAME = GetExeName()
	DAEMON_NAME = "ME"
	DAEMON_PARAM = []string{"-daemon"}
}

func envClear() {
	strOut := Rememory()
	fmt.Print(strOut)
	DebugPrintf(COMMON_MODULE_NAME, LOG_LEVEL_FATAL, "%s", strOut)
}

func startME() {
	out := bytes.Buffer{}
	exePath := WORK_DIR + "/" + DAEMON_NAME
	cmd := exec.Command(exePath, DAEMON_PARAM...)
	cmd.Stdout = &out
	count := 0

_ReExcute:
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Begin to start %s with params: %s\n", exePath, strings.Join(DAEMON_PARAM, " "))
	err := cmd.Start()
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Fatal Error: StartME  execute fail(times:%d), err:\n", count, err)
		time.Sleep(time.Duration(START_ERROR_RETRY_TIME_SECOND) * time.Second)
		goto _ReExcute

	}

	/// ME Core work as daemon mode and fg program would exit at once, so here cannot block and check whether ME core is running.
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Start %s and wait for %ds to let it work stablely...\n", EXE_NAME, RECONNECT_WAIT_TIME_SECOND)
	time.Sleep(RECONNECT_WAIT_TIME_SECOND * time.Second)
	fmt.Println(out.String())
}

func exitME(client *rpc.Client) error {
	var (
		param1, param2, param3, param4, param5, param6 string
	)

	command := "exitme"
	reply := Reply{}
	args := Args{Command: command, Param1: param1, Param2: param2, Param3: param3, Param4: param4, Param5: param5, Param6: param6}
	err := client.Call("Handler.CommandProc", args, &reply)
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "command(%s) execute fail, err: %s\n", command, err.Error())
		return err
	}
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Command(%s) execute and get result:\n%t\n", command, reply.Result)
	return nil
}

func getPID(client *rpc.Client) (pid int, err error) {
	var (
		param1, param2, param3, param4, param5, param6 string
	)

	command := "getPID"
	reply := PIDReply{-1}
	args := Args{Command: command, Param1: param1, Param2: param2, Param3: param3, Param4: param4, Param5: param5, Param6: param6}
	err = client.Call("Handler.GetPID", args, &reply)
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "command(%s) execute fail, err: %s\n", command, err.Error())
		return -1, err
	}
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Command(%s) execute and get result:\nPID = %d\n", command, reply.PID)
	return reply.PID, nil
}

func getFaulty(client *rpc.Client) (isFaulty bool, err error) {
	var (
		param1, param2, param3, param4, param5, param6 string
	)

	command := "faulty"
	reply := FaultyReply{false, "ok"}
	args := Args{Command: command, Param1: param1, Param2: param2, Param3: param3, Param4: param4, Param5: param5, Param6: param6}
	err = client.Call("Handler.GetFaulty", args, &reply)
	if err != nil {
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "command(%s) execute fail, err: %s\n", command, err.Error())
		return true, err
	}
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Command(%s) execute and get result %s:\nIsFaulty = %t\n", command, reply.Info, reply.IsFaulty)
	return reply.IsFaulty, nil
}

func killME() error {
	if DAEMON_PID == -1 {
		return fmt.Errorf("DAEMON_PID not initialized.")
	}

	process, err := os.FindProcess(DAEMON_PID)
	if err != nil {
		return fmt.Errorf("killME FindProcess(%d) fail.", DAEMON_PID)
	}
	process.Kill()
	DAEMON_PID = -1

	return nil
}

func CycleCheck() {
	var (
		count int = 0
	)

reConnectME:
	DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Connect to ME Core TCP Server(port:%s)\n", SERVER_PORT)
	client, err := rpc.Dial("tcp", SERVER_PORT)
	if err != nil {
		if !CMD_EXCUTE_BUSY {
			count++
			if count >= RECONNECT_RETRY_MAXTIMES {
				waitStartMeSignal <- StartMeSignal_DaemonFaulty
				count = 0
			}
		}

		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "%s Dial ME Core server fail, err:%s\nWait for %s second to retry...\n", EXE_NAME, err.Error(), RECONNECT_WAIT_TIME_SECOND)
		time.Sleep(RECONNECT_WAIT_TIME_SECOND * time.Second)
		goto reConnectME
	}

	/// Get ME's PID for kill signal ready
	pid, err := getPID(client)
	if err != nil {
		if !CMD_EXCUTE_BUSY {
			waitStartMeSignal <- StartMeSignal_Reconncect
		}

		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "CycleCheck getPID fail, err: %s\nWaiting %ds to reconnect to ME daemon\n", err.Error(), RECONNECT_WAIT_TIME_SECOND)
		client.Close()
		time.Sleep(RECONNECT_WAIT_TIME_SECOND * time.Second)
		goto reConnectME
	}
	DAEMON_PID = pid

	/// Check the ME Core Heart Beat
	isFaulty := false
	checkTimes := 0
	for {
		time.Sleep(time.Duration(CYCLECHECK_INTERVALTIME_SECOND) * time.Second)

		isFaulty, err = getFaulty(client)
		if err != nil && !CMD_EXCUTE_BUSY {
			if checkTimes >= CHECKFAULT_ACTION_TIMES {
				checkTimes = 0
				waitStartMeSignal <- StartMeSignal_DaemonQuit
				DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "CycleCheck getFaulty fail, err: %s\nTo reconnect to ME daemon\n", err.Error())
				break
			} else {
				checkTimes++
				continue
			}
		}

		if isFaulty && !CMD_EXCUTE_BUSY {
			if checkTimes >= CHECKFAULT_ACTION_TIMES {
				checkTimes = 0
				err = exitME(client)
				if err != nil {
					waitStartMeSignal <- StartMeSignal_Reconncect
				} else {
					waitStartMeSignal <- StartMeSignal_DaemonFaulty
				}
				break
			} else {
				checkTimes++
				continue
			}
		}

		if !isFaulty {
			checkTimes = 0
		}
	}

	/// Can not connect to ME Core
	if isFaulty {
		client.Close()
		goto reConnectME
	} else {
		panic(fmt.Errorf("CycleCheck Logic error!"))
	}
}

func daemonMode() {
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
}

func main() {
	fmt.Print(`
=====================Match Engine-Manager=====================
 The manager is a daemon, it manage ME daemon for running forever.
 When ME not run, it take charge to start ME
 When ME fall in faulty, it take charge to process it
 When these complete, it continue to manage ME's runing
==============================================================	
`,
	)
	daemonMode()

	envInit()
	startME()
	go CycleCheck()

	/// Control the manager process
	waitStartMeSignal = make(chan StartMeSignal)
	for {
		signal := <-waitStartMeSignal
		DebugPrintf(MODULE_NAME, LOG_LEVEL_ALWAYS, "Received %s signal to act\n", signal)
		switch signal {
		case StartMeSignal_DaemonQuit:
			CMD_EXCUTE_BUSY = true
			envClear()
			startME()
			CMD_EXCUTE_BUSY = false
		case StartMeSignal_Reconncect:
			CMD_EXCUTE_BUSY = true
			err := killME()
			if err != nil {
				DebugPrintln(MODULE_NAME, LOG_LEVEL_ALWAYS, err)
			}
			startME()
			CMD_EXCUTE_BUSY = false
		case StartMeSignal_DaemonFaulty:
			CMD_EXCUTE_BUSY = true
			envClear()
			startME()
			CMD_EXCUTE_BUSY = false
		}
	}
}

func init() {
	fmt.Printf("%s Begin to run...\n", os.Args[0])

	LOG_TO_FILE = true
	TEST_DEBUG_DISPLAY = true

	LOG_FILE_NAME_DEBUG = "startME.log"

	if LOG_TO_FILE {
		LogFile = NewLogToFile(WORK_DIR + "/" + LOG_FILE_NAME_DEBUG)
		LogFile.Start()
	}

}
