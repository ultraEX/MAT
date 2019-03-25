// multichans_out
package chansIO

import (
	"fmt"
	"sync/atomic"

	"../../comm"
)

const (
	MODULE_NAME_MULTICHANS_OUT string = "[MultiChans]: "
	OUT_MULTI_CHANS_SIZE       int    = 10
)

var (
	OUT_MULTI_CHANS_SIZE_VAR int = 10
)

type ChanProcess_Out func(int)
type MultiChans_Out struct {
	chans      [OUT_MULTI_CHANS_SIZE]*ChanUnit_Out
	proc       ChanProcess_Out
	chanUse    *ChannelUse
	idleChanNO int
}

func NewMultiChans_Out(p ChanProcess_Out) *MultiChans_Out {
	o := new(MultiChans_Out)
	for i := 0; i < OUT_MULTI_CHANS_SIZE; i++ {
		o.chans[i] = NewChanUnit_Out(i)
	}
	o.proc = p
	o.chanUse = NewChannelUse()
	/// prepare to work

	/// start work
	go func() { /// deal with chan nil problem
		for i := 0; i < OUT_MULTI_CHANS_SIZE; i++ {
			go o.proc(i)
		}
	}()

	return o
}

func (t *MultiChans_Out) InChannel(elem *OutElem) {
	switch elem.Type_ {
	case OUTPOOL_MATCHTRADE:
		askID := elem.Trade.AskTrade.ID
		bidID := elem.Trade.BidTrade.ID
		chSet := t.GetIdleChannel_Trade(askID, bidID)
		elem.Count = int64(chSet.Len())
		for _, v := range chSet.Elements() {
			t.chanUse.InChan(askID, v.(int))
			t.chanUse.InChan(bidID, v.(int))
			t.chans[v.(int)].In(elem)
		}
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_OUT, comm.LOG_LEVEL_TRACK, "MultiChans_Out OUTPOOL_MATCHTRADE InChannel(%v).\n", chSet)

	case OUTPOOL_CANCELORDER:
		id := elem.CancelOrder.Order.ID
		chSet := t.GetIdleChannel_Cancel(id)
		elem.Count = int64(chSet.Len())
		for _, v := range chSet.Elements() {
			t.chanUse.InChan(id, v.(int))
			t.chans[v.(int)].In(elem)
		}
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_OUT, comm.LOG_LEVEL_TRACK, "MultiChans_Out OUTPOOL_CANCELORDER InChannel(%v).\n", chSet)
	}
}

func (t *MultiChans_Out) OutChannel(chNO int) (*OutElem, bool) {
	elem := t.chans[chNO].Out()
	count := atomic.AddInt64(&elem.Count, -1)

	// chanuse manage
	switch elem.Type_ {
	case OUTPOOL_MATCHTRADE:
		t.chanUse.OutChan(elem.Trade.AskTrade.ID, chNO)
		t.chanUse.OutChan(elem.Trade.BidTrade.ID, chNO)
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_OUT, comm.LOG_LEVEL_TRACK,
			"MultiChans_Out OUTPOOL_MATCHTRADE OutChannel(%d): OutChan(ask(id=%d),bid(id=%d), chanNO=%d.\n",
			chNO, elem.Trade.AskTrade.ID, elem.Trade.BidTrade.ID, chNO)
	case OUTPOOL_CANCELORDER:
		t.chanUse.OutChan(elem.CancelOrder.Order.ID, chNO)
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_OUT, comm.LOG_LEVEL_TRACK,
			"MultiChans_Out OUTPOOL_CANCELORDER OutChannel(%d): OutChan(cancel order(id=%d), chanNO=%d.\n",
			chNO, elem.CancelOrder.Order.ID, chNO)
	}

	if count <= 0 {
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_OUT, comm.LOG_LEVEL_TRACK, "MultiChans_Out OutChannel(%d): %v.\n", chNO, elem)
		return elem, true
	} else {
		comm.DebugPrintf(MODULE_NAME_MULTICHANS_OUT, comm.LOG_LEVEL_TRACK, "MultiChans_Out OutChannel(%d) nil.\n", chNO)
		return nil, false
	}

}

func (t *MultiChans_Out) GetChanUseStatus() (IDs, CHs, chnums int) {
	return t.chanUse.Status()
}

func (t *MultiChans_Out) Dump() {
	fmt.Printf("==================[MultiChans_Out Dump Detail]==================\n")
	for k, v := range t.chans {
		fmt.Printf("Chan[%d]: cap = %d, len = %d\n", k, v.Cap(), v.Len())
	}
	fmt.Printf("idleChanNO: %d\n", t.idleChanNO)
	t.chanUse.Dump(true)
	fmt.Printf("==========================================================\n")
}

func (t *MultiChans_Out) Summary() {
	fmt.Printf("==================[MultiChans_Out Dump Summary]==================\n")
	for k, v := range t.chans {
		fmt.Printf("Chan[%d]: cap = %d, len = %d\n", k, v.Cap(), v.Len())
	}
	fmt.Printf("idleChanNO: %d\n", t.idleChanNO)
	t.chanUse.Dump(false)
	fmt.Printf("=============================================================\n")
}

func (t *MultiChans_Out) Len() int {
	return OUT_MULTI_CHANS_SIZE
}

func (t *MultiChans_Out) ChanCap() int {
	return OUT_CHANNEL_SIZE
}

func (t *MultiChans_Out) getANewChan() int {
	idleno := t.idleChanNO
	for i := 0; i < OUT_MULTI_CHANS_SIZE; i++ {
		no := t.idleChanNO + i
		if no >= OUT_MULTI_CHANS_SIZE {
			no = 0
		}
		if !t.chans[no].IsBusy() {
			idleno = no
			break
		}
	}

	/// update idleChanNO
	t.idleChanNO++
	if t.idleChanNO >= OUT_MULTI_CHANS_SIZE {
		t.idleChanNO = 0
	}
	return idleno
}

func (t *MultiChans_Out) GetIdleChannel_Trade(askID int64, bidID int64) comm.Set {
	/// if a secondary commer, use the original channel to ensure serialize
	askCh, okAsk := t.chanUse.GetChan(askID)
	bidCh, okBid := t.chanUse.GetChan(bidID)
	var chs comm.Set
	if okAsk || okBid {
		chs = comm.Union(askCh, bidCh)
		return chs
	}

	/// if a new commer
	if chs == nil {
		chs = comm.NewHashSet()
	}
	chs.Add(t.getANewChan())
	return chs
}

func (t *MultiChans_Out) GetIdleChannel_Cancel(id int64) comm.Set {
	/// if a secondary commer, use the original channel to ensure serialize
	chs, ok := t.chanUse.GetChan(id)
	if ok {
		return chs
	}

	/// if a new commer
	if chs == nil {
		chs = comm.NewHashSet()
	}
	chs.Add(t.getANewChan())
	return chs
}
