package host

import (
	"bufio"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"net"
	"strings"
	"syscall"
)

type Host struct {
	config      *viper.Viper
	certificate tls.Certificate
}

func NewHost(config *viper.Viper, certPath string, keyFilePath string) Host {
	log.WithFields(log.Fields{
		"config":      config,
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Traceln("--> host.NewHost")
	return Host{
		config:      config,
		certificate: loadCertKeyPair(certPath, keyFilePath),
	}
}

func (host Host) HandleConnection(socketFd uintptr, rAddr *net.TCPAddr) error {
	log.WithFields(log.Fields{
		"socketFd": socketFd,
		"rAddr":    *rAddr,
	}).Traceln("--> host.Host.HandleConnection")
	ErrorMsg := "Failed to handle connection."

	conn, err := connection.ConnFromFd(socketFd, host.certificate)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	log.WithField("rAddr", *rAddr).Infoln("Connected to client.")

	// Get the necessary information from the remote client
	remote := rAddr.IP.String()
	envs, err := host.getClientEnvs(conn, "TERM")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	rUser, err := host.requestClientEnv("USER", conn)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	username, err := host.requestClientEnv("GOSH_USER", conn)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	if err = host.stopTransfer(conn); err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	// Done gathering all the information.
	log.WithFields(log.Fields{
		"envs":     envs,
		"rUser":    rUser,
		"username": username,
		"remote":   remote,
	}).Infoln("Got all the information from the client.")

	ptm, pts, err := pty.Create()
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	pid, err := syscall.ForkExec("/bin/login", []string{"-h", remote}, &syscall.ProcAttr{
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

	go connection.Forward(ptm, conn, "ptm", "client")
	go connection.Forward(conn, ptm, "client", "ptm")

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
	connection.CloseConn(conn)
	return nil
}

func (host Host) getClientEnvs(conn net.Conn, envs ...string) ([]string, error) {
	log.WithFields(log.Fields{
		"conn": conn,
		"envs": envs,
	}).Traceln("--> host.Host.getClientEnvs")
	var filledEnvs []string
	for _, env := range envs {
		value, err := host.requestClientEnv(env, conn)
		if err != nil {
			return nil, err
		}
		filledEnvs = append(filledEnvs, fmt.Sprintf("%s=%s", env, value))
	}
	log.WithField("filledEnvs", filledEnvs).Debugln("Done gathering environment variables.")
	return filledEnvs, nil
}

func (host Host) requestClientEnv(env string, conn net.Conn) (string, error) {
	log.WithFields(log.Fields{
		"conn": conn,
		"env":  env,
	}).Traceln("--> host.Host.requestClientEnv")
	log.WithField("env", env).Debugln("Requesting environment variable from client.")
	bufIn := bufio.NewReader(conn)
	_, err := fmt.Fprint(conn, connection.EnvPacket{Request: env}.String())
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

func (host Host) stopTransfer(conn net.Conn) error {
	log.WithField("conn", conn).Traceln("--> host.Host.stopTransfer")
	_, err := fmt.Fprint(conn, connection.DonePacket{}.String())
	if err != nil {
		log.WithError(err).Errorln("Failed to send DonePacket.")
		return err
	}
	log.Debugln("Sent done packet.")
	return nil
}

func loadCertKeyPair(certPath string, keyFilePath string) tls.Certificate {
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
	return cert
}
