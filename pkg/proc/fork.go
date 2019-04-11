package proc

/*
#include <errno.h>
#include <unistd.h>
#include <linux/aio_abi.h>
int go_fork() {
	return fork();
}
int io_setup(unsigned nr_events, aio_context_t *ctx_idp);
*/
import "C"
import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
)

func Fork() (int, error) {
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
		log.WithField("error", err).Errorln("Failed to fork process.")
		return pid, err
	}
	return pid, nil
}

func SetIOContext() {
	nrEvents := 2
	ctxIdp := 3
	val := int(C.io_setup(C.uint(nrEvents), C.ulong(&ctxIdp)))
	println(val)

}
