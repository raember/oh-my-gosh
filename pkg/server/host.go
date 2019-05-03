package server

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/connection"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/passwd"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pty"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/utils"
	"golang.org/x/sys/unix"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

type Host struct {
	config      *viper.Viper
	certificate *tls.Certificate
	conn        net.Conn
	rAddr       *net.TCPAddr
	userEnvs    []string
	userName    string
	userPw      string
	rTerm       string
	rUser       string
	rHostname   string
	ptm         *os.File
	pts         *os.File
}

func NewHost(config *viper.Viper) Host {
	log.WithField("config", config).Traceln("--> host.NewHost")
	return Host{config: config}
}

func (host *Host) LoadCertKeyPair(certPath string, keyFilePath string) error {
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Traceln("--> host.loadCertKeyPair")
	cert, err := tls.LoadX509KeyPair(certPath, keyFilePath)
	if err != nil {
		log.WithError(err).Fatalln("Failed to load certificate key pair.")
	}
	log.WithFields(log.Fields{
		"certPath":    certPath,
		"keyFilePath": keyFilePath,
	}).Debugln("Loaded certificate key pair.")
	host.certificate = &cert
	return nil
}

func (host *Host) Connect(socketFd uintptr, rAddr *net.TCPAddr) error {
	log.WithFields(log.Fields{
		"socketFd": socketFd,
		"rAddr":    rAddr.String(),
	}).Traceln("--> host.Host.Connect")
	ErrorMsg := "Failed to handle connection."

	conn, err := utils.ConnFromFd(socketFd, host.certificate)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	log.WithFields(log.Fields{
		"conn":  conn,
		"rAddr": rAddr.String(),
	}).Infoln("Connected to client.")
	host.conn = conn
	host.rAddr = rAddr
	return nil
}

func (host *Host) Setup() error {
	log.Traceln("--> host.Host.Setup")
	ErrorMsg := "Failed to setup host."
	// Get the necessary information from the rAddr client
	err := host.getClientEnvs("TERM")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	host.rUser, err = host.requestClientEnv("USER")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("rUser", host.rUser).Debugln("Got remote user name.")
	}
	host.rHostname, err = host.requestClientEnv("HOSTNAME")
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		if host.rHostname == "" {
			log.Warnln("Client HOSTNAME was empty. Trying NAME next.")
			host.rHostname, err = host.requestClientEnv("NAME")
			if err != nil {
				log.WithError(err).Errorln(ErrorMsg)
				return err
			} else {
				if host.rHostname == "" {
					log.Warnln("Client NAME was empty. Using IP address instead.")
					host.rHostname = strings.Split(host.rAddr.String(), ":")[0]
				}
				log.WithField("rHostname", host.rHostname).Debugln("Got remote host name.")
			}
		}
		log.WithField("rHostname", host.rHostname).Debugln("Got remote host name.")
	}
	host.userName, err = host.requestClientEnv(common.ENV_GOSH_USER)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("userName", host.userName).Debugln("Got local user name.")
	}
	host.userPw, err = host.requestClientEnv(common.ENV_GOSH_PASSWORD)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	} else {
		log.WithField("userPw", host.userPw).Debugln("Got local user password.")
	}

	// Done gathering all the information.
	log.Infoln("Got all the information from the client.")

	host.ptm, host.pts, err = pty.Create()
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	log.Infoln("Set up host.")
	return nil
}

func (host *Host) Serve() error {
	log.Traceln("--> host.Host.Serve")
	cmd, err := host.StartShell()
	if err != nil {
		log.WithError(err).Errorln("Failed to serve client.")
	}
	defer func() {
		utils.CloseFile(host.pts)
		utils.CloseFile(host.ptm)
		utils.CloseConn(host.conn)
	}()

	// TODO: Handle forwarding yourself.
	go utils.Forward(host.ptm, host.conn, "ptm", "client")
	go utils.Forward(host.conn, host.ptm, "client", "ptm")

	//rFdSet := unix.FdSet{}
	//n, err := unix.Pselect(3, &rFdSet, &rFdSet, &rFdSet, nil, nil)

	wpid, err := syscall.Wait4(cmd.Process.Pid, nil, 0, nil)
	if err != nil {
		log.WithError(err).Errorln("Failed waiting for login.")
		return err
	} else {
		log.WithField("wpid", wpid).Debugln("Waited for login.")
	}
	return nil
}

