package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"io/ioutil"
	"os"
)

func PubKeyFromFile(path string) (*rsa.PublicKey, error) {
	log.WithField("path", path).Traceln("--> utils.PubKeyFromFile")
	block, err := BlockFromFile(path)
	if err != nil || block.Type != "PUBLIC KEY" {
		log.WithError(err).Errorln("Failed to load public key block.")
		return nil, err
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.WithError(err).Errorln("Failed to parse public key.")
		return nil, err
	}
	return pubInterface.(*rsa.PublicKey), nil
}

func PrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	log.WithField("path", path).Traceln("--> utils.PrivateKeyFromFile")
	block, err := BlockFromFile(path)
	if err != nil || block.Type != "PRIVATE KEY" {
		log.WithError(err).Errorln("Failed to load private key block.")
		return nil, err
	}
	pubInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.WithError(err).Errorln("Failed to parse private key.")
		return nil, err
	}
	return pubInterface.(*rsa.PrivateKey), nil
}

func BlockFromFile(path string) (*pem.Block, error) {
	log.WithField("path", path).Traceln("--> utils.BlockFromFile")
	pubKeyFile, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.WithError(err).Errorln("Failed to read key file.")
		return nil, err
	}
	pubKeyBytes, err := ioutil.ReadAll(pubKeyFile)
	if err != nil {
		log.WithError(err).Errorln("Failed to load key.")
		return nil, err
	}
	if err := pubKeyFile.Close(); err != nil {
		log.WithError(err).Errorln("Failed to close key file.")
		return nil, err
	}
	block, _ := pem.Decode(pubKeyBytes)
	log.WithField("block.Type", block.Type).Debugln("Decoded block.")
	return block, nil
}

func CreateSecret() ([]byte, int, error) {
	log.Traceln("--> utils.PubKeyFromFile")
	secret := make([]byte, common.SECRET_LENGTH)
	n, err := rand.Read(secret)
	if err != nil {
		log.WithError(err).Errorln("Failed to create secret.")
		return nil, 0, err
	} else {
		log.WithField("n", n).Debugln("Created secret.")
	}
	return secret, n, nil
}
