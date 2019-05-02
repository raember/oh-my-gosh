package server

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"strings"
	"syscall"
)

type Host struct {
	config      *viper.Viper
	certificate *tls.Certificate
	conn        *net.Conn
	rAddr       *net.TCPAddr
	userEnvs    []string
	userName    string
	userPw      string
	rTerm       string
	rUser       string
	rHostname   string
	rUserPubKey string
	ptm         *os.File
	pts         *os.File
}

func NewHost(config *viper.Viper) Host {
	log.WithField("config", config).Traceln("--> host.NewHost")
	return Host{config: config}
}

func (host *Host) LoadCertKeyPair(certPath string, keyFilePath string) error {
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Traceln("--> host.loadCertKeyPair")
	cert, err := tls.LoadX509KeyPair(certPath, keyFilePath)
	if err != nil {
		log.WithError(err).Fatalln("Failed to load certificate key pair.")
	}
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Debugln("Loaded certificate key pair.")
	host.certificate = &cert
	return nil
}

func (host *Host) Connect(socketFd uintptr, rAddr *net.TCPAddr) error {
	log.WithFields(log.Fields{
		"socketFd": socketFd,
		"rAddr":    *rAddr,
	}).Traceln("--> host.Host.Connect")
	ErrorMsg := "Failed to handle connection."

	conn, err := connection.ConnFromFd(socketFd, host.certificate)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	log.WithFields(log.Fields{
		"conn":  conn,
		"rAddr": *rAddr,
	}).Infoln("Connected to client.")
	host.conn = &conn
	host.rAddr = rAddr
	return nil
}

func (host *Host) Setup() error {
	log.Traceln("--> host.Host.Setup")
	ErrorMsg := "Failed to setup host."
	// Get the necessary information from the rAddr client
	err := host.getClientEnvs("TERM")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	host.rUser, err = host.requestClientEnv("USER")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("rUser", host.rUser).Debugln("Got remote user name.")
	}
	host.rHostname, err = host.requestClientEnv("HOSTNAME")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("rHostname", host.rHostname).Debugln("Got remote host name.")
	}
	host.userName, err = host.requestClientEnv(common.ENV_GOSH_USER)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("userName", host.userName).Debugln("Got local user name.")
	}
	host.userPw, err = host.requestClientEnv(common.ENV_GOSH_PASSWORD)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("userPw", host.userPw).Debugln("Got local user password.")
	}
	host.rUserPubKey, err = host.requestClientEnv(common.ENV_GOSH_PUBLIC_KEY)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("rUserPubKey", host.rUserPubKey).Debugln("Got remote user public key.")
	}

	// Done gathering all the information.
	log.Infoln("Got all the information from the client.")
	if err = host.stopTransfer(); err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	host.ptm, host.pts, err = pty.Create()
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	log.Infoln("Set up host.")
	return nil
}

func (host *Host) Serve() error {
	log.Traceln("--> host.Host.Serve")
	ErrorMsg := "Failed to serve."
	var pid int
	var err error
	if host.rUserPubKey != "" {
		pid, err = host.loginWithKey()
	} else {
		pid, err = host.login()
	}
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	// TODO: Handle forwarding yourself.
	go connection.Forward(host.ptm, *host.conn, "ptm", "client")
	go connection.Forward(*host.conn, host.ptm, "client", "ptm")

	//rFdSet := unix.FdSet{}
	//n, err := unix.Pselect(3, &rFdSet, &rFdSet, &rFdSet, nil, nil)

	wpid, err := syscall.Wait4(pid, nil, 0, nil)
	if err != nil {
		log.WithError(err).Errorln("Failed waiting for login.")
		return err
	} else {
		log.WithField("wpid", wpid).Debugln("Waited for login.")
	}
	// TODO: Fix breakdown of connection and files.
	connection.CloseFile(host.pts)
	connection.CloseFile(host.ptm)
	connection.CloseConn(*host.conn)
	return nil
}

