package main

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"net"
	"os"
)

func main() {
	config := server.Config()
	lookout, err := server.NewLookout(
		config.GetString("Server.Protocol"),
		config.GetInt("Server.Port"),
	)
	if err != nil {
		os.Exit(1)
	}
	listener, err := lookout.Listen()
	if err != nil {
		os.Exit(1)
	}
	defer listener.Close()
	err = server.WaitForConnections(listener, func(conn net.Conn) {
		srvr := server.NewServer(conn)
		srvr.Serve()
	})
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
