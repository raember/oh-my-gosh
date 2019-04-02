package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/client"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"os"
)

var configPath = flag.String("conf", common.CONFIGPATH, "Config path")

func main() {
	flag.Parse()
	address := flag.Arg(0)
	if address == "" {
		address = common.LOCALHOST
	}
	log.WithFields(log.Fields{"configPath": *configPath}).Debugln("Config path set.")
	log.WithFields(log.Fields{"address": address}).Debugln("Host address set.")
	config := client.Config(*configPath)
	dialer, err := client.NewDialer(
		config.GetString("Client.Protocol"),
		address,
		config.GetInt("Client.Port"),
	)
	if err != nil {
		os.Exit(1)
	}
	conn, err := dialer.Dial()
	if err != nil {
		os.Exit(1)
	}
	clnt := client.Client{}
	err = clnt.Communicate(conn)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
