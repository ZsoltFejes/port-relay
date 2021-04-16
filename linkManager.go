package main

import (
	"fmt"
	"net"
)

type LinkManager struct {
	publicAddr      string
	publicListener  net.Listener
	relayAddr       string
	relayLink       *Link
	register        chan net.Conn
	unregister      chan string
	relayLinkStatus chan Command
}

func (lm LinkManager) startLinkManager() {
	publicListener, err := net.Listen("tcp", lm.publicAddr)
	checkErr(err)
	lm.publicListener = publicListener
	lm.relayLink = &Link{
		publicSockets: make(map[string]net.Conn),
		relaySockets:  make(map[string]net.Conn),
	}
	go lm.relayLink.initLink(&lm)
	select {
	case newClient := <-lm.register:
		fmt.Printf("New Client: %s", newClient.RemoteAddr().String())
		clientID := randSeq(16)
		lm.relayLink.publicSockets[clientID] = newClient
		openNewClient := Command{NewClient: clientID}
		lm.relayLink.data <- openNewClient
	case relayID := <-lm.unregister:
		if _, ok := lm.relayLink.publicSockets[relayID]; ok {
			closeClient := Command{CloseClient: relayID}
			lm.relayLink.data <- closeClient
			delete(lm.relayLink.publicSockets, relayID)
			delete(lm.relayLink.relaySockets, relayID)
			fmt.Println("A connection has been terminated")
		}
	case command := <-lm.relayLinkStatus:
		if command.NewLink {
			lm.relayLink = &Link{
				publicSockets: make(map[string]net.Conn),
				relaySockets:  make(map[string]net.Conn),
			}
			go lm.relayLink.initLink(&lm)
		}
	}
}

func (lm LinkManager) listen() {
	for {
		newClient, err := lm.publicListener.Accept()
		if err != nil {
			fmt.Println(err)
			newClient.Close()
		}
		lm.register <- newClient
	}
}
