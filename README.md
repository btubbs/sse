# sse [![Build Status](https://travis-ci.org/btubbs/sse.svg?branch=master)](https://travis-ci.org/btubbs/sse) [![Coverage Status](https://coveralls.io/repos/github/btubbs/sse/badge.svg?branch=master)](https://coveralls.io/github/btubbs/sse?branch=master)

This package provides several tools for working with Server Sent Event streams in Go, including a
client library, stream parser, and stream generator.

# Client

To subscribe to an SSE stream, pass a `http.Request` object to `client.Subscribe`:

```go
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
```

TODO:
- Add automatic retry logic to client
- Add event emitter/publisher that will periodically push a comment line to the stream to act as a
  keepalive.
-  set headers
-  add Disconnect method to client
- Add examples.
