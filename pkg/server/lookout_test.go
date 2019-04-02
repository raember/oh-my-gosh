package server

import (
	"fmt"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"runtime"
	"strconv"
	"testing"
)

func TestNewLookout(t *testing.T) {
	lookout, err := NewLookout(common.TCP, common.PORT)
	if err != nil {
		t.Error("Failed creating a lookout")
	}
	if lookout.address.Host != common.LOCALHOST+":"+strconv.Itoa(common.PORT) {
		t.Error("Malformed address returned")
	}
	_, err = NewLookout(common.TCP4, common.PORT)
	if err != nil {
		t.Error("Failed creating a lookout with IP")
	}
	_, err = NewLookout("", common.PORT)
	if err == nil {
		t.Error("Created lookout with empty network")
	}
	_, err = NewLookout("http", common.PORT)
	if err == nil {
		t.Error("Created lookout with http network")
	}
	_, err = NewLookout(common.TCP, -1)
	if err == nil {
		t.Error("Created lookout with negative port")
	}
	_, err = NewLookout(common.TCP, 0)
	if err != nil {
		t.Error("Failed creating a lookout with arbitrary port")
	}
}

func TestLookout_Listen(t *testing.T) {
	_, path, _, _ := runtime.Caller(0)
	fmt.Println(path)
	lookout, _ := NewLookout(common.TCP, common.PORT)
	listener, err := lookout.Listen("../../test/certificate.pem", "../../test/key.pem")
	defer listener.Close()
	if err != nil {
		t.Error("Failed listening")
	}
}
