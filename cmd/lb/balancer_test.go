package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//func CalcHealthState(ServerHeap []ServerInfo, pq *healthyServersPool) {
//	for i, value := range ServerHeap {
//		server := value.server
//		number := i
//		go func() {
//			for range time.Tick(10 * time.Second) {
//				log.Println(server, health(server), number)
//				healthyState(server, number, pq)
//			}
//		}()
//	}
//}

func TestBalancer(t *testing.T) {
	addr := []string{
		"192.168.116.16",
		"102.178.186.98",
		"255.20.0.4",
		"172.20.0.2",
	}
	//all servers are healthy
	pq := &healthyServersPool{0, 1, 2}

	serverHeap := []ServerInfo{
		{"server1:8080", true, 0},
		{"server2:8080", true, 1},
		{"server3:8080", true, 2},
	}

	for i := 0; i < len(addr); i++ {
		serverName := addressHash(addr[i], pq, serverHeap)
		if i == 0 || i == 1 {
			assert.Equal(t, serverName, "server2:8080")
		}
		if i == 2 {
			assert.Equal(t, serverName, "server3:8080")
		}
		if i == 3 {
			assert.Equal(t, serverName, "server1:8080")
		}
	}

	//server2 is down
	pq = &healthyServersPool{0, 2}

	for i := 0; i < len(addr); i++ {
		serverName := addressHash(addr[i], pq, serverHeap)
		if i == 1 {
			assert.Equal(t, serverName, "server1:8080")
		} else {
			assert.Equal(t, serverName, "server3:8080")
		}
	}

	//server1 and server2 are down
	pq = &healthyServersPool{2}

	for i := 0; i < len(addr); i++ {
		serverName := addressHash(addr[i], pq, serverHeap)
		assert.Equal(t, serverName, "server3:8080")
	}
}
