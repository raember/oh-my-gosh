package common

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	TCP          = "tcp"
	TCP4         = "tcp4"
	TCP6         = "tcp6"
	UNIX         = "unix"
	UNIXPACKET   = "unixpacket"
	LOCALHOST    = "localhost"
	PORT         = 2222
	CLIENTNAME   = "gosh"
	SERVERNAME   = "goshd"
	CONFIGFORMAT = "toml"
	CERTFILE     = "/etc/gosh/certificate.pem"
	KEYFILE      = "/etc/gosh/key.pem"
	CONFIGPATH   = "/etc/gosh"
)

func init() {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		lvl = "info"
	}
	ll, err := log.ParseLevel(strings.ToLower(lvl))
	if err != nil {
		ll = log.DebugLevel
	}
	log.SetLevel(ll)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}
