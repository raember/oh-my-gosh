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
	"strconv"
	"strings"
	"syscall"
)

type User struct {
	Name        string
	Transaction *pam.Transaction
	PassWd      *pw.PassWd
}

type TransactionData struct {
	Service    string
	User       string
	Tty        string
	RHost      string
	Authtok    string
	Oldauthtok string
	RUser      string
	UserPrompt string
}

func (data TransactionData) String() string {
	return strings.Join([]string{
		data.Service,
		data.User,
		data.Tty,
		data.RHost,
		data.Authtok,
		data.Oldauthtok,
		data.RUser,
		data.UserPrompt,
	}, ", ")
}

func Authenticate(data *TransactionData, in io.Reader, out io.Writer, fd uintptr) (*User, error) {
	log.WithFields(log.Fields{
		"data": data,
		"in":   &in,
		"out":  &out,
		"fd":   fd,
	}).Traceln("--> login.Authenticate")
	if os.Getuid() != 0 {
		log.WithField("uid", os.Getuid()).Warnln("Process isn't root. Login as different loggedInUser won't work.")
	}

	bufIn := bufio.NewReader(in)
	//bufOut := bufio.NewWriter(out)
	var loggedInUser User
	transaction, err := pam.StartFunc(data.Service, data.User, func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			log.Traceln("PromptEchoOff")
			log.WithField("message", message).Debugln("Reading password.")
			_, _ = fmt.Fprint(out, message)
			//str, err := speakeasy.FAsk(out, message)
			_, _ = echoOff(fd)
			str, err := bufIn.ReadString('\n')
			echoOn(fd)
			//bytes, err := terminal.ReadPassword(int(fd))
			//str := string(bytes)
			if err != nil {
				log.WithError(err).Errorln("Failed to read password.")
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
				log.WithError(err).Errorln("Failed to read loggedInUser name.")
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
		log.WithError(err).Errorln("Failed to start authentication.")
		return &loggedInUser, err
	}

	if err = setupTransaction(transaction, data); err != nil {
		return nil, err
	}
	loggedInUser.Transaction = transaction
	err = transaction.Authenticate(0)
	if err != nil {
		log.WithField("status", transaction.Error()).Errorln("Err")
		authErr := &AuthError{Err: err.Error(), User: loggedInUser.Name}
		log.WithField("error", authErr.Error()).Errorln("Failed to authenticate.")
		return &loggedInUser, authErr
	}
	log.Infoln("Authentication succeeded.")

	err = loggedInUser.Transaction.AcctMgmt(pam.Silent)
	if err != nil {
		log.WithError(err).Errorln("Failed to validate the loggedInUser.")
	}

	if err = parseTransaction(loggedInUser.Transaction, data); err != nil {
		return nil, err
	}

	passWd, err := pw.GetPwByName(data.User)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Failed to setup loggedInUser.")
		return &loggedInUser, err
	}
	loggedInUser.PassWd = passWd

	return &loggedInUser, nil
}

// echoOff turns off the terminal echo.
func echoOff(fd uintptr) (int, error) {
	log.WithField("fd", fd).Traceln("--> login.echoOff")
	pid, err := syscall.ForkExec("/bin/stty", []string{"stty", "-F", strconv.Itoa(int(fd)), "-echo"}, &syscall.ProcAttr{Dir: "", Files: []uintptr{fd}})
	if err != nil {
		log.WithError(err).Errorln("Failed turning echo off.")
		return 0, err
	}
	log.WithField("pid", pid).Debugln("Turned echo off.")
	return pid, nil
}

