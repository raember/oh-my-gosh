package main

import (
	"../../pkg/client"
	"../../pkg/common"
)

func main() {
	client.TextExchangeLocal(common.TCP, common.LOCALHOST, common.PORT)
}
