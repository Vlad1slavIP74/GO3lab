package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"hash/fnv"
	"errors"


	"github.com/Vlad1slavIP74/GO3lab/httptools"
	"github.com/Vlad1slavIP74/GO3lab/signal"
)


var (
	port = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

var (
	timeout = time.Duration(*timeoutSec) * time.Second
	serversPool = map[string]bool {
		"server1:8080": true,
		"server2:8080": true,
		"server3:8080": true,
	}
)

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

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func addressHash(servers map[string]bool, clientAddress string) (string, error) {
	if (clientAddress == "") {
		return "", errors.New("BAD ADDRESS")
	}

	var availableServers []string

	for server, isAvailable := range servers {
		if isAvailable {
			availableServers = append(availableServers, server)
		}
	}


	if len(availableServers) == 0 {
		return "", errors.New("ALL SERVERS ARE NOT HEALTH")
	}

	checksum := hash(clientAddress)
	serverIndex := checksum % len(availableServers)
	availableServer := availableServers[serverIndex]
	return availableServer, nil
}

func main() {
	flag.Parse()

	for server := range serversPool {
		server := server
		go func() {
			for range time.Tick(10 * time.Second) {
				log.Println(server, health(server))
				serversPool[server] = health(server)
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		targetServer, err := addressHash(serversPool, r.Header.Get("X-Forwarded-For"))
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusServiceUnavailable)
		} else {
			forward(targetServer, rw, r)
		}
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}