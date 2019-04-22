// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux,!android netbsd openbsd
// +build cgo

// Package pty is a simple pseudo-terminal package for Unix systems,
// implemented by calling C functions via cgo.
// This is only used for testing the os/signal package.

package pty

/*
#define _XOPEN_SOURCE 600
#include <fcntl.h>
#include <stdlib.h>
#include <unistd.h>
*/
import "C"

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
)

type PtyError struct {
	FuncName    string
	ErrorString string
	Errno       syscall.Errno
}

func ptyError(name string, err error) *PtyError {
	log.WithFields(log.Fields{
		"name": name,
		"err":  err,
	}).Traceln("pty.ptyError")
	return &PtyError{name, err.Error(), err.(syscall.Errno)}
}

func (e *PtyError) Error() string {
	log.Traceln("pty.PtyError.Error")
	return fmt.Sprintf("%s: %s", e.FuncName, e.ErrorString)
}

// Open returns a master pty and the name of the linked slave tty.
func Open() (master *os.File, slave string, err error) {
	log.Traceln("pty.Open")
	m, err := C.posix_openpt(C.O_RDWR)
	if err != nil {
		err = ptyError("posix_openpt", err)
		log.WithField("error", err).Errorln("Couldn't open pseudo-terminal.")
		return nil, "", err
	}
	if _, err := C.grantpt(m); err != nil {
		err = ptyError("grantpt", err)
		log.WithField("error", err).Errorln("Couldn't grant pseudo-terminal perms.")
		C.close(m)
		return nil, "", err
	}
	if _, err := C.unlockpt(m); err != nil {
		err = ptyError("unlockpt", err)
		log.WithField("error", err).Errorln("Couldn't unlock pseudo-terminal.")
		C.close(m)
		return nil, "", err
	}
	slave = C.GoString(C.ptsname(m))
	file := os.NewFile(uintptr(m), "pty-master")
	log.WithFields(log.Fields{
		"slave": slave,
		"file":  file.Name(),
	}).Infoln("Opened pseudo terminal.")
	return file, slave, nil
}
