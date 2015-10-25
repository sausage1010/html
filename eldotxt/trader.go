// trader
package main

import (
	"net"
	"io"
	"bufio"
	"strings"
	"log"
	"strconv"
	"fmt"
)

type Trade struct {
	UserID		string
	BuySell		string
	Commodity	CommBase
	Amount		int64
	ConfirmChan	chan TradeConfirm
}

type TradeConfirm struct {
	TradeOK		bool
	Message		string
}

type TraderReg struct {
	UserID		string
	ConfirmChan	chan TraderRegConf
}

type TraderRegConf struct {
	RegOK		bool
	Message		string
}


func traderConnHandler(exch *Exchange) {
	
	log.Println("Starting Trader listener.")
	server, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		
		go traderHandler(exch, conn)
	}
	log.Println("Closing Trader listener")

}

func validateComm(c string) (CommBase, bool) {
	
	var tradeCom CommBase
	validComm := true

	switch c {
	case "C":
		tradeCom = COPPER
	case "G":
		tradeCom = GOLD
	case "S":
		tradeCom = SILVER
	case "Z":
		tradeCom = ZINC
	default:
		validComm = false
	}
	
	return tradeCom, validComm
}

func getTradeAmount(e *Exchange, c CommBase, userName string, BS string, divisor int64) int64 {
	req := PositionReq{
				UserID:	userName,
				Reply:	make(chan PositionUpdate),
			}
	e.PositionReqs <- req
	resp := <- req.Reply
	
	if BS == "B" {
		return (resp.Position.Balance / resp.Prices[c]) / divisor
	} else {
		return resp.Position.Holdings[c] / divisor
	}
}


func manageCommandLine(exch *Exchange, conn net.Conn, scanner *bufio.Scanner, userName string) {
	
	helpStr := "Commands:-\r\n" +
	           "[B|S] [C|G|S|Z] <Amount>  - Buy or Sell commodities.\r\n" +
               "                            If no amount, buy or sell max amount.\r\n" +
               "U                         - Get an account update.\r\n" +
               "H                         - Help. Print this again.\r\n\n"
	
	io.WriteString(conn, helpStr)
	
	for scanner.Scan() {
		ln := strings.ToUpper(scanner.Text())
		fs := strings.Fields(ln)

		var tradeCom CommBase
		var validOrder bool
		
		if len(fs) > 0 {
			switch fs[0] {
			case "B", "S":
				if len(fs) < 2 {
					io.WriteString(conn, "Invalid order. Please specify commodity.\r\n")
						
				} else if len(fs) <= 3 {  // Valid number of arguments
					
					tradeCom, validOrder = validateComm(fs[1])
					
					if validOrder {
						
						var amt int64
						
						if len(fs) == 2 {  // Buy or Sell max							

							amt = getTradeAmount(exch, tradeCom, userName, fs[0], 1)							

						} else { // An amount has been specified (possibly 'H' for half
							
							if fs[2] == "H" { // Buy or Sell half of max
								amt = getTradeAmount(exch, tradeCom, userName, fs[0], 2)
							} else { // Get user specified amount
								var numErr error
								amt, numErr = strconv.ParseInt(fs[2], 10, 64)
								if numErr != nil {
									io.WriteString(conn, "Invalid order. Could not read amount.\r\n")
									break
								}
							}
						}
						
						order := Trade{
									UserID:		userName,
									BuySell:		fs[0],
									Commodity:	tradeCom,
									Amount:		amt,
									ConfirmChan:	make(chan TradeConfirm),
						}
						
						exch.Trades <- order
						trConf := <- order.ConfirmChan
						io.WriteString(conn, trConf.Message + "\r\n")
					} else {
						io.WriteString(conn, "Invalid order. Please specify a valid commodity.\r\n")
					}
					
					
/*				} else if len(fs) == 3 {
					
					tradeCom, validOrder = validateComm(fs[1])
					
					if validOrder {
						
						amt, numErr := strconv.ParseInt(fs[2], 10, 64)
						
						if numErr == nil {
							order := Trade{
										UserID:		userName,
										BuySell:		fs[0],
										Commodity:	tradeCom,
										Amount:		amt,
										ConfirmChan:	make(chan TradeConfirm),
							}
							
							exch.Trades <- order
							trConf := <- order.ConfirmChan
							io.WriteString(conn, trConf.Message + "\n")
						} else {
							io.WriteString(conn, "Invalid order. Could not read amount.\n")
						}
						
					} else {
						io.WriteString(conn, "Invalid order. Please specify a valid commodity.\n")
					}
*/
				} else {
					io.WriteString(conn, "Invalid order. Too many arguments.\r\n")
				}
			
			case "U":
				req := PositionReq{
							UserID:	userName,
							Reply:	make(chan PositionUpdate),
						}
				exch.PositionReqs <- req
				resp := <-req.Reply
				
				displayStr := "Exchange Status: " + statusString(resp.ExchStatus) + "\r\n\n"
				titleStr :=  "    Trader | "
				pricesStr := "           | "
				
				// Create headers
				for _, c := range exch.commodities {
					titleStr += fmt.Sprintf("%10s | ", c.Name)
					pricesStr += fmt.Sprintf("%10d | ", resp.Prices[c.ComType])
				}
				titleStr += fmt.Sprintf("%10s |", "Balance")
				displayStr += titleStr + "\r\n" +
				              pricesStr +"\r\n" +
						      uLine + "\r\n"
				
				// List accounts
				accStr := fmt.Sprintf("%10s | ", userName)
				for _, c := range exch.commodities {
					accStr += fmt.Sprintf("%10d | ", resp.Position.Holdings[c.ComType])
				}
				accStr += fmt.Sprintf("%10d |", resp.Position.Balance)
				displayStr += accStr + "\r\n"
				displayStr += uLine +"\r\n"
				
				io.WriteString(conn, displayStr)
			
			case "H":
				io.WriteString(conn, helpStr)	
			
			default:
				io.WriteString(conn, "Invalid command.\r\n")
			}
		}
	}
}

func traderHandler(exch *Exchange, conn net.Conn) {
	
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	userName := ""
	for userName == "" {
		io.WriteString(conn, "Enter your Username: ")
	
		scanner.Scan()
		userName = strings.TrimSpace(scanner.Text())
		if len(userName) > 10 {userName = userName[:10]}
	}
	
	reg := TraderReg{
		UserID:   		userName,
		ConfirmChan:		make(chan TraderRegConf),
	}
	
	exch.TraderRegs <- reg
	conf := <-reg.ConfirmChan
	if conf.RegOK {
		defer func() {
			exch.TraderDeReg <- reg.UserID
		}()
	} else {
		io.WriteString(conn, conf.Message + "\r\n")
		return
	}
	
	manageCommandLine(exch, conn, scanner, userName)
	
}
