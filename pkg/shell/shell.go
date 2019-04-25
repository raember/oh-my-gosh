package shell

import (
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"io"
	"os"
	"os/exec"
	"strconv"
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
		Setsid: true,
		Credential: &syscall.Credential{
			Uid: passWd.Uid,
			Gid: passWd.Gid,
		},
	}
	shell.Dir = passWd.HomeDir
	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Errorln("Couldn't lookup hostname")
	}
	shell.Env = []string{
		"USER=" + passWd.Name,
		"UID=" + strconv.Itoa(int(passWd.Uid)),
		"GID=" + strconv.Itoa(int(passWd.Gid)),
		"HOME=" + passWd.HomeDir,
		"SHELL=" + passWd.Shell,
		"HOSTNAME=" + hostname,
	}
	shell.Stdin = in
	shell.Stdout = out
	shell.Stderr = out
	err = os.MkdirAll(passWd.HomeDir, os.ModeDir)
	if err != nil {
		log.WithError(err).WithField("path", passWd.HomeDir).Errorln("Couldn't create directory.")
	}
	// TODO: Tidy up fragments in shell.
	err = shell.Run()
	if err != nil {
		log.WithError(err).Errorln("An error occured.")
	} else {
		log.Debugln("Shell terminated.")
	}
	return err
}
