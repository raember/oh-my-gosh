package server

import (
	"bytes"
	"testing"
	"time"
)

func TestServer_PerformLogin(t *testing.T) {
	srvr := NewServer(Config(""))
	stdin := bytes.NewBufferString("test\nsecret\n")
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	timeout := make(chan bool, 1)
	loginChan := make(chan LoginResult)
	go func() {
		time.Sleep(time.Second)
		timeout <- true
	}()
	srvr.PerformLogin(loginChan, stdin, stdout, stderr)
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
	stdin := bytes.NewBufferString("test\nsecret\n")
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")

	_ = NewServer(Config("")).Serve(stdin, stdout, stderr)
}
