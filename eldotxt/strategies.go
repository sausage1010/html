// strategies
package main

import (
)


type roboStrategy interface {
	roboTrade(*Exchange, Robot, PositionUpdate) (BS string, comm CommBase, amount int64)
}


type easyTrade struct {
	divisor int64
}
func (e easyTrade)roboTrade(exch *Exchange, rob Robot, pos PositionUpdate) (BS string, comm CommBase, amount int64) {
	
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
			tmpAmount = (pos.Position.Balance / pos.Prices[thisComm.ComType]) / e.divisor
		}
		
		if tmpAmount > amount {
			BS = tmpBS
			comm = tmpComm
			amount = tmpAmount
		}
	}
	
	return BS, comm, amount
}


type profitTrade struct {
	divisor int64
}
func (p profitTrade)roboTrade(exch *Exchange, rob Robot, pos PositionUpdate) (BS string, comm CommBase, amount int64) {
	
	BS, tmpBS := "B", "B"
	comm, tmpComm := COPPER, COPPER
	amount, tmpAmount := int64(0), int64(0)
	
	for _, thisComm := range exch.commodities {
		tmpComm = thisComm.ComType

		if (pos.Position.Holdings[thisComm.ComType] == 0) ||
		   (rob.avgPrice[thisComm.ComType] > pos.Prices[thisComm.ComType]) {
			tmpBS = "B"
			tmpAmount = (pos.Position.Balance / pos.Prices[thisComm.ComType]) / p.divisor
		} else {
			tmpBS = "S"
			tmpAmount = pos.Position.Holdings[thisComm.ComType]
		}

		if tmpAmount > amount {
			BS = tmpBS
			comm = tmpComm
			amount = tmpAmount
		}
	}
	
	return BS, comm, amount
}


func movingAverage(hist []int64, periods int) int64 {
	if periods < 1 {
		return 0
	}
	var avg int64 = 0
	for _, v := range hist[:periods] {
		avg += v
	}
	return avg / int64(periods)
}

type mvgAvgTrade struct {
	divisor 		int64
	shortTime	int
	longTime		int
}
func (m mvgAvgTrade)roboTrade(exch *Exchange, rob Robot, pos PositionUpdate) (BS string, comm CommBase, amount int64) {
	
	BS, tmpBS := "B", "B"
	comm, tmpComm := COPPER, COPPER
	amount, tmpAmount := int64(0), int64(0)
	
	for _, thisComm := range exch.commodities {
		tmpComm = thisComm.ComType

		longAvg := movingAverage(pos.PriceHist[thisComm.ComType], m.longTime)
		shortAvg := movingAverage(pos.PriceHist[thisComm.ComType], m.shortTime)
		
		if shortAvg <= longAvg {
			tmpBS = "B"
			tmpAmount = (pos.Position.Balance / pos.Prices[thisComm.ComType]) / m.divisor
		} else {
			tmpBS = "S"
			tmpAmount = pos.Position.Holdings[thisComm.ComType]
		}

		if tmpAmount > amount {
			BS = tmpBS
			comm = tmpComm
			amount = tmpAmount
		}
	}
	
	return BS, comm, amount
}