func (host *Host) getClientEnvs(envs ...string) error {
	log.WithField("envs", envs).Traceln("--> host.Host.getClientEnvs")
	host.userEnvs = []string{}
	for _, env := range envs {
		value, err := host.requestClientEnv(env)
		if err != nil {
			return err
		}
		host.userEnvs = append(host.userEnvs, fmt.Sprintf("%s=%s", env, value))
	}
	log.WithField("host.userEnvs", host.userEnvs).Debugln("Done gathering environment variables.")
	return nil
}

func (host Host) requestClientEnv(env string) (string, error) {
	log.WithField("env", env).Traceln("--> host.Host.requestClientEnv")
	log.WithField("env", env).Debugln("Requesting environment variable from client.")
	bufIn := bufio.NewReader(*host.conn)
	_, err := fmt.Fprint(*host.conn, connection.EnvPacket{Request: env}.String())
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

func (host Host) stopTransfer() error {
	log.Traceln("--> host.Host.stopTransfer")
	_, err := fmt.Fprint(*host.conn, connection.DonePacket{}.String())
	if err != nil {
		log.WithError(err).Errorln("Failed to send DonePacket.")
		return err
	}
	log.Debugln("Sent done packet.")
	return nil
}

func (host Host) loginWithKey() (int, error) {
	log.Traceln("--> server.Host.loginWithKey")
	return 0, errors.New("not implemented yet")
}

func (host Host) login() (int, error) {
	log.Traceln("--> server.Host.login")
	pid, err := syscall.ForkExec("/bin/login", []string{"-h", host.rHostname}, &syscall.ProcAttr{
		Files: []uintptr{
			host.pts.Fd(),
			host.pts.Fd(),
			host.pts.Fd(),
		},
		Env: host.userEnvs,
		Sys: &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
		},
	})
	if err != nil {
		log.WithError(err).Errorln("Failed to fork login.")
	} else {
		log.WithField("pid", pid).Debugln("Forked login.")
	}
	if host.userName != "" {
		if err := host.answerPtyLoginRequest(pid); err != nil {
			return pid, err
		}
		if host.userPw != "" {
			if err := host.answerPtyPasswordRequest(pid); err != nil {
				return pid, err
			}
		}
	}
	return pid, err
}

func interrupt(pid int) error {
	err := unix.Kill(pid, unix.SIGINT)
	if err != nil {
		log.WithError(err).Errorln("Failed to interrupt process.")
		err = unix.Kill(pid, unix.SIGKILL)
		if err != nil {
			log.WithError(err).Errorln("Failed to kill process.")
		} else {
			log.WithField("pid", pid).Infoln("Killed process.")
		}
	} else {
		log.WithField("pid", pid).Infoln("Interrupted process.")
	}
	return err
}

func (host Host) answerPtyLoginRequest(pid int) error {
	log.WithField("pid", pid).Traceln("--> server.Host.answerPtyLoginRequest")
	str, err := bufio.NewReader(host.ptm).ReadString(':')
	if err != nil || !strings.HasSuffix(str, "login:") {
		log.WithError(err).WithField("str", str).Errorln("Failed to read 'login:' from pty")
		_ = interrupt(pid)
		return err
	}
	_, err = host.ptm.WriteString(host.userName + "\n")
	if err != nil {
		log.WithError(err).Errorln("Failed to send user name to pty.")
		_ = interrupt(pid)
		return err
	}
	log.Infoln("Sent user name to process")
	return nil
}

func (host Host) answerPtyPasswordRequest(pid int) error {
	log.WithField("pid", pid).Traceln("--> server.Host.answerPtyPasswordRequest")
	str, err := bufio.NewReader(host.ptm).ReadString(':')
	if err != nil || !strings.HasSuffix(str, "Password:") {
		log.WithError(err).WithField("str", str).Errorln("Failed to read 'Password:' from pty")
		_ = interrupt(pid)
		return err
	}
	_, err = host.ptm.WriteString(host.userPw + "\n")
	if err != nil {
		log.WithError(err).Errorln("Failed to send password to pty.")
		_ = interrupt(pid)
		return err
	}
	log.Infoln("Sent password to process")
	return nil
}
