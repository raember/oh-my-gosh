package server

import (
	"bufio"
	"io/ioutil"
	"testing"
)

func TestServer_PerformLogin(t *testing.T) {
	stdin, err := ioutil.TempFile("", "stdin")
	if err != nil {
		t.Error(err)
	}
	defer stdin.Close()
	//out := bufio.NewWriter(stdin)

	stdout, err := ioutil.TempFile("", "stdout")
	if err != nil {
		t.Error(err)
	}
	defer stdout.Close()
	out := bufio.NewReader(stdout)

	stderr, err := ioutil.TempFile("", "stderr")
	if err != nil {
		t.Error(err)
	}
	defer stderr.Close()
	//errp := bufio.NewReader(err_r)

	srvr := NewServer(Config(""))
	go func() {
		_, username, err := srvr.PerformLogin(stdin, stdout, stderr)
		if err != nil {
			t.Error(err)
		}
		if username != "test" {
			t.Error("Wrong username returned.")
		}
	}()
	println("INCOMING:::::")
	var bytes []byte
	n, err := stdout.Read(bytes)
	println(string(bytes[:n]))
	str, err := out.ReadString(':')
	if err != nil {
		t.Error("Couldn't read login request: " + err.Error())
	}
	if str != "Login:" {
		t.Error("Couldn't match login request.")
	}
	println(str)
	_, err = stdout.WriteString("test")
	if err != nil {
		t.Error("Couldn't send username: " + err.Error())
	}
	str, err = out.ReadString(' ')
	if err != nil {
		t.Error("Couldn't read password request: " + err.Error())
	}
	if str != "Password:" {
		t.Error("Couldn't ask for password request.")
	}
	println(str)
	_, err = stdin.WriteString("password")
	if err != nil {
		t.Error("Couldn't send password: " + err.Error())
	}
}

//func TestServer_Serve(t *testing.T) {
//	srv_in, srv_out, out, in := setupTest(t)
//
//	srvr := NewServer(conf)
//	srvr.Serve(srv_in, srv_out, srv_out)
//
//	str, err := in.ReadString(' ')
//	if err != nil {
//		t.Error(err)
//	}
//	if str != "Login:" {
//		t.Error("Didn't receive login request.")
//	}
//
//	_, err = out.WriteString("alan")
//	if err != nil {
//		t.Error(err)
//	}
//}
