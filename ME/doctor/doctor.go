// doctor
package doctor

import (
	"fmt"
	"time"
)

const (
	MAX_GROWING_DURATION    = int64(120 * time.Minute)
	MAX_HEART_BEAT_DURATION = int64(10 * time.Minute)
)

type GrowingProgress int64

const (
	Progress_NotStart     GrowingProgress = 0
	Progress_BeginInit    GrowingProgress = 1
	Progress_CompleteInit GrowingProgress = 2
	Progress_Working      GrowingProgress = 3
)

func (t GrowingProgress) String() string {
	switch t {
	case Progress_NotStart:
		return "NotStart"
	case Progress_BeginInit:
		return "BeginInit"
	case Progress_CompleteInit:
		return "CompleteInit"
	case Progress_Working:
		return "Working"
	}
	return "GrowingProgress-UNSET"
}

type RunningType int64

const (
	RunningType_Enorder     RunningType = 1
	RunningType_CancelOrder RunningType = 2
	RunningType_Match       RunningType = 3
	RunningType_Outpool     RunningType = 4
)

func (t RunningType) String() string {
	switch t {
	case RunningType_Enorder:
		return "Enorder Heart-Beat"
	case RunningType_CancelOrder:
		return "CancelOrder Heart-Beat"
	case RunningType_Match:
		return "MatchCore Heart-Beat"
	case RunningType_Outpool:
		return "OutPool Heart-Beat"
	}
	return "RunningType-UNSET"
}

type BeatHeart struct {
	name          RunningType
	startBeatTime int64
	lastBeatTime  int64

	lastBeatCount   int64
	lastBeatRate    float64
	averageBeatRate float64
}

func newBeatHeart(type_ RunningType) *BeatHeart {
	curTime := time.Now().UnixNano()
	inst := new(BeatHeart)
	inst.name = type_
	inst.startBeatTime = curTime
	inst.lastBeatTime = curTime
	inst.lastBeatCount = 1
	inst.lastBeatRate = -1    /*1.0 / ((time.Now().UnixNano() - inst.lastBeatTime) / 1 * time.Second)*/
	inst.averageBeatRate = -1 /*inst.lastBeatCount / ((inst.lastBeatTime - inst.startBeatTime) / 1 * time.Second)*/
	return inst
}
func (t *BeatHeart) Beat() {
	curTime := time.Now().UnixNano()
	t.lastBeatTime = curTime
	t.lastBeatCount += 1
	t.lastBeatRate = 1 / (float64((time.Now().UnixNano() - t.lastBeatTime)) / float64(time.Second))
	t.averageBeatRate = float64(t.lastBeatCount) / (float64(t.lastBeatTime-t.startBeatTime) / float64(time.Second))
}
func (t *BeatHeart) Dump() string {
	strBuff := fmt.Sprintf("===================[ME %s Work Beat Heart Info]===================\n", t.name)
	strBuff += fmt.Sprintf(" == Start Beat Time:\t%d\n", t.startBeatTime)
	strBuff += fmt.Sprintf(" == Last Beat Count:\t%d\n", t.lastBeatCount)
	strBuff += fmt.Sprintf(" == Last Beat Time :\t%d\n", t.lastBeatTime)
	strBuff += fmt.Sprintf(" == Last Beat Rate :\t%f\n", t.lastBeatRate)
	strBuff += fmt.Sprintf(" == Aver Beat Rate :\t%f\n", t.averageBeatRate)
	strBuff += fmt.Sprintf("--------------------------------------------------------------------------------\n")

	//fmt.Print(strBuff)
	return strBuff
}

type Running struct {
	enorderBeatHeart     *BeatHeart
	cancelOrderBeatHeart *BeatHeart
	matchCoreBeatHeat    *BeatHeart
	outpoolBeatHeat      *BeatHeart
}

func newRunning() *Running {
	o := new(Running)
	o.initRunning()
	return o
}

func (t *Running) initRunning() {
	t.enorderBeatHeart = newBeatHeart(RunningType_Enorder)
	t.cancelOrderBeatHeart = newBeatHeart(RunningType_CancelOrder)
	t.matchCoreBeatHeat = newBeatHeart(RunningType_Match)
	t.outpoolBeatHeat = newBeatHeart(RunningType_Outpool)
}
func (t *Running) recordBeatHeart(type_ RunningType) {
	switch type_ {
	case RunningType_Enorder:
		t.enorderBeatHeart.Beat()
	case RunningType_CancelOrder:
		t.cancelOrderBeatHeart.Beat()
	case RunningType_Match:
		t.matchCoreBeatHeat.Beat()
	case RunningType_Outpool:
		t.outpoolBeatHeat.Beat()
	}
}
func (t *Running) getBeatHeart(type_ RunningType) *BeatHeart {
	switch type_ {
	case RunningType_Enorder:
		return t.enorderBeatHeart
	case RunningType_CancelOrder:
		return t.cancelOrderBeatHeart
	case RunningType_Match:
		return t.matchCoreBeatHeat
	case RunningType_Outpool:
		return t.outpoolBeatHeat
	}
	return nil
}
func (t *Running) dumpBeatHeart(type_ RunningType) string {
	switch type_ {
	case RunningType_Enorder:
		return t.enorderBeatHeart.Dump()
	case RunningType_CancelOrder:
		return t.cancelOrderBeatHeart.Dump()
	case RunningType_Match:
		return t.matchCoreBeatHeat.Dump()
	case RunningType_Outpool:
		return t.outpoolBeatHeat.Dump()
	}
	return ""
}
func (t *Running) dumpAllBeatHeart() string {
	strBuff := t.enorderBeatHeart.Dump()
	strBuff += t.cancelOrderBeatHeart.Dump()
	strBuff += t.matchCoreBeatHeat.Dump()
	strBuff += t.outpoolBeatHeat.Dump()
	return strBuff
}

