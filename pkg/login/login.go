package login

import (
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
)

const (
	LOGIN_ACCEPT = 0
	LOGIN_FAIL   = 1
	LOGIN_EXCEED = 2
)

func Authenticate(user string, password string) (*pam.Transaction, error) {
	transaction, err := pam.StartFunc("", "", func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn:
			return user, nil
		case pam.ErrorMsg:
			log.WithField("error", message).Errorln("Error occurred.")
			return "", nil
		case pam.TextInfo:
			log.WithField("message", message).Errorln("Message from PAM.")
			return "", nil
		}
		return "", errors.New("unrecognized message style")
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't start authentication.")
		return nil, err
	}
	err = transaction.Authenticate(0)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't authenticate.")
		return nil, err
	}
	log.Infoln("Authentication succeeded!")
	return transaction, nil
}
