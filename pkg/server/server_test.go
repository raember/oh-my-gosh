package server

import (
	"bytes"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"testing"
	"time"
)

func TestServer_PerformLogin(t *testing.T) {
	srvr := NewServer(LoadConfig(""))
	stdin := bytes.NewBufferString("")
	stdout := bytes.NewBufferString("")
	timeout := make(chan bool, 1)
	loginChan := make(chan LoginResult)
	go func() {
		time.Sleep(time.Second)
		timeout <- true
	}()
	stdin.WriteString(connection.UsernamePacket{"test"}.String())
	stdin.WriteString(connection.PasswordPacket{"secret"}.String())
	srvr.PerformLogin(loginChan, stdin, stdout)
	select {
	case loginResult := <-loginChan:
		if loginResult.user.Name != "test" {
			t.Error("Wrong username returned: " + loginResult.user.String())
		}
	case <-timeout:
		t.Fail()
	}
}

func TestServer_Serve(t *testing.T) {
	stdin := bytes.NewBufferString("")
	stdout := bytes.NewBufferString("")
	stdin.WriteString(connection.UsernamePacket{"test"}.String())
	stdin.WriteString(connection.PasswordPacket{"secret"}.String())

	_ = NewServer(LoadConfig("")).Serve(stdin, stdout)
}
