package login

import (
	"bytes"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	stdin := bytes.NewBufferString("test\nsecret\n")
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	_, username, err := Authenticate(stdin, stdout, stderr)
	if err != nil {
		t.Error("Couldn't login: " + err.Error())
	}
	if username != "test" {
		t.Error("Wrong username returned.")
	}
}
