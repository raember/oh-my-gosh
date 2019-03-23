package server

import (
	"../common"
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"syscall"
)

type Server struct {
	address  *url.URL
	listener net.Listener
}

func NewServer(protocol string, port int) (Server, error) {
	if port < 0 {
		return Server{}, errors.New("port cannot be negative")
	}
	addrStr := protocol + "://" + common.LOCALHOST + ":" + strconv.Itoa(port)
	reqUrl, err := url.ParseRequestURI(addrStr)
	if err != nil {
		return Server{}, err
	}
	if reqUrl.Scheme != common.TCP && reqUrl.Scheme != common.TCP4 {
		return Server{}, errors.New("protocol has to be either tcp or udp")
	}
	return Server{address: reqUrl}, nil
}

func (server Server) Listen() (net.Listener, error) {
	listener, err := net.Listen(server.address.Scheme, server.address.Host)
	if err != nil {
		return nil, err
	}
	server.listener = listener
	return listener, nil
}

func TextReplyLocal(protocol string, port int) {
	server, err := NewServer(protocol, port)
	if err != nil {
		fmt.Println(err.Error())
		syscall.Exit(1)
	}
	fmt.Println("Initialized server.")

	ln, err := server.Listen()
	if err != nil {
		fmt.Println(err.Error())
		syscall.Exit(1)
	}
	fmt.Println("Listening for incoming requests on " + server.address.String())

	// accept connection on port
	conn, _ := ln.Accept()
	fmt.Println("Connection established.")

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Connection died.")
			break
		}
		// output message received
		fmt.Print("<- " + message)
		// sample process for string received
		newmessage := strings.ToUpper(message)
		// send new string back to client
		_, _ = conn.Write([]byte(newmessage + "\n"))
	}
}
