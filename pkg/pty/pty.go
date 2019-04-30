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
	log "github.com/sirupsen/logrus"
	"os"
)

// Creates a pty pair
func Create() (ptm *os.File, pts *os.File, err error) {
	log.Traceln("--> pty.Create")
	ptyFdC, err := C.posix_openpt(C.O_RDWR)
	if err != nil {
		log.WithError(err).Errorln("Failed to open pseudo-terminal.")
		return
	}
	if _, err = C.grantpt(ptyFdC); err != nil {
		C.close(ptyFdC)
		log.WithError(err).Errorln("Failed to grant pseudo-terminal perms.")
		return
	}
	if _, err = C.unlockpt(ptyFdC); err != nil {
		C.close(ptyFdC)
		log.WithError(err).Errorln("Failed to unlock pseudo-terminal.")
		return
	}
	ptyFd := uintptr(ptyFdC)
	ptsName := C.GoString(C.ptsname(ptyFdC))
	log.WithFields(log.Fields{
		"ptsName": ptsName,
		"ptyFd":   ptyFd,
	}).Debugln("Got pty information.")
	ptm = os.NewFile(ptyFd, "/dev/pts/ptmx")
	pts, err = os.OpenFile(ptsName, os.O_RDWR, 0755)
	if err != nil {
		log.WithError(err).WithField("ptsName", ptsName).Errorln("Failed to open pts.")
		if closeErr := ptm.Close(); closeErr != nil {
			log.WithError(closeErr).Errorln("Failed to close ptm.")
		}
	}
	log.WithFields(log.Fields{
		"ptm":      ptm,
		"ptm.Fd()": ptm.Fd(),
		"pts":      pts,
		"pts.Fd()": pts.Fd(),
	}).Infoln("Opened pseudo terminal.")
	return
}
