package common

import (
	log "github.com/sirupsen/logrus"
	"net"
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
	// LOG_LEVEL not set, let's default to debug
	if !ok {
		lvl = "trace"
	}
	// parse string, this is built-in feature of logrus
	ll, err := log.ParseLevel(strings.ToLower(lvl))
	if err != nil {
		ll = log.DebugLevel
	}
	// set global log level
	log.SetLevel(ll)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func AddrToStr(addr net.Addr) string {
	return addr.Network() + "://" + addr.String()
}
