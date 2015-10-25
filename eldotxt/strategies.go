// strategies
package main

import (
)

var	roboStrategies	[]func(*Exchange, Robot, PositionUpdate) (BS string, comm CommBase, amount int64)

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

func easyTrade(exch *Exchange, rob Robot, pos PositionUpdate) (BS string, comm CommBase, amount int64) {
	return
}

