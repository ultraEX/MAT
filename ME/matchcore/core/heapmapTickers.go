package core

import (
	"../../comm"
)

func (t *MEXCore) GetAskLevelOrders(limit int64) []*comm.Order {
	// to do
	var os []*comm.Order

	return os
}

func (t *MEXCore) GetBidLevelOrders(limit int64) []*comm.Order {
	// to do
	var os []*comm.Order

	return os
}

func (t *MEXCore) GetAskLevelsGroupByPrice(limit int64) []comm.OrderLevel {
	// to do
	var (
		ols []comm.OrderLevel
	)

	return ols
}

func (t *MEXCore) GetBidLevelsGroupByPrice(limit int64) []comm.OrderLevel {
	// to do
	var (
		ols []comm.OrderLevel
	)

	return ols
}
