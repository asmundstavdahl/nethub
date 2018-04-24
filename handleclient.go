package main

import (
	"container/list"
	"fmt"
	"log"
	"net"
)

var clientChannels list.List

func handleConnection(conn net.Conn, broadcast chan *Packet) {
	defer conn.Close()

	myChannel := make(chan []byte)
	myElement := clientChannels.PushBack(myChannel)
	defer func() {
		verbose("Removing channel", myChannel)
		clientChannels.Remove(myElement)
	}()

	open := make(chan bool, 2)
	open <- true

	go func(conn net.Conn, c chan []byte) {
		for {
			conn.Write(<-c)
		}
	}(conn, myChannel)

	for <-open {
		verbose("Link still open", myChannel, "\n")
		buf := make([]byte, maxReadBytes)
		n, err := conn.Read(buf)

		if err != nil {
			verbose("Connection closed by client (", conn.RemoteAddr(), ") (", err, ")")
			verbose("Marking link", myChannel, "as closed")
			open <- false
		} else {
			open <- true
			pak := NewPacket(myChannel, buf[:n])
			pak.Broadcast()
		}
	}
}

func acceptClients(end chan bool, broadcast chan *Packet) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalln("Failed to listen on port", port, ":", err)
		end <- true
	} else {
		verbose("Listening...")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			verbose("Accepting connection failed:", err)
			continue
		} else {
			go handleConnection(conn, broadcast)
			verbose("Got client, sent to handler")
		}
	}

	end <- true
}
