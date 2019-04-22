package shell

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"syscall"
)

func Execute(shellpath string, stdIn io.Reader, stdOut io.Writer) error {
	log.WithFields(log.Fields{
		"shellpath": shellpath,
		"stdIn":     stdIn,
		"stdOut":    stdOut,
	}).Traceln("shell.Execute")
	shell := exec.Command(shellpath, "--login")
	// TODO: Make shell transmit everything over to client CORRECTLY.
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
	log.Debugln("Shell terminated.")
	return err
}
