package main

import (
	"../../pkg/common"
	"../../pkg/server"
)

func main() {
	server.TextReplyLocal(common.TCP, common.PORT)
}
