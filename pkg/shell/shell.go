package shell

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

func Start(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) {
	shell := exec.Command("/bin/bash")
	shell.Stdin = stdIn
	shell.Stdout = stdOut
	shell.Stderr = stdErr
	log.Println("Running shell...")
	_ = shell.Run()
}
