package chansIO

import (
	"fmt"
	"sync"
	"sync/atomic"

	"../../comm"
)

const MODULE_NAME_MULTICHANS_IO string = "[MultiChans_IO]: "
const (
	OUT_CHANNEL_SIZE int = 6
)

type OutPoolType int64

const (
	OUTPOOL_MATCHTRADE  OutPoolType = 1
	OUTPOOL_CANCELORDER OutPoolType = 2
)

func (p OutPoolType) String() string {
	switch p {
	case OUTPOOL_MATCHTRADE:
		return "OUTPOOL_MATCHTRADE"
	case OUTPOOL_CANCELORDER:
		return "OUTPOOL_CANCELORDER"
	}
	return "<UNSET>"
}

type MatchTrade struct {
	BidTrade *comm.Trade
	AskTrade *comm.Trade
}

type CanceledOrder struct {
	Order *comm.Order
}

type Counter struct {
	count int64
}

func NewCounter() *Counter {
	o := new(Counter)
	o.count = 1
	return o
}

func (t *Counter) Inc() {
	atomic.AddInt64(&t.count, 1)
}

func (t *Counter) Dec() {
	atomic.AddInt64(&t.count, -1)
}

type ChansCount struct {
	countsMap map[int]*Counter
}

func NewChansCount() *ChansCount {
	o := new(ChansCount)
	o.countsMap = make(map[int]*Counter)
	return o
}

func (t *ChansCount) ChanCountInc(no int) {

	if _, ok := t.countsMap[no]; ok {
		t.countsMap[no].Inc()
	} else {
		counter := NewCounter()
		t.countsMap[no] = counter
	}

}

func (t *ChansCount) ChanCountDec(no int) {
	_, ok := t.countsMap[no]
	if ok {
		t.countsMap[no].Dec()
		if t.countsMap[no].count <= 0 {
			delete(t.countsMap, no)
		}
	} else {
		panic(fmt.Errorf("ChansCount.ChanCountDec logic error."))
	}
}

func (t *ChansCount) Len() int {
	return len(t.countsMap)
}

func (t *ChansCount) Size() int64 {
	var size int64 = 0
	for _, v := range t.countsMap {
		size += v.count
	}
	return size
}

func (t *ChansCount) Dump() {
	fmt.Printf("----------------------ChansCount Info----------------------------\n")
	fmt.Printf("len=%d\n", t.Len())

	for k, v := range t.countsMap {
		fmt.Printf("\tchanNO = %d count = %d\n", k, v.count)
	}

	fmt.Printf("------------------------------------------------------------------\n")
}

type ChannelUse struct {
	chansMap map[int64]*ChansCount
	conMutex *sync.RWMutex
}

func NewChannelUse() *ChannelUse {
	o := new(ChannelUse)
	o.chansMap = make(map[int64]*ChansCount)
	o.conMutex = new(sync.RWMutex)
	return o
}

func (t *ChannelUse) IsChanEmpty(id int64) bool {
	t.conMutex.RLock()
	defer t.conMutex.RUnlock()

	if chans, ok := t.chansMap[id]; ok {
		if chans.Len() <= 0 {
			return true
		} else {
			return false
		}
	}
	return true
}

func (t *ChannelUse) InChan(id int64, no int) {
	t.conMutex.Lock()
	defer t.conMutex.Unlock()

	chans, ok := t.chansMap[id]
	if ok {
		chans.ChanCountInc(no)
	} else {
		chans = NewChansCount()
		chans.ChanCountInc(no)
		t.chansMap[id] = chans
	}

	comm.DebugPrintf(MODULE_NAME_MULTICHANS_IO, comm.LOG_LEVEL_TRACK,
		"ChannelUse id(%d) InChan(%d): len(chansMap)=%d, len(chansMap[id])=%d.\n",
		id, no, len(t.chansMap), chans.Len())
}

func (t *ChannelUse) OutChan(id int64, no int) {
	t.conMutex.Lock()
	defer t.conMutex.Unlock()

	chans, ok := t.chansMap[id]
	if !ok {
		panic(fmt.Errorf("id(%d) not exist in chanusemap, Logic error, Check it!!!", id))
	}

	chans.ChanCountDec(no)
	size := chans.Len()
	if size <= 0 {
		delete(t.chansMap, id)
	}

	comm.DebugPrintf(MODULE_NAME_MULTICHANS_IO, comm.LOG_LEVEL_TRACK,
		"ChannelUse id(%d) OutChan(%d): len(chansMap)=%d, len(t.chansMap[id])=%d.\n",
		id, no, len(t.chansMap), size)

}

func (t *ChannelUse) RemoveID(id int64) {
	t.conMutex.Lock()
	defer t.conMutex.Unlock()

	delete(t.chansMap, id)

}

func (t *ChannelUse) GetChan(id int64) (*comm.HashSet, bool) {
	t.conMutex.RLock()
	defer t.conMutex.RUnlock()

	var chSet *comm.HashSet = comm.NewHashSet()
	var res bool = false
	if cs, ok := t.chansMap[id]; ok {
		for k, _ := range cs.countsMap {
			// chSet = append(chSet, k)
			chSet.Add(k)
		}
		res = true
	} else {
		res = false
	}

	return chSet, res
}

func (t *ChannelUse) Len() int {
	return len(t.chansMap)
}

func (t *ChannelUse) Status() (int, int, int) {
	t.conMutex.RLock()
	defer t.conMutex.RUnlock()

	IDs := len(t.chansMap)
	CHs := 0
	chnums := 0

	for _, v := range t.chansMap {
		CHs += v.Len()
		chnums += int(v.Size())
	}

	return IDs, CHs, chnums
}

func (t *ChannelUse) Dump(detail bool) {
	fmt.Printf("==================[ChannelUse Dump]==================\n")
	fmt.Printf("len=%d\n", t.Len())
	if detail {
		for k, v := range t.chansMap {
			fmt.Printf("---------------------------------------------------------\n")
			fmt.Printf("channel id = [%d]: ", k)
			v.Dump()
		}
	}
	fmt.Printf("======================================================\n")
}
