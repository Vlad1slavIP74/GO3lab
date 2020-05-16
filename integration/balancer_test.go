package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {


	req3, err3 := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", baseAddress), nil)
	req3.Header.Add("X-Forwarded-For", "192.168.116.16")
	resp3, err3 := client.Do(req3)
	if err3 != nil {
		t.Error(err3)
	} else {
		assert.Equal(t, resp3.Header.Get("lb-from"), "server3:8080", "BAD ADDRESS")
	}

	req1, err1 := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", baseAddress), nil)
	req1.Header.Add("X-Forwarded-For", "192.150.119.16")
	resp1, err1 := client.Do(req1)
	if err1 != nil {
		t.Error(err1)
	} else {
		assert.Equal(t, resp1.Header.Get("lb-from"), "server1:8080", "BAD ADDRESS")
	}

	req2, err2 := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", baseAddress), nil)
	req2.Header.Add("X-Forwarded-For", "192.167.114.12")
	resp2, err2 := client.Do(req2)
	if err2 != nil {
		t.Error(err2)
	} else {
		assert.Equal(t, resp2.Header.Get("lb-from"), "server2:8080", "BAD ADDRESS")
	}
}

func BenchmarkBalancer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _= client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
	}
}