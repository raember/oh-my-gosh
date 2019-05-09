package main

import "C"
import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
)

func main() {
	log.WithField("args", os.Args).Traceln("--> goshd.main")
	configPath := flag.String("conf", common.CONFIGPATH, "Config path.")
	authPath := flag.String("auth", common.AUTHPATH, "Authorized keys path.")
	certFile := flag.String("cert", common.CERTFILE, "Certificate file.")
	keyFile := flag.String("key", common.KEYFILE, "Key file.")

	flag.Parse()
	log.WithFields(log.Fields{
		"certFile":   *certFile,
		"authPath":   *authPath,
		"keyFile":    *keyFile,
		"configPath": *configPath,
	}).Debugln("Parsed arguments.")

	config := server.LoadConfig(*configPath)
	config.Set("Authentication.KeyStore", *authPath)
	srvr := server.NewServer(config)

	var children []*exec.Cmd
	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, unix.SIGINT)
		log.WithField("sig", (<-sigChan).String()).Warnln("Received signal. Shutting down.")
		cleanup(children)
	}()
	remoteChan := make(chan server.RemoteHandle)
	defer close(remoteChan)
	go srvr.AwaitConnections(remoteChan)
	for remote := range remoteChan {
		log.WithField("remote", remote).Debugln("Got a remote handle.")
		host := exec.Command(path.Join(path.Dir(os.Args[0]), "goshh"),
			"--conf", *configPath,
			"--auth", *authPath,
			"--cert", *certFile,
			"--key", *keyFile,
			"--fd", strconv.FormatUint(uint64(remote.Fd), 10),
			"--remote", remote.RemoteAddr.String())
		host.Env = []string{fmt.Sprintf("LOG_LEVEL=%s", log.GetLevel().String())}
		host.Stdin = os.Stdin
		host.Stdout = os.Stdout
		host.Stderr = os.Stderr
		err := host.Start()
		if err != nil {
			log.WithError(err).Errorln("Failed to start child")
		} else {
			log.WithField("pid", host.Process.Pid).Infoln("Started child.")
			children = append(children, host)
			log.WithFields(log.Fields{
				"children": children,
				"host":     host,
			}).Debugln("Added child process to array.")
		}
	}
}

func cleanup(children []*exec.Cmd) {
	log.WithField("children", children).Traceln("--> main.cleanup")
	for _, host := range children {
		if err := host.Process.Signal(unix.SIGINT); err != nil {
			log.WithError(err).WithField("host", host).Warnln("Failed to interrupt to child.")
		} else {
			log.WithField("host", host).Infoln("Sent interrupt to child.")
		}
	}
	os.Exit(0)
}
