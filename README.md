# sse

This package provides several tools for working with Server Sent Event streams in Go.  Specifically:

- An `Event` struct with the fields specified by the SSE specification, using idiomatic Go types
  (e.g. the `retry` field in the spec is represented by a Go `time.Duration`).
- A `Bytes()` method on that struct, which will serialize the event into a byte slice that conforms
  to the SSE spec, and can be written to any `Writer` (including an `http.ResponseWriter`).
- A `Client` struct with all the same methods as an `http.Client`, with an added `Subscribe` method
  that takes a prepared http request and an event callback.
- A `Parse` function that takes any `io.Reader` that contains a stream of SSE bytes, and a callback
  function.  For each event in the stream, an `Event` object will be created and passed to the
  provided callback.

TODO:
- Add examples.
- Add automatic retry logic
- Add event emitter/publisher that will periodically push a comment line to the stream to act as a
  keepalive.