type Diagnose interface {
	RecordBeatHeart(type_ RunningType)
	GetBeatHeart(type_ RunningType) *BeatHeart
	DumpBeatHeart(type_ RunningType)
	DumpAllBeatHeart()
	SetProgress(progress GrowingProgress)
	GetProgress() GrowingProgress

	IsNormal() bool

	isLaunchFault() bool
	isEnorderFault() bool
	isCancelOrderFault() bool
	isMatchCoreFault() bool
	isOutpoolFault() bool
}

type FaultType int64

const (
	FaultType_Normal          FaultType = 1
	FaultType_Dead            FaultType = -1
	FaultType_LaunchFault     FaultType = -2
	FaultType_EnorderFault    FaultType = -3
	FaultType_CancelOderFault FaultType = -4
	FaultType_MatchCoreFault  FaultType = -5
	FaultType_OutpoolFault    FaultType = -6
)

func (t FaultType) String() string {
	switch t {
	case FaultType_Normal:
		return "Work Normally"
	case FaultType_Dead:
		return "Dead"
	case FaultType_LaunchFault:
		return "LaunchFault"
	case FaultType_EnorderFault:
		return "EnorderFault"
	case FaultType_CancelOderFault:
		return "CancelOderFault"
	case FaultType_MatchCoreFault:
		return "MatchCoreFault"
	case FaultType_OutpoolFault:
		return "OutpoolFault"
	}
	return "FaultType-UNSET"
}

type Doctor struct {
	progress GrowingProgress
	running  *Running
}

func NewDoctor() *Doctor {
	o := new(Doctor)
	o.progress = Progress_NotStart
	o.running = newRunning()
	return o
}

func (t *Doctor) SetProgress(progress GrowingProgress) {
	t.progress = progress
}

func (t *Doctor) GetProgress() GrowingProgress {
	return t.progress
}

func (t *Doctor) RecordBeatHeart(type_ RunningType) {
	t.running.recordBeatHeart(type_)
}

func (t *Doctor) GetBeatHeart(type_ RunningType) *BeatHeart {
	return t.running.getBeatHeart(type_)
}

func (t *Doctor) DumpBeatHeart(type_ RunningType) string {
	return t.running.dumpBeatHeart(type_)
}

func (t *Doctor) DumpAllBeatHeart() string {
	return t.running.dumpAllBeatHeart()
}

func (t *Doctor) IsLaunchFault() bool {
	beat := t.running.getBeatHeart(RunningType_Enorder)
	if t.GetProgress() < Progress_Working && (time.Now().UnixNano()-beat.startBeatTime) > MAX_GROWING_DURATION {
		return true
	} else {
		return false
	}
}
func (t *Doctor) IsEnorderFault() bool {
	beat := t.running.getBeatHeart(RunningType_Enorder)
	if (time.Now().UnixNano() - beat.lastBeatTime) > MAX_HEART_BEAT_DURATION {
		return true
	} else {
		return false
	}
}
func (t *Doctor) IsCancelOrderFault() bool {
	beat := t.running.getBeatHeart(RunningType_CancelOrder)
	if (time.Now().UnixNano() - beat.lastBeatTime) > MAX_HEART_BEAT_DURATION {
		return true
	} else {
		return false
	}
}
func (t *Doctor) IsMatchCoreFault() bool {
	beat := t.running.getBeatHeart(RunningType_Match)
	if (time.Now().UnixNano() - beat.lastBeatTime) > MAX_HEART_BEAT_DURATION {
		return true
	} else {
		return false
	}
}
func (t *Doctor) IsOutpoolFault() bool {
	beat := t.running.getBeatHeart(RunningType_Outpool)
	if (time.Now().UnixNano() - beat.lastBeatTime) > MAX_HEART_BEAT_DURATION {
		return true
	} else {
		return false
	}
}
func (t *Doctor) IsNormal() bool {
	if t.IsEnorderFault() || t.IsCancelOrderFault() || t.IsMatchCoreFault() || t.IsOutpoolFault() || t.IsLaunchFault() {
		return false
	} else {
		return true
	}
}

type FaultChekList struct {
	isLaunchFalult     bool
	isEnorderFault     bool
	isCancelOrderFault bool
	isMatchCoreFault   bool
	isOutpoolFault     bool
}

func (t *Doctor) FaultReport() (report FaultChekList) {
	report.isLaunchFalult = t.IsLaunchFault()
	report.isEnorderFault = t.IsEnorderFault()
	report.isCancelOrderFault = t.IsCancelOrderFault()
	report.isMatchCoreFault = t.IsMatchCoreFault()
	report.isOutpoolFault = t.IsOutpoolFault()
	return report
}

func init() {

}
