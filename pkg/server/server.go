package server

import (
	"bufio"
	"errors"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/willdonnelly/passwd"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/speakeasier"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// The Server structure is used to handle a connection.
type Server struct {
	config *viper.Viper
	conn   net.Conn
}

func NewServer(config *viper.Viper) Server {
	return Server{
		config: config,
		conn:   nil,
	}
}

// Serve a newly established connection.
func (server Server) Serve(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) {
	client := bufio.NewReader(stdIn)

	transaction, username, err := server.PerformLogin(stdIn, stdOut, stdErr)
	if err != nil {
		return
	}

	message := "Authentication was successful"
	log.WithFields(log.Fields{
		"message": message,
	}).Infoln("Outbound")
	_, _ = server.conn.Write([]byte(message + "\n"))

	err = transaction.SetCred(pam.Silent)
	if err != nil {
		log.Errorln("Couldn't set credentials for the user.")
	}
	err = transaction.AcctMgmt(pam.Silent)
	if err != nil {
		log.Errorln("Couldn't validate the user.")
	}
	err = transaction.OpenSession(pam.Silent)
	if err != nil {
		log.WithField("error", err.Error()).Errorln("Couldn't open a session.")
	}
	defer transaction.CloseSession(pam.Silent)
	str, err := transaction.GetItem(pam.Service)
	if err == nil {
		log.Infoln("Service: " + str)
	}
	str, err = transaction.GetItem(pam.User)
	if err == nil {
		log.Infoln("User: " + str)
	}
	str, err = transaction.GetItem(pam.Tty)
	if err == nil {
		log.Infoln("Tty: " + str)
	}
	str, err = transaction.GetItem(pam.Rhost)
	if err == nil {
		log.Infoln("Rhost: " + str)
	}
	str, err = transaction.GetItem(pam.Authtok)
	if err == nil {
		log.Infoln("Authtok: " + str)
	}
	str, err = transaction.GetItem(pam.Oldauthtok)
	if err == nil {
		log.Infoln("Oldauthtok: " + str)
	}
	str, err = transaction.GetItem(pam.Ruser)
	if err == nil {
		log.Infoln("Ruser: " + str)
	}
	str, err = transaction.GetItem(pam.UserPrompt)
	if err == nil {
		log.Infoln("UserPrompt: " + str)
	}
	strs, err := transaction.GetEnvList()
	if err == nil {
		log.Println("#envs: " + strconv.Itoa(len(strs)))
		for _, str := range strs {
			log.Infoln(str + ": " + strs[str])
		}
	}
	file, name, err := pty.Open()
	log.WithField("slavename", name).Infoln("Pts")
	log.WithField("filename", file.Name()).Infoln("File")

	users, err := passwd.Parse()
	if err != nil {
		log.WithField("error", err).Errorln("Couldn't lookup users.")
		return
	}
	myUser := users[username]
	USER := username
	UID := myUser.Uid
	GID := myUser.Gid
	HOME := myUser.Home
	SHELL := myUser.Shell
	log.WithFields(log.Fields{
		"USER":  USER,
		"UID":   UID,
		"GID":   GID,
		"HOME":  HOME,
		"SHELL": SHELL,
	}).Println("Looked up user.")

	if _, err := os.Stat("/etc/motd"); err == nil {
		motd, err := ioutil.ReadFile("/etc/motd")
		if err != nil {
			log.WithField("error", err).Errorln("Couldn't read message of the day.")
			return
		}
		log.Println(motd)
	}

	uidInt, _ := strconv.ParseUint(UID, 10, 32)
	gidInt, _ := strconv.ParseUint(GID, 10, 32)
	creds := syscall.Credential{
		Uid:    uint32(uidInt),
		Gid:    uint32(gidInt),
		Groups: []uint32{},
	}

	sysattr := syscall.SysProcAttr{
		Credential: &creds,
	}

	attr := syscall.ProcAttr{
		Dir:   HOME,
		Env:   []string{},
		Files: []uintptr{},
		Sys:   &sysattr,
	}
	pid, err := syscall.ForkExec("sup?", []string{}, &attr)
	log.WithField("pid", pid).Println("Forked.")

	err = server.checkForNologinFile(stdIn, stdOut, stdErr)
	if err != nil {
		return
	}

	//shell.Start(client, conn, conn)

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := client.ReadString('\n')
		if err != nil {
			log.Infoln("Connection died.")
			break
		}
		// output message received
		log.WithFields(log.Fields{
			"message": message,
		}).Infoln("Inbound")
		// sample process for string received
		answer := strings.ToUpper(message)
		// send new string back to client
		log.WithFields(log.Fields{
			"answer": answer,
		}).Infoln("Outbound")
		_, _ = server.conn.Write([]byte(answer + "\n"))
	}
}

// Performs login attempts until either the attempt succeeds or the limit of tries has been reached.
// Receives 2 lines from the client:
// user\n
// password\n
// Sends one byte as answer.
func (server Server) PerformLogin(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) (*pam.Transaction, string, error) {
	transaction, err := pam.StartFunc("goshd", "", func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			log.WithField("msg", message).Debugln("Will read password.")
			//str, err := bufio.NewReader(stdIn).ReadString('\n')
			str, err := speakeasy.FAsk(stdIn, stdOut, message)
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read password.")
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read password.")
			return str, nil
		case pam.PromptEchoOn:
			_, _ = stdOut.Write([]byte(message)) // "login:"
			log.WithField("msg", message).Debugln("Will read user name.")
			str, err := bufio.NewReader(stdIn).ReadString('\n')
			if err != nil {
				log.WithField("error", err).Errorln("Couldn't read user name.")
			}
			str = strings.TrimSpace(str)
			log.WithField("str", str).Debugln("Read user name.")
			return str, nil
		case pam.ErrorMsg:
			log.WithField("msg", message).Debugln("Will send error.")
			_, _ = stdErr.Write([]byte(message))
			return "", nil
		case pam.TextInfo:
			log.WithField("msg", message).Debugln("Will send text.")
			_, _ = stdOut.Write([]byte(message))
			return "", nil
		}
		return "", errors.New("unrecognized message style")
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't start authentication.")
		return nil, "", err
	}
	err = transaction.Authenticate(0)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorln("Couldn't authenticate.")
		return nil, "", err
	}
	log.Infoln("Authentication succeeded!")
	return transaction, "", nil
}

func (server Server) checkForNologinFile(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) error {
	bytes, err := ioutil.ReadFile("/etc/nologin")
	if err != nil {
		log.Debugln("/etc/nologin file not found. Login permitted.")
		return nil
	}
	err = errors.New("/etc/nologin file exists: no login allowed")
	log.WithField("error", err).Infoln("/etc/nologin file exists. Login not permitted.")
	_, _ = stdOut.Write(bytes)
	return err
}
