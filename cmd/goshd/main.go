package main

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
)

func main() {
	server.TextReplyLocal(common.TCP, common.PORT)
}
