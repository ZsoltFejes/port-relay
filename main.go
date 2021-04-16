package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

// Random string generator to identify Sockets
func randSeq(n int) string {
	var characters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}

type Command struct {
	NewClient    string `json:"new,omitempty"`
	CloseClient  string `json:"close,omitempty"`
	newLink      bool
	closedClient bool
}

type Link struct {
	mgtSocket     net.Conn
	data          chan Command
	publicSockets map[string]net.Conn
	relaySockets  map[string]net.Conn
}

type LinkManager struct {
	publicAddr      string
	publicListener  net.Listener
	relayAddr       string
	relayLink       Link
	register        chan net.Conn
	unregister      chan string
	relayLinkStatus chan Command
}

func (lm LinkManager) startLinkManager() {
	publicListener, err := net.Listen("tcp", lm.publicAddr)
	checkErr(err)
	lm.publicListener = publicListener
	link := Link{
		publicSockets: make(map[string]net.Conn),
		relaySockets:  make(map[string]net.Conn),
	}
	lm.relayLink = link
	intiLink := Command{newLink: true}
	lm.relayLinkStatus <- intiLink
	select {
	case newClient := <-lm.register:
		fmt.Printf("New Client: %s", newClient.RemoteAddr().String())
		clientID := randSeq(16)
		lm.relayLink.publicSockets[clientID] = newClient
		openNewClient := Command{NewClient: clientID}
		lm.relayLink.data <- openNewClient
	case relayID := <-lm.unregister:
		if _, ok := lm.relayLink.publicSockets[relayID]; ok {
			delete(lm.relayLink.publicSockets, relayID)
			fmt.Println("A connection has been terminated")
		}
	case command := <-lm.relayLinkStatus:
		if command.newLink {
			lm.relayLink = Link{}
			go link.initLinkSocket(&lm)
		}
	}
}

func (l Link) initLinkSocket(lm *LinkManager) {
	relayListener, err := net.Listen("tcp", lm.relayAddr)
	checkErr(err)

	fmt.Printf("Waiting for Relay to connect %s\n", lm.relayAddr)
	linkSocket, err := relayListener.Accept()
	if err != nil {
		fmt.Println(err)
		linkSocket.Close()
	}
	l.mgtSocket = linkSocket
	go lm.listen()
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

func (l Link) listen() {
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

func (l Link) send() {
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

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed random with current time
	manager := LinkManager{
		publicAddr:      ":12345",
		relayAddr:       ":12346",
		register:        make(chan net.Conn),
		unregister:      make(chan string),
		relayLinkStatus: make(chan Command),
	}
	manager.startLinkManager()
}
