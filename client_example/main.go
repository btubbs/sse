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
		"https://infinite-mountain-77592.herokuapp.com/events/",
		nil,
	)
	if err != nil {
		panic(err.Error())
	}

	client := sse.Client{}
	log.Fatal(client.Subscribe(req, func(e sse.Event) {
		fmt.Println(
			client.LastID,
			string(e.Data),
		)
	}))
}
