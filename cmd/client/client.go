package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"io/ioutil"
)

var target = flag.String("target", "http://localhost:8090", "request target")

func main() {
	flag.Parse()
	client := new(http.Client)
	client.Timeout = 10 * time.Second

	for range time.Tick(1 * time.Second) {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", *target))
		if err == nil {
			body, _ := ioutil.ReadAll(resp.Body)
			bodyString := string(body)
			fmt.Printf(bodyString)
			log.Printf("response %d", resp.StatusCode)
		} else {
			log.Printf("error %s", err)
		}
	}
}
