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
	log.Traceln("--> connection.UsernamePacket.Error")
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
	log.Traceln("--> connection.PasswordPacket.Error")
	return nil
}

// =============== Done Packet ===============

type DonePacket struct{}

func (req DonePacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.DonePacket.Ask")
	err := errors.New("nothing to ask")
	log.WithError(err).Errorln("Not implemented!")
	return err
}

func (req DonePacket) String() string {
	log.Traceln("--> connection.DonePacket.String")
	return "?D:\n"
}

func (req DonePacket) Done() bool {
	log.Traceln("--> connection.DonePacket.Done")
	return true
}

func (req DonePacket) Field() string {
	log.Traceln("--> connection.DonePacket.Field")
	return ""
}

func (req DonePacket) Error() error {
	log.Traceln("--> connection.DonePacket.Error")
	return nil
}

// =============== Timeout Packet ===============

type TimeoutPacket struct{}

func (req TimeoutPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.TimeoutPacket.Ask")
	os.Getenv(req.Field())
	return nil
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
	log.Traceln("--> connection.TimeoutPacket.Error")
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
	log.Traceln("--> connection.MaxTriesExceededPacket.Error")
	err := errors.New("maximum tries reached")
	log.WithError(err).Errorln("Login failed because of maximum tries reached.")
	return err
}

// =============== Environment Variable Packet ===============

type EnvPacket struct {
	Request string
}

func (req EnvPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> connection.EnvPacket.Ask")
	log.WithField("msg", req.Request).Debugln("Reading environment variable.")

	envvar := os.Getenv(req.Field())
	log.WithField("envvar", envvar).Debugln("Read environment variable. Sending to server.")
	_, err := fmt.Fprintln(out, envvar+"\n")
	return err
}

func (req EnvPacket) String() string {
	log.Traceln("--> connection.EnvPacket.String")
	return "?E:" + req.Request + "\n"
}

func (req EnvPacket) Done() bool {
	log.Traceln("--> connection.EnvPacket.Done")
	return false
}

func (req EnvPacket) Field() string {
	log.Traceln("--> connection.EnvPacket.Field")
	return req.Request
}

func (req EnvPacket) Error() error {
	log.Traceln("--> connection.EnvPacket.Error")
	return nil
}

// Parser:

func Parse(str string) (Packet, error) {
	log.WithField("str", str).Traceln("--> connection.Parse")
	if strings.HasPrefix(str, "?U:") {
		return UsernamePacket{str[3:]}, nil
	} else if strings.HasPrefix(str, "?P:") {
		return PasswordPacket{str[3:]}, nil
	} else if str == "?D:" {
		return DonePacket{}, nil
	} else if str == "?T:" {
		return TimeoutPacket{}, nil
	} else if str == "?X:" {
		return MaxTriesExceededPacket{}, nil
	} else if strings.HasPrefix(str, "?E:") {
		return EnvPacket{str[3:]}, nil
	}
	return nil, errors.New("couldn't parse packet")
}
