package server

import (
	"bufio"
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/willdonnelly/passwd"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	config *viper.Viper
}

func NewServer(config *viper.Viper) Server {
	return Server{config: config}
}

func (server Server) Serve(conn net.Conn) {
	defer conn.Close()
	log.WithFields(log.Fields{
		"remote": common.AddrToStr(conn.RemoteAddr()),
	}).Debugln("Serving new connection.")
	client := bufio.NewReader(conn)

	transaction, username, err := server.performLogin(conn)
	if err != nil {
		return
	}

	message := "Authentication was successful"
	log.WithFields(log.Fields{
		"message": message,
	}).Infoln("Outbound")
	_, _ = conn.Write([]byte(message + "\n"))

	err = transaction.SetCred(pam.Silent)
	if err != nil {
		log.Errorln("Couldn't set credentials for the user.")
	}
	err = transaction.AcctMgmt(pam.Silent)
	if err != nil {
		log.Errorln("Couldn't validate the user.")
	}
	err = transaction.OpenSession(pam.Silent)
	if err != nil {
		log.WithField("error", err.Error()).Errorln("Couldn't open a session.")
	}
	defer transaction.CloseSession(pam.Silent)
	str, err := transaction.GetItem(pam.Service)
	if err == nil {
		log.Infoln("Service: " + str)
	}
	str, err = transaction.GetItem(pam.User)
	if err == nil {
		log.Infoln("User: " + str)
	}
	str, err = transaction.GetItem(pam.Tty)
	if err == nil {
		log.Infoln("Tty: " + str)
	}
	str, err = transaction.GetItem(pam.Rhost)
	if err == nil {
		log.Infoln("Rhost: " + str)
	}
	str, err = transaction.GetItem(pam.Authtok)
	if err == nil {
		log.Infoln("Authtok: " + str)
	}
	str, err = transaction.GetItem(pam.Oldauthtok)
	if err == nil {
		log.Infoln("Oldauthtok: " + str)
	}
	str, err = transaction.GetItem(pam.Ruser)
	if err == nil {
		log.Infoln("Ruser: " + str)
	}
	str, err = transaction.GetItem(pam.UserPrompt)
	if err == nil {
		log.Infoln("UserPrompt: " + str)
	}
	strs, err := transaction.GetEnvList()
	if err == nil {
		log.Println("#envs: " + strconv.Itoa(len(strs)))
		for _, str := range strs {
			log.Infoln(str + ": " + strs[str])
		}
	}
	file, name, err := pty.Open()
	log.WithField("slavename", name).Infoln("Pts")
	log.WithField("filename", file.Name()).Infoln("File")

	users, err := passwd.Parse()
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't lookup users.")
		return
	}
	myUser := users[username]
	USER := username
	UID := myUser.Uid
	GID := myUser.Gid
	HOME := myUser.Home
	SHELL := myUser.Shell
	log.WithFields(log.Fields{
		"USER":  USER,
		"UID":   UID,
		"GID":   GID,
		"HOME":  HOME,
		"SHELL": SHELL,
	}).Println("Looked up user.")

	if _, err := os.Stat("/etc/motd"); err == nil {
		motd, err := ioutil.ReadFile("/etc/motd")
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't read message of the day.")
			return
		}
		log.Println(motd)
	}

	//shell.Start(client, conn, conn)

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := client.ReadString('\n')
		if err != nil {
			log.Infoln("Connection died.")
			break
		}
		// output message received
		log.WithFields(log.Fields{
			"message": message,
		}).Infoln("Inbound")
		// sample process for string received
		answer := strings.ToUpper(message)
		// send new string back to client
		log.WithFields(log.Fields{
			"answer": answer,
		}).Infoln("Outbound")
		_, _ = conn.Write([]byte(answer + "\n"))
	}
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
// Receives 2 lines from the client:
// user\n
// password\n
// Sends one byte as answer.
func (server Server) performLogin(conn net.Conn) (*pam.Transaction, string, error) {
	client := bufio.NewReader(conn)
	tries := 1
	for {
		if tries > server.config.GetInt("Authentication.MaxTries") {
			err := errors.New("maximum tries reached")
			log.WithField("error", err).Errorln("User reached maximum tries.")
			_, _ = conn.Write([]byte{login.LOGIN_EXCEED})
			return nil, "", err
		}
		tries++
		username, err := client.ReadString('\n')
		if err != nil {
			log.Errorln(err.Error())
		}
		username = strings.TrimRight(username, "\n")
		password, err := client.ReadString('\n')
		if err != nil {
			log.Errorln(err.Error())
		}
		password = strings.TrimRight(password, "\n")
		transaction, err := login.Authenticate(username, password)
		if err == nil {
			log.WithField("username", username).Infoln("User successfully authenticated himself.")
			_, err = conn.Write([]byte{login.LOGIN_ACCEPT})
			if err != nil {
				log.Errorln(err.Error())
			}
			return transaction, username, nil
		}
		log.WithField("username", username).Errorln("User failed to authenticate himself.")
		_, _ = conn.Write([]byte{login.LOGIN_FAIL})
	}
}
