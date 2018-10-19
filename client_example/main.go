package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/btubbs/sse"
)

func main() {
	req, err := http.NewRequest(
		http.MethodGet,
		// This app will output a stream of incrementing integers.
		"http://localhost:8080",
		nil,
	)
	if err != nil {
		panic(err.Error())
	}

	log.Fatal(sse.Subscribe(req, func(e sse.Event) {
		fmt.Println(string(e.Data))
	}))
}
