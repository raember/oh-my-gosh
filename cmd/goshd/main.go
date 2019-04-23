package main

import "C"
import (
	"errors"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"io"
	"os"
	"os/exec"
)

var certFile = flag.String("cert", common.CERTFILE, "Certificate file")
var keyFile = flag.String("key", common.KEYFILE, "Key file")
var configPath = flag.String("conf", common.CONFIGPATH, "LoadConfig path")
var fd = flag.Uint("fd", 0, "Handle connection using supplied fd")
var std = flag.Bool("std", false, "Handle connection using standard pipes")
var host = flag.Bool("host", false, "Act as Host for user session")
var config = server.LoadConfig(*configPath)

func main() {
	log.Traceln("--> goshd.main")
	flag.Parse()
	log.WithField("certFile", *certFile).Debugln("Certificate file")
	log.WithField("keyFile", *keyFile).Debugln("Key file")
	log.WithField("configPath", *configPath).Debugln("LoadConfig path")
	log.WithField("fd", *fd).Debugln("Connection fd")
	log.WithField("std", *std).Debugln("Standard pipes")
	log.WithField("host", *host).Debugln("User session")

	log.Traceln("Determine control flow")
	if *fd != 0 || *std {
		in, out, err := getFile(fd, std)
		if err != nil {
			log.WithError(err).Fatalln("Cannot prepare pipes.")
		}
		if *host {
			log.Traceln("Act as Host for user session")
			log.Fatalln("Not implemented yet!")
		} else {
			log.Traceln("Serve connection")
			if err := server.NewServer(config).Serve(in, out); err != nil {
				log.WithError(err).Fatalln("Failed serving connection.")
			}
		}
	} else {
		log.Traceln("Listen for connections")
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
			fdUint := uint(connFd)
			bin := os.Args[0]
			args := os.Args[1:]
			args = append(args, fmt.Sprintf("--fd %d", fdUint))
			err = exec.Command(bin, args...).Start()
			if err != nil {
				log.WithError(err).Errorln("Couldn't execute child")
			}
		})
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func getFile(fd *uint, std *bool) (io.Reader, io.Writer, error) {
	log.WithFields(log.Fields{
		"fd":  *fd,
		"std": *std,
	}).Traceln("--> main.getFile")
	if *fd != 0 && !*std {
		log.Traceln("Read from socket fd.")
		conn, err := connection.FromFd(uintptr(*fd))
		if err != nil {
			log.WithError(err).Errorln("Couldn't get connection from fd.")
			return nil, nil, err
		}
		defer func() {
			err = conn.Close()
			if err != nil {
				log.WithError(err).Errorln("Couldn't close connection.")
			}
		}()
		return conn, conn, nil
	} else if *std && *fd == 0 {
		log.Traceln("Read from standard pipes.")
		return os.Stdin, os.Stdout, nil
	} else if *std && *fd != 0 {
		log.Fatalln("Cannot have --fd and --std set at the same time!")
	}
	return nil, nil, errors.New("couldn't decide how to get the files")
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
