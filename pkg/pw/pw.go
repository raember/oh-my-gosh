package pw

// https://linux.die.net/man/3/getpwnam

/*
#include <sys/types.h>
#include <pwd.h>

struct passwd *getpwnam(const char *name);

//struct passwd *getpwuid(uid_t uid);

//int getpwnam_r(const char *name, struct passwd *pwd,
//            char *buf, size_t buflen, struct passwd **result);

//int getpwuid_r(uid_t uid, struct passwd *pwd,
//            char *buf, size_t buflen, struct passwd **result);
*/
import "C"
import (
	"errors"
	log "github.com/sirupsen/logrus"
	_ "github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
)

type PassWd struct {
	Name     string
	Password string
	Uid      uint32
	Gid      uint32
	Gecos    string
	HomeDir  string
	Shell    string
}

// The GetPwByName() function returns a pointer to a structure containing the broken-out fields of the record in the
// password database (e.g., the local password file /etc/passwd, NIS, and LDAP) that matches the username name.
func GetPwByName(username string) (*PassWd, error) {
	log.WithField("username", username).Traceln("--> pw.GetPwByName")
	return convertToPasswd(C.getpwnam(C.CString(username)))
}

// The GetPwByUid() function returns a pointer to a structure containing the broken-out fields of the record in the
// password database that matches the user ID uid.
func GetPwByUid(uid uint32) (*PassWd, error) {
	log.WithField("uid", uid).Traceln("--> pw.GetPwByUid")
	return convertToPasswd(C.getpwuid(C.uint(uid)))
}

func convertToPasswd(cpasswd *C.struct_passwd) (*PassWd, error) {
	log.WithField("cpasswd", cpasswd).Traceln("--> pw.convertToPasswd")
	if cpasswd == nil {
		err := errors.New("got null pointer instead of *C.struct_passwd")
		log.WithError(err).Errorln("Lookup failed.")
		return nil, err
	}
	passWd := &PassWd{
		Name:     C.GoString(cpasswd.pw_name),
		Password: C.GoString(cpasswd.pw_passwd),
		Uid:      uint32(cpasswd.pw_uid),
		Gid:      uint32(cpasswd.pw_gid),
		Gecos:    C.GoString(cpasswd.pw_gecos),
		HomeDir:  C.GoString(cpasswd.pw_dir),
		Shell:    C.GoString(cpasswd.pw_shell),
	}
	log.WithFields(log.Fields{
		"name":     passWd.Name,
		"password": passWd.Password,
		"uid":      passWd.Uid,
		"gid":      passWd.Gid,
		"gecos":    passWd.Gecos,
		"homeDir":  passWd.HomeDir,
		"shell":    passWd.Shell,
	}).Debugln("Looked up user.")
	return passWd, nil
}
