package server

import (
	"bufio"
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"net"
	"strings"
)

type Server struct {
	conn net.Conn
}

func (server Server) Serve(conn net.Conn) {
	defer conn.Close()
	log.WithFields(log.Fields{
		"remote": common.AddrToStr(conn.RemoteAddr()),
	}).Debugln("Serving new connection.")
	client := bufio.NewReader(conn)

	transaction, err := login.Authenticate(func(style pam.Style, outBound string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			//return speakeasy.Ask(outBound)
			outBound += " \n"
			_, _ = conn.Write([]byte(outBound))
			inBound, err := client.ReadString('\n')
			if err != nil {
				log.Infoln("Connection died.")
				return "", err
			}
			log.WithFields(log.Fields{"inBound": inBound}).Debugln("Received message.")
			return inBound[:len(inBound)-1], nil
		case pam.PromptEchoOn:
			outBound += " \n"
			_, _ = conn.Write([]byte(outBound))
			log.WithFields(log.Fields{"outBound": outBound}).Debugln("Sent message.")
			inBound, err := client.ReadString('\n')
			if err != nil {
				log.Infoln("Connection died.")
				return "", err
			}
			log.WithFields(log.Fields{"inBound": inBound}).Debugln("Received message.")
			return inBound[:len(inBound)-1], nil
		case pam.ErrorMsg:
			_, _ = conn.Write([]byte(outBound))
			log.WithFields(log.Fields{"outBound": outBound}).Debugln("Sent message.")
			return "", nil
		case pam.TextInfo:
			_, _ = conn.Write([]byte(outBound))
			log.WithFields(log.Fields{"outBound": outBound}).Debugln("Sent message.")
			return "", nil
		}
		return "", errors.New("unrecognized outBound style")
	})
	if err != nil {
		return
	}
	defer transaction.CloseSession(0)

	message := "Authentication was successful"
	log.WithFields(log.Fields{
		"message": message,
	}).Infoln("Outbound")
	_, _ = conn.Write([]byte(message + "\n"))

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := client.ReadString('\n')
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
		_, _ = conn.Write([]byte(answer + "\n"))
	}
}
