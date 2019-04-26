package connection

import (
	"testing"
)

func TestParseUsernamePacket(t *testing.T) {
	pkg, err := Parse("?U:login: ")
	if err != nil {
		t.Error(err)
	}
	switch pkg.(type) {
	case UsernamePacket:
		if pkg.Field() != "login: " {
			t.Fail()
		}
	default:
		t.Fail()
	}
}

func TestParsePasswordPacket(t *testing.T) {
	pkg, err := Parse("?P:password: ")
	if err != nil {
		t.Error(err)
	}
	switch pkg.(type) {
	case PasswordPacket:
		if pkg.Field() != "password: " {
			t.Fail()
		}
	default:
		t.Fail()
	}
}

func TestParseAuthSucceededPacket(t *testing.T) {
	pkg, err := Parse("?S:")
	if err != nil {
		t.Error(err)
	}
	switch pkg.(type) {
	case DonePacket:
		if pkg.Field() != "" {
			t.Fail()
		}
	default:
		t.Fail()
	}
}

func TestParseTimeoutPacket(t *testing.T) {
	pkg, err := Parse("?T:")
	if err != nil {
		t.Error(err)
	}
	switch pkg.(type) {
	case TimeoutPacket:
		if pkg.Field() != "" {
			t.Fail()
		}
	default:
		t.Fail()
	}
}
