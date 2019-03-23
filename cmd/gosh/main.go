package main

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/client"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
)

func main() {
	client.TextExchangeLocal(common.TCP, common.LOCALHOST, common.PORT)
}
