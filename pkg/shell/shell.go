package shell

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"syscall"
)

func Execute(shellpath string, stdIn io.Reader, stdOut io.Writer) error {
	log.Trace("Executing shell")
	shell := exec.Command(shellpath, "--login")
	pid, err := syscall.Setsid()
	if err != nil {
		log.WithField("error", err).Errorln("Failed setting sid.")
	} else {
		log.WithField("pid", pid).Infoln("Set sid.")
	}
	shell.Stdin = stdIn
	shell.Stdout = stdOut
	shell.Stderr = stdOut
	err = shell.Run()
	if err != nil {
		log.WithField("error", err).Errorln("An error occured.")
	}
	return err
}
