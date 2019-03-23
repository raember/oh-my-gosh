package client

import (
	"../common"
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"syscall"
)

type Client struct {
	address *url.URL
	conn    net.Conn
}

func NewClient(protocol string, address string, port int) (Client, error) {
	if port < 0 {
		return Client{}, errors.New("port cannot be negative")
	}
	if address == "" {
		return Client{}, errors.New("address cannot be empty")
	}
	addrStr := protocol + "://" + address + ":" + strconv.Itoa(port)
	reqUrl, err := url.ParseRequestURI(addrStr)
	if err != nil {
		return Client{}, err
	}
	if reqUrl.Scheme != common.TCP &&
		reqUrl.Scheme != common.TCP4 &&
		reqUrl.Scheme != common.TCP6 &&
		reqUrl.Scheme != common.UNIX &&
		reqUrl.Scheme != common.UNIXPACKET {
		return Client{}, errors.New("protocol has to be either tcp, tcp4, tcp6, unix or unixpacket")
	}
	return Client{address: reqUrl}, nil
}

func (client Client) Connect() (net.Conn, error) {
	conn, err := net.Dial(client.address.Scheme, client.address.Host)
	if err != nil {
		return nil, err
	}
	client.conn = conn
	return conn, nil
}

func TextExchangeLocal(protocol string, address string, port int) {
	client, err := NewClient(protocol, address, port)
	if err != nil {
		fmt.Println(err.Error())
		syscall.Exit(1)
	}
	fmt.Println("Initialized client.")

	conn, err := client.Connect()
	if err != nil {
		fmt.Println(err.Error())
		syscall.Exit(1)
	}
	fmt.Println("Established connection to " + client.address.String())

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		// send to socket
		_, err = fmt.Fprintln(conn, text)
		if err != nil {
			fmt.Println("Connection died.")
			break
		}
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("-> " + message)
	}
}
