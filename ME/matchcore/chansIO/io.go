package chansIO

import (
	"fmt"
	"sync"

	"../../comm"
)

const MODULE_NAME_MULTICHANS_IO string = "[MultiChans_IO]: "

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
	count int
}

func NewCounter() *Counter {
	o := new(Counter)
	o.count = 1
	return o
}

func (t *Counter) Inc() {
	t.count++
}

func (t *Counter) Dec() {
	t.count--
}

type MutexMap struct {
	mutexMap sync.Map
}

func (t *MutexMap) Lock(id int64) {
	if lockItf, ok := t.mutexMap.Load(id); ok {
		lock := lockItf.(*sync.Mutex)
		lock.Lock()
	} else {
		lock := sync.Mutex{}
		t.mutexMap.Store(id, &lock)
		lock.Lock()
	}
}

func (t *MutexMap) Unlock(id int64) {
	if lockItf, ok := t.mutexMap.Load(id); ok {
		lock := lockItf.(*sync.Mutex)
		lock.Unlock()
	} else {
		panic(fmt.Errorf("MutexMap.Unlock  a not exist lock, logic error, check it!!!"))
	}
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

func (t *ChansCount) Size() int {
	size := 0
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
	conMutex sync.Mutex
	// chansMap sync.Map
}

func NewChannelUse() *ChannelUse {
	o := new(ChannelUse)
	o.chansMap = make(map[int64]*ChansCount)
	return o
}

func (t *ChannelUse) IsSubEmpty(id int64) bool {
	// if chans, ok := t.chansMap.Load(id); ok {
	// 	cs := chans.(*sync.Map)
	// 	is := true
	// 	cs.Range(func(k, v interface{}) bool {
	// 		is = false
	// 		return false
	// 	})
	// 	return is
	// }

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

	// // chans, ok := t.chansMap[id]
	// var cs *sync.Map
	// chans, isID := t.chansMap.Load(id)
	// chans.ConMutex.
	// if isID {
	// 	cs = chans.(*sync.Map)
	// 	if count, ok := cs.Load(no); ok {
	// 		// t.chansMap[id][no]++
	// 		// cs.Store(no, count.(int)+1)
	// 		pCount := count.(*int32)
	// 		atomic.AddInt32(pCount, 1)
	// 	} else {
	// 		// t.chansMap[id][no] = 1
	// 		// cs.Store(no, 1)
	// 		pCount := count.(*int32)
	// 		atomic.StoreInt32(pCount, 1)
	// 	}
	// } else {
	// 	var u sync.Map
	// 	var count int32 = 1
	// 	u.Store(no, &count)
	// 	cs = &u
	// 	t.chansMap.Store(id, &u)
	// }

	t.conMutex.Lock()
	chans, ok := t.chansMap[id]
	if ok {
		chans.ChanCountInc(no)
	} else {
		chans = NewChansCount()
		chans.ChanCountInc(no)
		t.chansMap[id] = chans
	}
	t.conMutex.Unlock()

	comm.DebugPrintf(MODULE_NAME_MULTICHANS_IO, comm.LOG_LEVEL_TRACK,
		"ChannelUse id(%d) InChan(%d): len(chansMap)=%d, len(chansMap[id])=%d.\n",
		id, no, len(t.chansMap), chans.Len())
}

func (t *ChannelUse) OutChan(id int64, no int) {
	// // set := t.chansMap[id]
	// chans, ok := t.chansMap.Load(id)
	// if !ok {
	// 	panic(fmt.Errorf("id(%d) not exist in chanusemap, Logic error, Check it!!!", id))
	// }

	// //if _, ok := t.chansMap[id][no]; ok {
	// cs := chans.(*sync.Map)
	// if count, ok := cs.Load(no); ok {
	// 	//t.chansMap[id][no]--
	// 	// cs.Store(no, count.(int)-1)
	// 	pCount := count.(*int32)
	// 	atomic.AddInt32(pCount, -1)
	// 	//if t.chansMap[id][no] <= 0 {
	// 	if count, ok := cs.Load(no); ok && count.(int) <= 0 {
	// 		// delete(t.chansMap[id], no)
	// 		cs.Delete(no)
	// 	}
	// 	// if len(t.chansMap[id]) == 0 {
	// 	if t.isSubEmpty(id) {
	// 		// delete(t.chansMap, id)
	// 		t.chansMap.Delete(id)
	// 	}

	// 	comm.DebugPrintf(MODULE_NAME_MULTICHANS_IO, comm.LOG_LEVEL_TRACK,
	// 		"ChannelUse id(%d) OutChan(%d): len(chansMap)=%d, len(t.chansMap[id])=%d.\n",
	// 		id, no, comm.LenOfSyncMap(&t.chansMap), comm.LenOfSyncMap(cs))
	// } else {
	// 	panic(fmt.Errorf("ChannelUse logic error2."))
	// }

	t.conMutex.Lock()

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

	t.conMutex.Unlock()
}

func (t *ChannelUse) RemoveFromChan(id int64) {
	t.conMutex.Lock()
	delete(t.chansMap, id)
	t.conMutex.Unlock()
	// t.chansMap.Delete(id)
}

func (t *ChannelUse) GetChan(id int64) ([]int, bool) {
	t.conMutex.Lock()
	var chSet []int = nil
	var res bool = false
	if cs, ok := t.chansMap[id]; ok {
		for k, _ := range cs.countsMap {
			chSet = append(chSet, k)
		}
		res = true
	} else {
		res = false
	}
	t.conMutex.Unlock()
	// v, ok = t.chansMap.Load(id)
	return chSet, res
}

func (t *ChannelUse) Len() int {
	return len(t.chansMap)
}

func (t *ChannelUse) Status() (int, int, int) {
	// IDs := comm.LenOfSyncMap(&t.chansMap)
	// CHs := 0
	// chnums := 0
	// // for _, v := range t.chansMap {
	// t.chansMap.Range(func(k, v interface{}) bool {
	// 	cs := v.(*sync.Map)
	// 	CHs += comm.LenOfSyncMap(v.(*sync.Map))
	// 	// for _, value := range v {
	// 	cs.Range(func(k, v interface{}) bool {
	// 		chnums += v.(int)
	// 		return true
	// 	})
	// 	return true
	// })

	IDs := len(t.chansMap)
	CHs := 0
	chnums := 0
	t.conMutex.Lock()
	for _, v := range t.chansMap {
		CHs += v.Len()
		chnums += v.Size()
	}
	t.conMutex.Unlock()

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
