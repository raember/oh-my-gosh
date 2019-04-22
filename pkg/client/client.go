package client

import (
	"bufio"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	address *url.URL
}

func NewClient(protocol string, address string, port int) (*Client, error) {
	log.WithFields(log.Fields{
		"protocol": protocol,
		"address":  address,
		"port":     port,
	}).Traceln("client.NewClient")
	if port < 0 {
		err := errors.New("connection cannot be negative")
		log.WithFields(log.Fields{
			"connection": port,
		}).Errorln(err.Error())
		return nil, err
	}
	if address == "" {
		err := errors.New("address cannot be empty")
		log.WithFields(log.Fields{
			"connection": port,
		}).Errorln(err.Error())
		return nil, err
	}
	addrStr := protocol + "://" + address + ":" + strconv.Itoa(port)
	reqUrl, err := url.ParseRequestURI(addrStr)
	if err != nil {
		log.WithFields(log.Fields{
			"addrStr": addrStr,
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
	return &Client{address: reqUrl}, nil
}

func (client Client) Dial() (net.Conn, error) {
	log.Traceln("client.Client.Dial")
	address := client.address
	//tlsConfig := &tls.Config{InsecureSkipVerify: true}
	log.WithFields(log.Fields{
		"scheme": address.Scheme,
		"host":   address.Host,
	}).Infoln("Dialing server.")
	//conn, err := tls.Dial(address.Scheme, address.Host, tlsConfig)
	conn, err := net.Dial(address.Scheme, address.Host)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": address.Scheme,
			"host":     address.Host,
			"error":    err.Error(),
		}).Errorln("Couldn't connect to host.")
		return nil, err
	}
	log.WithField("remote", conn.RemoteAddr()).Infoln("Connection established.")
	return conn, nil
}

func PerformLogin(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("client.PerformLogin")
	bIn := bufio.NewReader(in)
	for {
		str, err := bIn.ReadString('\n')
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't read from server.")
			return err
		}
		pkg, err := connection.Parse(strings.TrimSpace(str))
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't parse request.")
			return err
		}
		if pkg.Done() {
			return pkg.Error()
		}
		err = pkg.Ask(os.Stdin, out)
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't perform request.")
			return err
		}
	}
}
