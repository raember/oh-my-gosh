package shell

import (
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	err := Execute("/usr/sbin/nologin", os.Stdin, os.Stdout)
	if err == nil {
		t.Error("Could run nologin.")
	}

	err = Execute("/bin/bash", os.Stdin, os.Stdout)
	if err != nil {
		t.Error("Couldn't run bash.")
	}
}
