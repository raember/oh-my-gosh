package client

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"strconv"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(common.TCP, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a client")
	}
	if client.address.Host != common.LOCALHOST+":"+strconv.Itoa(common.PORT) {
		t.Error("Malformed address returned")
	}
	_, err = NewClient(common.TCP4, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a client with TCP4 scheme")
	}
	_, err = NewClient(common.TCP6, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a client with TCP6 scheme")
	}
	_, err = NewClient(common.UNIX, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a client with UNIX scheme")
	}
	_, err = NewClient(common.UNIXPACKET, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a client with UNIXPACKET scheme")
	}
	_, err = NewClient(common.TCP, "127.0.0.1", common.PORT)
	if err != nil {
		t.Error("Failed creating a client with IP")
	}
	_, err = NewClient("", common.LOCALHOST, common.PORT)
	if err == nil {
		t.Error("Created client with empty network")
	}
	_, err = NewClient("http", common.LOCALHOST, common.PORT)
	if err == nil {
		t.Error("Created client with http network")
	}
	_, err = NewClient(common.TCP, "", common.PORT)
	if err == nil {
		t.Error("Created client with empty address")
	}
	_, err = NewClient(common.TCP, "not a domain", common.PORT)
	if err == nil {
		t.Error("Created client with faulty address")
	}
	_, err = NewClient(common.TCP, common.LOCALHOST, -1)
	if err == nil {
		t.Error("Created client with negative port")
	}
	_, err = NewClient(common.TCP, common.LOCALHOST, 0)
	if err != nil {
		t.Error("Failed creating a client with arbitrary port")
	}
}
