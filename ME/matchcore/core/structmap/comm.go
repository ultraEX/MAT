package structmap

import "../../../comm"

type OrdersByScoreItf interface {
	Push(order *comm.Order)
	Pop() *comm.Order
	GetTop() *comm.Order
	Set(order *comm.Order)
	// Get(score float64) *comm.Order
	// Remove(score float64) *comm.Order
	Len() int
	Dump()
	Pump()
}
