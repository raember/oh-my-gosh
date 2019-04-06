package client

import (
	"net"
	"os"
)

type Client struct {
	conn net.Conn
}

func (client Client) Communicate(stdIn *os.File, stdOut *os.File, stdErr *os.File) error {
	println("test")
	os.Stdin = stdIn
	os.Stdout = stdOut
	os.Stderr = stdErr
	return nil
}
