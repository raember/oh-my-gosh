package client

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net"
	"net/url"
	"os"
	"strconv"
	"syscall"
)

type Client struct {
	address *url.URL
	conn    net.Conn
	config  *viper.Viper
}

func NewClient(protocol string, address string, port int) (*Client, error) {
	if port < 0 {
		err := errors.New("port cannot be negative")
		log.WithFields(log.Fields{
			"port": port,
		}).Errorln(err.Error())
		return nil, err
	}
	if address == "" {
		err := errors.New("address cannot be empty")
		log.WithFields(log.Fields{
			"port": port,
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

func (client Client) Connect() (net.Conn, error) {
	conn, err := net.Dial(client.address.Scheme, client.address.Host)
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": client.address.Scheme,
			"host":     client.address.Host,
		}).Errorln(err.Error())
		return nil, err
	}
	client.conn = conn
	return conn, nil
}

var config = viper.New()

func init() {
	config.SetConfigName(common.APPNAME + "_config")
	config.AddConfigPath("/etc/" + common.APPNAME + "/")
	config.SetConfigType(common.CONFIGFORMAT)
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.Warnln("Config file changed:", e.Name)
	})
	err := config.ReadInConfig()
	if err != nil {
		log.Fatalf("Couldn't read config file: %s\n", err)
	}
}

func TextExchangeLocal(protocol string, address string, port int) {
	client, err := NewClient(protocol, address, port)
	if err != nil {
		syscall.Exit(1)
	}
	log.Infoln("Initialized client.")

	conn, err := client.Connect()
	if err != nil {
		syscall.Exit(1)
	}
	log.Infoln("Established connection to " + client.address.String())

	for {
		// listen for reply
		answer, _ := bufio.NewReader(conn).ReadString('\n')
		log.WithFields(log.Fields{
			"rawtext": answer,
		}).Debugln("Inbound")
		log.Println(answer)

		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		message, err := reader.ReadString('\n')
		log.WithFields(log.Fields{
			"rawtext": message,
		}).Debugln("Outbound")
		// send to socket
		_, err = fmt.Fprintln(conn, message)
		if err != nil {
			log.Infoln("Connection died.")
			break
		}
	}
}
