package pw

import "testing"

func TestGetPwByName(t *testing.T) {
	passwd, err := GetPwByName("test")
	if err != nil {
		t.Error("Failed to look up test user: " + err.Error())
		t.Fail()
		return
	}
	if passwd.Name != "test" {
		t.Error("Name of looked up test user doesn't match.")
	}
	if passwd.Shell != "/bin/bash" {
		t.Error("Shell of looked up test user doesn't match.")
	}
	passwd, err = GetPwByName("")
	if err == nil {
		t.Error("No error after empty look up.")
	}
}

func TestGetPwByUid(t *testing.T) {
	pwd, err := GetPwByName("test")
	if err != nil {
		t.Error("Lookup failed: Cannot test function.")
		return
	}
	uid := pwd.Uid
	passwd, err := GetPwByUid(uid)
	if err != nil {
		t.Error("Failed to look up test user: " + err.Error())
		t.Fail()
		return
	}
	if passwd.Name != "test" {
		t.Error("Name of looked up test user doesn't match.")
	}
	if passwd.Shell != "/bin/bash" {
		t.Error("Shell of looked up test user doesn't match.")
	}
}
