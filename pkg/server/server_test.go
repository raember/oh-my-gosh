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
	user, err := srvr.PerformLogin(stdin, stdout)
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

	_ = NewServer(LoadConfig("")).Serve(stdin, stdout)
}
