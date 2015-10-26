// eldoServer
package main

import (
	"net"
	"log"
	"math/rand"
	"time"
	"strconv"
)


type PositionReq struct {
	UserID		string
	Reply		chan PositionUpdate
}

type PositionUpdate struct {
	Prices		map[CommBase]int64
	PriceHist	map[CommBase][]int64
	Position		Account
	ExchStatus	ExchangeStatus
}



type ExchangeStatus int
const (
	SUSPEND ExchangeStatus = iota
	OPEN
	STOP
)

func statusString(s ExchangeStatus) string{
	switch s {
	case SUSPEND:
		return "Suspended"
	case OPEN:
		return "Open"
	case STOP:
		return "Closed"
	default:
		return "ERROR"
	}
}

type RoboDiffLevel int
const (
	EASY RoboDiffLevel = iota
	MEDIUM
	HARD
)

func roboDiffString(d RoboDiffLevel) string {
	switch d {
	case EASY:
		return "Easy"
	case MEDIUM:
		return "Medium"
	case HARD:
		return "Hard"
	default:
		return "ERROR"
	}
}

type CommBase int
const (
	COPPER CommBase = iota
	GOLD
	SILVER
	ZINC
)
const NumComm int = 4
const StartPrice int64 = 100
const StartBalance int64 = 1000
const PriceHistLen int = 60
const WinningBal int64 = 1000000000
const MaxRobots int = 10
//const WinningBal int64 = 10000

type Commodity struct {
	ComType		CommBase
	Name			string
	Price		int64
}

type Account struct {
	Holdings		map[CommBase]int64
	Balance		int64
}

type Exchange struct {
	// External channels
	Commands			chan Command
	DisplayReg		chan net.Conn
	DisplayDeReg		chan net.Conn
	TraderRegs		chan TraderReg
	TraderDeReg		chan string
	Trades			chan Trade
	PositionReqs		chan PositionReq
	
	// Internal channels
	
	// Internal Variables
	status			ExchangeStatus
	statusMessage	string
	displays			[]net.Conn
	commodities		[]Commodity
	priceHist		map[CommBase][]int64
	accounts			map[string]Account
	joinOrder		[]string
	roboCount		int
	roboDiff			RoboDiffLevel
}


func (exch *Exchange)handleCommand(command *Command) {
	
	switch command.Command {
	case "START":
	{
		switch exch.status {
		case SUSPEND:
			exch.status = OPEN
			command.Reply <- "Exchange started\r\n"
			close(command.Reply)
			sendDisplayUpdate(exch)
		case OPEN:
			command.Reply <- "Invalid command: Exchange already running\r\n"
			close(command.Reply)
		case STOP:
		default:
			log.Fatalln("Invalid Exchange status")
		}
	}
	case "PAUSE":
	{
		switch exch.status {
		case SUSPEND:
			command.Reply <- "Invalid command: Exchange already suspended.\r\n"
			close(command.Reply)
		case OPEN:
			exch.status = SUSPEND
			command.Reply <- "Exchange suspended\r\n"
			close(command.Reply)
			sendDisplayUpdate(exch)
		case STOP:
		default:
			log.Fatalln("Invalid Exchange status")
		}
	}
	case "STOP":
	{
		switch exch.status {
		case SUSPEND:
			exch.status = STOP
			command.Reply <- "Closing exchange ...\r\n"
			close(command.Reply)
		case OPEN:
			command.Reply <- "Invalid Command: Please pause the exchange before stopping\r\n"
			close(command.Reply)
		case STOP:
		default:
			log.Fatalln("Invalid Exchange status")
		}
	}
	case "ROBOT":
	{
		n, numErr := strconv.ParseInt(command.Argument1, 10, 8)
		numRobots := int(n)
		if numErr != nil {
			command.Reply <- "Invalid argument: Should be a number between 1 and " + strconv.Itoa(MaxRobots) + ".\r\n"
			close(command.Reply)
		} else if numRobots < 1 {
			command.Reply <- "Invalid argument: Should be a positive number\r\n"
			close(command.Reply)
		} else if (numRobots + exch.roboCount) > MaxRobots {
			if exch.roboCount == 0 {
				command.Reply <- "Sorry, max number of robots is " + strconv.Itoa(MaxRobots) + "\r\n"
				close(command.Reply)				
			} else {
				command.Reply <- "Sorry, there are already " + strconv.Itoa(exch.roboCount) + " robots. " +
								"Max is " + strconv.Itoa(MaxRobots) + "\r\n"
				close(command.Reply)
			}
		} else {
			go AddRobots(exch, numRobots, exch.roboCount, exch.roboDiff)
			exch.roboCount += numRobots
			command.Reply <- "OK\r\n"
			close(command.Reply)
		}
	}
	case "DIFF":
	{
		switch command.Argument1 {
		case "E":
			exch.roboDiff = EASY
			command.Reply <- "Robot level set to EASY\r\n"
			close(command.Reply)
		case "M":
			exch.roboDiff = MEDIUM
			command.Reply <- "Robot level set to MEDIUM\r\n"
			close(command.Reply)
		case "H":
			exch.roboDiff = HARD
			command.Reply <- "Robot level set to HARD\r\n"
			close(command.Reply)
		case "1":  // If no second argument was given, '1' is used as default
			command.Reply <- "Robot difficulty level is " + roboDiffString(exch.roboDiff) + "\r\n"
			close(command.Reply)
		default:
			command.Reply <- "Invalid level. Please use E, M or H\r\n"
			close(command.Reply)
		}
	}
	default:
		command.Reply <- "Invalid Command.\r\n"
		close(command.Reply)
	}
}

