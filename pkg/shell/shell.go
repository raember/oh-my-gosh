package shell

import (
	log "github.com/sirupsen/logrus"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/pw"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func Execute(passWd *pw.PassWd, in io.Reader, out io.Writer) error {
	log.WithFields(log.Fields{
		"passWd": passWd,
		"in":     &in,
		"out":    &out,
	}).Traceln("--> shell.Execute")
	shell := exec.Command(passWd.Shell, "--login")
	shell.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
		Credential: &syscall.Credential{
			Uid: passWd.Uid,
			Gid: passWd.Gid,
		},
	}
	shell.Dir = passWd.HomeDir
	envs, err := getEnv(passWd, in, out)
	if err != nil {
		return err
	}
	shell.Env = envs
	shell.Stdin = in
	shell.Stdout = out
	shell.Stderr = out
	err = os.MkdirAll(passWd.HomeDir, os.ModeDir)
	if err != nil {
		log.WithError(err).WithField("path", passWd.HomeDir).Errorln("Couldn't create directory.")
	}
	// TODO: Tidy up fragments in shell.
	err = shell.Run()
	if err != nil {
		log.WithError(err).Errorln("An error occured.")
	} else {
		log.Debugln("Shell terminated.")
	}
	return err
}

func getEnv(passWd *pw.PassWd, in io.Reader, out io.Writer) ([]string, error) {
	log.WithFields(log.Fields{
		"passWd": passWd,
		"in":     &in,
		"out":    &out,
	}).Traceln("--> shell.getEnv")
	//bufIn := bufio.NewReader(in)
	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Errorln("Couldn't lookup hostname")
		return nil, err
	}
	envs := []string{
		"USER=" + passWd.Name,
		"UID=" + strconv.Itoa(int(passWd.Uid)),
		"GID=" + strconv.Itoa(int(passWd.Gid)),
		"HOME=" + passWd.HomeDir,
		"SHELL=" + passWd.Shell,
		"HOSTNAME=" + hostname,
	}
	//for _, env := range []string{"TERM"} {
	//	log.WithField("env", env).Debugln("Requesting environment variable.")
	//	_, err := fmt.Fprint(out, connection.EnvPacket{Request: env}.String())
	//	if err != nil {
	//		return nil, err
	//	}
	//	value, err := bufIn.ReadString('\n')
	//	if err != nil {
	//		log.WithError(err).Errorln("Couldn't read environment variable value.")
	//		return nil, err
	//	}
	//	value = strings.TrimSpace(value)
	//	log.WithField("value", value).Debugln("Read environment variable value.")
	//	envs = append(envs, fmt.Sprintf("%s=%s", env, value))
	//}
	//log.Debugln("Sending DonePacket.")
	//_, err = fmt.Fprint(out, connection.DonePacket{}.String())
	//if err != nil {
	//	log.WithError(err).Errorln("Couldn't send DonePacket.")
	//	return nil, err
	//}
	log.WithField("envs", envs).Debugln("Done gathering environment variables.")
	return envs, nil
}
