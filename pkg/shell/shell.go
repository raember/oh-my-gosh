package shell

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

func Execute(shellpath string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) error {
	shell := exec.Command(shellpath)
	shell.Stdin = stdIn
	shell.Stdout = stdOut
	shell.Stderr = stdErr
	err := shell.Run()
	if err != nil {
		log.WithField("error", err).Errorln("An error occured.")
	}
	return err
}
