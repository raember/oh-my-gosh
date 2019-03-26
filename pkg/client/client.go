package client

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
)

type Client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn}
}

func (client Client) Communicate() error {
	conn := client.conn
	defer conn.Close()
	log.WithFields(log.Fields{
		"remote": conn.RemoteAddr(),
	}).Debugln("Communicating with server.")
	readwriter := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		// listen for server
		answer, err := readwriter.ReadString('\n')
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Debugln("Couldn't read from connection.")
		}
		log.WithFields(log.Fields{
			"rawtext": answer,
		}).Debugln("Inbound")
		log.Println(answer)

		// read in input from stdin
		input := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		message, err := input.ReadString('\n')
		log.WithFields(log.Fields{
			"rawtext": message,
		}).Debugln("Outbound")
		// send to socket
		_, err = readwriter.WriteString(message + "\n")
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Debugln("Couldn't send message.")
			break
		}
	}
	return nil
}
