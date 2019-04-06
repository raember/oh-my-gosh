package main

/*
#include <unistd.h>
int go_fork() {
	return fork();
}
*/
import "C"

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"net"
	"os"
)

var certFile = flag.String("cert", common.CERTFILE, "Certificate file")
var keyFile = flag.String("key", common.KEYFILE, "Key file")
var configPath = flag.String("conf", common.CONFIGPATH, "Config path")

func main() {
	flag.Parse()
	log.WithFields(log.Fields{"certFile": *certFile}).Debugln("Certificate file set.")
	log.WithFields(log.Fields{"keyFile": *keyFile}).Debugln("Key file set.")
	log.WithFields(log.Fields{"configPath": *configPath}).Debugln("Config path set.")

	config := server.Config(*configPath)
	lookout, err := server.NewLookout(
		config.GetString("Server.Protocol"),
		config.GetInt("Server.Port"),
	)
	if err != nil {
		os.Exit(1)
	}
	listener, err := lookout.Listen(*certFile, *keyFile)
	if err != nil {
		os.Exit(1)
	}
	defer listener.Close()
	err = server.WaitForConnections(listener, func(conn net.Conn) {
		log.WithFields(log.Fields{
			"remote": common.AddrToStr(conn.RemoteAddr()),
		}).Debugln("Serving new connection.")
		srvr := server.NewServer(config)
		pid := C.go_fork()
		if pid == 0 { // Child
			log.WithField("pid", pid).Debugln("Forked.")
			srvr.Serve(conn, conn, conn)
		} else if pid > 0 { // Parent
			log.WithField("pid", pid).Debugln("Forked off child.")
		} else {
			log.WithField("pid", pid).Errorln("Failed to fork process.")
		}
	})
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)

	//shell.Execute("/bin/bash", os.Stdin, os.Stdout, os.Stderr)
	//
	//srvr := server.NewServer(config)
	//_, user, _ := srvr.PerformLogin(os.Stdin, os.Stdout, os.Stderr)
	//log.Infoln(user)
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
