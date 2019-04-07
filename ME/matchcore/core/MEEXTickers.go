package core

import (
	"../../comm"
)

func (t *MEEXCore) GetAskLevelOrders(limit int64) []*comm.Order {
	// to do
	var os []*comm.Order

	return os
}

func (t *MEEXCore) GetBidLevelOrders(limit int64) []*comm.Order {
	// to do
	var os []*comm.Order

	return os
}

func (t *MEEXCore) GetAskLevelsGroupByPrice(limit int64) []comm.OrderLevel {
	// to do
	var (
		ols []comm.OrderLevel
	)

	return ols
}

func (t *MEEXCore) GetBidLevelsGroupByPrice(limit int64) []comm.OrderLevel {
	// to do
	var (
		ols []comm.OrderLevel
	)

	return ols
}
