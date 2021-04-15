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

type Link struct {
	mgtSocket   net.Conn
	dataSockets map[string]net.Conn
}

type LinkManager struct {
	publicAddr     string
	publicListener net.Listener
	relayAddr      string
	relayLink      Link
}

func main() {
	rand.Seed(time.Now().UnixNano())

}
