package login

import (
	"bufio"
	"errors"
	"fmt"
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
	username := ""
	transaction, err := pam.StartFunc("goshd", "", func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			log.WithField("msg", message).Debugln("Reading password.")
			_, _ = fmt.Fprint(stdOut, message) // "Password: "
			// Dirty hack in case `tput invis` doesn't work
			_ = tput(stdIn, stdOut, stdErr, "setaf", "0")
			_ = tput(stdIn, stdOut, stdErr, "setab", "0")
			_ = tput(stdIn, stdOut, stdErr, "invis")
			str, err := in.ReadString('\n')
			//str, err := speakeasy.FAsk(stdIn, stdOut, message)
			_ = tput(stdIn, stdOut, stdErr, "sgr0")
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read password.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read password.")
			return str, nil
		case pam.PromptEchoOn:
			log.WithField("msg", message).Debugln("Reading user name.")
			_, _ = fmt.Fprint(stdOut, message) // "login:"
			str, err := in.ReadString('\n')
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read user name.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read user name.")

			username = str
			return str, nil
		case pam.ErrorMsg:
			log.WithField("msg", message).Debugln("Will send error.")
			_, _ = fmt.Fprint(stdErr, message)
			return "", nil
		case pam.TextInfo:
			log.WithField("msg", message).Debugln("Will send text.")
			_, _ = fmt.Fprint(stdOut, message)
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
	//cmd.Stdin = stdIn
	cmd.Stdout = stdOut
	//cmd.Stderr = stdErr
	return cmd.Run()
}

type AuthError struct {
	Err  string
	User string
}

func (ae AuthError) Error() string {
	return fmt.Sprintf(ae.Err, ae.User)
}
