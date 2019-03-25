package login

import (
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
)

func Authenticate(callback func(s pam.Style, msg string) (string, error)) {
	t, err := pam.StartFunc("", "", callback)
	if err != nil {
		log.Fatalf("Start: %s", err.Error())
	}
	err = t.Authenticate(0)
	if err != nil {
		log.Fatalf("Authenticate: %s", err.Error())
	}
	log.Infoln("Authentication succeeded!")
}
