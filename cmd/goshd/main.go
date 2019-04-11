package main

import "C"
import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/proc"
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
		// TODO: Fix usage corruption of conn struct after forking.

		proc.SetIOContext(2, &conn)
		pid, err := proc.Fork()
		if err != nil {
			return
		}
		if pid == 0 { // Child
			srvr := server.NewServer(config)
			err = srvr.Serve(conn, conn, conn)
			if err != nil {
				_ = conn.Close()
			}
		} else { // Parent, child-pid recieved
			return
		}
	})
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
