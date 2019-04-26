package main

import (
	"bufio"
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/client"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"golang.org/x/crypto/ssh/terminal"
	"os"
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
	//err = client.PerformLogin(conn, conn)
	//if err != nil {
	//	os.Exit(1)
	//}
	//err = client.PerformEnvTransfer(conn, conn)
	//if err != nil {
	//	os.Exit(1)
	//}
	// TODO: Test
	go connection.Forward(os.Stdin, conn, "stdin", "server")
	//go connection.Forward(conn, os.Stdout, "server", "stdout")
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.WithError(err).Fatalln("Couldn't set terminal into raw mode.")
	}
	defer func() {
		err = terminal.Restore(int(os.Stdin.Fd()), oldState)
		if err != nil {
			log.WithError(err).Fatalln("Couldn't set terminal into raw mode.")
		}
	}()

	n, err := bufio.NewReader(conn).WriteTo(os.Stdout)
	if err != nil {
		log.WithError(err).Errorln("Couldn't write from stdin to server.")
		return
	}
	log.WithField("n", n).Debugln("Wrote from stdin to server.")
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
