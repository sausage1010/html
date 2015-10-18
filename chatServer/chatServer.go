// redisServer
package main

import (
	"fmt"
	"net"
	"bufio"
//	"strings"
	"io"
	"log"
)

var clients int


type Message struct {
	Text string
	User string
}
//type Board struct {
//	Result chan string
//}

func chatServer(boardIn chan Message, boardOut chan Message) {
	for message := range boardIn {
		fmt.Println("Got message: " + message.User + " : " + message.Text)
		for i := 0; i < clients; i++ {
			boardOut <- message
		}
	}

/*
	var data = make(map[string]string)
	for cmd := range commands {
		if len(cmd.Fields) < 2 {
			cmd.Result <- "Expected at least two arguments"
			continue
		}
		
		fmt.Println("GOT COMMAND", cmd)
		
		switch cmd.Fields[0] {
		// GET <KEY>
		case "GET":Board
			key := cmd.Fields[1]
			value := data[key]
			
			cmd.Result <- value
			
		// SET <KEY> <VALUE>
		case "SET":
			if len(cmd.Fields) != 3 {
				cmd.Result <- "EXPECTED A VALUE"
				continue
			}
			key := cmd.Fields[1]
			value := cmd.Fields[2]
			data[key] = value
			cmd.Result <- ""
			
		//DEL <KEY>
		case "DEL":
			key := cmd.Fields[1]
			delete(data, key)
			cmd.Result <- ""
		
		default:
			cmd.Result <- "INVALID COMMAND" + cmd.Fields[0] + "\n"
		}
	}
*/
}

func boardRepeater(boardOut chan Message, conn net.Conn, name string) {
	for message := range boardOut {
		if message.User != name {
			io.WriteString(conn, message.User + " : " + message.Text +"\n")
		}
	}
}

func handleClient(boardIn chan Message, boardOut chan Message, conn net.Conn, name string) {
	defer conn.Close()
	
	go boardRepeater(boardOut, conn, name)
	
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var message Message
		message.Text = scanner.Text()
		message.User = name
//		fs := strings.Fields(ln)
		
//		result := make(chan string)
//		commands <- Command {
//			Fields: fs,
//			Result: result,
//		}
		
		boardIn <- message
//		io.WriteString(conn, <-result+"\n")
	}
}

func main() {
	li, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal(err)
	}
	defer li.Close()
	
	boardIn := make(chan Message)
	boardOut := make(chan Message)
	go chatServer(boardIn, boardOut)
	clients = 0
	
	for {
		conn, err := li.Accept()
		if err != nil {
			log.Fatal(err)
		}
		
		clients++
		go handleClient(boardIn, boardOut, conn, fmt.Sprintf("User %v", clients))
	}
}
