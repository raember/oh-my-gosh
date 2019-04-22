package main

import "C"
import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"os"
)

var certFile = flag.String("cert", common.CERTFILE, "Certificate file")
var keyFile = flag.String("key", common.KEYFILE, "Key file")
var configPath = flag.String("conf", common.CONFIGPATH, "LoadConfig path")
var local = flag.Bool("local", false, "Don't host an ssh server. Instead, login and use shell locally.")

func main() {
	log.Traceln("goshd.main")
	flag.Parse()
	log.WithFields(log.Fields{"certFile": *certFile}).Debugln("Certificate file set.")
	log.WithFields(log.Fields{"keyFile": *keyFile}).Debugln("Key file set.")
	log.WithFields(log.Fields{"configPath": *configPath}).Debugln("LoadConfig path set.")
	log.WithFields(log.Fields{"local": *local}).Debugln("Local shell set.")

	config := server.LoadConfig(*configPath)

	if !*local {
		lookout, err := server.NewLookout(
			config.GetString("Server.Protocol"),
			config.GetInt("Server.Port"),
		)
		if err != nil {
			os.Exit(1)
		}

		socketFd, err := lookout.Listen(*certFile, *keyFile)
		if err != nil {
			os.Exit(1)
		}
		err = server.WaitForConnections(socketFd, func(connFd uintptr) {
			// TODO: Fix usage corruption of conn struct after forking.
			conn, err := connection.FromFd(connFd)
			if err != nil {
				return
			}
			if err = server.NewServer(config).Serve(conn, conn); err != nil {
				_ = conn.Close()
			}
			//pid, err := proc.Fork()
			//if err != nil {
			//	return
			//}
			//if pid == 0 { // Child
			//	conn, err := net.FileConn(os.NewFile(connFd, ""))
			//	if err != nil {
			//		log.WithFields(log.Fields{
			//			"error": err,
			//			"connFd": connFd,
			//		}).Errorln("Couldn't make a conn object from file descriptor.")
			//		return
			//	}
			//	srvr := server.NewServer(config)
			//	err = srvr.Serve(conn, conn, conn)
			//	if err != nil {
			//		_ = conn.Close()
			//	}
			//} else { // Parent, child-pid recieved
			//	return
			//}
		})
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	} else {
		if err := server.NewServer(config).Serve(os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
