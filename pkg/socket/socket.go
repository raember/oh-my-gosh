package socket

/*
#include <sys/socket.h>
#include <arpa/inet.h>

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
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net"
)

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
