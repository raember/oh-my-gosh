package login

import (
	"bufio"
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	LOGIN_ACCEPT = 0
	LOGIN_FAIL   = 1
	LOGIN_EXCEED = 2
)

func Authenticate(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) (*pam.Transaction, string, error) {
	if os.Getuid() != 0 {
		log.WithField("uid", os.Getuid()).Warnln("Process isn't root. Login as different user won't work.")
	}

	in := bufio.NewReader(stdIn)
	out := bufio.NewWriter(stdOut)
	username := ""
	transaction, err := pam.StartFunc("goshd", "", func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			_, _ = out.WriteString(message) // "Password: "
			_ = out.Flush()
			// Dirty hack in case `tput invis` doesn't work
			_ = tput(stdIn, stdOut, stdErr, "setaf", "0")
			_ = tput(stdIn, stdOut, stdErr, "setab", "0")
			_ = tput(stdIn, stdOut, stdErr, "invis")
			log.WithField("msg", message).Debugln("Will read password.")
			str, err := in.ReadString('\n')
			//str, err := speakeasy.FAsk(stdIn, stdOut, message)
			_ = tput(stdIn, stdOut, stdErr, "sgr0")
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read password.")
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read password.")
			return str, nil
		case pam.PromptEchoOn:
			_, _ = out.WriteString(message) // "login:"
			_ = out.Flush()
			log.WithField("msg", message).Debugln("Will read user name.")
			str, err := in.ReadString('\n')
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read user name.")
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read user name.")
			username = str
			return str, nil
		case pam.ErrorMsg:
			log.WithField("msg", message).Debugln("Will send error.")
			_, _ = stdErr.Write([]byte(message))
			return "", nil
		case pam.TextInfo:
			log.WithField("msg", message).Debugln("Will send text.")
			_ = out.Flush()
			_, _ = out.Write([]byte(message))
			return "", nil
		}
		return "", errors.New("unrecognized message style")
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't start authentication.")
		return nil, username, err
	}
	err = transaction.Authenticate(0)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't authenticate.")
		return nil, username, err
	}
	log.Infoln("Authentication succeeded!")
	return transaction, username, nil
}

func tput(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, args ...string) error {
	cmd := exec.Command("tput", args...)
	cmd.Stdin = stdIn
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	return cmd.Run()
}
