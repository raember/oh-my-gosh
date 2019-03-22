package gosh

import (
	_ "github.com/kubernetes/pkg/ssh"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh.git/pkg/pty"
	_ "github.engineering.zhaw.ch/oh-my-gosh/oh-my-gosh/pkg/pty"
	_ "k8s.io/kubernetes/pkg/kubectl/cmd"
)
