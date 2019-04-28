package shell

import (
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	pwd := &pw.PassWd{
		Name:  "test",
		Shell: "/usr/sbin/nologin",
	}
	err := Execute(pwd, os.Stdin, os.Stdout)
	if err == nil {
		t.Error("Could run nologin.")
	}

	pwd = &pw.PassWd{
		Name:  "test",
		Shell: "/bin/bash",
	}
	err = Execute(pwd, os.Stdin, os.Stdout)
	if err != nil {
		t.Error("Failed to run bash.")
	}
}
