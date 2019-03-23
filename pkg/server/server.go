package server

import (
	"bufio"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
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

func NewServer(protocol string, port int) (*Server, error) {
	if port < 0 {
		err := errors.New("port cannot be negative")
		log.WithFields(log.Fields{
			"port": port,
		}).Errorln(err.Error())
		return nil, err
	}
	addrStr := protocol + "://" + common.LOCALHOST + ":" + strconv.Itoa(port)
	reqUrl, err := url.ParseRequestURI(addrStr)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": protocol,
			"port":     port,
			"addrStr":  addrStr,
		}).Errorln(err.Error())
		return nil, err
	}
	if reqUrl.Scheme != common.TCP && reqUrl.Scheme != common.TCP4 {
		err := errors.New("protocol has to be either tcp or udp")
		log.WithFields(log.Fields{
			"protocol": reqUrl.Scheme,
		}).Errorln(err.Error())
		return nil, err
	}
	return &Server{address: reqUrl}, nil
}

func (server Server) Listen() (net.Listener, error) {
	listener, err := net.Listen(server.address.Scheme, server.address.Host)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": server.address.Scheme,
			"host":     server.address.Host,
		}).Errorln(err.Error())
		return nil, err
	}
	server.listener = listener
	return listener, nil
}

func (server Server) StartListening(callback func(conn net.Conn)) error {
	listener, err := net.Listen(server.address.Scheme, server.address.Host)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": server.address.Scheme,
			"host":     server.address.Host,
		}).Errorln(err.Error())
		return err
	}
	server.listener = listener
	log.WithFields(log.Fields{
		"address": server.address.String(),
	}).Infoln("Listening for incoming requests.")
	return nil
}

func TextReplyLocal(protocol string, port int) {
	server, err := NewServer(protocol, port)
	if err != nil {
		syscall.Exit(1)
	}
	log.Infoln("Initialized server.")

	ln, err := server.Listen()
	if err != nil {
		log.Errorln(err.Error())
		syscall.Exit(1)
	}
	log.Infoln("Listening for incoming requests on " + server.address.String())

	// accept connection on port
	conn, _ := ln.Accept()
	log.Infoln("Connection established.")

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Infoln("Connection died.")
			break
		}
		// output message received
		log.WithFields(log.Fields{
			"rawtext": message,
		}).Infoln("Inbound")
		// sample process for string received
		answer := strings.ToUpper(message)
		// send new string back to client
		log.WithFields(log.Fields{
			"rawtext": answer,
		}).Infoln("Outbound")
		_, _ = conn.Write([]byte(answer))
	}
}
