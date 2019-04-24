package server

import (
	"bufio"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type Server struct {
	config *viper.Viper
}

func NewServer(config *viper.Viper) Server {
	log.WithField("config", config).Traceln("--> server.NewServer")
	return Server{config: config}
}

// Serve a newly established connection.
func (server Server) Serve(in io.Reader, out io.Writer) (*os.File, string, uint32, error) {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> server.Server.Serve")
	user, err := server.PerformLogin("", in, out)
	if err != nil {
		return nil, "", 0, err
	}

	err = server.checkForNologinFile(in, out)
	if err != nil {
		return nil, "", 0, err
	}

	ptyFd, ptsName, err := pty.Create()
	if err != nil {
		return nil, "", 0, err
	}
	ptyFile := os.NewFile(ptyFd, "pty")

	if err = server.printMotD(in, out); err != nil {
		return nil, "", 0, err
	}

	// Forward client to shell
	go func() {
		bufIn := bufio.NewReader(in)
		for {
			n, err := bufIn.WriteTo(ptyFile)
			if err != nil {
				log.WithError(err).Errorln("Couldn't read from client.")
				break
			}
			if n > 0 {
				log.WithField("n", n).Debugln("Wrote to pty.")
			}
		}
	}()

	// Forward shell output to client
	go func() {
		bufIn := bufio.NewReader(ptyFile)
		for {
			n, err := bufIn.WriteTo(out)
			if err != nil {
				log.WithError(err).Errorln("Couldn't read from pty.")
				break
			}
			if n > 0 {
				log.WithField("n", n).Debugln("Wrote to client.")
			}
		}
	}()

	return ptyFile, ptsName, user.PassWd.Uid, nil
}

type LoginResult struct {
	user  *login.User
	error error
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
func (server Server) PerformLogin(userName string, in io.Reader, out io.Writer) (*login.User, error) {
	log.WithFields(log.Fields{
		"userName": userName,
		"in":       in,
		"out":      out,
	}).Traceln("--> server.Server.PerformLogin")
	timeout := make(chan bool, 1)
	loginChan := make(chan LoginResult)
	go func() {
		dur := time.Second * time.Duration(server.config.GetInt("Authentication.LoginGraceTime"))
		log.WithField("duration", dur).Traceln("Go authentication timeout")
		time.Sleep(dur)
		timeout <- true
		log.Traceln("End authentication timeout")
	}()
	go func() {
		log.Traceln("Go authentication try")
		try := 0
		maxTries := server.config.GetInt("Authentication.MaxTries")
		for {
			try++
			log.WithFields(log.Fields{
				"try":      try,
				"maxTries": maxTries,
			}).Debugln("Start authentication.")
			user, err := login.Authenticate(userName, in, out)
			if err != nil {
				switch err.(type) {
				case *login.AuthError: // Auth error -> continue
					log.WithField("username", user.Name).Errorln("User failed to authenticate himself.")
					if try >= maxTries {
						err := errors.New("maximum try reached")
						log.WithError(err).Errorln("User reached maximum try.")
						_, _ = out.Write([]byte(connection.MaxTriesExceededPacket{}.String()))
						loginChan <- LoginResult{user, err}
						return
					}
					continue
				case error: // i.E. connection error -> abort
					log.WithError(err).Errorln("Failed to authenticate user.")
					loginChan <- LoginResult{user, err}
					return
				}
			}
			log.WithField("username", user.Name).Infoln("User successfully authenticated himself.")
			loginChan <- LoginResult{user, nil}
			log.Traceln("End authentication try")
			return
		}
	}()
	select {
	case res := <-loginChan:
		return res.user, res.error
	case <-timeout:
		err := errors.New("login timed out")
		log.WithError(err).Errorln("Login grace time exceeded.")
		_, _ = out.Write([]byte(connection.TimeoutPacket{}.String()))
		return nil, err
	}
}

func (server Server) checkForNologinFile(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> server.Server.checkForNologinFile")
	bytes, err := ioutil.ReadFile("/etc/nologin")
	if err != nil {
		log.Debugln("/etc/nologin file not found. Login permitted.")
		return nil
	}
	err = errors.New("/etc/nologin file exists: no login allowed")
	log.WithError(err).Errorln("/etc/nologin file exists. Login not permitted.")
	_, _ = out.Write(bytes)
	return err
}

func (server Server) printMotD(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> server.Server.printMotD")
	if _, err := os.Stat("/etc/motd"); err == nil {
		motd, err := ioutil.ReadFile("/etc/motd")
		if err != nil {
			log.WithError(err).Errorln("Couldn't read message of the day.")
			return err
		}
		log.WithField("motd", string(motd)).Debugln("Read message of the day.")
		_, _ = out.Write(motd)
	}
	return nil
}
