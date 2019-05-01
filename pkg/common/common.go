package common

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	TCP               = "tcp"
	TCP4              = "tcp4"
	TCP6              = "tcp6"
	UNIX              = "unix"
	UNIXPACKET        = "unixpacket"
	LOCALHOST         = "localhost"
	PORT              = 2222
	CLIENTNAME        = "gosh"
	SERVERNAME        = "goshd"
	CONFIGFORMAT      = "toml"
	CERTFILE          = "/etc/gosh/certificate.pem"
	KEYFILE           = "/etc/gosh/key.pem"
	CONFIGPATH        = "/etc/gosh"
	DEFAULT_LOG_LEVEL = log.InfoLevel
)

//TODO: Use global loggers

func init() {
	log.SetLevel(getLogLevel())
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func getLogLevel() log.Level {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		lvl = DEFAULT_LOG_LEVEL.String()
		_ = os.Setenv("LOG_LEVEL", lvl)
	}
	ll, err := log.ParseLevel(strings.ToLower(lvl))
	if err != nil {
		ll = DEFAULT_LOG_LEVEL
	}
	return ll
}
