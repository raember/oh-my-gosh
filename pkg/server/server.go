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
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/shell"
	"io"
	"io/ioutil"
	"os"
	"strconv"
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
func (server Server) Serve(stdIn io.Reader, stdOut io.Writer) error {
	log.WithFields(log.Fields{
		"stdIn":  stdIn,
		"stdOut": stdOut,
	}).Traceln("--> server.Server.Serve")
	user, err := server.PerformLogin(stdIn, stdOut)
	if err != nil {
		return err
	}

	err = server.checkForNologinFile(stdIn, stdOut)
	if err != nil {
		return err
	}

	// TODO: Redirect traffic through pts to the pty
	ptyFile, ptsName, err := pty.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err = ptyFile.Close(); err != nil {
			log.WithError(err).Errorln("Couldn't close pty file.")
		} else {
			log.Debugln("Closed pty file.")
		}
	}()

	if err = server.printMotD(stdIn, stdOut); err != nil {
		return err
	}

	ptsFile, err := os.Create(ptsName)
	if err != nil {
		log.WithError(err).Errorln("Couldn't open pts file.")
		return err
	}
	defer func() {
		if err = ptsFile.Close(); err != nil {
			log.WithError(err).Errorln("Couldn't close pts file.")
		} else {
			log.Debugln("Closed pts file.")
		}
	}()

	// Forward client to pts(shell)
	go func() {
		bufIn := bufio.NewReader(stdIn)
		for {
			n, err := bufIn.WriteTo(ptsFile)
			if err != nil {
				log.WithError(err).Errorln("Couldn't read from client.")
				break
			}
			if n > 0 {
				log.Debugln("Written " + strconv.Itoa(int(n)) + " bytes to pts.")
			}
		}
	}()

	// Forward pts(shell) output to client
	go func() {
		bufIn := bufio.NewReader(ptsFile)
		for {
			n, err := bufIn.WriteTo(stdOut)
			if err != nil {
				log.WithError(err).Errorln("Couldn't read from pts.")
				break
			}
			if n > 0 {
				log.Debugln("Read " + strconv.Itoa(int(n)) + " bytes from pts.")
			}
		}
	}()

	return shell.Execute(user.PassWd.Shell, ptyFile, ptyFile)
}

type LoginResult struct {
	user  *login.User
	error error
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
func (server Server) PerformLogin(stdIn io.Reader, stdOut io.Writer) (*login.User, error) {
	log.WithFields(log.Fields{
		"stdIn":  stdIn,
		"stdOut": stdOut,
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
			user, err := login.Authenticate(stdIn, stdOut)
			if err != nil {
				switch err.(type) {
				case *login.AuthError: // Auth error -> continue
					log.WithField("username", user.Name).Errorln("User failed to authenticate himself.")
					if try >= maxTries {
						err := errors.New("maximum try reached")
						log.WithError(err).Errorln("User reached maximum try.")
						_, _ = stdOut.Write([]byte(connection.MaxTriesExceededPacket{}.String()))
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
		_, _ = stdOut.Write([]byte(connection.TimeoutPacket{}.String()))
		return nil, err
	}
}

func (server Server) checkForNologinFile(stdIn io.Reader, stdOut io.Writer) error {
	log.WithFields(log.Fields{
		"stdIn":  stdIn,
		"stdOut": stdOut,
	}).Traceln("--> server.Server.checkForNologinFile")
	bytes, err := ioutil.ReadFile("/etc/nologin")
	if err != nil {
		log.Debugln("/etc/nologin file not found. Login permitted.")
		return nil
	}
	err = errors.New("/etc/nologin file exists: no login allowed")
	log.WithError(err).Errorln("/etc/nologin file exists. Login not permitted.")
	_, _ = stdOut.Write(bytes)
	return err
}

func (server Server) printMotD(stdIn io.Reader, stdOut io.Writer) error {
	log.WithFields(log.Fields{
		"stdIn":  stdIn,
		"stdOut": stdOut,
	}).Traceln("--> server.Server.printMotD")
	if _, err := os.Stat("/etc/motd"); err == nil {
		motd, err := ioutil.ReadFile("/etc/motd")
		if err != nil {
			log.WithError(err).Errorln("Couldn't read message of the day.")
			return err
		}
		log.WithField("motd", string(motd)).Debugln("Read message of the day.")
		_, _ = stdOut.Write(motd)
	}
	return nil
}
