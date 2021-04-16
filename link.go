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

func (l Link) initLink(lm *LinkManager) {
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
		if data.NewLink {
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
		go l.initNewRelay(newRelay)
	}
}

func (l Link) copyIO(relayID string, public bool) {
	for {
		if public {
			_, err := io.Copy(l.publicSockets[relayID], l.relaySockets[relayID])
			if err != nil {
				fmt.Println(err)
				// TODO: Think about closing the socket, server should never close
			}
		} else {
			_, err := io.Copy(l.relaySockets[relayID], l.publicSockets[relayID])
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (l Link) initNewRelay(socket net.Conn) {
	data := Command{}
	decoder := json.NewDecoder(socket)
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Printf("New Relay Decoding error: %s", err)
	}
	if _, ok := l.publicSockets[data.Connecting]; ok {
		l.relaySockets[data.Connecting] = socket
		go l.copyIO(data.Connecting, true)
		go l.copyIO(data.Connecting, false)
	}
}
