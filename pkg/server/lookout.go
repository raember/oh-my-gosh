package server

import (
	"crypto/tls"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/socket"
	"net/url"
	"strconv"
)

type Lookout struct {
	address *url.URL
}

func NewLookout(protocol string, port int) (*Lookout, error) {
	log.WithFields(log.Fields{
		"protocol": protocol,
		"port":     port,
	}).Traceln("server.NewLookout")
	if port < 0 {
		err := errors.New("port cannot be negative")
		log.WithFields(log.Fields{
			"port": port,
		}).Errorln(err.Error())
		return nil, err
	}
	addrStr := protocol + "://localhost:" + strconv.Itoa(port)
	reqUrl, err := url.ParseRequestURI(addrStr)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": protocol,
			"port":     port,
			"addrStr":  addrStr,
		}).Errorln(err.Error())
		return nil, err
	}
	if reqUrl.Scheme != common.TCP &&
		reqUrl.Scheme != common.TCP4 &&
		reqUrl.Scheme != common.TCP6 &&
		reqUrl.Scheme != common.UNIX &&
		reqUrl.Scheme != common.UNIXPACKET {
		err := errors.New("protocol has to be either tcp, tcp4, tcp6, unix or unixpacket")
		log.WithFields(log.Fields{
			"protocol": reqUrl.Scheme,
		}).Errorln(err.Error())
		return nil, err
	}
	return &Lookout{address: reqUrl}, nil
}

func (lookout Lookout) Listen(certpath string, keypath string) (uintptr, error) {
	log.WithFields(log.Fields{
		"certpath": certpath,
		"keypath":  keypath,
	}).Traceln("server.Lookout.Listen")
	protocol := lookout.address.Scheme
	portStr := lookout.address.Port()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.WithFields(log.Fields{
			"port":  lookout.address.Port(),
			"error": err.Error(),
		}).Fatalln("Couldn't convert port to int.")
		return 0, err
	}
	socketFd, err := socket.Listen(port)
	if err != nil {
		return 0, err
	}
	log.WithFields(log.Fields{
		"protocol": protocol,
		"port":     portStr,
	}).Infoln("Listening for incoming requests.")
	return socketFd, nil
}

func (lookout Lookout) loadCertKeyPair(certPath string, keyFilePath string) (tls.Certificate, error) {
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Traceln("server.Lookout.loadCertKeyPair")
	cert, err := tls.LoadX509KeyPair(certPath, keyFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"certPath":    certPath,
			"keyFilePath": keyFilePath,
			"error":       err.Error(),
		}).Fatalln("Couldn't load certificate key pair.")
		return tls.Certificate{}, err
	}
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Infoln("Loaded certificate key pair.")
	return cert, nil
}

func WaitForConnections(socketFd uintptr, handler func(uintptr)) error {
	log.WithFields(log.Fields{
		"socketFd": socketFd,
		"handler":  handler,
	}).Traceln("server.WaitForConnections")
	for {
		connFd, err := socket.Accept(socketFd)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Errorln("Failed opening connection.")
			return err
		}

		addr, err := socket.GetPeerName(connFd)
		if err != nil {
			return err
		}
		log.Infoln(addr.String())
		log.WithField("connFd", connFd).Debugln("Handle connection")
		go handler(connFd)
	}
}
