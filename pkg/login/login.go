package login

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"io"
	"os"
	"strings"
)

type User struct {
	Name        string
	Transaction *pam.Transaction
	PassWd      *pw.PassWd
}

func Authenticate(userName string, in io.Reader, out io.Writer) (*User, error) {
	log.WithFields(log.Fields{
		"userName": userName,
		"in":       &in,
		"out":      &out,
	}).Traceln("--> login.Authenticate")
	if os.Getuid() != 0 {
		log.WithField("uid", os.Getuid()).Warnln("Process isn't root. Login as different loggedInUser won't work.")
	}

	bufIn := bufio.NewReader(in)
	//bufOut := bufio.NewWriter(out)
	var loggedInUser User
	transaction, err := pam.StartFunc(os.Args[0], userName, func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			log.Traceln("PromptEchoOff")
			log.WithField("message", message).Debugln("Reading password.")
			_, _ = fmt.Fprint(out, message)
			str, err := bufIn.ReadString('\n')
			if err != nil {
				log.WithError(err).Errorln("Couldn't read password.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("password", str).Debugln("Read password.")
			return str, nil
		case pam.PromptEchoOn:
			log.Traceln("PromptEchoOn")
			log.WithField("message", message).Debugln("Reading loggedInUser name.")
			_, _ = fmt.Fprint(out, message)
			str, err := bufIn.ReadString('\n')
			if err != nil {
				log.WithError(err).Errorln("Couldn't read loggedInUser name.")
				return "", err
			}
			str = strings.TrimSpace(str)
			log.WithField("loggedInUser", str).Debugln("Read loggedInUser name.")
			loggedInUser.Name = str
			return str, nil
		case pam.ErrorMsg:
			log.Traceln("ErrorMsg")
			log.WithField("message", message).Errorln("An error occurred.")
			return "", nil
		case pam.TextInfo:
			log.Traceln("ErrorMsg")
			log.WithField("message", message).Debugln("Text received.")
			return "", nil
		default:
			log.Traceln("default")
			log.WithField("message", message).Errorln("Not supported message style.")
			return "", errors.New("unrecognized message style")
		}
	})
	if err != nil {
		log.WithError(err).Errorln("Couldn't start authentication.")
		return &loggedInUser, err
	}

	loggedInUser.Transaction = setupTransaction(transaction)
	err = transaction.Authenticate(0)
	if err != nil {
		log.WithField("status", transaction.Error()).Errorln("Err")
		authErr := &AuthError{Err: err.Error(), User: loggedInUser.Name}
		log.WithField("error", authErr.Error()).Errorln("Couldn't authenticate.")
		return &loggedInUser, authErr
	}
	log.Infoln("Authentication succeeded.")
	//_, _ = bufOut.WriteString(connection.DonePacket{}.String())
	//_ = bufOut.Flush()

	// Print all the transaction commands:
	err = loggedInUser.Transaction.AcctMgmt(pam.Silent)
	if err != nil {
		log.WithError(err).Errorln("Couldn't validate the loggedInUser.")
	}
	service, err := loggedInUser.Transaction.GetItem(pam.Service)
	if err != nil {
		log.WithError(err).Errorln("Failed getting Service from transaction.")
	} else {
		log.WithField("service", service).Infoln("Got service from pam.")
	}
	username, err := loggedInUser.Transaction.GetItem(pam.User)
	if err != nil {
		log.WithError(err).Errorln("Failed getting User from transaction.")
	} else {
		log.WithField("username", username).Infoln("Got username from pam.")
	}
	tty, err := loggedInUser.Transaction.GetItem(pam.Tty)
	if err != nil {
		log.WithError(err).Errorln("Failed getting Tty from transaction.")
	} else {
		log.WithField("tty", tty).Infoln("Got tty from pam.")
	}
	rhost, err := loggedInUser.Transaction.GetItem(pam.Rhost)
	if err != nil {
		log.WithError(err).Errorln("Failed getting Rhost from transaction.")
	} else {
		log.WithField("rhost", rhost).Infoln("Got rhost from pam.")
	}
	ruser, err := loggedInUser.Transaction.GetItem(pam.Ruser)
	if err != nil {
		log.WithError(err).Errorln("Failed getting Ruser from transaction.")
	} else {
		log.WithField("ruser", ruser).Infoln("Got ruser from pam.")
	}
	userPrompt, err := loggedInUser.Transaction.GetItem(pam.UserPrompt)
	if err != nil {
		log.WithError(err).Errorln("Failed getting UserPrompt from transaction.")
	} else {
		log.WithField("userPrompt", userPrompt).Infoln("Got userPrompt from pam.")
	}
	envs, err := loggedInUser.Transaction.GetEnvList()
	if err != nil {
		log.WithError(err).Errorln("Failed getting Env from transaction.")
	} else {
		log.WithField("envs", envs).Infoln("Got env list from pam.")
	}

	passWd, err := pw.GetPwByName(username)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't setup loggedInUser.")
		return &loggedInUser, err
	}
	loggedInUser.PassWd = passWd

	return &loggedInUser, nil
}

func (user User) String() string {
	log.Traceln("--> login.User.String")
	return user.Name
}

type AuthError struct {
	Err  string
	User string
}

func (e *AuthError) Error() string {
	log.Traceln("--> login.AuthError.Error")
	return fmt.Sprintf("%s: %s", e.User, e.Err)
}

func setupTransaction(transaction *pam.Transaction) *pam.Transaction {
	// TODO: Get information about the peer
	return transaction
}
