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
	"io"
	"net"
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
	fd := flag.Uint("fd", 0, "Use file descriptor. Conflicts with --std.")
	std := flag.Bool("std", false, "Use standard I/O. Conflicts with --fd.")

	flag.Parse()
	log.WithFields(log.Fields{
		"certFile":   *certFile,
		"keyFile":    *keyFile,
		"configPath": *configPath,
		"fd":         *fd,
		"std":        *std,
	}).Debugln("Parsed arguments.")

	config = server.LoadConfig(*configPath)

	if *fd != 0 || *std {
		if *fd != 0 && !*std {
			// Get the reader, writer and remote address from a file descriptor.
			conn, err := connection.FromFd(uintptr(*fd))
			if err != nil {
				log.WithError(err).Fatalln("Couldn't get connection from fd.")
			}
			defer func() {
				if err = conn.Close(); err != nil {
					log.WithError(err).Errorln("Couldn't close connection.")
				} else {
					log.Debugln("Closed connection.")
				}
			}()
			tlsConn := tls.Server(conn, &tls.Config{
				Certificates: []tls.Certificate{server.LoadCertKeyPair(*certFile, *keyFile)},
			})
			server.NewServer(config).Serve(tlsConn.RemoteAddr(), tlsConn, tlsConn)
		} else if *std && *fd == 0 {
			server.NewServer(config).Serve(fromStdIO())
		} else {
			log.Fatalln("Cannot have --fd and --std set at the same time!")
		}
	} else {
		listen()
	}
}

// Get the reader, writer and remote address from the standard I/O.
func fromStdIO() (net.Addr, io.Reader, io.Writer) {
	log.Traceln("--> fromStdIO")
	mockAddr, err := net.ResolveTCPAddr(common.TCP, fmt.Sprintf("%s:%d", common.LOCALHOST, 0))
	if err != nil {
		log.WithError(err).Fatalln("Couldn't create address mock.")
	}
	return mockAddr, os.Stdin, os.Stdout
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
