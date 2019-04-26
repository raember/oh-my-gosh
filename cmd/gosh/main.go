package main

import (
	"bufio"
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/client"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strconv"
	"time"
)

var configPath = flag.String("conf", common.CONFIGPATH, "Config path")

func main() {
	log.Traceln("--> gosh.main")
	flag.Parse()
	address := flag.Arg(0)
	if address == "" {
		address = common.LOCALHOST
	}
	log.WithFields(log.Fields{"configPath": *configPath}).Debugln("Config path set.")
	log.WithFields(log.Fields{"address": address}).Debugln("Host address set.")
	config := client.Config(*configPath)
	clnt, err := client.NewClient(
		config.GetString("Client.Protocol"),
		address,
		config.GetInt("Client.Port"),
	)
	if err != nil {
		os.Exit(1)
	}
	conn, err := clnt.Dial()
	if err != nil {
		os.Exit(1)
	}
	err = client.PerformLogin(conn, conn)
	if err != nil {
		os.Exit(1)
	}
	err = client.PerformEnvTransfer(conn, conn)
	if err != nil {
		os.Exit(1)
	}
	// TODO: Test
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.WithError(err).Errorln("Couldn't set terminal into raw mode.")
		os.Exit(1)
	}

	go func() {
		bufIn := bufio.NewReader(conn)
		n, err := bufIn.WriteTo(os.Stdout)
		if err != nil {
			log.WithError(err).Errorln("Couldn't read from connection.")
			os.Exit(1)
		}
		if n > 0 {
			log.Debugln("Read " + strconv.Itoa(int(n)) + " bytes from server.")
		}
		err = terminal.Restore(int(os.Stdin.Fd()), oldState)
		if err != nil {
			log.WithError(err).Errorln("Couldn't cook terminal.")
			os.Exit(1)
		}
		os.Exit(0)
	}()
	go func() {
		bufIn := bufio.NewReader(os.Stdin)
		n, err := bufIn.WriteTo(conn)
		if err != nil {
			log.WithError(err).Errorln("Couldn't read from stdin.")
			os.Exit(1)
		}
		if n > 0 {
			log.Debugln("Written " + strconv.Itoa(int(n)) + " bytes to server.")
		}
		os.Exit(0)
	}()
	for {
		time.Sleep(time.Second)
	}
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
