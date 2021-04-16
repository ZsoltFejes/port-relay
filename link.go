package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Link struct {
	listener      net.Listener
	mgtSocket     net.Conn
	data          chan Command
	publicSockets map[string]net.Conn
	relaySockets  map[string]net.Conn
}

func (l Link) initLinkSocket(lm *LinkManager) {
	relayListener, err := net.Listen("tcp", lm.relayAddr)
	checkErr(err)
	for {
		fmt.Printf("Waiting for Relay to connect %s\n", lm.relayAddr)
		linkSocket, err := relayListener.Accept()
		if err != nil {
			fmt.Println(err)
			linkSocket.Close()
		}
		decoder := json.NewDecoder(linkSocket)
		data := Command{}
		err = decoder.Decode(&data)
		if err != nil {
			fmt.Println("Error decoding new link's message")
			linkSocket.Close()
		}
		if data.newLink {
			l.mgtSocket = linkSocket
			go lm.listen()
			go l.mgtListen()
			go l.mgtSend()
			go l.startNewRelayListener()
		}
	}

}

func (l Link) mgtListen() {
	command := Command{}
	decoder := json.NewDecoder(l.mgtSocket)
	for {
		err := decoder.Decode(&command)
		if err != nil {
			fmt.Println("Link manager socket read error, closing!")
			close(l.data)
			l.mgtSocket.Close()
			// TODO: Initiate new socet listener
		}
	}
}

func (l Link) mgtSend() {
	encoder := json.NewEncoder(l.mgtSocket)
	for {
		select {
		case newCommand := <-l.data:
			err := encoder.Encode(newCommand)
			if err != nil {
				fmt.Println("Link manager socket write error, closing!")
				close(l.data)
				l.mgtSocket.Close()
			}
		}
	}
}

func (l Link) startNewRelayListener() {
	for {
		newRelay, err := l.listener.Accept()
		if err != nil {
			newRelay.Close()
		}
	}
}

func (l Link) copyIO(relayID string) {
	for {
		_, err := io.Copy(l.publicSockets[relayID], l.relaySockets[relayID])
		if err != nil {
			fmt.Println(err)
		}
		_, err = io.Copy(l.relaySockets[relayID], l.publicSockets[relayID])
		if err != nil {
			fmt.Println(err)
		}
	}
}
