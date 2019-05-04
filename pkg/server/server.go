package server

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
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
		Port: server.config.GetInt("Serve.Port"),
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
	maxSessions := server.config.GetInt("Serve.MaxSessions")
	err = unix.Listen(fd, maxSessions)
	if err != nil {
		log.WithError(err).Fatalln("Failed to listen on socket.")
	} else {
		log.WithFields(log.Fields{
			"fd":          fd,
			"maxSessions": maxSessions,
		}).Infoln("Listening on socket.")
	}
	for {
		socketFd, peer, err := unix.Accept(fd) // Can't use peer Sockaddr because Go...
		if err != nil {
			log.WithError(err).Fatalln("Failed opening connection.")
		} else {
			peerInet4 := peer.(*unix.SockaddrInet4)
			rAddr := net.TCPAddr{
				IP: net.IPv4(
					peerInet4.Addr[0],
					peerInet4.Addr[1],
					peerInet4.Addr[2],
					peerInet4.Addr[3],
				),
				Port: peerInet4.Port,
			}
			log.WithFields(log.Fields{
				"socketFd": socketFd,
				"rAddr":    rAddr.String(),
			}).Infoln("Accepted connection from peer.")
			fdChan <- RemoteHandle{
				Fd:         uintptr(socketFd),
				RemoteAddr: &rAddr,
			}
		}
	}
}

type RemoteHandle struct {
	Fd         uintptr
	RemoteAddr *net.TCPAddr
}
