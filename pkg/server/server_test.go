package server

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"strconv"
	"testing"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer(common.TCP, common.PORT)
	if err != nil {
		t.Error("Failed creating a server")
	}
	if server.address.Host != common.LOCALHOST+":"+strconv.Itoa(common.PORT) {
		t.Error("Malformed address returned")
	}
	_, err = NewServer(common.TCP4, common.PORT)
	if err != nil {
		t.Error("Failed creating a server with IP")
	}
	_, err = NewServer("", common.PORT)
	if err == nil {
		t.Error("Created server with empty network")
	}
	_, err = NewServer("http", common.PORT)
	if err == nil {
		t.Error("Created server with http network")
	}
	_, err = NewServer(common.TCP, -1)
	if err == nil {
		t.Error("Created server with negative port")
	}
	_, err = NewServer(common.TCP, 0)
	if err != nil {
		t.Error("Failed creating a server with arbitrary port")
	}
}

func TestServer_Listen(t *testing.T) {
	server, _ := NewServer(common.TCP, common.PORT)
	_, err := server.Listen()
	if err != nil {
		t.Error("Failed listening")
	}
}
