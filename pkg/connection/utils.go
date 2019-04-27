package connection

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
)

func FromFd(fd uintptr) (net.Conn, error) {
	log.WithField("fd", fd).Traceln("--> connection.FromFd")
	conn, err := net.FileConn(os.NewFile(fd, ""))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"fd":    fd,
		}).Errorln("Couldn't make a conn object from file descriptor.")
		return nil, err
	}
	return conn, nil
}

func CloseFile(file *os.File, fileStr string) {
	log.WithFields(log.Fields{
		"file":    file,
		"fileStr": fileStr,
	}).Traceln("--> connection.CloseFile")
	if err := file.Close(); err != nil {
		log.WithError(err).Errorln(fmt.Sprintf("Couldn't close %s file.", fileStr))
	} else {
		log.Debugln(fmt.Sprintf("Closed %s file.", fileStr))
	}
}

func CloseConn(conn net.Conn, fileStr string) {
	log.WithFields(log.Fields{
		"conn":    conn,
		"fileStr": fileStr,
	}).Traceln("--> connection.CloseFile")
	if err := conn.Close(); err != nil {
		log.WithError(err).Errorln(fmt.Sprintf("Couldn't close %s connection.", fileStr))
	} else {
		log.Debugln(fmt.Sprintf("Closed %s connection.", fileStr))
	}
}

func Forward(in io.Reader, out io.Writer, inStr string, outStr string) {
	GOROUTINE := fmt.Sprintf("%s->%s", inStr, outStr)
	log.WithFields(log.Fields{
		"GoRoutine": GOROUTINE,
		"&in":       &in,
		"&out":      &out,
		"inStr":     inStr,
		"outStr":    outStr,
	}).Traceln("==> Go server.Forward")
	n, err := bufio.NewReader(in).WriteTo(out)
	if err != nil {
		log.WithError(err).WithField("GoRoutine", GOROUTINE).Errorln(fmt.Sprintf("Couldn't write from %s to %s.", inStr, outStr))
		return
	}
	log.WithFields(log.Fields{
		"GoRoutine": GOROUTINE,
		"n":         n,
	}).Debugln(fmt.Sprintf("Wrote from %s to %s.", inStr, outStr))
}
