package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/btubbs/sse"
)

func main() {
	http.ListenAndServe(
		":8080",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Type", "text/event-stream")
			// loop forever and emit a counter
			n := 0
			for {
				n++
				e := sse.Event{Data: []byte(strconv.Itoa(n))}
				if _, err := w.Write(e.Bytes()); err != nil {
					panic(err.Error())
				}
				flusher := w.(http.Flusher)
				flusher.Flush()
				time.Sleep(time.Second)
			}
		}))
}
