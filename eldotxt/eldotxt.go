// eldotxt
package main

import (
//	"net"
	"log"
	"math/rand"
	"time"
)

func main() {
	
	playAgain := true
	
	for playAgain {
		
		log.Println("Initialising exchange ...")
		
		rand.Seed(time.Now().UTC().UnixNano())
		
		exchange := &Exchange{
			Commands:		make(chan Command),
			DisplayReg:		make(chan Display),
			DisplayDeReg:	make(chan Display),
			TraderRegs:		make(chan TraderReg),
			TraderDeReg:		make(chan string),
			Trades:			make(chan Trade),
			PositionReqs:	make(chan PositionReq),
			
			status:			SUSPEND,
			statusMessage:	"",
			displays:		make(map[int]Display),
			
			commodities:		[]Commodity{
								{ComType:	COPPER,
								 Name:		"Copper",
								 Price:		StartPrice},
								
								{ComType:	GOLD,
								 Name:		"Gold",
								 Price:		StartPrice},
								
								{ComType:	SILVER,
								 Name:		"Silver",
								 Price:		StartPrice},
								
								{ComType:	ZINC,
								 Name:		"Zinc",
								 Price:		StartPrice},
							},
							
			priceHist:		map[CommBase][]int64 {
								COPPER:	make([]int64, PriceHistLen, PriceHistLen),
								GOLD:	make([]int64, PriceHistLen, PriceHistLen),
								SILVER:	make([]int64, PriceHistLen, PriceHistLen),
								ZINC:	make([]int64, PriceHistLen, PriceHistLen),
							},
							
			accounts:		map[string]Account{},
			joinOrder:		[]string{},
			roboCount:		0,
			roboDiff:		MEDIUM,
			nextDisplayNum:	0,
		}
		
		for _, c := range exchange.commodities {
			for i, _ := range exchange.priceHist[c.ComType] {
				exchange.priceHist[c.ComType][i] = StartPrice
			}
		}
		
		log.Println("Starting exchange ...")
		playAgain = exchange.Run()
		playAgain = false // The whole restart thing is very complicated - need to keep net channels open ....
	}
	
	log.Println("Exiting.")
}
