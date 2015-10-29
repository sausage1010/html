// display
package main

import (
	"net"
	"log"
	"io"
	"fmt"
	"time"
)

type Display struct {
	displayID		int
	terminal			net.Conn
	displayChan		chan DisplayUpdate
	displayRegConf	chan int
	displayTime		time.Time
}

type DisplayUpdate struct {
	data			string
	timeStamp	time.Time
}

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
		
		 
		disp := Display {
					displayID:		-1,
					terminal:		conn,
					displayChan:		make(chan DisplayUpdate),
					displayRegConf:	make(chan int),
					displayTime:		time.Now(),
				}

		exch.DisplayReg <- disp
		disp.displayID = <- disp.displayRegConf
				
		go disp.UpdateManager(exch.DisplayDeReg)
		}
	
	log.Println("Closing Display listener")
}



const cls string = "\r\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"
const uLine string = "-----------------------------------------------------------------------------"

func sendDisplayUpdate(exch *Exchange) {
	
	updateTime := time.Now()
	
	displayStr := "Exchange Status: " + statusString(exch.status) +
	              "           Time: " + fmt.Sprintf("%02d:%02d:%02d", updateTime.Hour(), updateTime.Minute(), updateTime.Second()) +
				  "\r\n\n"
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
	
	for _, conn := range exch.displays {
		go SendData(conn, displayStr, updateTime)
	}
}

func SendData(conn Display, displayStr string, updateTime time.Time) {
	conn.displayChan <- DisplayUpdate {
							data:		displayStr,
							timeStamp:	updateTime,
					    }
}

func (d *Display)UpdateManager(deRegChan chan Display) {
	defer d.terminal.Close()
	
	for {
		update := <- d.displayChan
		
		if update.timeStamp.After(d.displayTime) {
			
			_, err := io.WriteString(d.terminal, cls)
			if err !=nil {
				log.Println("Display connection lost.  Deregistering display")
				deRegChan <- *d
				break
			} else {
				io.WriteString(d.terminal, update.data)
				d.displayTime = update.timeStamp
			}

		} else {
			log.Println("Out of order diaplay update")
		}
		
	}
}

/*func updateDisplays(deRegChan chan net.Conn, displayList []net.Conn, displayStr string) {
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
*/