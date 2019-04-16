package login

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type User struct {
	Name        string
	Transaction *pam.Transaction
	PassWd      *pw.PassWd
}

func Authenticate(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) (*User, error) {
	if os.Getuid() != 0 {
		log.WithField("uid", os.Getuid()).Warnln("Process isn't root. Login as different user won't work.")
	}

	in := bufio.NewReader(stdIn)
	user := &User{"", nil, nil}
	transaction, err := pam.StartFunc("goshd", "", func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			log.WithField("msg", message).Debugln("Reading password.")
			_, _ = fmt.Fprint(stdOut, connection.PasswordPacket{message}.String())
			str, err := in.ReadString('\n')
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read password.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read password.")
			return str, nil
		case pam.PromptEchoOn:
			log.WithField("msg", message).Debugln("Reading user name.")
			_, _ = fmt.Fprint(stdOut, connection.UsernamePacket{message}.String())
			str, err := in.ReadString('\n')
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read user name.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read user name.")

			user.Name = str
			return str, nil
		case pam.ErrorMsg:
			log.WithField("msg", message).Errorln("An error occurred.")
			return "", nil
		case pam.TextInfo:
			log.WithField("msg", message).Debugln("Text received.")
			return "", nil
		}
		return "", errors.New("unrecognized message style")
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't start authentication.")
		return user, err
	}

	user.Transaction = transaction
	err = transaction.Authenticate(0)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't authenticate.")
		return user, err
	}
	log.Infoln("Authentication succeeded.")
	_, _ = fmt.Fprint(stdOut, connection.AuthSucceededPacket{}.String())

	// Print all the transaction commands:

	err = user.Transaction.SetCred(pam.Silent)
	if err != nil {
		log.Errorln("Couldn't set credentials for the user.")
	}
	err = user.Transaction.AcctMgmt(pam.Silent)
	if err != nil {
		log.Errorln("Couldn't validate the user.")
	}
	err = user.Transaction.OpenSession(pam.Silent)
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't open a session.")
	}
	defer user.Transaction.CloseSession(pam.Silent)
	str, err := user.Transaction.GetItem(pam.Service)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Service from transaction.")
	} else {
		log.Infoln("Service: " + str)
	}
	str, err = user.Transaction.GetItem(pam.User)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting User from transaction.")
	} else {
		log.Infoln("User: " + str)
	}
	str, err = user.Transaction.GetItem(pam.Tty)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Tty from transaction.")
	} else {
		log.Infoln("Tty: " + str)
	}
	str, err = user.Transaction.GetItem(pam.Rhost)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Rhost from transaction.")
	} else {
		log.Infoln("Rhost: " + str)
	}
	str, err = user.Transaction.GetItem(pam.Authtok)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Authtok from transaction.")
	} else {
		log.Infoln("Authtok: " + str)
	}
	str, err = user.Transaction.GetItem(pam.Oldauthtok)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Oldauthtok from transaction.")
	} else {
		log.Infoln("Oldauthtok: " + str)
	}
	str, err = user.Transaction.GetItem(pam.Ruser)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Ruser from transaction.")
	} else {
		log.Infoln("Ruser: " + str)
	}
	str, err = user.Transaction.GetItem(pam.UserPrompt)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting UserPrompt from transaction.")
	} else {
		log.Infoln("UserPrompt: " + str)
	}
	strs, err := user.Transaction.GetEnvList()
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Env from transaction.")
	} else {
		log.Println("#envs: " + strconv.Itoa(len(strs)))
		for _, str := range strs {
			log.Infoln(str + ": " + strs[str])
		}
	}

	return user, nil
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

func (user User) Setup() error {
	passWd, err := pw.GetPwByName(user.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't setup user.")
		return err
	}
	user.PassWd = passWd
	return nil
}

func (user User) String() string {
	return user.Name
}
