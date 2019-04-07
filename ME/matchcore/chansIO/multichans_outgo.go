// multichans_out
package chansIO

import (
	"fmt"
	"sync/atomic"
)

const (
	MODULE_NAME_MULTICHANS_OUTGO string = "[MultiChans_Outgo]: "
	OUTGO_MULTI_CHANS_SIZE              = 10
	OUTGO_CHANNEL_SIZE                  = 6
)

var (
	OUTGO_MULTI_CHANS_SIZE_VAR int = 10
)

type ChanProcess_OutGo func(int)
type MultiChans_OutGo struct {
	chans      [OUTGO_MULTI_CHANS_SIZE]*ChanUnit_Out
	proc       ChanProcess_OutGo
	chanUse    *ChannelUse
	idleChanNO int64
}

func NewMultiChans_OutGo(p ChanProcess_OutGo) *MultiChans_OutGo {
	o := new(MultiChans_OutGo)
	for i := 0; i < OUTGO_MULTI_CHANS_SIZE; i++ {
		o.chans[i] = NewChanUnit_Out(i)
	}
	o.proc = p
	o.chanUse = NewChannelUse()
	/// prepare to work

	/// start work
	go func() { /// deal with chan nil problem
		for i := 0; i < OUTGO_MULTI_CHANS_SIZE; i++ {
			go o.proc(i)
		}
	}()

	return o
}

func (t *MultiChans_OutGo) InChannel(elem *OutElem) {
	switch elem.Type_ {
	case OUTPOOL_MATCHTRADE:
		elem.Count = 1
		t.chans[t.GetANewChan()].In(elem)

	case OUTPOOL_CANCELORDER:
		elem.Count = 1
		t.chans[t.GetANewChan()].In(elem)
	}
}

func (t *MultiChans_OutGo) OutChannel(chNO int) (*OutElem, bool) {
	elem := t.chans[chNO].Out()
	return elem, true
}

func (t *MultiChans_OutGo) GetChanUseStatus() (IDs, CHs, chnums int) {
	return t.chanUse.Status()
}

func (t *MultiChans_OutGo) Dump() {
	fmt.Printf("==================[MultiChans_OutGo Dump Detail]==================\n")
	for k, v := range t.chans {
		fmt.Printf("Chan[%d]: cap = %d, len = %d\n", k, v.Cap(), v.Len())
	}
	fmt.Printf("idleChanNO: %d\n", t.idleChanNO)
	t.chanUse.Dump(true)
	fmt.Printf("==========================================================\n")
}

func (t *MultiChans_OutGo) Summary() {
	fmt.Printf("==================[MultiChans_OutGo Dump Summary]==================\n")
	for k, v := range t.chans {
		fmt.Printf("Chan[%d]: cap = %d, len = %d\n", k, v.Cap(), v.Len())
	}
	fmt.Printf("idleChanNO: %d\n", t.idleChanNO)
	t.chanUse.Dump(false)
	fmt.Printf("=============================================================\n")
}

func (t *MultiChans_OutGo) Len() int {
	return OUTGO_MULTI_CHANS_SIZE
}

func (t *MultiChans_OutGo) ChanCap() int {
	return OUTGO_CHANNEL_SIZE
}

func (t *MultiChans_OutGo) GetANewChan() int64 {
	idleno := atomic.LoadInt64(&t.idleChanNO)
	for i := int64(0); i < OUTGO_MULTI_CHANS_SIZE; i++ {
		no := idleno + i
		if no >= OUTGO_MULTI_CHANS_SIZE {
			no = 0
		}
		if !t.chans[no].IsBusy() {
			idleno = no
			break
		}
	}

	/// update idleChanNO
	chNO := atomic.AddInt64(&t.idleChanNO, 1)
	if chNO >= OUTGO_MULTI_CHANS_SIZE {
		atomic.StoreInt64(&t.idleChanNO, 0)
	}

	if idleno >= OUTGO_MULTI_CHANS_SIZE {
		return 0
	} else {
		return idleno
	}
}
