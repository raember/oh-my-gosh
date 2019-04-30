package connection

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
)

func ConnFromFd(fd uintptr, certificate tls.Certificate) (net.Conn, error) {
	log.WithFields(log.Fields{
		"fd":          fd,
		"certificate": certificate,
	}).Traceln("--> utils.ConnFromFd")
	conn, err := net.FileConn(os.NewFile(fd, "conn"))
	if err != nil {
		log.WithError(err).Errorln("Failed to make a connection from file descriptor.")
		return nil, err
	}
	tlsConn := tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{certificate},
	})
	return tlsConn, nil
}

func CloseFile(file *os.File) {
	log.WithFields(log.Fields{
		"file":        file,
		"file.Name()": file.Name(),
	}).Traceln("--> connection.CloseFile")
	if err := file.Close(); err != nil {
		log.WithError(err).Errorln(fmt.Sprintf("Failed to close %s file.", file.Name()))
	} else {
		log.Debugln(fmt.Sprintf("Closed %s file.", file.Name()))
	}
}

func CloseConn(conn net.Conn) {
	log.WithField("conn", conn).Traceln("--> connection.CloseFile")
	if err := conn.Close(); err != nil {
		log.WithError(err).Errorln("Failed to close connection.")
	} else {
		log.Debugln("Closed connection.")
	}
}

func Forward(in io.Reader, out io.Writer, inStr string, outStr string) {
	GoRoutine := fmt.Sprintf("%s->%s", inStr, outStr)
	log.WithFields(log.Fields{
		"GoRoutine": GoRoutine,
		"&in":       &in,
		"&out":      &out,
		"inStr":     inStr,
		"outStr":    outStr,
	}).Traceln("==> Go server.Forward")
	n, err := bufio.NewReader(in).WriteTo(out)
	if err != nil {
		log.WithError(err).WithField("GoRoutine", GoRoutine).Errorln(fmt.Sprintf("Failed to write from %s to %s.", inStr, outStr))
		return
	}
	log.WithFields(log.Fields{
		"GoRoutine": GoRoutine,
		"n":         n,
	}).Debugln(fmt.Sprintf("Wrote from %s to %s.", inStr, outStr))
}
