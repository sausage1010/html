// eldotxt
package main

import (
	"net"
	"log"
)

func main() {
	
	log.Println("Initialising exchange ...")
	exchange := &Exchange{
		Commands:		make(chan Command),
		DisplayReg:		make(chan net.Conn),
		DisplayDeReg:	make(chan net.Conn),
		TraderRegs:		make(chan TraderReg),
		TraderDeReg:		make(chan string),
		Trades:			make(chan Trade),
		PositionReqs:	make(chan PositionReq),
		
		status:			SUSPEND,
		statusMessage:	"",
		displays:		[]net.Conn{},
		
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
	}
	
	for _, c := range exchange.commodities {
		for i, _ := range exchange.priceHist[c.ComType] {
			exchange.priceHist[c.ComType][i] = StartPrice
		}
	}
	
	log.Println("Starting exchange ...")
	exchange.Run()
	
	log.Println("Exiting.")
}
