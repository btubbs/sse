package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/btubbs/sse"
)

func main() {
	log.Fatal(http.ListenAndServe(
		":8080",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ew := sse.NewEventWriter(w)
			n := 0
			for {
				n++
				e := sse.Event{Data: []byte(strconv.Itoa(n))}
				if err := ew.Write(e); err != nil {
					// web browsers are expected to disconnect at any time.  That will raise an error here.
					fmt.Println(err)
					return
				}
				time.Sleep(time.Second)
			}
		})))
}
