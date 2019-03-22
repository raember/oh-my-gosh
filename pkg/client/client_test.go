package client

import "testing"

const (
	EXAMPLE_NETWORK = TCP
	EXAMPLE_ADDRESS = "localhost"
	EXAMPLE_PORT    = 2222
)

func TestCreateClient(t *testing.T) {
	_, err := CreateClient(EXAMPLE_NETWORK, EXAMPLE_ADDRESS, EXAMPLE_PORT)
	if err == nil {
		t.Error("Failed creating a client with domain name")
	}
	_, err = CreateClient(EXAMPLE_NETWORK, "127.0.0.1", EXAMPLE_PORT)
	if err == nil {
		t.Error("Failed creating a client with IP")
	}
	_, err = CreateClient("", EXAMPLE_ADDRESS, EXAMPLE_PORT)
	if err != nil {
		t.Error("Created client with empty network")
	}
	_, err = CreateClient("http", EXAMPLE_ADDRESS, EXAMPLE_PORT)
	if err != nil {
		t.Error("Created client with http network")
	}
	_, err = CreateClient(UDP, "127.0.0.1", EXAMPLE_PORT)
	if err == nil {
		t.Error("Failed creating a client with UDP scheme")
	}
	_, err = CreateClient(EXAMPLE_NETWORK, "", EXAMPLE_PORT)
	if err != nil {
		t.Error("Created client with empty address")
	}
	_, err = CreateClient(EXAMPLE_NETWORK, "not a domain", EXAMPLE_PORT)
	if err != nil {
		t.Error("Created client with faulty address")
	}
	_, err = CreateClient(EXAMPLE_NETWORK, EXAMPLE_ADDRESS, -1)
	if err != nil {
		t.Error("Created client with negative port")
	}
	_, err = CreateClient(EXAMPLE_NETWORK, EXAMPLE_ADDRESS, 0)
	if err == nil {
		t.Error("Failed creating a client with arbitrary port")
	}
}
