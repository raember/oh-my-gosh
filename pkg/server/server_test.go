package server

import (
	"bytes"
	"testing"
)

func TestServer_PerformLogin(t *testing.T) {
	srvr := NewServer(Config(""))
	stdin := bytes.NewBufferString("test\nsecret\n")
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	_, username, err := srvr.PerformLogin(stdin, stdout, stderr)
	if err != nil {
		t.Error("Couldn't login: " + err.Error())
	}
	if username != "test" {
		t.Error("Wrong username returned.")
	}
}

func TestServer_Serve(t *testing.T) {
	stdin := bytes.NewBufferString("test\nsecret\n")
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")

	_ = NewServer(Config("")).Serve(stdin, stdout, stderr)
}
