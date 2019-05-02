package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/client"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

var configPath = flag.String("conf", common.CONFIGPATH, "Config path.")
var authPath = flag.String("auth", common.AUTHPATH, "Path of the public and private keys.")

func main() {
	log.Traceln("--> gosh.main")
	flag.Parse()

	log.WithField("configPath", *configPath).Debugln("Config path set.")
	log.WithField("authPath", *authPath).Debugln("Auth path set.")

	clnt := client.NewClient(client.LoadConfig(*configPath))
	if err := clnt.ParseArgument(flag.Arg(0)); err != nil {
		os.Exit(1)
	}
	if err := clnt.Setup(); err != nil {
		os.Exit(1)
	}
	conn, err := clnt.Dial()
	if err != nil {
		os.Exit(1)
	}
	defer connection.CloseConn(conn)
	err = client.PerformEnvTransfer(conn, conn)
	if err != nil {
		os.Exit(1)
	}
	go connection.Forward(os.Stdin, conn, "stdin", "server")
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.WithError(err).Fatalln("Failed to set terminal into raw mode.")
	} else {
		log.WithField("oldState", oldState).Debugln("Set terminal into raw mode.")
	}
	defer func() {
		err = terminal.Restore(int(os.Stdin.Fd()), oldState)
		if err != nil {
			log.WithError(err).Fatalln("Failed to set terminal into cooked mode.")
		}
	}()

	connection.Forward(conn, os.Stdout, "server", "stdout")
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
