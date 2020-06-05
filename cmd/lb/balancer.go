package main

import (
	"container/heap"
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Vlad1slavIP74/GO3lab/httptools"
	"github.com/Vlad1slavIP74/GO3lab/signal"
)

var (
	port       = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https      = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

var (
	timeout = time.Duration(*timeoutSec) * time.Second
)

type ServerInfo struct {
	server  string
	healthy bool
	number  int
}

type healthyServersPool []int

func (h healthyServersPool) Len() int {
	return len(h)
}

func (h healthyServersPool) Less(i, j int) bool {
	return h[i] < h[j]
}
func (h healthyServersPool) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *healthyServersPool) Pop() interface{} {
	old := *h
	ret := old[len(old)-1]
	*h = old[:len(old)-1]
	return ret
}

func (h *healthyServersPool) Push(x interface{}) {
	*h = append(*h, x.(int))
}

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(dst string) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func forward(dst string, rw http.ResponseWriter, r *http.Request) error {
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = dst

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err == nil {
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			fmt.Println(dst, "ffnsdajnfljkdsanfkljs")
			rw.Header().Set("lb-from", dst)
		}
		log.Println("fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
}

func healthyState(server string, number int, pq *healthyServersPool) {
	if health(server) {
		heap.Push(pq, number)
		*pq = append(*pq, number)
		fmt.Println(pq.Len(), number)
	}
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func addressHash(clientAddress string, pq *healthyServersPool, heap []ServerInfo) string {
	checksum := hash(clientAddress)
	serverIndex := checksum % pq.Len()
	availableServer := heap[(*pq)[serverIndex]].server
	return availableServer
}

func main() {
	flag.Parse()

	pq := &healthyServersPool{}
	serverHeap := []ServerInfo{
		{"server1:8080", true, 0},
		{"server2:8080", true, 1},
		{"server3:8080", true, 2},
	}

	for _, value := range serverHeap {
		server := value.server
		number := value.number
		go func() {
			for range time.Tick(10 * time.Second) {
				log.Println(server, health(server), number)
				healthyState(server, number, pq)
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if pq.Len() != 0 {
			targetServer := addressHash(r.Header.Get("X-Forwarded-For"), pq, serverHeap)
			forward(targetServer, rw, r)
		}
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}
