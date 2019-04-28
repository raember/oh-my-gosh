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
	"os"
	"strings"
	"syscall"
)

type Host struct {
	config      *viper.Viper
	certificate tls.Certificate
	conn        net.Conn
}

func NewHost(config *viper.Viper, certPath string, keyFilePath string) Host {
	log.WithField("config", config).Traceln("--> host.NewHost")
	return Host{
		config:      config,
		certificate: loadCertKeyPair(certPath, keyFilePath),
		conn:        nil,
	}
}

func (host Host) HandleConnection(fd uintptr, rAddr *net.TCPAddr) {
	log.WithFields(log.Fields{
		"fd":    fd,
		"rAddr": *rAddr,
	}).Traceln("--> host.Host.HandleConnection")

	conn, err := net.FileConn(os.NewFile(fd, "conn"))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"fd":    fd,
		}).Errorln("Failed to make a connection from file descriptor.")
		return
	}
	host.conn = tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{host.certificate},
	})
	log.WithField("rAddr", *rAddr).Infoln("Connected to client.")

	// Get the necessary information from the remote client
	envs, err := host.getClientEnvs("TERM")
	if err != nil {
		return
	}
	rUser, err := host.requestClientEnv("USER")
	if err != nil {
		return
	}
	username, err := host.requestClientEnv("GOSH_USER")
	if err != nil {
		return
	}
	if host.stopTransfer() != nil {
		return
	}

	// Done gathering all the information.
	log.WithFields(log.Fields{
		"envs":     envs,
		"rUser":    rUser,
		"username": username,
	}).Infoln("Got all the information from the client.")

	ptmFd, ptsName := pty.Create()
	ptmFile := os.NewFile(ptmFd, "ptm")
	defer connection.CloseFile(ptmFile, "ptm")
	go connection.Forward(host.conn, ptmFile, "client", "ptm")
	go connection.Forward(ptmFile, host.conn, "ptm", "client")
	ptsFile, err := os.Create(ptsName)
	if err != nil {
		log.WithError(err).Fatalln("Failed to open pts file.")
	}
	defer connection.CloseFile(ptsFile, "pts")

	//user := host.performLogin(data, ptsFile, ptsFile, fd)
	//
	//host.checkForNologinFile(ptsFile, ptsFile)
	//
	//host.printMotD(ptsFile, ptsFile)
	//
	//pwd, err := pw.GetPwByUid(user.PassWd.Uid)
	//if err != nil {
	//	log.WithError(err).Fatalln("Failed to create new file.")
	//}
	//_ = cmd.Execute(pwd, envs, ptsFile, ptsFile)

	ptsFd, err := syscall.Dup(int(ptsFile.Fd()))
	if err != nil {
		log.WithError(err).Fatalln("Failed to duplicate fd.")
	}
	pid, err := syscall.ForkExec("/bin/login", []string{"-h", rAddr.IP.String()}, &syscall.ProcAttr{
		Files: []uintptr{
			uintptr(ptsFd),
			uintptr(ptsFd),
			uintptr(ptsFd),
		},
		Env: envs,
		Sys: &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
		},
	})
	if err != nil {
		log.WithError(err).Fatalln("Failed to fork login.")
	} else {
		log.WithField("pid", pid).Debugln("Forked login.")
	}

	wpid, err := syscall.Wait4(pid, nil, 0, nil)
	if err != nil {
		log.WithError(err).Errorln("Failed waiting for login.")
	} else {
		log.WithField("wpid", wpid).Debugln("Waited for login.")
	}
	// TODO: Fix breakdown of connection and files.
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
	println(host.conn)
	bufIn := bufio.NewReader(host.conn)
	_, err := fmt.Fprint(host.conn, connection.EnvPacket{Request: env}.String())
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
	_, err := fmt.Fprint(host.conn, connection.DonePacket{}.String())
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
