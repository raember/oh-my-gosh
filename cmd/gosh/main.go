package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/client"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/utils"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

func main() {
	log.WithField("args", os.Args).Traceln("--> gosh.main")
	configPath := flag.String("conf", common.CONFIGPATH, "Config path.")
	authPath := flag.String("auth", common.AUTHPATH, "Authorized keys path.")

	flag.Parse()
	log.WithFields(log.Fields{
		"configPath": *configPath,
		"authPath":   *authPath,
	}).Debugln("Parsed arguments.")

	config := client.LoadConfig(*configPath)
	config.Set("Authentication.KeyStore", *authPath)
	clnt := client.NewClient(config)

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
	defer utils.CloseConn(conn)
	err = clnt.PerformTransfer(conn, conn)
	if err != nil {
		os.Exit(1)
	}
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

	go utils.Forward(os.Stdin, conn, "stdin", "server")
	utils.Forward(conn, os.Stdout, "server", "stdout")
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
