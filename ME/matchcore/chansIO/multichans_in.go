package chansIO

import (
	"fmt"
	"sync/atomic"

	"../../comm"
)

const (
	MODULE_NAME_MULTICHANS_IN string = "[MultiChans_In]: "
	IN_MULTI_CHANS_SIZE       int    = 10
	IN_CHANNEL_SIZE           int    = 6
)

var (
	IN_MULTI_CHANS_SIZE_VAR int = 10
)

type ChanProcess_In func(int)
type MultiChans_In struct {
	chans      [IN_MULTI_CHANS_SIZE]*ChanUnit_In
	proc       ChanProcess_In
	chanUse    *ChannelUse
	idleChanNO int
}

func NewMultiChans_In(p ChanProcess_In) *MultiChans_In {
	o := new(MultiChans_In)
	for i := 0; i < IN_MULTI_CHANS_SIZE; i++ {
		o.chans[i] = NewChanUnit_In(i)
	}
	o.proc = p
	o.chanUse = NewChannelUse()
	/// prepare to work

	/// start work
	go func() { /// deal with chan nil problem
		for i := 0; i < IN_MULTI_CHANS_SIZE; i++ {
			go o.proc(i)
		}
	}()

	return o
}

func (t *MultiChans_In) InChannel(elem *InElem) {
	switch elem.Type_ {
	case InElemType_EnOrder:
		id := elem.Order.ID
		chSet := t.GetIdleChannel(id)
		elem.Count = int64(chSet.Len())
		for _, v := range chSet.Elements() {
			t.chanUse.InChan(id, v.(int))
			t.chans[v.(int)].In(elem)
		}
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_IN, comm.LOG_LEVEL_TRACK, "MultiChans_In InElemType_EnOrder InChannel(%v).\n", chSet)

	case InElemType_CancelID:
		id := elem.CancelId
		chSet := t.GetIdleChannel(id)
		elem.Count = int64(chSet.Len())
		for _, v := range chSet.Elements() {
			t.chanUse.InChan(id, v.(int))
			t.chans[v.(int)].In(elem)
		}
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_IN, comm.LOG_LEVEL_TRACK, "MultiChans_In InElemType_CancelID InChannel(%v).\n", chSet)
	}
}

func (t *MultiChans_In) OutChannel(chNO int) (*InElem, bool) {
	elem := t.chans[chNO].Out()
	count := atomic.AddInt64(&elem.Count, -1)

	// chanuse manage
	switch elem.Type_ {
	case InElemType_EnOrder:
		t.chanUse.OutChan(elem.Order.ID, chNO)
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_IN, comm.LOG_LEVEL_TRACK,
			"MultiChans_In InElemType_EnOrder OutChannel(%d): OutChan(ask(id=%d),bid(id=%d), chanNO=%d.\n",
			chNO, elem.Order.ID, elem.Order.ID, chNO)
	case InElemType_CancelID:
		t.chanUse.OutChan(elem.CancelId, chNO)
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_IN, comm.LOG_LEVEL_TRACK,
			"MultiChans_In OUTPOOL_CANCELORDER OutChannel(%d): OutChan(cancel order(id=%d), chanNO=%d.\n",
			chNO, elem.CancelId, chNO)
	}

	if count <= 0 {
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_IN, comm.LOG_LEVEL_TRACK, "MultiChans_In OutChannel(%d): %v.\n", chNO, elem)
		return elem, true
	} else {
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_IN, comm.LOG_LEVEL_TRACK, "MultiChans_In OutChannel(%d) nil.\n", chNO)
		return nil, false
	}

}

func (t *MultiChans_In) Dump() {
	fmt.Printf("==================[MultiChans_In Dump Detail]==================\n")
	for k, v := range t.chans {
		fmt.Printf("Chan[%d]: cap = %d, len = %d\n", k, v.Cap(), v.Len())
	}
	fmt.Printf("idleChanNO: %d\n", t.idleChanNO)
	t.chanUse.Dump(true)
	fmt.Printf("=============================================================\n")
}

func (t *MultiChans_In) Summary() {
	fmt.Printf("==================[MultiChans_In Dump Summary]==================\n")
	for k, v := range t.chans {
		fmt.Printf("Chan[%d]: cap = %d, len = %d\n", k, v.Cap(), v.Len())
	}
	fmt.Printf("idleChanNO: %d\n", t.idleChanNO)
	t.chanUse.Dump(false)
	fmt.Printf("=============================================================\n")
}

func (t *MultiChans_In) Len() int {
	return IN_MULTI_CHANS_SIZE
}

func (t *MultiChans_In) ChanCap() int {
	return IN_CHANNEL_SIZE
}

func (t *MultiChans_In) getANewChan() int {
	idleno := t.idleChanNO
	for i := 0; i < IN_MULTI_CHANS_SIZE; i++ {
		no := t.idleChanNO + i
		if no >= IN_MULTI_CHANS_SIZE {
			no = 0
		}
		if !t.chans[no].IsBusy() {
			idleno = no
			break
		}

	}

	/// update idleChanNO
	t.idleChanNO++
	if t.idleChanNO >= IN_MULTI_CHANS_SIZE {
		t.idleChanNO = 0
	}
	return idleno
}

func (t *MultiChans_In) GetIdleChannel(id int64) comm.Set {
	/// if a secondary commer, use the original channel to ensure serialize
	chans, ok := t.chanUse.GetChan(id)
	if ok {
		return chans
	}

	/// if a new commer
	chans.Add(t.getANewChan())
	return chans
}
