package server

import (
	"bufio"
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"net"
	"strings"
)

type Server struct {
	conn net.Conn
}

func NewServer(conn net.Conn) *Server {
	return &Server{conn: conn}
}

func (server Server) Serve() {
	conn := server.conn
	defer conn.Close()
	log.WithFields(log.Fields{
		"remote": conn.RemoteAddr(),
	}).Debugln("Serving new connection.")
	readwriter := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	transaction, err := login.Authenticate(func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			//return speakeasy.Ask(message + " \n")
			message = message
			log.WithFields(log.Fields{
				"message": message,
			}).Debugln("Outbound")
			_, _ = readwriter.WriteString(message + " \n")
			//fmt.Print(message + " ")
			input, err := readwriter.ReadString('\n')
			if err != nil {
				log.Infoln("Connection died.")
				return "", err
			}
			log.WithFields(log.Fields{
				"input": input,
			}).Debugln("Inbound")
			return input[:len(input)-1], nil
		case pam.PromptEchoOn:
			message = message
			log.WithFields(log.Fields{
				"message": message,
			}).Debugln("Outbound")
			_, _ = readwriter.WriteString(message + " \n")
			//fmt.Print(message + " ")
			input, err := readwriter.ReadString('\n')
			if err != nil {
				log.Infoln("Connection died.")
				return "", err
			}
			log.WithFields(log.Fields{
				"input": input,
			}).Debugln("Inbound")
			return input[:len(input)-1], nil
		case pam.ErrorMsg:
			message = message
			log.WithFields(log.Fields{
				"message": message,
			}).Debugln("Outbound")
			_, _ = readwriter.WriteString(message + " \n")
			return "", nil
		case pam.TextInfo:
			message = message
			log.WithFields(log.Fields{
				"message": message,
			}).Debugln("Outbound")
			_, _ = readwriter.WriteString(message + " \n")
			return "", nil
		}
		return "", errors.New("unrecognized message style")
	})
	if err != nil {
		return
	}
	defer transaction.CloseSession(0)

	message := "Authentication was successful"
	log.WithFields(log.Fields{
		"message": message,
	}).Infoln("Outbound")
	_, _ = readwriter.WriteString(message + "\n")

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := readwriter.ReadString('\n')
		if err != nil {
			log.Infoln("Connection died.")
			break
		}
		// output message received
		log.WithFields(log.Fields{
			"message": message,
		}).Infoln("Inbound")
		// sample process for string received
		answer := strings.ToUpper(message)
		// send new string back to client
		log.WithFields(log.Fields{
			"answer": answer,
		}).Infoln("Outbound")
		_, _ = readwriter.WriteString(answer + "\n")
	}
}
