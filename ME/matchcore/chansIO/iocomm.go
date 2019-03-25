package chansIO

import "../../comm"

type InElemType int64

const (
	InElemType_EnOrder  InElemType = 1
	InElemType_CancelID InElemType = 2
)

func (p InElemType) String() string {
	switch p {
	case InElemType_EnOrder:
		return "InElemType: EnOrder"
	case InElemType_CancelID:
		return "InElemType: CancelID"
	}
	return "<InElemType UNSET>"
}

type InElem struct {
	Type_    InElemType
	Order    *comm.Order
	CancelId int64
	Count    int64
}

type ChanUnit_In struct {
	ch chan *InElem
	no int
}

func NewChanUnit_In(no int) *ChanUnit_In {
	o := new(ChanUnit_In)
	o.ch = make(chan *InElem, IN_CHANNEL_SIZE)
	o.no = no

	return o
}

func (t *ChanUnit_In) Len() int {
	return len(t.ch)
}

func (t *ChanUnit_In) Cap() int {
	return cap(t.ch)
}

func (t *ChanUnit_In) IsBusy() bool {
	return len(t.ch) == cap(t.ch)
}

func (t *ChanUnit_In) In(elem *InElem) {
	t.ch <- elem
}

func (t *ChanUnit_In) Out() (elem *InElem) {
	return <-t.ch
}

type OutElem struct {
	Type_       OutPoolType
	Trade       *MatchTrade
	CancelOrder *CanceledOrder
	Count       int64
}

type ChanUnit_Out struct {
	ch chan *OutElem
	no int
}

func NewChanUnit_Out(no int) *ChanUnit_Out {
	o := new(ChanUnit_Out)
	o.ch = make(chan *OutElem, OUT_CHANNEL_SIZE)
	o.no = no

	return o
}

func (t *ChanUnit_Out) Len() int {
	return len(t.ch)
}

func (t *ChanUnit_Out) Cap() int {
	return cap(t.ch)
}

func (t *ChanUnit_Out) IsBusy() bool {
	return len(t.ch) == cap(t.ch)
}

func (t *ChanUnit_Out) In(elem *OutElem) {
	t.ch <- elem
}

func (t *ChanUnit_Out) Out() (elem *OutElem) {
	return <-t.ch
}
