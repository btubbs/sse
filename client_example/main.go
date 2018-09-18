package main

import (
	"fmt"
	"net/http"

	"github.com/btubbs/sse"
)

func main() {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://infinite-mountain-77592.herokuapp.com/events/",
		nil,
	)
	if err != nil {
		panic(err.Error())
	}

	client := sse.Client{}
	client.Subscribe(req, func(e sse.Event) {
		fmt.Println(
			client.LastID,
			string(e.Data),
		)
	})
}
