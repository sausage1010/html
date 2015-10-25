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
	sleepTime	time.Duration
	tradeStrat	func(*Exchange, Robot, PositionUpdate) (BS string, comm CommBase, amount int64)
}


func AddRobots(exch *Exchange, numRobots int, existingRobots int, d RoboDiffLevel) {
	
	// Set parameters for random sleep time
	var shortTime, longTime int64
	switch d {
	case EASY:
		shortTime = 15
		longTime = 60
	case MEDIUM:
		shortTime = 5
		longTime = 20
	case HARD:
		shortTime = 1
		longTime = 5
	}
	
	for i := 0; i < numRobots; i++ {
		
		rob := Robot{
			Name:		"Robot" + strconv.Itoa(existingRobots + i + 1),
			avgPrice:	make(map[CommBase]int64),
			sleepTime:	time.Duration(rand.Int63n(longTime - shortTime) + shortTime) * time.Second,
				}
		
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
	log.Println("Robot: ", rob.Name, " running")
}
