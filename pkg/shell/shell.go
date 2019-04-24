package shell

import (
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"io"
	"os/exec"
	"syscall"
)

func Execute(passWd *pw.PassWd, in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"passWd": passWd,
		"in":     in,
		"out":    out,
	}).Traceln("--> shell.Execute")
	shell := exec.Command(passWd.Shell, "--login")
	shell.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: passWd.Uid,
			Gid: passWd.Gid,
		},
	}
	shell.Stdin = in
	shell.Stdout = out
	shell.Stderr = out
	// TODO: Tidy up fragments in shell.
	err := shell.Run()
	if err != nil {
		log.WithError(err).Errorln("An error occured.")
	} else {
		log.Debugln("Shell terminated.")
	}
	return err
}
