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
	}).Traceln("--> pty.ptyError")
	return &PtyError{name, err.Error(), err.(syscall.Errno)}
}

func (e *PtyError) Error() string {
	log.Traceln("--> pty.PtyError.Error")
	return fmt.Sprintf("%s: %s", e.FuncName, e.ErrorString)
}

// Creates a master pty and the name of the linked slave pts.
func Create() (uintptr, string) {
	log.Traceln("--> pty.Create")
	ptyFdC, err := C.posix_openpt(C.O_RDWR)
	if err != nil {
		err = ptyError("posix_openpt", err)
		log.WithError(err).Fatalln("Couldn't open pseudo-terminal.")
	}
	if _, err := C.grantpt(ptyFdC); err != nil {
		err = ptyError("grantpt", err)
		C.close(ptyFdC)
		log.WithError(err).Fatalln("Couldn't grant pseudo-terminal perms.")
	}
	if _, err := C.unlockpt(ptyFdC); err != nil {
		err = ptyError("unlockpt", err)
		C.close(ptyFdC)
		log.WithError(err).Fatalln("Couldn't unlock pseudo-terminal.")
	}
	ptyFd := uintptr(ptyFdC)
	ptsName := C.GoString(C.ptsname(ptyFdC))
	log.WithFields(log.Fields{
		"ptsName": ptsName,
		"ptyFd":   ptyFd,
	}).Infoln("Opened pseudo terminal.")
	return ptyFd, ptsName
}
