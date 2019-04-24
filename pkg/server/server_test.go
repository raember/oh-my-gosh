package server

import (
	"bytes"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"testing"
)

func TestServer_PerformLogin(t *testing.T) {
	srvr := NewServer(LoadConfig(""))
	stdin := bytes.NewBufferString("")
	stdout := bytes.NewBufferString("")
	stdin.WriteString(connection.UsernamePacket{"test"}.String())
	stdin.WriteString(connection.PasswordPacket{"secret"}.String())
	user, err := srvr.PerformLogin("", stdin, stdout)
	if err != nil {
		t.Error("Couldn't lookup user.")
		t.Fail()
		return
	}
	if user.Name != "test" {
		t.Error("Wrong username returned: " + user.Name)
	}
}

func TestServer_Serve(t *testing.T) {
	stdin := bytes.NewBufferString("")
	stdout := bytes.NewBufferString("")
	stdin.WriteString(connection.UsernamePacket{"test"}.String())
	stdin.WriteString(connection.PasswordPacket{"secret"}.String())

	pty, ptsName, uid, err := NewServer(LoadConfig("")).Serve(stdin, stdout)
	if err != nil {
		t.Error("Couldn't serve.")
		t.Fail()
		return
	}
	if uid == 0 {
		t.Error("Uid is 0.")
		t.Fail()
		return
	}
	if ptsName == "" {
		t.Error("Pts is empty.")
		t.Fail()
		return
	}
	if pty == nil {
		t.Error("Pty is nil.")
		t.Fail()
		return
	}
}