func (host *Host) StartShell() (cmd *exec.Cmd, err error) {
	log.Traceln("--> host.Host.StartShell")
	ErrorMsg := "Failed to start shell."
	if host.userName != "" {
		err = host.authenticateWithKeys(host.userName)
		if err != nil {
			log.WithError(err).Infoln("Failed to log in user with keys. Proceed to login command.")
			err = host.stopTransfer(true)
			if err == nil {
				cmd, err = host.login()
			}
		} else {
			pwd, err := passwd.GetPwByName(host.userName)
			if err != nil {
				log.WithError(err).Errorln("Failed to log in with keys.")
			} else {
				err = host.stopTransfer(true)
				if err == nil {
					//TODO: Make entry in utmx
					//TODO: Display MOTD
					cmd, err = host.spawnShell(pwd)
				}
			}
		}
	} else {
		err = host.stopTransfer(true)
		if err == nil {
			cmd, err = host.login()
		}
	}
	if err != nil {
		log.WithError(err).WithField("cmd", cmd).Errorln(ErrorMsg)
	}
	return
}

func (host *Host) getClientEnvs(envs ...string) error {
	log.WithField("envs", envs).Traceln("--> host.Host.getClientEnvs")
	host.userEnvs = []string{}
	for _, env := range envs {
		value, err := host.requestClientEnv(env)
		if err != nil {
			return err
		}
		host.userEnvs = append(host.userEnvs, fmt.Sprintf("%s=%s", env, value))
	}
	log.WithField("host.userEnvs", host.userEnvs).Debugln("Done gathering environment variables.")
	return nil
}

func (host Host) requestClientEnv(env string) (string, error) {
	log.WithField("env", env).Traceln("--> host.Host.requestClientEnv")
	log.WithField("env", env).Debugln("Requesting environment variable from client.")
	bufIn := bufio.NewReader(host.conn)
	_, err := fmt.Fprint(host.conn, connection.EnvPacket{Request: env}.String())
	if err != nil {
		log.WithError(err).Errorln("Failed to send packet.")
		return "", err
	}
	value, err := bufIn.ReadString('\n')
	if err != nil {
		log.WithError(err).Errorln("Failed to read environment variable value.")
		return "", err
	}
	value = strings.TrimSpace(value)
	log.WithField("value", value).Debugln("Read environment variable value.")
	return value, nil
}

func (host Host) stopTransfer(success bool) error {
	log.Traceln("--> host.Host.stopTransfer")
	_, err := fmt.Fprint(host.conn, connection.DonePacket{Success: success}.String())
	if err != nil {
		log.WithError(err).Errorln("Failed to send DonePacket.")
		return err
	}
	log.Debugln("Sent done packet.")
	return nil
}

func (host Host) authenticateWithKeys(user string) error {
	log.WithField("user", user).Traceln("--> server.Host.authenticateWithKeys")
	ErrorMsg := "Failed login with key."
	filename := url.PathEscape(host.rUser) + ".pub"
	keyStorePath := path.Join(host.config.GetString("Authentication.KeyStore"), common.AUTHKEYSDIR, user, filename)
	if _, err := os.Stat(keyStorePath); os.IsNotExist(err) {
		log.WithError(err).Warnln("Failed to locate public key file.")
		return err
	}
	pubKey, err := utils.PubKeyFromFile(keyStorePath)
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	secret, nSecret, err := utils.CreateSecret()
	if err != nil {
		log.WithError(err).Errorln(ErrorMsg)
		return err
	}
	encryptedSecret, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, secret[:nSecret])
	if err != nil {
		log.WithError(err).Errorln("Failed to encrypt secret.")
		return err
	} else {
		log.WithField("encryptedSecret", string(encryptedSecret)).Infoln("Encrypted secret. Sending to client.")
	}

	_, err = host.conn.Write([]byte(connection.RsaPacket{EncryptedSecret: encryptedSecret, EncryptedSecretN: len(encryptedSecret)}.String()))
	if err != nil {
		log.WithError(err).Errorln("Failed to send RSA packet.")
		return err
	}
	_, err = host.conn.Write(encryptedSecret)
	if err != nil {
		log.WithError(err).Errorln("Failed to send encrypted secret.")
		return err
	}

	answer := make([]byte, common.SECRET_LENGTH)
	nAnswer, err := host.conn.Read(answer)
	if err != nil {
		log.WithError(err).Errorln("Failed to receive decrypted answer.")
		return err
	}
	if nAnswer != nSecret || !bytes.Equal(secret[:nSecret], answer[:nAnswer]) {
		log.WithError(err).Errorln("The answer does not match the secret.")
		return err
	} else {
		log.Infoln("Client authenticated itself using keys.")

	}
	return nil
}

