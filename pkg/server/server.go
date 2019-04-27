package server

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/login"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/socket"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
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
func (server Server) Serve(addr net.Addr, in io.Reader, out io.Writer, conn *net.Conn) {
	log.WithFields(log.Fields{
		"addr": addr,
		"in":   &in,
		"out":  &out,
		"conn": &conn,
	}).Traceln("--> server.Server.Serve")

	envs, err := server.getClientEnvs([]string{"TERM"}, in, out)
	if err != nil {
		return
	}
	rUser, err := server.requestClientEnv("USER", in, out)
	if err != nil {
		return
	}
	username, err := server.requestClientEnv("GOSH_USER", in, out)
	if err != nil {
		return
	}
	data := &login.TransactionData{
		Service:    os.Args[0],
		User:       username,
		Tty:        "",
		RHost:      addr.String(),
		Authtok:    "",
		Oldauthtok: "",
		RUser:      rUser,
		UserPrompt: "",
	}

	if server.stopTransfer(in, out) != nil {
		return
	}

	ptmFd, ptsName := pty.Create()
	ptmFile := os.NewFile(ptmFd, "ptm")
	defer connection.CloseFile(ptmFile, "ptm")
	go connection.Forward(in, ptmFile, "client", "ptm")
	go connection.Forward(ptmFile, out, "ptm", "client")
	ptsFile, err := os.Create(ptsName)
	if err != nil {
		log.WithError(err).Fatalln("Couldn't open pts file.")
	}
	defer connection.CloseFile(ptsFile, "pts")

	//user := server.performLogin(data, ptsFile, ptsFile, fd)
	//
	//server.checkForNologinFile(ptsFile, ptsFile)
	//
	//server.printMotD(ptsFile, ptsFile)
	//
	//pwd, err := pw.GetPwByUid(user.PassWd.Uid)
	//if err != nil {
	//	log.WithError(err).Fatalln("Couldn't create new file.")
	//}
	//_ = cmd.Execute(pwd, envs, ptsFile, ptsFile)

	cmd := exec.Command("/bin/login", "-h", data.RHost)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	//hostname, err := os.Hostname()
	//if err != nil {
	//	log.WithError(err).Fatalln("Couldn't lookup hostname")
	//}
	//envs = append([]string{
	//	"USER=" + passWd.Name,
	//	"UID=" + strconv.Itoa(int(passWd.Uid)),
	//	"GID=" + strconv.Itoa(int(passWd.Gid)),
	//	"HOME=" + passWd.HomeDir,
	//	"SHELL=" + passWd.Shell,
	//	"HOSTNAME=" + hostname,
	//}, envs...)
	cmd.Env = envs
	cmd.Stdin = ptsFile
	cmd.Stdout = ptsFile
	cmd.Stderr = ptsFile
	err = cmd.Run()
	if err != nil {
		log.WithError(err).Errorln("An error occured.")
	} else {
		log.Debugln("Shell terminated.")
	}
	// TODO: Fix breakdown of connection and files.
}

type LoginResult struct {
	user  *login.User
	error error
}

func (server Server) getClientEnvs(envs []string, in io.Reader, out io.Writer) ([]string, error) {
	log.WithFields(log.Fields{
		"envs": envs,
		"in":   &in,
		"out":  &out,
	}).Traceln("--> server.getClientEnvs")
	var filledEnvs []string
	for _, env := range envs {
		value, err := server.requestClientEnv(env, in, out)
		if err != nil {
			return nil, err
		}
		filledEnvs = append(filledEnvs, fmt.Sprintf("%s=%s", env, value))
	}
	log.WithField("filledEnvs", filledEnvs).Debugln("Done gathering environment variables.")
	return filledEnvs, nil
}

func (server Server) stopTransfer(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> server.stopTransfer")
	_, err := fmt.Fprint(out, connection.DonePacket{}.String())
	if err != nil {
		log.WithError(err).Errorln("Couldn't send DonePacket.")
		return err
	}
	log.Debugln("Sent done packet.")
	return nil
}

