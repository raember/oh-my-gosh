package main

import "C"
import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"os/signal"
)

func main() {
	log.WithField("args", os.Args).Traceln("--> goshh.main")
	configPath := flag.String("conf", common.CONFIGPATH, "Config path.")
	authPath := flag.String("auth", common.AUTHPATH, "Authorized keys path.")
	certFile := flag.String("cert", common.CERTFILE, "Certificate file.")
	keyFile := flag.String("key", common.KEYFILE, "Key file.")
	fd := flag.Uint("fd", 0, "The file descriptor for the connection.")
	rAddr := flag.String("remote", fmt.Sprintf("%s:%d", common.LOCALHOST, common.PORT), "The address of the remote.")

	flag.Parse()
	log.WithFields(log.Fields{
		"certFile":   *certFile,
		"keyFile":    *keyFile,
		"configPath": *configPath,
		"authPath":   *authPath,
		"fd":         *fd,
		"rAddr":      *rAddr,
	}).Debugln("Parsed arguments.")

	config := server.LoadConfig(*configPath)
	config.Set("Authentication.KeyStore", *authPath)
	host := server.NewHost(config)

	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, unix.SIGINT)
		log.WithField("sig", (<-sigChan).String()).Warnln("Received signal. Shutting down.")
		cleanup(host)
	}()

	if err := host.LoadCertKeyPair(*certFile, *keyFile); err != nil {
		log.WithError(err).Fatalln("Failed to prepare hosting.")
	}
	peerAddr, err := net.ResolveTCPAddr(common.TCP, *rAddr)
	if err != nil {
		log.WithError(err).Fatalln("Failed to resolve remote client address.")
	}
	if err := host.Connect(uintptr(*fd), peerAddr); err != nil {
		log.WithError(err).Fatalln("Failed to prepare hosting.")
	}
	if err := host.Setup(); err != nil {
		log.WithError(err).Fatalln("Failed to setup host.")
	}
	if err := host.Serve(); err != nil {
		log.WithError(err).Fatalln("Hosting exited with an error.")
	}
}

func cleanup(host server.Host) {
	log.WithField("host", host).Traceln("--> main.cleanup")
	if err := host.Kill(); err != nil {
		log.WithError(err).WithField("host", host).Warnln("Failed to send interrupt to shell.")
	} else {
		log.WithField("host", host).Infoln("Sent interrupt to shell.")
	}
	os.Exit(0)
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
