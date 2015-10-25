// display
package main

import (
	"net"
	"log"
	"io"
	"fmt"
)

func displayConnHandler(exch *Exchange) {
	
	log.Println("Starting Display listener.")
	server, err := net.Listen("tcp", ":4001")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		
		exch.DisplayReg <- conn
	}
	log.Println("Closing Display listener")
}

const cls string = "\r\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"
const uLine string = "-----------------------------------------------------------------------------"

func sendDisplayUpdate(exch *Exchange) {
	
	displayStr := "Exchange Status: " + statusString(exch.status) + "\r\n\n"
	titleStr :=  "    Trader | "
	pricesStr := "           | "
	
	// Create headers
	for _, c := range exch.commodities {
		titleStr += fmt.Sprintf("%10s | ", c.Name)
		pricesStr += fmt.Sprintf("%10d | ", c.Price)
	}
	titleStr += fmt.Sprintf("%10s |", "Balance")
	displayStr += titleStr + "\r\n" +
	              pricesStr +"\r\n" +
			      uLine + "\r\n"
	
	// List accounts
	for _, userID := range exch.joinOrder {
		acc := exch.accounts[userID]
		accStr := fmt.Sprintf("%10s | ", userID)
		for _, c := range exch.commodities {
			accStr += fmt.Sprintf("%10d | ", acc.Holdings[c.ComType])
		}
		accStr += fmt.Sprintf("%10d |", acc.Balance)
		displayStr += accStr + "\r\n"
	}
	displayStr += uLine + "\r\n" + exch.statusMessage + "\r\n"
	
	go updateDisplays(exch.DisplayDeReg, exch.displays, displayStr)
}


func updateDisplays(deRegChan chan net.Conn, displayList []net.Conn, displayStr string) {
	for i, conn := range displayList {
		_, err := io.WriteString(conn, cls)
		if err !=nil {
			log.Println("Display connection lost.  Deregistering display: ", i)
			deRegChan <- conn
		} else {
			io.WriteString(conn, displayStr)
		}
	}
}