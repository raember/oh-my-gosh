package client

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"strconv"
	"testing"
)

func TestNewDialer(t *testing.T) {
	dialer, err := NewDialer(common.TCP, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a dialer")
	}
	if dialer.address.Host != common.LOCALHOST+":"+strconv.Itoa(common.PORT) {
		t.Error("Malformed address returned")
	}
	_, err = NewDialer(common.TCP4, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a dialer with TCP4 scheme")
	}
	_, err = NewDialer(common.TCP6, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a dialer with TCP6 scheme")
	}
	_, err = NewDialer(common.UNIX, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a dialer with UNIX scheme")
	}
	_, err = NewDialer(common.UNIXPACKET, common.LOCALHOST, common.PORT)
	if err != nil {
		t.Error("Failed creating a dialer with UNIXPACKET scheme")
	}
	_, err = NewDialer(common.TCP, "127.0.0.1", common.PORT)
	if err != nil {
		t.Error("Failed creating a dialer with IP")
	}
	_, err = NewDialer("", common.LOCALHOST, common.PORT)
	if err == nil {
		t.Error("Created dialer with empty network")
	}
	_, err = NewDialer("http", common.LOCALHOST, common.PORT)
	if err == nil {
		t.Error("Created dialer with http network")
	}
	_, err = NewDialer(common.TCP, "", common.PORT)
	if err == nil {
		t.Error("Created dialer with empty address")
	}
	_, err = NewDialer(common.TCP, "not a domain", common.PORT)
	if err == nil {
		t.Error("Created dialer with faulty address")
	}
	_, err = NewDialer(common.TCP, common.LOCALHOST, -1)
	if err == nil {
		t.Error("Created dialer with negative port")
	}
	_, err = NewDialer(common.TCP, common.LOCALHOST, 0)
	if err != nil {
		t.Error("Failed creating a dialer with arbitrary port")
	}
}
