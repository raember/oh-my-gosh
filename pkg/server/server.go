package server

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/socket"
	"golang.org/x/sys/unix"
	"net"
)

type Server struct {
	config *viper.Viper
}

func NewServer(config *viper.Viper) Server {
	log.WithField("config", config).Traceln("--> server.NewServer")
	return Server{config: config}
}

func (server Server) AwaitConnections(fdChan chan RemoteHandle) {
	log.WithField("fdChan", fdChan).Traceln("--> server.AwaitConnections")

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		log.WithError(err).Fatalln("Failed to create socket.")
	} else {
		log.WithField("fd", fd).Debugln("Created socket.")
	}

	sockAddr := &unix.SockaddrInet4{
		Addr: [4]byte{},
		Port: server.config.GetInt("Server.Port"),
	}
	err = unix.Bind(fd, sockAddr)
	if err != nil {
		log.WithError(err).Fatalln("Failed to bind socket.")
	} else {
		log.WithFields(log.Fields{
			"fd":       fd,
			"sockAddr": sockAddr,
		}).Debugln("Bound socket.")
	}
	maxSessions := server.config.GetInt("Server.MaxSessions")
	err = unix.Listen(fd, maxSessions)
	if err != nil {
		log.WithError(err).Fatalln("Failed to listen on socket.")
	} else {
		log.WithFields(log.Fields{
			"fd":          fd,
			"maxSessions": maxSessions,
		}).Debugln("Listening on socket.")
	}

	for {
		socketFd, _, err := unix.Accept(fd)
		if err != nil {
			log.WithError(err).Fatalln("Failed opening connection.")
		} else {
			rAddr := socket.GetPeerName(uintptr(socketFd))
			log.WithFields(log.Fields{
				"socketFd": socketFd,
				"rAddr":    rAddr,
			}).Debugln("Accepted connection from peer.")
			fdChan <- RemoteHandle{
				Fd:         uintptr(socketFd),
				RemoteAddr: rAddr,
			}
		}
	}
}

type RemoteHandle struct {
	Fd         uintptr
	RemoteAddr *net.TCPAddr
}