func (server Server) requestClientEnv(env string, in io.Reader, out io.Writer) (string, error) {
	log.WithFields(log.Fields{
		"env": env,
		"in":  &in,
		"out": &out,
	}).Traceln("--> server.requestClientEnv")
	log.WithField("env", env).Debugln("Requesting environment variable from client.")
	bufIn := bufio.NewReader(in)
	_, err := fmt.Fprint(out, connection.EnvPacket{Request: env}.String())
	if err != nil {
		log.WithError(err).Errorln("Failed to send packet.")
		return "", err
	}
	value, err := bufIn.ReadString('\n')
	if err != nil {
		log.WithError(err).Errorln("Failed to read environment variable value.")
		return "", err
	}
	value = strings.TrimSpace(value)
	log.WithField("value", value).Debugln("Read environment variable value.")
	return value, nil
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
func (server Server) performLogin(data *login.TransactionData, in io.Reader, out io.Writer, fd uintptr) *login.User {
	log.WithFields(log.Fields{
		"data": data,
		"in":   &in,
		"out":  &out,
		"fd":   fd,
	}).Traceln("--> server.Server.performLogin")
	timeout := make(chan bool, 1)
	loginChan := make(chan LoginResult)
	go func() {
		GOROUTINE := "timeout"
		log.WithField("GoRoutine", GOROUTINE).Traceln("==> Go")
		dur := time.Second * time.Duration(server.config.GetInt("Authentication.LoginGraceTime"))
		log.WithFields(log.Fields{
			"GoRoutine": GOROUTINE,
			"duration":  dur,
		}).Debugln("Set timeout")
		time.Sleep(dur)
		timeout <- true
		log.WithField("GoRoutine", GOROUTINE).Traceln("==> End")
	}()
	go func() {
		GOROUTINE := "authtry"
		log.WithField("GoRoutine", GOROUTINE).Traceln("==> Go")
		try := 0
		maxTries := server.config.GetInt("Authentication.MaxTries")
		log.WithFields(log.Fields{
			"GoRoutine": GOROUTINE,
			"maxTries":  maxTries,
		}).Debugln("Set maximum tries")
		for {
			try++
			log.WithFields(log.Fields{
				"GoRoutine": GOROUTINE,
				"try":       try,
				"maxTries":  maxTries,
			}).Debugln("Start authentication.")
			user, err := login.Authenticate(data, in, out, fd)
			if err != nil {
				switch err.(type) {
				case *login.AuthError:
					log.WithField("GoRoutine", GOROUTINE).WithError(err).Errorln("User failed to authenticate himself.")
					if err.Error() == ": Authentication failure" { // General error
						log.WithField("GoRoutine", GOROUTINE).Traceln("==> End")
						return
					}
					if try >= maxTries {
						err := errors.New("maximum try reached")
						log.WithField("GoRoutine", GOROUTINE).WithError(err).Errorln("User reached maximum try.")
						_, _ = fmt.Fprint(out, err.Error()+"\n")
						loginChan <- LoginResult{user, err}
						return
					}
					continue
				case error: // i.E. connection error -> abort
					log.WithField("GoRoutine", GOROUTINE).WithError(err).Errorln("Failed to authenticate user.")
					loginChan <- LoginResult{user, err}
					log.WithField("GoRoutine", GOROUTINE).Traceln("==> End")
					return
				}
			}
			log.WithFields(log.Fields{
				"GoRoutine": GOROUTINE,
				"username":  user.Name,
			}).Infoln("User successfully authenticated himself.")
			loginChan <- LoginResult{user, nil}
			log.WithField("GoRoutine", GOROUTINE).Traceln("==> End")
			return
		}
	}()
	select {
	case res := <-loginChan:
		return res.user
	case <-timeout:
		err := errors.New("login timed out")
		_, _ = fmt.Fprint(out, err.Error()+"\n")
		log.WithError(err).Fatalln("Login grace time exceeded.")
	}
	return nil
}

func (server Server) checkForNologinFile(in io.Reader, out io.Writer) {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> server.Server.checkForNologinFile")
	bytes, err := ioutil.ReadFile("/etc/nologin")
	if err != nil {
		log.Debugln("/etc/nologin file not found. Login permitted.")
		return
	}
	err = errors.New("/etc/nologin file exists: no login allowed")
	_, _ = out.Write(bytes)
	log.WithError(err).Fatalln("/etc/nologin file exists. Login not permitted.")
}

func (server Server) printMotD(in io.Reader, out io.Writer) {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> server.Server.printMotD")
	if _, err := os.Stat("/etc/motd"); err == nil {
		motd, err := ioutil.ReadFile("/etc/motd")
		if err != nil {
			log.WithError(err).Fatalln("Couldn't read message of the day.")
		}
		log.WithField("motd", string(motd)).Debugln("Read message of the day.")
		_, _ = out.Write(motd)
	}
}
func LoadCertKeyPair(certPath string, keyFilePath string) tls.Certificate {
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Traceln("--> server.Lookout.loadCertKeyPair")
	cert, err := tls.LoadX509KeyPair(certPath, keyFilePath)
	if err != nil {
		log.WithError(err).Fatalln("Couldn't load certificate key pair.")
	}
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Debugln("Loaded certificate key pair.")
	return cert
}

func WaitForConnections(listenerFd uintptr, fdChan chan uintptr) {
	log.WithFields(log.Fields{
		"listenerFd": listenerFd,
		"fdChan":     fdChan,
	}).Traceln("--> server.WaitForConnections")
	for {
		socketFd, err := socket.Accept(listenerFd)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Errorln("Failed opening connection.")
		} else {
			log.WithField("socketFd", socketFd).Debugln("Handle connection.")
			fdChan <- socketFd
		}
	}
}
