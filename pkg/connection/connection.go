package connection

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/utils"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

func Parse(str string) (Packet, error) {
	log.WithField("str", str).Traceln("--> connection.Parse")
	if str == "?D1:" {
		return DonePacket{Success: true}, nil
	} else if str == "?D0:" {
		return DonePacket{Success: false}, nil
	} else if strings.HasPrefix(str, "?E:") {
		return EnvPacket{str[3:]}, nil
	} else if strings.HasPrefix(str, "?K") {
		colonIdx := strings.Index(str, ":")
		length, err := strconv.Atoi(str[2:(colonIdx)])
		if err != nil {
			log.WithError(err).WithField("str", str).Errorln("Failed to parse length from packet")
		}
		return RsaPacket{EncryptedSecretN: length}, nil
	}
	err := errors.New("unknown packet format")
	log.WithError(err).WithField("str", str).Errorln("Failed to parse packet.")
	return nil, err
}

type Packet interface {
	Ask(io.Reader, io.Writer) error
	String() string
	Done() bool
}

// =============== Environment Variable Packet ===============

type EnvPacket struct {
	Request string
}

func (req EnvPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> connection.EnvPacket.Ask")
	log.WithField("request", req.Request).Debugln("Reading environment variable.")

	envvar := os.Getenv(req.Request)
	log.WithField("envvar", envvar).Debugln("Read environment variable. Sending to server.")
	_, err := fmt.Fprintln(out, envvar+"\n")
	return err
}

func (req EnvPacket) String() string {
	log.Traceln("--> connection.EnvPacket.String")
	return fmt.Sprintf("?E:%s\n", req.Request)
}

func (req EnvPacket) Done() bool {
	log.WithField("done", false).Traceln("--> connection.EnvPacket.Done")
	return false
}

// =============== RSA Packet ===============

type RsaPacket struct {
	EncryptedSecret  []byte
	EncryptedSecretN int
	KeyPath          string
}

func (req RsaPacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> connection.RsaPacket.Ask")
	log.WithFields(log.Fields{
		"KeyPath":          req.KeyPath,
		"EncryptedSecretN": req.EncryptedSecretN,
	}).Debugln("Decrypting secret.")
	filename := url.PathEscape(os.Getenv("USER")) + ".pem"
	privateKey, err := utils.PrivateKeyFromFile(path.Join(req.KeyPath, filename))
	if err != nil {
		log.WithError(err).Errorln("Failed to decrypt secret.")
		return err
	}
	encryptedSecret := make([]byte, req.EncryptedSecretN) // or common.SECRET_LENGTH
	nEncryptedSecret, err := in.Read(encryptedSecret)
	if err != nil {
		log.WithError(err).Errorln("Failed to receive decrypted secret.")
		return err
	} else {
		log.WithFields(log.Fields{
			"nEncryptedSecret": nEncryptedSecret,
			"encryptedSecret":  string(encryptedSecret),
		}).Debugln("Read encrypted secret.")
	}
	secret, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedSecret[:nEncryptedSecret])
	if err != nil {
		log.WithError(err).Errorln("Failed to decrypt secret.")
		return err
	}
	_, err = out.Write(secret)
	if err != nil {
		log.WithError(err).Errorln("Failed to send decrypted secret.")
		return err
	}
	return nil
}

func (req RsaPacket) String() string {
	log.Traceln("--> connection.RsaPacket.String")
	return fmt.Sprintf("?K%d:\n", req.EncryptedSecretN)
}

func (req RsaPacket) Done() bool {
	log.WithField("done", false).Traceln("--> connection.EnvPacket.Done")
	return false
}

// =============== Done Packet ===============

type DonePacket struct {
	Success bool
}

func (req DonePacket) Ask(in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"in":  &in,
		"out": &out,
	}).Traceln("--> connection.DonePacket.Ask")
	err := errors.New("nothing to ask")
	log.WithError(err).Errorln("Not implemented!")
	return err
}

func (req DonePacket) String() string {
	log.Traceln("--> connection.DonePacket.String")
	if req.Success {
		return "?D1:\n"
	} else {
		return "?D0:\n"
	}
}

func (req DonePacket) Done() bool {
	log.WithField("done", true).Traceln("--> connection.DonePacket.Done")
	return true
}
