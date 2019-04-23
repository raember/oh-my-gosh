package proc

/*
#include <errno.h>
#include <unistd.h>
int go_fork() {
	return fork();
}
*/
import "C"
import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
)

func Fork() (int, error) {
	log.Traceln("--> proc.Fork")
	pid := int(C.go_fork())
	if pid == 0 { // Child
		log.WithFields(log.Fields{
			"pid":  pid,
			"ppid": os.Getppid(),
		}).Debugln("Forked.")
	} else if pid > 0 { // Parent
		log.WithField("pid", pid).Debugln("Forked off child.")
	} else {
		err := errors.New("failed to fork")
		log.WithError(err).Errorln("Failed to fork process.")
		return pid, err
	}
	return pid, nil
}
