package client

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	config *viper.Viper
	rUri   *url.URL
}

func NewClient(config *viper.Viper) *Client {
	log.WithField("config", config).Traceln("--> client.NewClient")
	return &Client{config: config}
}

func (client Client) Dial() (net.Conn, error) {
	log.Traceln("--> client.Client.Dial")
	address := client.rUri
	log.WithFields(log.Fields{
		"scheme": address.Scheme,
		"host":   address.Host,
	}).Infoln("Dialing server.")
	//TODO: Don't skip unsecure certificates. Ask the user instead.
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial(address.Scheme, address.Host, tlsConfig)
	//conn, err := net.Dial(rUri.Scheme, rUri.Host)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": address.Scheme,
			"host":     address.Host,
			"error":    err.Error(),
		}).Errorln("Failed to connect to host.")
		return nil, err
	}
	log.WithField("remote", conn.RemoteAddr()).Infoln("Connection established.")
	return conn, nil
}

func (client Client) PerformTransfer(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> client.PerformTransfer")
	bIn := bufio.NewReader(in)
	for {
		str, err := bIn.ReadString('\n')
		if err != nil {
			log.WithError(err).Errorln("Failed to read from server.")
			return err
		} else {
			log.WithField("str", str).Debugln("Read a line from the server.")
		}
		packet, err := connection.Parse(strings.TrimSpace(str))
		if err != nil {
			log.WithError(err).Errorln("Failed to parse request.")
			return err
		}
		if packet.Done() {
			return nil
		}
		switch pckt := packet.(type) {
		case connection.RsaPacket:
			log.Debugln("Detected RSA packet.")
			pckt.KeyPath = client.config.GetString("Authentication.KeyStore")
			err = pckt.Ask(in, out)
		default:
			err = pckt.Ask(os.Stdin, out)
		}
		if err != nil {
			log.WithError(err).Errorln("Failed to perform request.")
			return err
		}
	}
}

func (client Client) checkAddress(address string, port int) (*url.URL, error) {
	log.WithFields(log.Fields{
		"rUri": address,
		"port": port,
	}).Traceln("--> client.NewClient")
	if port < 0 {
		err := errors.New("connection cannot be negative")
		log.WithFields(log.Fields{
			"connection": port,
		}).Errorln(err.Error())
		return nil, err
	}
	if address == "" {
		err := errors.New("rUri cannot be empty")
		log.WithFields(log.Fields{
			"connection": port,
		}).Errorln(err.Error())
		return nil, err
	}
	addrStr := address + ":" + strconv.Itoa(port)
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
	return reqUrl, nil
}

func (client *Client) ParseArgument(arg string) error {
	log.WithField("arg", arg).Traceln("--> client.Client.ParseArgument")
	ErrorMsg := "Failed to parse argument"
	if arg == "" {
		arg = fmt.Sprintf("%s:%d", common.LOCALHOST, client.config.GetInt("Client.Port"))
	}
	rUri, err := url.ParseRequestURI("tcp://" + arg)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	if err = client.checkUrl(rUri); err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}

	client.rUri = rUri
	log.WithField("rUri", rUri).Debugln("Parsed URI.")
	return nil
}

func (client *Client) checkUrl(rUri *url.URL) error {
	log.WithField("rUri", rUri).Traceln("--> client.Client.checkUrl")
	ErrorMsg := "Failed to check url"
	if rUri.Fragment != "" {
		err := errors.New("fragment cannot be set")
		log.WithError(err).WithField("Fragment", rUri.Fragment).Errorln(ErrorMsg)
		return err
	}
	if rUri.Opaque != "" {
		err := errors.New("opaque cannot be set")
		log.WithError(err).WithField("Opaque", rUri.Opaque).Errorln(ErrorMsg)
		return err
	}
	if rUri.RawPath != "" {
		err := errors.New("raw path cannot be set")
		log.WithError(err).WithField("RawPath", rUri.RawPath).Errorln(ErrorMsg)
		return err
	}
	if rUri.RawQuery != "" {
		err := errors.New("raw query cannot be set")
		log.WithError(err).WithField("RawQuery", rUri.RawQuery).Errorln(ErrorMsg)
		return err
	}
	if rUri.Port() == "" {
		newUri, _ := url.ParseRequestURI(fmt.Sprintf("%s:%d", rUri.String(), client.config.GetInt("Client.Port")))
		rUri.Host = newUri.Host
		return client.checkUrl(rUri)
	}
	return nil
}

func (client Client) Setup() error {
	log.Traceln("--> client.Client.Setup")
	if err := os.Setenv(common.ENV_GOSH_USER, client.rUri.User.Username()); err != nil {
		log.WithError(err).Errorln(fmt.Sprintf("Failed to set %s environment variable.", common.ENV_GOSH_USER))
		return err
	}
	pw, pwSet := client.rUri.User.Password()
	if pwSet {
		if err := os.Setenv(common.ENV_GOSH_PASSWORD, pw); err != nil {
			log.WithError(err).Errorln(fmt.Sprintf("Failed to set %s environment variable.", common.ENV_GOSH_USER))
			return err
		}
	}
	//TODO: Load the public key and set it as environment variable.
	return nil
}
