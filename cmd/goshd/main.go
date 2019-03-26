package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"net"
	"os"
	"strconv"
)

func main() {
	config := server.Config()
	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatalln("Couldn't load certificates.")
		return
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	protocol := config.GetString("Server.Protocol")
	address := ":" + strconv.Itoa(config.GetInt("Server.Port"))
	listener, err := tls.Listen(protocol, address, tlsConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": protocol,
			"address":  address,
			"error":    err.Error(),
		}).Fatalln("Couldn't setup listener.")
		return
	}
	defer listener.Close()
	log.WithFields(log.Fields{
		"address": listener.Addr(),
	}).Infoln("Listening for incoming requests.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalln("Failed accepting connection request.")
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.WithFields(log.Fields{
		"remote": conn.RemoteAddr(),
	}).Infoln("Connection established.")
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalln("Failed to read.")
			return
		}

		log.Println(msg)

		input := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		message, err := input.ReadString('\n')
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalln("Failed to read from input.")
			break
		}
		_, err = conn.Write([]byte(message))
		if err != nil {
			log.WithFields(log.Fields{
				"message": message,
				"error":   err.Error(),
			}).Fatalln("Failed to write to peer.")
			break
		}
	}
}
