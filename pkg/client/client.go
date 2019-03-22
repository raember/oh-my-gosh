package client

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

const (
	TCP = "tcp"
	UDP = "udp"
)

type Client struct {
	address *url.URL
	conn    net.Conn
}

func CreateClient(network string, address string, port int64) (Client, error) {
	addrStr := network + "://" + address + ":" + strconv.FormatInt(port, 10)
	reqUrl, err := url.ParseRequestURI(addrStr)
	return Client{reqUrl, nil}, err
}

func (client Client) Connect() {
	conn, _ := net.Dial(client.address.Scheme, client.address.Host)
	client.conn = conn
}

func main() {

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}
}
