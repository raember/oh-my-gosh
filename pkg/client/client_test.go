package client

import (
	"fmt"
	"github.engineering.zhaw.ch/neut/oh-my-gosh/pkg/common"
	"net/url"
	"os"
	"strconv"
	"testing"
)

var config = LoadConfig("")

func TestNewClient(t *testing.T) {
	clnt := NewClient(config)
	if clnt == nil {
		t.Error("Client was nil.")
		t.FailNow()
	}
	if clnt.config == nil {
		t.Error("Client config was nil.")
		t.FailNow()
	}
}

func TestClient_ParseArgument_Full(t *testing.T) {
	clnt := NewClient(config)
	if err := clnt.ParseArgument("test:secret@localhost:22"); err != nil {
		t.Fail()
	}
	checkUrl(t, clnt.rUri, "test", "secret", "localhost", 22)
}

func TestClient_ParseArgument_No_Password(t *testing.T) {
	clnt := NewClient(config)
	if err := clnt.ParseArgument("test@localhost:22"); err != nil {
		t.Fail()
	}
	checkUrl(t, clnt.rUri, "test", "", "localhost", 22)
}

func TestClient_ParseArgument_No_User(t *testing.T) {
	clnt := NewClient(config)
	if err := clnt.ParseArgument("localhost:22"); err != nil {
		t.Fail()
	}
	checkUrl(t, clnt.rUri, "", "", "localhost", 22)
}

func TestClient_ParseArgument_No_Port(t *testing.T) {
	clnt := NewClient(config)
	if err := clnt.ParseArgument("localhost"); err != nil {
		t.Fail()
	}
	checkUrl(t, clnt.rUri, "", "", "localhost", 2222)
}

func TestClient_ParseArgument_No_Nothing(t *testing.T) {
	clnt := NewClient(config)
	if err := clnt.ParseArgument(""); err != nil {
		t.Fail()
	}
	checkUrl(t, clnt.rUri, "", "", "localhost", 2222)
}

func checkUrl(t *testing.T, url *url.URL, user string, password string, host string, port uint) {
	if user != "" {
		uUser := url.User.Username()
		if uUser != user {
			t.Error(fmt.Sprintf("User mismatch: %s != %s", uUser, user))
			t.Fail()
		}
		if password != "" {
			uPassword, uPwSet := url.User.Password()
			if !uPwSet || uPassword != password {
				t.Error(fmt.Sprintf("Password mismatch: %s != %s", uPassword, password))
				t.Fail()
			}
		}
	}
	if host != "" {
		uHost := url.Hostname()
		if uHost != host {
			t.Error(fmt.Sprintf("Hostname mismatch: %s != %s", uHost, host))
			t.Fail()
		}
	}
	if port != 0 {
		uPort, _ := strconv.Atoi(url.Port())
		if uPort != int(port) {
			t.Error(fmt.Sprintf("Port mismatch: %d != %d", uPort, port))
			t.Fail()
		}
	}
}

func TestClient_Setup(t *testing.T) {
	clnt := NewClient(config)
	if err := clnt.ParseArgument("test:secret@localhost:22"); err != nil {
		t.FailNow()
	}
	_ = os.Unsetenv(common.ENV_GOSH_USER)
	_ = os.Unsetenv(common.ENV_GOSH_PASSWORD)
	if err := clnt.Setup(); err != nil {
		t.Fail()
	} else {
		if os.Getenv(common.ENV_GOSH_USER) != "test" {
			t.Error("User env not set")
			t.Fail()
		}
		if os.Getenv(common.ENV_GOSH_PASSWORD) != "secret" {
			t.Error("Password env not set")
			t.Fail()
		}
	}
	_ = os.Unsetenv(common.ENV_GOSH_USER)
	_ = os.Unsetenv(common.ENV_GOSH_PASSWORD)
}
