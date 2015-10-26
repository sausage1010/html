// strategies
package main

import (
)


type roboStrategy interface {
	roboTrade(*Exchange, Robot, PositionUpdate) (BS string, comm CommBase, amount int64)
}

func movingAverage(exch *Exchange, comm CommBase, periods int) int64 {
	if periods < 1 {
		return 0
	}
	var avg int64 = 0
	for _, v := range exch.priceHist[comm][:periods] {
		avg += v
	}
	return avg / int64(periods)
}


func easyTrade(exch *Exchange, rob Robot, pos PositionUpdate, divisor int64) (BS string, comm CommBase, amount int64) {
	
	BS, tmpBS := "B", "B"
	comm, tmpComm := COPPER, COPPER
	amount, tmpAmount := int64(0), int64(0)
	
	for _, thisComm := range exch.commodities {
		tmpComm = thisComm.ComType
		if pos.Position.Holdings[thisComm.ComType] > 0 {
			tmpBS = "S"
			tmpAmount = pos.Position.Holdings[thisComm.ComType]
		} else {
			tmpBS = "B"
			tmpAmount = (pos.Position.Balance / pos.Prices[thisComm.ComType]) / divisor
		}
		
		if tmpAmount > amount {
			BS = tmpBS
			comm = tmpComm
			amount = tmpAmount
		}
	}
	
	return BS, comm, amount
}

type easyTradeFull struct {}
func (e easyTradeFull)roboTrade(exch *Exchange, rob Robot, pos PositionUpdate) (BS string, comm CommBase, amount int64) {
	return easyTrade(exch, rob, pos, 1)
}

type easyTradeHalf struct {}
func (e easyTradeHalf)roboTrade(exch *Exchange, rob Robot, pos PositionUpdate) (BS string, comm CommBase, amount int64) {
	return easyTrade(exch, rob, pos, 2)
}

/*
type PositionUpdate struct {
	Prices		map[CommBase]int64
	PriceHist	map[CommBase][]int64
	Position		Account
	ExchStatus	ExchangeStatus
}

type Account struct {
	Holdings		map[CommBase]int64
	Balance		int64
}
*/