package proc

/*
#include <errno.h>
#include <unistd.h>
int go_fork() {
	return fork();
}
#include <stdio.h>
#include <string.h>
#include <inttypes.h>
#include <fcntl.h>
#include <sys/syscall.h>
#include <linux/aio_abi.h>
int io_setup(unsigned nr, aio_context_t *ctxp) {
	return syscall(__NR_io_setup, nr, ctxp);
}
*/
import "C"
import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"unsafe"
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

func SetIOContext(nrEvents uint, conn *net.Conn) {
	//lCtxIdp := C.ulong(&conn)
	val := int(C.io_setup(C.uint(nrEvents), (*C.ulong)(unsafe.Pointer(conn))))
	if val == -1 {
		C.perror(C.CString("error detected"))
	}
	//switch val {
	//case C.EAGAIN:
	//	log.Errorln("The specified nr_events exceeds the user's limit of available events.")
	//case C.EFAULT:
	//	log.Errorln("An invalid pointer is passed for ctx_idp.")
	//case C.EINVAL:
	//	log.Errorln("ctx_idp is not initialized, or the specified nr_events exceeds internal limits. nr_events should be greater than 0.")
	//case C.ENOMEM:
	//	log.Errorln("Insufficient kernel resources are available.")
	//case C.ENOSYS:
	//	log.Errorln("io_setup() is not implemented on this architecture.")
	//default:
	//	log.WithField("errno", val).Errorln("idefk -_-")
	//}

}
