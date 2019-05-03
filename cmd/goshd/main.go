package main

import "C"
import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"os"
	"path"
	"strconv"
	"syscall"
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

	var children []uintptr
	fdChan := make(chan server.RemoteHandle)
	defer close(fdChan)
	go srvr.AwaitConnections(fdChan)
	for remoteHandle := range fdChan {
		log.WithField("remoteHandle", remoteHandle).Debugln("Got a remote handle.")
		bin := path.Join(path.Dir(os.Args[0]), "goshh")
		args := []string{bin,
			"--conf", *configPath,
			"--auth", *authPath,
			"--cert", *certFile,
			"--key", *keyFile,
			"--fd", strconv.FormatUint(uint64(remoteHandle.Fd), 10),
			"--remote", remoteHandle.RemoteAddr.String(),
		}
		pid, err := syscall.ForkExec(bin, args, &syscall.ProcAttr{
			Env: []string{
				fmt.Sprintf("LOG_LEVEL=%s", log.GetLevel().String()),
			},
			Files: []uintptr{
				uintptr(os.Stdin.Fd()),
				uintptr(os.Stdout.Fd()),
				uintptr(os.Stderr.Fd()),
			},
		})
		if err != nil {
			log.WithError(err).Fatalln("Failed to forkexec child")
		}
		log.WithField("pid", pid).Infoln("Forkexeced child.")
		children = append(children, uintptr(pid))
		log.WithFields(log.Fields{
			"children": children,
			"pid":      pid,
		}).Debugln("Added child pid to array.")
	}
	for pid := range children {
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			log.WithError(err).Errorln("Failed to kill child.")
		}
	}
}
