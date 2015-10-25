// controller
package main

import (
	"net"
	"log"
	"io"
	"bufio"
	"sync"
	"strings"
)


type Command struct {
	Command		string
	Argument1	string
	Reply		chan string
}



func controllerConnHandler(exch *Exchange) {
	
	var m sync.Mutex
	controlEstab := false
	
	log.Println("Starting Control listener.")
	server, err := net.Listen("tcp", ":4002")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		
		m.Lock()
		if !controlEstab {
			controlEstab = true
			m.Unlock()
			go controller(exch, conn, &controlEstab, &m)
		} else {
			m.Unlock()
			io.WriteString(conn, "Controller already logged in. Exiting.\r\n")
			conn.Close()
		}
	}
	log.Println("Closing Control listener")
}

func controlPrompt(exch *Exchange, conn net.Conn) {
	
	io.WriteString(conn, "Status: " + statusString(exch.status) + "\r\n")
	
	io.WriteString(conn, "Available commands:\r\n")
	switch exch.status {
	case SUSPEND:
		io.WriteString(conn, "START | STOP | ROBOT <n> | DIFF <E|M|H>\r\n")
	case OPEN:
		io.WriteString(conn, "PAUSE | ROBOT <n> | DIFF <E|M|H>\r\n")
	default:
		panic("Invalid Exchange Status")
	}
}


func controller(exch *Exchange, conn net.Conn, controlEstab *bool, m *sync.Mutex) {
	
	defer conn.Close()
	
	log.Println("Starting Control terminal session.")
	scanner := bufio.NewScanner(conn)
	
	io.WriteString(conn, "Control console ready.\r\n")
	controlPrompt(exch, conn)

	for scanner.Scan() {
		
		ln := strings.ToUpper(scanner.Text())
		fs := strings.Fields(ln)
		if len(fs) > 0 {
			cmd := ""
			arg1 := ""
			if len(fs) == 1 {
				cmd = fs[0]
				arg1 = "1"
			} else {
				cmd = fs[0]
				arg1 = fs[1]
			}
		
			command := Command{cmd, arg1, make(chan string)}
			exch.Commands <- command
			io.WriteString(conn, <-command.Reply)
			controlPrompt(exch, conn)
		}
	}
	
	log.Println("Closing Control terminal session.")
	m.Lock()
	*controlEstab = false
	m.Unlock()
}


