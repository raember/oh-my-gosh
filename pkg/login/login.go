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
	"strconv"
	"strings"
)

type User struct {
	Name        string
	Transaction *pam.Transaction
	PassWd      *pw.PassWd
}

func Authenticate(stdIn io.Reader, stdOut io.Writer) (*User, error) {
	log.WithFields(log.Fields{
		"stdIn":  stdIn,
		"stdOut": stdOut,
	}).Traceln("login.Authenticate")
	if os.Getuid() != 0 {
		log.WithField("uid", os.Getuid()).Warnln("Process isn't root. Login as different user won't work.")
	}

	in := bufio.NewReader(stdIn)
	out := bufio.NewWriter(stdOut)
	user := &User{"", nil, nil}
	transaction, err := pam.StartFunc("goshd", "", func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			log.WithField("message", message).Debugln("Reading password.")
			_, _ = out.WriteString(connection.PasswordPacket{message}.String())
			_ = out.Flush()
			str, err := in.ReadString('\n')
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read password.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("password", str).Debugln("Read password.")
			return str, nil
		case pam.PromptEchoOn:
			log.WithField("message", message).Debugln("Reading user name.")
			_, _ = out.WriteString(connection.UsernamePacket{message}.String())
			_ = out.Flush()
			str, err := in.ReadString('\n')
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read user name.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("user", str).Debugln("Read user name.")
			user.Name = str
			return str, nil
		case pam.ErrorMsg:
			log.WithField("message", message).Errorln("An error occurred.")
			return "", nil
		case pam.TextInfo:
			log.WithField("message", message).Debugln("Text received.")
			return "", nil
		}
		return "", errors.New("unrecognized message style")
	})
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't start authentication.")
		return user, err
	}

	user.Transaction = transaction
	err = transaction.Authenticate(0)
	if err != nil {
		authErr := &AuthError{Err: err.Error(), User: user.Name}
		log.WithField("error", authErr.Error()).Errorln("Couldn't authenticate.")
		return user, authErr
	}
	log.Infoln("Authentication succeeded.")
	_, _ = out.WriteString(connection.AuthSucceededPacket{}.String())
	_ = out.Flush()

	// Print all the transaction commands:
	err = user.Transaction.SetCred(pam.Silent)
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't set credentials for the user.")
	}
	err = user.Transaction.AcctMgmt(pam.Silent)
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't validate the user.")
	}
	err = user.Transaction.OpenSession(pam.Silent)
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't open a session.")
	}
	defer func() {
		err := user.Transaction.CloseSession(pam.Silent)
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't close transaction session.")
			return
		}
	}()
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

func (user User) Setup() error {
	log.Traceln("login.User.Setup")
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
	log.Traceln("login.User.String")
	return user.Name
}

type AuthError struct {
	Err  string
	User string
}

func (e *AuthError) Error() string {
	log.Traceln("login.AuthError.Error")
	return fmt.Sprintf("%s: %s", e.User, e.Err)
}
