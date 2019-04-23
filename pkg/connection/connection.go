package connection

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/bgentry/speakeasy"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"strings"
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

type Packet interface {
	Ask(io.Reader, io.Writer) error
	String() string
	Done() bool
	Field() string
	Error() error
}

// =============== Username Packet ===============

type UsernamePacket struct {
	Request string
}

func (req UsernamePacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.UsernamePacket.Ask")
	log.WithField("msg", req.Request).Debugln("Reading user name.")
	_, _ = os.Stdout.WriteString(req.Request)
	bIn := bufio.NewReader(in)
	str, err := bIn.ReadString('\n')
	if err != nil {
		log.WithError(err).Errorln("Couldn't read user name.")
		return err
	}
	username := strings.TrimSpace(str)
	log.WithField("username", username).Debugln("Read user name. Sending user name to server.")
	_, err = fmt.Fprintln(out, username)
	return err
}

func (req UsernamePacket) String() string {
	log.Traceln("--> connection.UsernamePacket.String")
	return "?U:" + req.Request + "\n"
}

func (req UsernamePacket) Done() bool {
	log.Traceln("--> connection.UsernamePacket.Done")
	return false
}

func (req UsernamePacket) Field() string {
	log.Traceln("--> connection.UsernamePacket.Field")
	return req.Request
}

func (req UsernamePacket) Error() error {
	log.Traceln("--> connection.UsernamePacket.Field")
	return nil
}

// =============== Password Packet ===============

type PasswordPacket struct {
	Request string
}

func (req PasswordPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.PasswordPacket.Ask")
	log.WithField("msg", req.Request).Debugln("Reading password.")
	str, err := speakeasy.Ask(req.Request)
	if err != nil {
		log.WithError(err).Errorln("Couldn't read password.")
		return err
	}
	password := strings.TrimSpace(str)
	log.WithField("password", password).Debugln("Read password. Sending password to server.")
	_, err = fmt.Fprintln(out, password)
	return err
}

func (req PasswordPacket) String() string {
	log.Traceln("--> connection.PasswordPacket.String")
	return "?P:" + req.Request + "\n"
}

func (req PasswordPacket) Done() bool {
	log.Traceln("--> connection.PasswordPacket.Done")
	return false
}

func (req PasswordPacket) Field() string {
	log.Traceln("--> connection.PasswordPacket.Field")
	return req.Request
}

func (req PasswordPacket) Error() error {
	log.Traceln("--> connection.PasswordPacket.Field")
	return nil
}

// =============== Authentication Succeeded Packet ===============

type AuthSucceededPacket struct{}

func (req AuthSucceededPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.AuthSucceededPacket.Ask")
	err := errors.New("nothing to ask")
	log.WithError(err).Errorln("Not implemented!")
	return err
}

func (req AuthSucceededPacket) String() string {
	log.Traceln("--> connection.AuthSucceededPacket.String")
	return "?S:\n"
}

func (req AuthSucceededPacket) Done() bool {
	log.Traceln("--> connection.AuthSucceededPacket.Done")
	return true
}

func (req AuthSucceededPacket) Field() string {
	log.Traceln("--> connection.AuthSucceededPacket.Field")
	return ""
}

func (req AuthSucceededPacket) Error() error {
	log.Traceln("--> connection.AuthSucceededPacket.Field")
	return nil
}

// =============== Timeout Packet ===============

type TimeoutPacket struct{}

func (req TimeoutPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.TimeoutPacket.Ask")
	err := errors.New("nothing to ask")
	log.WithError(err).Errorln("Not implemented!")
	return err
}

func (req TimeoutPacket) String() string {
	log.Traceln("--> connection.TimeoutPacket.String")
	return "?T:\n"
}

func (req TimeoutPacket) Done() bool {
	log.Traceln("--> connection.TimeoutPacket.Done")
	return true
}

func (req TimeoutPacket) Field() string {
	log.Traceln("--> connection.TimeoutPacket.Field")
	return ""
}

func (req TimeoutPacket) Error() error {
	log.Traceln("--> connection.TimeoutPacket.Field")
	err := errors.New("timeout reached")
	log.WithError(err).Errorln("Login failed because of timeout.")
	return err
}

// =============== Maximum Tries Exceeded Packet ===============

type MaxTriesExceededPacket struct{}

func (req MaxTriesExceededPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.MaxTriesExceededPacket.Ask")
	err := errors.New("nothing to ask")
	log.WithError(err).Errorln("Not implemented!")
	return err
}

func (req MaxTriesExceededPacket) String() string {
	log.Traceln("--> connection.MaxTriesExceededPacket.String")
	return "?X:\n"
}

func (req MaxTriesExceededPacket) Done() bool {
	log.Traceln("--> connection.MaxTriesExceededPacket.Done")
	return true
}

func (req MaxTriesExceededPacket) Field() string {
	log.Traceln("--> connection.MaxTriesExceededPacket.Field")
	return ""
}

func (req MaxTriesExceededPacket) Error() error {
	log.Traceln("--> connection.MaxTriesExceededPacket.Field")
	err := errors.New("maximum tries reached")
	log.WithError(err).Errorln("Login failed because of maximum tries reached.")
	return err
}

// Parser:

func Parse(str string) (Packet, error) {
	log.WithField("str", str).Traceln("--> connection.Parse")
	if strings.HasPrefix(str, "?U:") {
		str = str[3:]
		return UsernamePacket{str}, nil
	} else if strings.HasPrefix(str, "?P:") {
		str = str[3:]
		return PasswordPacket{str}, nil
	} else if str == "?S:" {
		return AuthSucceededPacket{}, nil
	} else if str == "?T:" {
		return TimeoutPacket{}, nil
	} else if str == "?X:" {
		return MaxTriesExceededPacket{}, nil
	}
	return nil, errors.New("couldn't parse packet")
}
