package main

import (
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
	NewClient   string `json:"new,omitempty"`
	CloseClient string `json:"close,omitempty"`
	NewLink     bool   `json:"newLink,omitempty"`
	newLink     bool
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
