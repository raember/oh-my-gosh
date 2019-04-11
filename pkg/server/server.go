package server

import (
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/shell"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"time"
)

type Server struct {
	config *viper.Viper
}

func NewServer(config *viper.Viper) Server {
	return Server{config: config}
}

// Serve a newly established connection.
func (server Server) Serve(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) error {
	timeout := make(chan bool, 1)
	loginChan := make(chan LoginResult)
	go func() {
		time.Sleep(time.Second * time.Duration(server.config.GetInt("Authentication.LoginGraceTime")))
		timeout <- true
	}()
	server.PerformLogin(loginChan, stdIn, stdOut, stdErr)
	var username string
	var transaction *pam.Transaction
	var err error
	select {
	case loginResult := <-loginChan:
		// a read from ch has occurred
		username = loginResult.username
		transaction = loginResult.transaction
		err = loginResult.error
	case <-timeout:
		err = errors.New("login timed out")
		log.WithField("error", err).Errorln("Login grace time exceeded.")
		_, _ = stdErr.Write([]byte(err.Error() + ""))
		return err
	}
	if err != nil {
		return err
	}

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
		log.WithField("error", err).Errorln("Couldn't open a session.")
	}
	defer transaction.CloseSession(pam.Silent)
	str, err := transaction.GetItem(pam.Service)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Service from transaction.")
	} else {
		log.Infoln("Service: " + str)
	}
	str, err = transaction.GetItem(pam.User)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting User from transaction.")
	} else {
		log.Infoln("User: " + str)
	}
	str, err = transaction.GetItem(pam.Tty)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Tty from transaction.")
	} else {
		log.Infoln("Tty: " + str)
	}
	str, err = transaction.GetItem(pam.Rhost)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Rhost from transaction.")
	} else {
		log.Infoln("Rhost: " + str)
	}
	str, err = transaction.GetItem(pam.Authtok)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Authtok from transaction.")
	} else {
		log.Infoln("Authtok: " + str)
	}
	str, err = transaction.GetItem(pam.Oldauthtok)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Oldauthtok from transaction.")
	} else {
		log.Infoln("Oldauthtok: " + str)
	}
	str, err = transaction.GetItem(pam.Ruser)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Ruser from transaction.")
	} else {
		log.Infoln("Ruser: " + str)
	}
	str, err = transaction.GetItem(pam.UserPrompt)
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting UserPrompt from transaction.")
	} else {
		log.Infoln("UserPrompt: " + str)
	}
	strs, err := transaction.GetEnvList()
	if err != nil {
		log.WithField("error", err).Errorln("Failed getting Env from transaction.")
	} else {
		log.Println("#envs: " + strconv.Itoa(len(strs)))
		for _, str := range strs {
			log.Infoln(str + ": " + strs[str])
		}
	}
	file, name, err := pty.Open()
	log.WithField("slavename", name).Infoln("Pts")
	log.WithField("filename", file.Name()).Infoln("File")

	user, err := pw.GetPwByName(username)
	log.WithFields(log.Fields{
		"USER":  user.Name,
		"UID":   user.Uid,
		"GID":   user.Gid,
		"HOME":  user.HomeDir,
		"SHELL": user.Shell,
	}).Println("Looked up user.")

	if _, err := os.Stat("/etc/motd"); err == nil {
		motd, err := ioutil.ReadFile("/etc/motd")
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't read message of the day.")
			return err
		}
		log.Println(motd)
	}

	creds := syscall.Credential{
		Uid:    user.Uid,
		Gid:    user.Gid,
		Groups: []uint32{},
	}

	sysattr := syscall.SysProcAttr{
		Credential: &creds,
	}

	attr := syscall.ProcAttr{
		Dir:   user.HomeDir,
		Env:   []string{},
		Files: []uintptr{},
		Sys:   &sysattr,
	}
	pid, err := syscall.ForkExec("sup?", []string{}, &attr)
	log.WithField("pid", pid).Println("Forked.")

	err = server.checkForNologinFile(stdIn, stdOut, stdErr)
	if err != nil {
		return err
	}

	// TODO: Make shell transmit everything over to client CORRECTLY.
	return shell.Execute(user.Shell, stdIn, stdOut, stdErr)
}

type LoginResult struct {
	transaction *pam.Transaction
	username    string
	error       error
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
func (server Server) PerformLogin(loginChan chan LoginResult, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) {
	go func() {
		tries := 1
		for {
			tries++
			transaction, username, err := login.Authenticate(stdIn, stdOut, stdErr)
			if err != nil {
				switch err.(type) {
				case login.AuthError: // Auth error -> continue
					log.WithField("username", username).Errorln("User failed to authenticate himself.")
					if tries > server.config.GetInt("Authentication.MaxTries") {
						err := errors.New("maximum tries reached")
						log.WithField("error", err).Errorln("User reached maximum tries.")
						loginChan <- LoginResult{nil, "", err}
						return
					}
					continue
				case error: // i.E. connection error -> abort
					loginChan <- LoginResult{transaction, username, err}
					return
				}
			} else {
				log.WithField("username", username).Infoln("User successfully authenticated himself.")
				loginChan <- LoginResult{transaction, username, nil}
				return
			}
		}
	}()
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
