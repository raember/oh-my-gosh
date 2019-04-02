package server

import (
	"crypto/tls"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net"
	"net/url"
	"os"
	"strconv"
)

func init() {
	err := os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't set variable GODEBUG for TLS1.3 support")
		os.Exit(1)
	}
}

type Socket struct {
	address  *url.URL
	listener net.Listener
}

func NewSocket(protocol string, port int) (*Socket, error) {
	if port < 0 {
		err := errors.New("port cannot be negative")
		log.WithFields(log.Fields{
			"port": port,
		}).Errorln(err.Error())
		return nil, err
	}
	addrStr := protocol + "://" + common.LOCALHOST + ":" + strconv.Itoa(port)
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
	return &Socket{address: reqUrl}, nil
}

func (socket Socket) Listen() error {
	_, _ = tls.Listen(socket.address.Scheme, socket.address.Host, nil)
	listener, err := net.Listen(socket.address.Scheme, socket.address.Host)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": socket.address.Scheme,
			"host":     socket.address.Host,
		}).Errorln(err.Error())
		return err
	}
	socket.listener = listener
	log.WithFields(log.Fields{
		"address": socket.address.String(),
	}).Infoln("Listening for incoming requests.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"protocol": socket.address.Scheme,
				"host":     socket.address.Host,
			}).Errorln(err.Error())
			return err
		}
		log.WithFields(log.Fields{
			"remoteaddress": conn.RemoteAddr(),
		}).Infoln("Connection established.")
		server := Server{}
		go server.Serve(conn)
	}
}