func (host Host) spawnShell(pwd *passwd.PassWd) (*exec.Cmd, error) {
	log.Traceln("--> server.Host.spawnShell")
	shell := exec.Command(pwd.Shell, "--login")
	shell.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
		//Setctty: true, //TODO: Fix error when using ctty
		Credential: &syscall.Credential{
			Uid: pwd.Uid,
			Gid: pwd.Gid,
		},
	}
	shell.Dir = pwd.HomeDir
	shell.Env = host.userEnvs
	shell.Stdin = host.pts
	shell.Stdout = host.pts
	shell.Stderr = host.pts
	err := shell.Start()
	if err != nil {
		log.WithError(err).Errorln("Failed to fork shell.")
	} else {
		log.WithField("shell", shell).Debugln("Forked shell.")
	}
	return shell, err
}

//TODO: Fix "operation not supported" error or drop this function
func dropPrivilege(pwd *passwd.PassWd) error {
	log.Traceln("--> server.dropPrivilege")
	if err := unix.Setuid(int(pwd.Uid)); err != nil {
		log.WithError(err).Errorln("Failed to set UID.")
		return err
	} else {
		log.WithField("uid", pwd.Uid).Infoln("Dropped user privilege.")
	}
	if err := unix.Setgid(int(pwd.Gid)); err != nil {
		log.WithError(err).Errorln("Failed to set UID.")
		return err
	} else {
		log.WithField("gid", pwd.Gid).Infoln("Dropped group privilege.")
	}
	return nil
}

func (host Host) login() (*exec.Cmd, error) {
	log.Traceln("--> server.Host.login")
	login := exec.Command("/bin/login", "-h", host.rHostname)
	login.Stdin = host.pts
	login.Stdout = host.pts
	login.Stderr = host.pts
	login.Env = host.userEnvs
	login.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	err := login.Start()
	if err != nil {
		log.WithError(err).Errorln("Failed to fork login.")
	} else {
		log.WithField("login", login).Debugln("Forked login.")
	}
	if host.userName != "" {
		if err := host.answerPtyLoginRequest(login.Process.Pid); err != nil {
			return login, err
		}
		if host.userPw != "" {
			if err := host.answerPtyPasswordRequest(login.Process.Pid); err != nil {
				return login, err
			}
		}
	}
	return login, err
}

func interrupt(pid int) error {
	err := unix.Kill(pid, unix.SIGINT)
	if err != nil {
		log.WithError(err).Errorln("Failed to interrupt process.")
		err = unix.Kill(pid, unix.SIGKILL)
		if err != nil {
			log.WithError(err).Errorln("Failed to kill process.")
		} else {
			log.WithField("pid", pid).Infoln("Killed process.")
		}
	} else {
		log.WithField("pid", pid).Infoln("Interrupted process.")
	}
	return err
}

func (host Host) answerPtyLoginRequest(pid int) error {
	log.WithField("pid", pid).Traceln("--> server.Host.answerPtyLoginRequest")
	str, err := bufio.NewReader(host.ptm).ReadString(':')
	if err != nil || !strings.HasSuffix(str, "login:") {
		log.WithError(err).WithField("str", str).Errorln("Failed to read 'login:' from pty")
		_ = interrupt(pid)
		return err
	}
	_, err = host.ptm.WriteString(host.userName + "\n")
	if err != nil {
		log.WithError(err).Errorln("Failed to send user name to pty.")
		_ = interrupt(pid)
		return err
	}
	log.Infoln("Sent user name to process")
	return nil
}

func (host Host) answerPtyPasswordRequest(pid int) error {
	log.WithField("pid", pid).Traceln("--> server.Host.answerPtyPasswordRequest")
	str, err := bufio.NewReader(host.ptm).ReadString(':')
	if err != nil || !strings.HasSuffix(str, "Password:") {
		log.WithError(err).WithField("str", str).Errorln("Failed to read 'Password:' from pty")
		_ = interrupt(pid)
		return err
	}
	_, err = host.ptm.WriteString(host.userPw + "\n")
	if err != nil {
		log.WithError(err).Errorln("Failed to send password to pty.")
		_ = interrupt(pid)
		return err
	}
	log.Infoln("Sent password to process")
	return nil
}
