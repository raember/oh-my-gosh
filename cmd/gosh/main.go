package main

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/client"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"os"
)

func main() {
	config := client.Config()
	dialer, err := client.NewDialer(
		config.GetString("Client.Protocol"),
		common.LOCALHOST,
		config.GetInt("Client.Port"),
	)
	if err != nil {
		os.Exit(1)
	}
	conn, err := dialer.Dial()
	if err != nil {
		os.Exit(1)
	}
	defer conn.Close()
	clnt := client.NewClient(conn)
	err = clnt.Communicate()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
