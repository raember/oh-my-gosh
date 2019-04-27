package main

import "C"
import (
	"crypto/tls"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/socket"
	"os"
	"os/exec"
	"strconv"
)

var config *viper.Viper
var certFile = flag.String("cert", common.CERTFILE, "Certificate file.")
var keyFile = flag.String("key", common.KEYFILE, "Key file.")
var configPath = flag.String("conf", common.CONFIGPATH, "LoadConfig path.")

/*
The goshd service has 2 possible control flows:
1. Set up a listener and wait for incoming connections.
2. Serve an individual connection.
*/
func main() {
	log.WithField("args", os.Args).Traceln("--> goshd.main")

	// I/O control:
	fd := flag.Uint("fd", 0, "Use file descriptor for socket.")

	flag.Parse()
	log.WithFields(log.Fields{
		"certFile":   *certFile,
		"keyFile":    *keyFile,
		"configPath": *configPath,
		"fd":         *fd,
	}).Debugln("Parsed arguments.")

	config = server.LoadConfig(*configPath)

	if *fd != 0 {
		srvr := server.NewServer(config)
		conn, err := connection.FromFd(uintptr(*fd))
		if err != nil {
			log.WithError(err).Fatalln("Couldn't get connection from fd.")
		}
		defer connection.CloseConn(conn, "client")
		tlsConn := tls.Server(conn, &tls.Config{
			Certificates: []tls.Certificate{server.LoadCertKeyPair(*certFile, *keyFile)},
		})
		srvr.Serve(tlsConn.RemoteAddr(), tlsConn, tlsConn, &conn)
	} else {
		listen()
	}
	log.Traceln("Exiting")
}

// Listen for incoming connections and recurse for every new connection.
func listen() {
	log.Traceln("--> main.listen")
	fdChan := make(chan uintptr)
	defer close(fdChan)
	go server.WaitForConnections(socket.Listen(common.PORT), fdChan)
	for fd := range fdChan {
		bin := os.Args[0]
		args := []string{
			"--cert", *certFile,
			"--key", *keyFile,
			"--conf", *configPath,
			"--fd", strconv.FormatUint(uint64(fd), 10),
		}
		cmd := exec.Command(bin, args...)
		cmd.Env = append(cmd.Env, fmt.Sprintf("LOG_LEVEL=%s", log.GetLevel().String()))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.WithError(err).Fatalln("Couldn't execute child")
		}
		log.WithField("cmd", cmd).Infoln("Started child.")
	}
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
