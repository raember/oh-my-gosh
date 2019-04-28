package login

import (
	"bytes"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	stdin := bytes.NewBufferString("test\nsecret\n")
	stdout := bytes.NewBufferString("")
	user, err := Authenticate("", stdin, stdout)
	if err != nil {
		t.Error("Failed to login: " + err.Error())
		return
	}
	if user.Name != "test" {
		t.Error("Wrong username returned.")
	}
}
