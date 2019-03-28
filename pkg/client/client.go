package client

import (
	"bufio"
	"errors"
	"github.com/bgentry/speakeasy"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"net"
	"os"
)

type Client struct {
	conn net.Conn
}

func (client Client) Communicate(conn net.Conn) error {
	defer conn.Close()
	log.WithField("remote", common.AddrToStr(conn.RemoteAddr())).Debugln("Communicating with host.")
	host := bufio.NewReader(conn)
	input := bufio.NewReader(os.Stdin)

	err := client.performLogin(conn)
	if err != nil {
		return err
	}

	for {
		// listen for server
		inBound, err := host.ReadString('\n')
		if err != nil {
			log.WithField("error", err.Error()).Debugln("Couldn't read from connection.")
		}
		log.WithField("inBound", inBound).Debugln("Received message.")
		log.Println(inBound)

		// read in input from stdin
		log.Print("> ")
		outBound, err := input.ReadString('\n')

		// send to socket
		_, err = conn.Write([]byte(outBound))
		log.WithFields(log.Fields{"outBound": outBound}).Debugln("Sent message.")
		if err != nil {
			log.WithField("error", err.Error()).Debugln("Couldn't send message.")
			break
		}
	}
	return nil
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
// Sends 2 lines:
// user\n
// password\n
// Reads one byte to determine the outcome.
func (client Client) performLogin(conn net.Conn) error {
	host := bufio.NewReader(conn)
	input := bufio.NewReader(os.Stdin)

	for {
		print("Login: ")
		user, err := input.ReadString('\n')
		if err != nil {
			log.Error("Failed reading user from stdin.")
			return err
		}
		password, err := speakeasy.Ask("Password: ")
		if err != nil {
			log.Error("Failed reading password from stdin.")
			return err
		}
		log.WithField("user", user).Debugln("Sending credentials to host.")
		_, _ = conn.Write([]byte(user))
		_, _ = conn.Write([]byte(password + "\n"))
		answer, err := host.ReadByte()
		if err != nil {
			log.Error("Couldn't get answer from host.")
			return err
		}
		switch answer {
		case login.LOGIN_ACCEPT:
			return nil
		case login.LOGIN_FAIL:
			err = errors.New("login attempt failed")
			log.WithFields(log.Fields{"error": err}).Errorln("Couldn't authenticate user.")
			continue
		case login.LOGIN_EXCEED:
			err = errors.New("exceeded allowed amount of user attempts")
			log.WithFields(log.Fields{"error": err}).Errorln("Couldn't authenticate user.")
			return err
		}
	}
}
