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
		break
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
		break
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
	case AuthSucceededPacket:
		break
	default:
		t.Fail()
	}
}
