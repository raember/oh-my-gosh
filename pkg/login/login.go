package login

import (
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
)

func Authenticate(callback func(s pam.Style, msg string) (string, error)) (*pam.Transaction, error) {
	transaction, err := pam.StartFunc("", "", callback)
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
