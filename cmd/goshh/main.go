package main

import "C"
import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/host"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"net"
	"os"
)

func main() {
	log.WithField("args", os.Args).Traceln("--> goshh.main")
	configPath := flag.String("conf", common.CONFIGPATH, "Config path.")
	certFile := flag.String("cert", common.CERTFILE, "Certificate file.")
	keyFile := flag.String("key", common.KEYFILE, "Key file.")
	fd := flag.Uint("fd", 0, "The file descriptor for the connection.")
	rAddr := flag.String("remote", fmt.Sprintf("%s:%d", common.LOCALHOST, common.PORT), "The address of the remote.")

	flag.Parse()
	log.WithFields(log.Fields{
		"certFile":   *certFile,
		"keyFile":    *keyFile,
		"configPath": *configPath,
		"fd":         *fd,
		"rAddr":      *rAddr,
	}).Debugln("Parsed arguments.")

	hst := host.NewHost(server.LoadConfig(*configPath))
	if err := hst.LoadCertKeyPair(*certFile, *keyFile); err != nil {
		log.WithError(err).Fatalln("Failed to prepare hosting.")
	}
	peerAddr, err := net.ResolveTCPAddr(common.TCP, *rAddr)
	if err != nil {
		log.WithError(err).Fatalln("Failed to resolve remote client address.")
	}
	if err := hst.Connect(uintptr(*fd), peerAddr); err != nil {
		log.WithError(err).Fatalln("Failed to prepare hosting.")
	}
	if err := hst.Serve(); err != nil {
		log.WithError(err).Fatalln("Hosting exited with an error.")
	}
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
