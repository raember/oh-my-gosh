package socket

/*
#include <unistd.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <stdio.h>
#include <arpa/inet.h>

int go_listen(int port) {
	int sockfd;
	sockfd = socket(AF_INET, SOCK_STREAM, 0);

	struct sockaddr_in servaddr;
	memset(&servaddr, 0, sizeof(servaddr));
	servaddr.sin_family = AF_INET;
	servaddr.sin_addr.s_addr = htons(INADDR_ANY);
	servaddr.sin_port = htons(port);

	// https://linux.die.net/man/3/bind
	// int bind(int socket, const struct sockaddr *address, socklen_t address_len);
	int retval;
	retval = bind(sockfd, (struct sockaddr *) &servaddr, sizeof(servaddr));
	if (retval < 0) { return retval; }
	retval = listen(sockfd, 10);
	if (retval < 0) { return retval; }
	return sockfd;
}

int go_accept(int socket) {
	// https://linux.die.net/man/3/accept
	// int accept(int socket, struct sockaddr *restrict address, socklen_t *restrict address_len);
	return accept(socket, (struct sockaddr*) NULL, NULL);
}
struct sockaddr_in go_getpeername(int socket) {
	struct sockaddr address;
    socklen_t addressLength;
	addressLength = sizeof(address);

	// https://linux.die.net/man/3/getpeername
	// int getpeername(int socket, struct sockaddr *restrict address, socklen_t *restrict address_len);
	getpeername(socket, &address, &addressLength);
	return *(struct sockaddr_in*)&address;
}
*/
import "C"
import (
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net"
	"strconv"
)

func Listen(port int) (uintptr, error) {
	log.WithField("port", port).Traceln("socket.Listen")
	retval, err := C.go_listen(C.int(port))
	if err != nil {
		log.WithField("error", err).Errorln("Failed to setup listener.")
		return 0, err
	}
	return uintptr(retval), nil
}

func Accept(socketFd uintptr) (uintptr, error) {
	log.WithField("socketFd", socketFd).Traceln("socket.Accept")
	retval, err := C.go_accept(C.int(socketFd))
	if err != nil {
		log.WithField("error", err).Errorln("Failed to accept connection.")
		return 0, err
	}
	return uintptr(retval), nil
}

func GetPeerName(socketFd uintptr) (*net.TCPAddr, error) {
	log.WithField("socketFd", socketFd).Traceln("socket.GetPeerName")
	addr, err := C.go_getpeername(C.int(socketFd))
	if err != nil {
		log.WithField("error", err).Errorln("Failed to get peer address.")
		return nil, err
	}
	addrStr := C.GoString(C.inet_ntoa(addr.sin_addr))
	port := int(addr.sin_port)
	tcpAddr, err := net.ResolveTCPAddr(common.TCP, addrStr+":"+strconv.Itoa(port))
	if err != nil {
		log.WithField("error", err).Errorln("Failed to parse address.")
		return nil, err
	}
	log.WithField("address", tcpAddr).Infoln("Got peer address.")
	return tcpAddr, nil
}