func (exch *Exchange)handleTrade(trade Trade) {
	
// Test for invalid trade details	
	acc, accOk := exch.accounts[trade.UserID]
	if !accOk {
		trade.ConfirmChan <- TradeConfirm{
								TradeOK:		false,
								Message:		"Invalid UserID: " + trade.UserID,
							}
		close(trade.ConfirmChan)
		return
	}
	
	if trade.BuySell != "B" && trade.BuySell != "S" {
		trade.ConfirmChan <- TradeConfirm{
								TradeOK:		false,
								Message:		"Invalid order type: " + trade.BuySell,
							}
		close(trade.ConfirmChan)
		return
	}
	
	if trade.Amount < 0 {
		trade.ConfirmChan <- TradeConfirm{
								TradeOK:		false,
								Message:		"Negative trade volume",
							}
		close(trade.ConfirmChan)
		return
	}
	
	if (trade.BuySell == "S") && (acc.Holdings[trade.Commodity] < trade.Amount) {
		trade.ConfirmChan <- TradeConfirm{
								TradeOK:		false,
								Message:		"You do not hold enough " +
											exch.commodities[trade.Commodity].Name +
											" to fill the order",
							}
		close(trade.ConfirmChan)
		return
	}
	
	price := trade.Amount * exch.commodities[trade.Commodity].Price
	if (trade.BuySell == "B") && (price > acc.Balance) {
		trade.ConfirmChan <- TradeConfirm{
								TradeOK:		false,
								Message:		"Insufficient funds",
							}
		close(trade.ConfirmChan)
		return
	}
	
	bsMult := int64(1)
	if trade.BuySell == "S" {
		bsMult = -1
	}
	acc = exch.accounts[trade.UserID] // Don't understand why I have to do this
	acc.Balance -= price * bsMult
	exch.accounts[trade.UserID] = acc
	exch.accounts[trade.UserID].Holdings[trade.Commodity] += trade.Amount * bsMult

	trade.ConfirmChan <- TradeConfirm{
						TradeOK:		true,
						Message:		"OK",
						}
	close(trade.ConfirmChan)
	
	exch.calcPriceEffect(trade)
	sendDisplayUpdate(exch)
	
	return
}

func (exch *Exchange)calcPriceEffect(trade Trade) {
	
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	currPrice := exch.commodities[trade.Commodity].Price
	bsMult := int64(1)
	if trade.BuySell == "S" {bsMult = -1}
	suggPrice := int64Max(currPrice + ((r.Int63n(50) - 10) * trade.Amount * bsMult / 4), 1)
	exch.commodities[trade.Commodity].Price = int64Min(suggPrice, currPrice * 10)
}

