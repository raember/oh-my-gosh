package connection

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

func Parse(str string) (Packet, error) {
	log.WithField("str", str).Traceln("--> connection.Parse")
	if str == "?D:" {
		return DonePacket{}, nil
	} else if strings.HasPrefix(str, "?E:") {
		return EnvPacket{str[3:]}, nil
	}
	return nil, errors.New("couldn't parse packet")
}

type Packet interface {
	Ask(io.Reader, io.Writer) error
	String() string
	Done() bool
}

// =============== Environment Variable Packet ===============

type EnvPacket struct {
	Request string
}

func (req EnvPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> connection.EnvPacket.Ask")
	log.WithField("request", req.Request).Debugln("Reading environment variable.")

	envvar := os.Getenv(req.Request)
	log.WithField("envvar", envvar).Debugln("Read environment variable. Sending to server.")
	_, err := fmt.Fprintln(out, envvar+"\n")
	return err
}

func (req EnvPacket) String() string {
	log.Traceln("--> connection.EnvPacket.String")
	return fmt.Sprintf("?E:%s\n", req.Request)
}

func (req EnvPacket) Done() bool {
	log.WithField("done", false).Traceln("--> connection.EnvPacket.Done")
	return false
}

// =============== Done Packet ===============

type DonePacket struct{}

func (req DonePacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
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
	log.WithField("done", true).Traceln("--> connection.DonePacket.Done")
	return true
}
