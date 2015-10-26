// robots
package main

import (
	"strconv"
	"time"
	"math/rand"
	"log"
)


type Robot struct {
	Name			string
	avgPrice		map[CommBase]int64
	minTime		int64
	duration		int64
	tradeStrat	roboStrategy
}


func AddRobots(exch *Exchange, numRobots int, existingRobots int, d RoboDiffLevel) {
	
	// Set parameters for random sleep time
	var mt, dur int64
	switch d {
	case EASY:
		mt = 15
		dur = 45
	case MEDIUM:
		mt = 5
		dur = 15
	case HARD:
		mt = 1
		dur = 4
	}
	
	// Select a random trade strategy
	strategies := []roboStrategy{
		easyTradeFull{},
		easyTradeHalf{},
	}
	
	for i := 0; i < numRobots; i++ {
		
		stratNum := rand.Intn(len(strategies))
		
		rob := Robot{
			Name:		"Robot" + strconv.Itoa(existingRobots + i + 1),
			avgPrice:	make(map[CommBase]int64),
			minTime:		mt,
			duration:	dur,
			tradeStrat:	strategies[stratNum],
				}
		
		log.Println(rob.Name, " strategy is ", stratNum)
		
		reg := TraderReg{
					UserID:		rob.Name,
					ConfirmChan:	make(chan TraderRegConf),
				}
				
		exch.TraderRegs <- reg
		conf := <- reg.ConfirmChan
		
		if conf.RegOK {
			go rob.Run(exch)
		}
	}
}

func (rob Robot)Run(exch *Exchange) {
	
	for {
		time.Sleep(time.Duration(rand.Int63n(rob.duration) + rob.minTime) * time.Second)
		
		
		req := PositionReq {
					UserID:	rob.Name,
					Reply:	make(chan PositionUpdate),
				}
		exch.PositionReqs <- req
		pos := <- req.Reply
		
		if pos.ExchStatus == OPEN {
			BS, comm, amt := rob.tradeStrat.roboTrade(exch, rob, pos)
			if amt > 0 {
				trd := Trade {
							UserID:			rob.Name,
							BuySell:			BS,
							Commodity:		comm,
							Amount:			amt,
							ConfirmChan:		make(chan TradeConfirm),
						}
				exch.Trades <- trd
				conf := <- trd.ConfirmChan
				
				if !conf.TradeOK {
					log.Println(rob.Name, " trade failed: ", conf.Message)
				}
			} else {
				log.Println(rob.Name, ": Zero trade")
			}
		}
	}
}
