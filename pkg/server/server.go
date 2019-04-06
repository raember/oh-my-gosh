package server

import (
	"bufio"
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/willdonnelly/passwd"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Server struct {
	config *viper.Viper
}

func NewServer(config *viper.Viper) Server {
	return Server{config: config}
}

// Serve a newly established connection.
func (server Server) Serve(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) {
	client := bufio.NewReader(stdIn)

	transaction, username, err := server.PerformLogin(stdIn, stdOut, stdErr)
	if err != nil {
		return
	}

	message := "Authentication was successful"
	log.WithFields(log.Fields{
		"message": message,
	}).Infoln("Outbound")
	_, _ = stdOut.Write([]byte(message + "\n"))

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

	uidInt, _ := strconv.ParseUint(UID, 10, 32)
	gidInt, _ := strconv.ParseUint(GID, 10, 32)
	creds := syscall.Credential{
		Uid:    uint32(uidInt),
		Gid:    uint32(gidInt),
		Groups: []uint32{},
	}

	sysattr := syscall.SysProcAttr{
		Credential: &creds,
	}

	attr := syscall.ProcAttr{
		Dir:   HOME,
		Env:   []string{},
		Files: []uintptr{},
		Sys:   &sysattr,
	}
	pid, err := syscall.ForkExec("sup?", []string{}, &attr)
	log.WithField("pid", pid).Println("Forked.")

	err = server.checkForNologinFile(stdIn, stdOut, stdErr)
	if err != nil {
		return
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
		_, _ = stdOut.Write([]byte(answer + "\n"))
	}
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
func (server Server) PerformLogin(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) (*pam.Transaction, string, error) {
	tries := 1
	for {
		tries++
		transaction, username, err := login.Authenticate(stdIn, stdOut, stdErr)
		if err == nil {
			log.WithField("username", username).Infoln("User successfully authenticated himself.")
			return transaction, username, nil
		}
		log.WithField("username", username).Errorln("User failed to authenticate himself.")
		if tries > server.config.GetInt("Authentication.MaxTries") {
			err := errors.New("maximum tries reached")
			log.WithField("error", err).Errorln("User reached maximum tries.")
			return nil, "", err
		}
	}
}

func (server Server) checkForNologinFile(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) error {
	bytes, err := ioutil.ReadFile("/etc/nologin")
	if err != nil {
		log.Debugln("/etc/nologin file not found. Login permitted.")
		return nil
	}
	err = errors.New("/etc/nologin file exists: no login allowed")
	log.WithField("error", err).Infoln("/etc/nologin file exists. Login not permitted.")
	_, _ = stdOut.Write(bytes)
	return err
}