// echoOn turns back on the terminal echo.
func echoOn(fd uintptr) {
	log.WithField("fd", fd).Traceln("--> login.echoOn")
	pid, e := syscall.ForkExec("/bin/stty", []string{"stty", "-F", strconv.Itoa(int(fd)), "echo"}, &syscall.ProcAttr{Dir: "", Files: []uintptr{fd}})
	if e == nil {
		log.WithField("pid", pid).Debugln("Turned echo on.")
		wpid, err := syscall.Wait4(pid, nil, 0, nil)
		if err != nil {
			log.WithError(err).Errorln("Failed waiting for subprocess.")
		} else {
			log.WithField("wpid", wpid).Debugln("Waited for subprocess.")
		}
	} else {
		log.WithError(e).Errorln("Failed turning echo on.")
	}
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

func setupTransaction(transaction *pam.Transaction, data *TransactionData) error {
	log.WithFields(log.Fields{
		"transaction": transaction,
		"data":        data,
	}).Traceln("--> login.setupTransaction")
	// TODO: Get information about the peer
	if err := transaction.SetItem(pam.Service, data.Service); err != nil {
		log.WithError(err).Errorln("Failed to set Service.")
		return err
	}
	if err := transaction.SetItem(pam.User, data.User); err != nil {
		log.WithError(err).Errorln("Failed to set User.")
		return err
	}
	if err := transaction.SetItem(pam.Tty, data.Tty); err != nil {
		log.WithError(err).Errorln("Failed to set Tty.")
		return err
	}
	if err := transaction.SetItem(pam.Rhost, data.RHost); err != nil {
		log.WithError(err).Errorln("Failed to set RHost.")
		return err
	}
	if data.Authtok != "" {
		if err := transaction.SetItem(pam.Authtok, data.Authtok); err != nil {
			log.WithError(err).Errorln("Failed to set Authtok.")
			return err
		}
	}
	if data.Oldauthtok != "" {
		if err := transaction.SetItem(pam.Oldauthtok, data.Oldauthtok); err != nil {
			log.WithError(err).Errorln("Failed to set Oldauthtok.")
			return err
		}
	}
	if err := transaction.SetItem(pam.Ruser, data.RUser); err != nil {
		log.WithError(err).Errorln("Failed to set RUser.")
		return err
	}
	if err := transaction.SetItem(pam.UserPrompt, data.UserPrompt); err != nil {
		log.WithError(err).Errorln("Failed to set UserPrompt.")
		return err
	}
	log.WithFields(log.Fields{
		"service":    data.Service,
		"user":       data.User,
		"tty":        data.Tty,
		"rhost":      data.RHost,
		"authtok":    data.Authtok,
		"olauthtok":  data.Oldauthtok,
		"ruser":      data.RUser,
		"userprompt": data.UserPrompt,
	}).Debugln("Setup transaction object.")
	return nil
}

func parseTransaction(transaction *pam.Transaction, data *TransactionData) error {
	log.WithFields(log.Fields{
		"transaction": transaction,
		"data":        data,
	}).Traceln("--> login.parseTransaction")
	service, err := transaction.GetItem(pam.Service)
	if err != nil {
		log.WithError(err).Errorln("Failed to set Service.")
		return err
	}
	data.Service = service

	userName, err := transaction.GetItem(pam.User)
	if err != nil {
		log.WithError(err).Errorln("Failed to set User.")
		return err
	}
	data.User = userName

	tty, err := transaction.GetItem(pam.Tty)
	if err != nil {
		log.WithError(err).Errorln("Failed to set Tty.")
		return err
	}
	data.Tty = tty

	rHost, err := transaction.GetItem(pam.Rhost)
	if err != nil {
		log.WithError(err).Errorln("Failed to set RHost.")
		return err
	}
	data.RHost = rHost

	authtok, err := transaction.GetItem(pam.Authtok)
	if err != nil {
		log.WithError(err).Errorln("Failed to set Authtok.")
		return err
	}
	data.Authtok = authtok

	oldauthtok, err := transaction.GetItem(pam.Oldauthtok)
	if err != nil {
		log.WithError(err).Errorln("Failed to set Oldauthtok.")
		return err
	}
	data.Oldauthtok = oldauthtok

	rUser, err := transaction.GetItem(pam.Ruser)
	if err != nil {
		log.WithError(err).Errorln("Failed to set RUser.")
		return err
	}
	data.RUser = rUser

	userPrompt, err := transaction.GetItem(pam.UserPrompt)
	if err != nil {
		log.WithError(err).Errorln("Failed to set UserPrompt.")
		return err
	}
	data.UserPrompt = userPrompt
	log.WithFields(log.Fields{
		"service":    data.Service,
		"user":       data.User,
		"tty":        data.Tty,
		"rhost":      data.RHost,
		"authtok":    data.Authtok,
		"olauthtok":  data.Oldauthtok,
		"ruser":      data.RUser,
		"userprompt": data.UserPrompt,
	}).Debugln("Updated data.")
	return nil
}
