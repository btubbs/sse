package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/btubbs/sse"
)

func main() {
	// If you run server_example/main.go, then it will be listening on this host/port.
	url := "http://localhost:8080"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err.Error())
	}

	log.Fatal(sse.Subscribe(req, func(e sse.Event) {
		fmt.Println(string(e.Data))
	}))
}
