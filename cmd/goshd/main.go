package main

import "C"
import (
	"crypto/tls"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/server"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/shell"
	"io"
	"os"
	"os/exec"
	"strconv"
)

var certFile = flag.String("cert", common.CERTFILE, "Certificate file.")
var keyFile = flag.String("key", common.KEYFILE, "Key file.")
var configPath = flag.String("conf", common.CONFIGPATH, "LoadConfig path.")
var fd = flag.Uint("fd", 0, "Handle connection using supplied fd. Conflicts with --std and --pty.")
var std = flag.Bool("std", false, "Handle connection using standard pipes. Conflicts with --fd and --pty.")
var file = flag.String("file", "", "Handle connection using pty. Conflicts with --fd and --std.")
var uid = flag.Uint("uid", 0, "Start user session using the provided UID. Needs either --fd, --std or --pty.")
var config *viper.Viper

func main() {
	log.WithField("args", os.Args).Traceln("--> goshd.main")
	flag.Parse()
	log.WithField("certFile", *certFile).Debugln("Set certificate file")
	log.WithField("keyFile", *keyFile).Debugln("Set key file")
	log.WithField("configPath", *configPath).Debugln("Set configuration file path")
	log.WithField("fd", *fd).Debugln("Set socket file descriptor")
	log.WithField("std", *std).Debugln("Set standard pipes")
	log.WithField("file", *file).Debugln("Set file name")
	log.WithField("uid", *file).Debugln("Set user id")

	config = server.LoadConfig(*configPath)

	log.Traceln("Determine control flow")
	if *fd != 0 || *std || *file != "" {

		var in io.Reader
		var out io.Writer

		log.Traceln("Determine input and output pipes")
		if *fd != 0 && !*std && *file == "" {
			log.Traceln("Read from socket fd.")
			conn, err := connection.FromFd(uintptr(*fd))
			if err != nil {
				log.WithError(err).Fatalln("Couldn't get connection from fd.")
			}
			defer func() {
				err = conn.Close()
				if err != nil {
					log.WithError(err).Errorln("Couldn't close connection.")
				}
			}()
			cert, err := server.LoadCertKeyPair(*certFile, *keyFile)
			if err != nil {
				log.WithError(err).Fatalln("Couldn't load TLS certificate.")
			}
			tlsConn := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{cert}})
			in = tlsConn
			out = tlsConn
		} else if *std && *fd == 0 && *file == "" {
			log.Traceln("Read from standard pipes.")
			in = os.Stdin
			out = os.Stdout
		} else if *file != "" && !*std && *fd == 0 {
			log.Traceln("Read from pty.")
			ptsFile, err := os.Create(*file)
			if err != nil {
				log.WithError(err).Fatalln("Couldn't open pts file.")
			}
			defer func() {
				if err = ptsFile.Close(); err != nil {
					log.WithError(err).Errorln("Couldn't close pts file.")
				} else {
					log.Debugln("Closed pts file.")
				}
			}()
			in = ptsFile
			out = ptsFile
		} else if *std && *fd != 0 {
			log.Fatalln("Cannot have --fd and --std set at the same time!")
		}

		// Continue control flow
		if *file != "" && *uid != 0 {
			hostUser(uint32(*uid), in, out)
		} else {
			serveConnection(in, out)
		}
	} else {
		listen(*certFile, *keyFile)
	}
}

func listen(certFilename string, keyFilename string) {
	log.WithFields(log.Fields{
		"certFilename": certFilename,
		"keyFilename":  keyFilename,
	}).Traceln("--> main.listen")
	lookout, err := server.NewLookout(
		config.GetString("Server.Protocol"),
		config.GetInt("Server.Port"),
	)
	if err != nil {
		log.WithError(err).Fatalln("Couldn't set up lookout.")
	}

	socketFd, err := lookout.Listen(certFilename, keyFilename)
	if err != nil {
		os.Exit(1)
	}
	err = server.WaitForConnections(socketFd, func(connFd uintptr) {
		cmd := recursiveExec("--fd", strconv.FormatUint(uint64(connFd), 10))
		if err := cmd.Start(); err != nil {
			log.WithError(err).Fatalln("Couldn't execute child")
		}
	})
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func recursiveExec(additionalArgs ...string) *exec.Cmd {
	log.WithField("additionalArgs", additionalArgs).Traceln("--> main.recursiveExec")
	bin := os.Args[0]
	args := []string{
		"--cert", *certFile,
		"--key", *keyFile,
		"--conf", *configPath,
	}
	args = append(args, additionalArgs...)
	cmd := exec.Command(bin, args...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("LOG_LEVEL=%s", log.GetLevel().String()))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func hostUser(uid uint32, in io.Reader, out io.Writer) {
	log.WithFields(log.Fields{
		"uid": uid,
		"in":  in,
		"out": out,
	}).Traceln("--> main.hostUser")
	pwd, err := pw.GetPwByUid(uid)
	if err != nil {
		log.WithError(err).Fatalln("Couldn't create new file.")
	}
	if err = shell.Execute(pwd, in, out); err != nil {
		log.WithError(err).Fatalln("Shell returned an error.")
	}
}

func serveConnection(in io.Reader, out io.Writer) {
	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Traceln("--> main.serveConnection")
	ptyFile, ptsName, uid, err := server.NewServer(config).Serve(in, out)
	if err != nil {
		log.WithError(err).Fatalln("Failed serving connection.")
	}
	defer func() {
		if err = ptyFile.Close(); err != nil {
			log.WithError(err).Errorln("Couldn't close pty file.")
		} else {
			log.Debugln("Closed pty file.")
		}
	}()
	cmd := recursiveExec("--uid", strconv.FormatUint(uint64(uid), 10), "--file", ptsName)
	if err := cmd.Run(); err != nil {
		log.WithError(err).Fatalln("Couldn't execute child")
	}
}

func init() {
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}
