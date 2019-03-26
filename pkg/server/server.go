package server

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"net"
)

type Server struct {
	conn       net.Conn
	readwriter *bufio.ReadWriter
}

func (server Server) Serve() {
	log.WithFields(log.Fields{
		"remote": server.conn.RemoteAddr(),
	}).Infoln("Serving new connection.")
	server.readwriter = bufio.NewReadWriter(bufio.NewReader(server.conn), bufio.NewWriter(server.conn))
	defer server.conn.Close()
}

//func TextReplyLocal(protocol string, port int) {
//	server, err := NewServer(protocol, port)
//	if err != nil {
//		syscall.Exit(1)
//	}
//	log.Infoln("Initialized server.")
//
//	ln, err := server.Listen()
//	if err != nil {
//		log.Errorln(err.Error())
//		syscall.Exit(1)
//	}
//	log.Infoln("Listening for incoming requests on " + server.address.String())
//
//	// accept connection on port
//	conn, _ := ln.Accept()
//	log.WithFields(log.Fields{
//		"remoteaddress": conn.RemoteAddr(),
//	}).Infoln("Connection established.")
//
//	login.Authenticate(func(s pam.Style, message string) (string, error) {
//		switch s {
//		case pam.PromptEchoOff:
//			//return speakeasy.Ask(message + " \n")
//			message = message + " \n"
//			log.WithFields(log.Fields{
//				"message": message,
//			}).Debugln("Outbound")
//			_, _ = conn.Write([]byte(message))
//			//fmt.Print(message + " ")
//			input, err := bufio.NewReader(conn).ReadString('\n')
//			if err != nil {
//				log.Infoln("Connection died.")
//				return "", err
//			}
//			log.WithFields(log.Fields{
//				"input": input,
//			}).Debugln("Inbound")
//			return input[:len(input)-1], nil
//		case pam.PromptEchoOn:
//			message = message + " \n"
//			log.WithFields(log.Fields{
//				"message": message,
//			}).Debugln("Outbound")
//			_, _ = conn.Write([]byte(message))
//			//fmt.Print(message + " ")
//			input, err := bufio.NewReader(conn).ReadString('\n')
//			if err != nil {
//				log.Infoln("Connection died.")
//				return "", err
//			}
//			log.WithFields(log.Fields{
//				"input": input,
//			}).Debugln("Inbound")
//			return input[:len(input)-1], nil
//		case pam.ErrorMsg:
//			message = message + " \n"
//			log.WithFields(log.Fields{
//				"message": message,
//			}).Debugln("Outbound")
//			_, _ = conn.Write([]byte(message))
//			return "", nil
//		case pam.TextInfo:
//			message = message + " \n"
//			log.WithFields(log.Fields{
//				"message": message,
//			}).Debugln("Outbound")
//			_, _ = conn.Write([]byte(message))
//			return "", nil
//		}
//		return "", errors.New("unrecognized message style")
//	})
//
//	message := "Authentication was successful\n"
//	log.WithFields(log.Fields{
//		"message": message,
//	}).Infoln("Outbound")
//	_, _ = conn.Write([]byte(message))
//
//	// run loop forever (or until ctrl-c)
//	for {
//		// will listen for message to process ending in newline (\n)
//		message, err := bufio.NewReader(conn).ReadString('\n')
//		if err != nil {
//			log.Infoln("Connection died.")
//			break
//		}
//		// output message received
//		log.WithFields(log.Fields{
//			"message": message,
//		}).Infoln("Inbound")
//		// sample process for string received
//		answer := strings.ToUpper(message)
//		// send new string back to client
//		log.WithFields(log.Fields{
//			"answer": answer,
//		}).Infoln("Outbound")
//		_, _ = conn.Write([]byte(answer))
//	}
//}
