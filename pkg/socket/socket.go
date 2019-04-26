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

int go_shutdown(int socket) {
	// https://linux.die.net/man/3/shutdown
	// int shutdown(int socket, int how);
	return shutdown(socket, SHUT_RDWR);
}
*/
import "C"
import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net"
)

func Listen(port int) uintptr {
	log.WithField("port", port).Traceln("--> socket.Listen")
	listenerFd, err := C.go_listen(C.int(port))
	if err != nil {
		log.WithError(err).Fatalln("Failed to setup listener.")
	}
	log.WithField("listenerFd", listenerFd).Debugln("Setup Listener.")
	return uintptr(listenerFd)
}

func Accept(listenerFd uintptr) (uintptr, error) {
	log.WithField("listenerFd", listenerFd).Traceln("--> socket.Accept")
	socketFd, err := C.go_accept(C.int(listenerFd))
	if err != nil {
		log.WithError(err).Errorln("Failed to accept connection.")
		return 0, err
	}
	log.WithField("socketFd", socketFd).Debugln("Accepted connection.")
	return uintptr(socketFd), nil
}

func GetPeerName(socketFd uintptr) *net.TCPAddr {
	log.WithField("socketFd", socketFd).Traceln("--> socket.GetPeerName")
	ERROR_MSG := "Failed to get peer address."
	addr, err := C.go_getpeername(C.int(socketFd))
	if err != nil {
		log.WithError(err).Fatalln(ERROR_MSG)
	}
	addrStr := C.GoString(C.inet_ntoa(addr.sin_addr))
	port := int(addr.sin_port)
	tcpAddr, err := net.ResolveTCPAddr(common.TCP, fmt.Sprintf("%s:%d", addrStr, port))
	if err != nil {
		log.WithError(err).Fatalln(ERROR_MSG)
	}
	log.WithField("address", tcpAddr).Infoln("Got peer address.")
	return tcpAddr
}

func Shutdown(socketFd uintptr) error {
	log.WithField("socketFd", socketFd).Traceln("--> socket.Shutdown")
	_, err := C.go_shutdown(C.int(socketFd))
	if err != nil {
		log.WithError(err).Errorln("Failed to shutdown socket.")
	}
	return err
}
