package server

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"net"
	"strings"
	"syscall"
)

type Host struct {
	config      *viper.Viper
	certificate *tls.Certificate
	conn        *net.Conn
	rAddr       *net.TCPAddr
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

func (host *Host) Serve() error {
	log.Traceln("--> host.Host.Serve")
	ErrorMsg := "Failed to handle connection."

	// Get the necessary information from the rHost client
	rHost := host.rAddr.IP.String()
	envs, err := host.getClientEnvs("TERM")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	rUser, err := host.requestClientEnv("USER")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	rHostname, err := host.requestClientEnv("HOSTNAME")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	username, err := host.requestClientEnv(common.ENV_GOSH_USER)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	password, err := host.requestClientEnv(common.ENV_GOSH_PASSWORD)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	// Done gathering all the information.
	log.WithFields(log.Fields{
		"envs":      envs,
		"rUser":     rUser,
		"username":  username,
		"password":  password,
		"rHost":     rHost,
		"rHostname": rHostname,
	}).Infoln("Got all the information from the client.")
	if err = host.stopTransfer(); err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	ptm, pts, err := pty.Create()
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	pid, err := syscall.ForkExec("/bin/login", []string{"-h", rHost}, &syscall.ProcAttr{
		Files: []uintptr{
			pts.Fd(),
			pts.Fd(),
			pts.Fd(),
		},
		Env: envs,
		Sys: &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
		},
	})
	if err != nil {
		log.WithError(err).Errorln("Failed to fork login.")
		return err
	} else {
		log.WithField("pid", pid).Debugln("Forked login.")
	}

	// TODO: Handle forwarding yourself.
	go connection.Forward(ptm, *host.conn, "ptm", "client")
	go connection.Forward(*host.conn, ptm, "client", "ptm")

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
	connection.CloseFile(pts)
	connection.CloseFile(ptm)
	connection.CloseConn(*host.conn)
	return nil
}

func (host Host) getClientEnvs(envs ...string) ([]string, error) {
	log.WithField("envs", envs).Traceln("--> host.Host.getClientEnvs")
	var filledEnvs []string
	for _, env := range envs {
		value, err := host.requestClientEnv(env)
		if err != nil {
			return nil, err
		}
		filledEnvs = append(filledEnvs, fmt.Sprintf("%s=%s", env, value))
	}
	log.WithField("filledEnvs", filledEnvs).Debugln("Done gathering environment variables.")
	return filledEnvs, nil
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
