package server

import (
	"crypto/tls"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net"
	"net/url"
	"strconv"
)

type Lookout struct {
	address *url.URL
}

func NewLookout(protocol string, port int) (*Lookout, error) {
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

func (lookout Lookout) Listen(certpath string, keypath string) (net.Listener, error) {
	cert, err := lookout.loadCertKeyPair(certpath, keypath)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	protocol := lookout.address.Scheme
	address := ":" + lookout.address.Port()
	listener, err := tls.Listen(protocol, address, tlsConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": protocol,
			"address":  address,
			"error":    err.Error(),
		}).Fatalln("Couldn't setup listener.")
		return nil, err
	}
	log.WithFields(log.Fields{
		"protocol": protocol,
		"address":  address,
	}).Infoln("Listening for incoming requests.")
	return listener, nil
}

func (lookout Lookout) loadCertKeyPair(certPath string, keyFilePath string) (tls.Certificate, error) {
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

func WaitForConnections(listener net.Listener, handler func(net.Conn)) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Errorln("Failed opening connection.")
			return err
		}
		log.WithFields(log.Fields{
			"remote": common.AddrToStr(conn.RemoteAddr()),
		}).Infoln("Connection established.")

		go handler(conn)
	}
}
