package chansIO

import (
	"fmt"
	"sync"

	"../../comm"
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

type ChannelUse_Out struct {
	// chansMap map[int64]map[int]int
	chansMap sync.Map
}

func NewChannelUse_Out() *ChannelUse_Out {
	o := new(ChannelUse_Out)
	// o.chansMap = make(map[int64]map[int]int)
	return o
}

func (t *ChannelUse_Out) isSubEmpty(id int64) bool {
	if chans, ok := t.chansMap.Load(id); ok {
		cs := chans.(*sync.Map)
		is := true
		cs.Range(func(k, v interface{}) bool {
			is = false
			return false
		})
		return is
	}

	return true
}

func (t *ChannelUse_Out) InChan(id int64, no int) {

	// chans, ok := t.chansMap[id]
	var cs *sync.Map
	chans, isID := t.chansMap.Load(id)
	if isID {
		cs = chans.(*sync.Map)
		if count, ok := cs.Load(no); ok {
			// t.chansMap[id][no]++
			cs.Store(no, count.(int)+1)
		} else {
			// t.chansMap[id][no] = 1
			cs.Store(no, 1)
		}
	} else {
		var u sync.Map
		u.Store(no, 1)
		cs = &u
		t.chansMap.Store(id, &u)
	}

	comm.DebugPrintf(MODULE_NAME_MULTICHANS, comm.LOG_LEVEL_TRACK,
		"ChannelUse_Out id(%d) InChan(%d): len(chansMap)=%d, len(t.chansMap[id])=%d.\n",
		id, no, comm.LenOfSyncMap(&t.chansMap), comm.LenOfSyncMap(cs))
}

func (t *ChannelUse_Out) OutChan(id int64, no int) {
	// set := t.chansMap[id]
	chans, ok := t.chansMap.Load(id)
	if !ok {
		panic(fmt.Errorf("ChannelUse_Out logic error1."))
	}

	//if _, ok := t.chansMap[id][no]; ok {
	cs := chans.(*sync.Map)
	if count, ok := cs.Load(no); ok {
		//t.chansMap[id][no]--
		cs.Store(no, count.(int)-1)
		//if t.chansMap[id][no] <= 0 {
		if count, ok := cs.Load(no); ok && count.(int) <= 0 {
			// delete(t.chansMap[id], no)
			cs.Delete(no)
		}
		// if len(t.chansMap[id]) == 0 {
		if t.isSubEmpty(id) {
			// delete(t.chansMap, id)
			t.chansMap.Delete(id)
		}

		comm.DebugPrintf(MODULE_NAME_MULTICHANS, comm.LOG_LEVEL_TRACK,
			"ChannelUse_Out id(%d) OutChan(%d): len(chansMap)=%d, len(t.chansMap[id])=%d.\n",
			id, no, comm.LenOfSyncMap(&t.chansMap), comm.LenOfSyncMap(cs))
	} else {
		panic(fmt.Errorf("ChannelUse_Out logic error2."))
	}
}

func (t *ChannelUse_Out) RemoveFromChan(id int64) {
	// delete(t.chansMap, id)
	t.chansMap.Delete(id)
}

func (t *ChannelUse_Out) GetChan(id int64) (v interface{}, ok bool) {
	// v, ok = t.chansMap[id]
	v, ok = t.chansMap.Load(id)
	return v, ok
}

func (t *ChannelUse_Out) Status() (int, int, int) {
	IDs := comm.LenOfSyncMap(&t.chansMap)
	CHs := 0
	chnums := 0
	// for _, v := range t.chansMap {
	t.chansMap.Range(func(k, v interface{}) bool {
		cs := v.(*sync.Map)
		CHs += comm.LenOfSyncMap(v.(*sync.Map))
		// for _, value := range v {
		cs.Range(func(k, v interface{}) bool {
			chnums += v.(int)
			return true
		})
		return true
	})

	return IDs, CHs, chnums
}

func (t *ChannelUse_Out) Dump(detail bool) {
	fmt.Printf("==================[ChannelUse_Out Dump]==================\n")
	fmt.Printf("len=%d\n", comm.LenOfSyncMap(&t.chansMap))
	if detail {
		// for k, v := range t.chansMap {
		t.chansMap.Range(func(k, v interface{}) bool {
			fmt.Printf("---------------------------------------------------------\n")
			fmt.Printf("channel id = [%d]: ", k.(int64))
			// for key, value := range v {
			cs := v.(*sync.Map)
			cs.Range(func(key, value interface{}) bool {
				fmt.Printf("id[%d]: chanNO = %v, count = %d\n", k.(int64), key.(int), value.(int))
				return true
			})
			return true
		})
	}
	fmt.Printf("======================================================\n")
}
