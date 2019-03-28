package client

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net"
	"os"
)

type Client struct {
	conn net.Conn
}

func (client Client) Communicate(conn net.Conn) error {
	defer conn.Close()
	log.WithFields(log.Fields{
		"remote": common.AddrToStr(conn.RemoteAddr()),
	}).Debugln("Communicating with host.")
	host := bufio.NewReader(conn)
	input := bufio.NewReader(os.Stdin)

	for {
		// listen for server
		inBound, err := host.ReadString('\n')
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Debugln("Couldn't read from connection.")
		}
		log.WithFields(log.Fields{"inBound": inBound}).Debugln("Received message.")
		log.Println(inBound)

		// read in input from stdin
		log.Print("> ")
		outBound, err := input.ReadString('\n')

		// send to socket
		_, err = conn.Write([]byte(outBound))
		log.WithFields(log.Fields{"outBound": outBound}).Debugln("Sent message.")
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Debugln("Couldn't send message.")
			break
		}
	}
	return nil
}
