package main

import (
	"fmt"
	"net"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

type Link struct {
	mgtSocket   net.Conn
	dataSockets map[int]net.Conn
}

type LinkManager struct {
	publicAddr     string
	publicListener net.Listener
	relayAddr      string
	relayLink      Link
}

func main() {

}
