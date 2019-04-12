package socket

/*
#include <unistd.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <stdio.h>

int go_listen(int port) {
	int sockfd;
	sockfd = socket(AF_INET, SOCK_STREAM, 0);

	struct sockaddr_in servaddr;
	memset(&servaddr, 0, sizeof(servaddr));
	servaddr.sin_family = AF_INET;
	servaddr.sin_addr.s_addr = htons(INADDR_ANY);
	servaddr.sin_port = htons(port);

	int retval;
	// int bind(int socket, const struct sockaddr *address, socklen_t address_len);
	retval = bind(sockfd, (struct sockaddr *) &servaddr, sizeof(servaddr));
	if (retval < 0) { return retval; }
	retval = listen(sockfd, 10);
	if (retval < 0) { return retval; }
	return sockfd;
}

int go_accept(int sockfd) {
	// int accept(int socket, struct sockaddr *restrict address, socklen_t *restrict address_len);
	return accept(sockfd, (struct sockaddr*) NULL, NULL);
}
*/
import "C"
import (
	log "github.com/sirupsen/logrus"
)

func Listen(port int) (uintptr, error) {
	retval, err := C.go_listen(C.int(port))
	if err != nil {
		log.WithField("error", err).Errorln("Failed to setup listener.")
		return 0, err
	}
	return uintptr(retval), nil
}

func Accept(socketFd uintptr) (uintptr, error) {
	retval, err := C.go_accept(C.int(socketFd))
	if err != nil {
		log.WithField("error", err).Errorln("Failed to accept connection.")
		return 0, err
	}
	return uintptr(retval), nil
}