func (exch *Exchange)Run() {
	
	log.Println("Launching Control handler.")
	go controllerConnHandler(exch)

	log.Println("Launching Display connection handler.")
	go displayConnHandler(exch)
	
	log.Println("Launching Trader connection handler.")
	go traderConnHandler(exch)
	
	
	for exch.status != STOP {
		select {
			
		case command := <-exch.Commands:
			//Handle command
			exch.handleCommand(&command)
			
		case regDispConn := <-exch.DisplayReg:
			// Register new display
			log.Println("Received Display connection request.")
			exch.displays = append(exch.displays, regDispConn)
			sendDisplayUpdate(exch)
			
		case deregDispConn := <-exch.DisplayDeReg:
			// Deregister a display
			log.Println("Received Display disconnection request.")
			// Try and find the connection
			connIndex := -1
			for i, conn := range exch.displays {
				if conn == deregDispConn {
					connIndex = i
					break
				}
			}
			// Remove from the list of display connections and close
			if connIndex == -1 {
				log.Println("Can't find connection.  Not deregistered.")
			} else {
				exch.displays = append(exch.displays[:connIndex], exch.displays[connIndex+1:]...)
				deregDispConn.Close()
				log.Println("Closed display connection at index ", connIndex)
			}
			
		case traderReg := <-exch.TraderRegs:
			// Register a trader
			
			if _, ok := exch.accounts[traderReg.UserID]; ok {
				log.Println("Registration failed. UserID ", traderReg.UserID, " already registered.")
				traderReg.ConfirmChan <- TraderRegConf {
											RegOK:	false,
											Message:	"Registration failed. UserID " +
												     traderReg.UserID +
													" already registered.",
										}
				close(traderReg.ConfirmChan)
			} else {
				
				acc := Account{
						Holdings:	make(map[CommBase]int64),
						Balance:		1000,
					}
				
				for _, c := range exch.commodities {
					acc.Holdings[c.ComType] = 0
				}
				acc.Balance = StartBalance
				
				exch.accounts[traderReg.UserID] = acc
				exch.joinOrder = append(exch.joinOrder, traderReg.UserID)
				
				traderReg.ConfirmChan <- TraderRegConf {
											RegOK:	true,
											Message:	"OK",
										}
				close(traderReg.ConfirmChan)
				
				sendDisplayUpdate(exch)
			}
			
		case deregTraderName := <-exch.TraderDeReg:
			// Deregister a trader
			
			if _, ok := exch.accounts[deregTraderName]; ok {
				delete(exch.accounts, deregTraderName)
				
				foundID := false
				for i, v := range exch.joinOrder {
					if v == deregTraderName {
						foundID = true
						exch.joinOrder = append(exch.joinOrder[:i], exch.joinOrder[i+1:]...)
						break
					}
				}
				if !foundID {
					log.Println("ERROR: exch.joinOrder out of synch. Could not find: ", deregTraderName)
				}
				
				sendDisplayUpdate(exch)
				
			} else {
				log.Println("De-registration failed. UserID ", deregTraderName, " not found.")
			}
			
		case trade := <-exch.Trades:
			// Execute a trade
			
			if exch.status == OPEN {
				exch.handleTrade(trade)
				if exch.accounts[trade.UserID].Balance > WinningBal {
					log.Println(trade.UserID, " is the winner!")
					time.Sleep(250 * time.Millisecond)  // Maybe I need to allow other display updates to run?
					exch.statusMessage = trade.UserID + " is the winner!"
					exch.status = SUSPEND
					sendDisplayUpdate(exch)
				}
			} else {
				reply := TradeConfirm {
							TradeOK:		false,
							Message:		"Trade failed. Exchange not open.",
						}
				trade.ConfirmChan <- reply
				close(trade.ConfirmChan)
			}
			
		case positionReq := <-exch.PositionReqs:
			// Return trading position
			
			pos := PositionUpdate{
					Prices:		make(map[CommBase]int64),
					PriceHist:	exch.priceHist,
					Position:	exch.accounts[positionReq.UserID],
					ExchStatus:	exch.status,
					}
			
			for _, c := range exch.commodities {
				pos.Prices[c.ComType] = c.Price
			}
			
			positionReq.Reply <- pos
			close(positionReq.Reply)
		}
	}
}
