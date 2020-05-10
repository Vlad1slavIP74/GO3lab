package main

import (
	 "testing"
	 "github.com/stretchr/testify/assert"
)

func TestWhenAllServersOff(t *testing.T) {
	_, err := addressHash(map[string]bool {
		"server1:8080": false,
		"server2:8080": false,
		"server3:8080": false,
	}, "192.168.116.16")
	
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "ALL SERVERS ARE NOT HEALTH")
}

func TestWhenOneServerOn(t *testing.T) {
	server,_ := addressHash(map[string]bool {
		"server1:8080": true,
		"server2:8080": false,
		"server3:8080": false,
	}, "192.168.116.16")
	assert.Equal(t, server, "server1:8080", "Wrong url")
}

func TestForDifferentHash(t *testing.T) {
	serverFirst,_ := addressHash(map[string]bool {
		"server1:8080": true,
		"server2:8080": true,
		"server3:8080": true,
	}, "192.168.116.16")

	serverSecond,_ := addressHash(map[string]bool {
		"server1:8080": true,
		"server2:8080": true,
		"server3:8080": true,
	}, "102.178.186.28")

	if serverFirst == serverSecond {
		t.Fail()
	}
}