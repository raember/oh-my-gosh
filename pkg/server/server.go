package server

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/shell"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type Server struct {
	config *viper.Viper
}

func NewServer(config *viper.Viper) Server {
	return Server{config: config}
}

// Serve a newly established connection.
func (server Server) Serve(stdIn io.Reader, stdOut io.Writer) error {
	timeout := make(chan bool, 1)
	loginChan := make(chan LoginResult)
	go func() {
		time.Sleep(time.Second * time.Duration(server.config.GetInt("Authentication.LoginGraceTime")))
		timeout <- true
	}()
	server.PerformLogin(loginChan, stdIn, stdOut)
	var user *login.User
	var err error
	select {
	case loginResult := <-loginChan:
		user = loginResult.user
		err = loginResult.error
	case <-timeout:
		err = errors.New("login timed out")
		log.WithField("error", err).Errorln("Login grace time exceeded.")
		_, _ = stdOut.Write([]byte(connection.TimeoutPacket{}.String()))
		return err
	}
	if err != nil {
		return err
	}
	err = user.Setup()
	if err != nil {
		return err
	}
	file, name, err := pty.Open()
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't open pseudo-terminal.")
		return err
	}
	log.WithField("slavename", name).Infoln("Pts")
	log.WithField("filename", file.Name()).Infoln("File")

	if _, err := os.Stat("/etc/motd"); err == nil {
		motd, err := ioutil.ReadFile("/etc/motd")
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't read message of the day.")
			return err
		}
		log.Println(motd)
	}

	err = server.checkForNologinFile(stdIn, stdOut)
	if err != nil {
		return err
	}

	// TODO: Make shell transmit everything over to client CORRECTLY.
	// TODO: Fix segfault
	return shell.Execute(user.PassWd.Shell, stdIn, stdOut)
}

type LoginResult struct {
	user  *login.User
	error error
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
func (server Server) PerformLogin(loginChan chan LoginResult, stdIn io.Reader, stdOut io.Writer) {
	go func() {
		tries := 1
		for {
			tries++
			user, err := login.Authenticate(stdIn, stdOut)
			if err != nil {
				switch err.(type) {
				case login.AuthError: // Auth error -> continue
					log.WithField("username", user).Errorln("User failed to authenticate himself.")
					if tries > server.config.GetInt("Authentication.MaxTries") {
						err := errors.New("maximum tries reached")
						log.WithField("error", err).Errorln("User reached maximum tries.")
						loginChan <- LoginResult{user, err}
						return
					}
					continue
				case error: // i.E. connection error -> abort
					loginChan <- LoginResult{user, err}
					return
				}
			} else {
				log.WithField("username", user).Infoln("User successfully authenticated himself.")
				loginChan <- LoginResult{user, nil}
				return
			}
		}
	}()
}

func (server Server) checkForNologinFile(stdIn io.Reader, stdOut io.Writer) error {
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
