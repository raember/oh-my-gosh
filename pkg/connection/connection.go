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

func FromFD(fd uintptr) (net.Conn, error) {
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
}

type UsernamePacket struct {
	Request string
}

func (req UsernamePacket) Ask(in io.Reader, out io.Writer) error {
	log.WithField("msg", req.Request).Debugln("Reading user name.")
	_, _ = os.Stdout.WriteString(req.Request)
	bIn := bufio.NewReader(in)
	str, err := bIn.ReadString('\n')
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't read user name.")
		return err
	}
	username := strings.TrimSpace(str)
	log.WithField("username", username).Debugln("Read user name. Sending user name to server.")
	_, err = fmt.Fprintln(out, username)
	return err
}

func (req UsernamePacket) String() string {
	return "?U:" + req.Request + "\n"
}

func (req UsernamePacket) Done() bool {
	return false
}

func (req UsernamePacket) Field() string {
	return req.Request
}

type PasswordPacket struct {
	Request string
}

func (req PasswordPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithField("msg", req.Request).Debugln("Reading password.")
	str, err := speakeasy.Ask(req.Request)
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't read password.")
		return err
	}
	password := strings.TrimSpace(str)
	log.WithField("password", password).Debugln("Read password. Sending password to server.")
	_, err = fmt.Fprintln(out, password)
	return err
}

func (req PasswordPacket) String() string {
	return "?P:" + req.Request + "\n"
}

func (req PasswordPacket) Done() bool {
	return false
}

func (req PasswordPacket) Field() string {
	return req.Request
}

// Authentication succeeded. No need to wait for more packets
type AuthSucceededPacket struct{}

func (req AuthSucceededPacket) Ask(in io.Reader, out io.Writer) error {
	return errors.New("nothing to ask")
}

func (req AuthSucceededPacket) String() string {
	return "?S:\n"
}

func (req AuthSucceededPacket) Done() bool {
	return true
}

func (req AuthSucceededPacket) Field() string {
	return ""
}

// Authentication succeeded. No need to wait for more packets
type TimeoutPacket struct{}

func (req TimeoutPacket) Ask(in io.Reader, out io.Writer) error {
	return errors.New("nothing to ask")
}

func (req TimeoutPacket) String() string {
	return "?T:\n"
}

func (req TimeoutPacket) Done() bool {
	return true
}

func (req TimeoutPacket) Field() string {
	return ""
}

func Parse(str string) (Packet, error) {
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
	}
	return nil, errors.New("couldn't parse packet")
}
