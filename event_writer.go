package sse

import (
	"net/http"
)

// EventWriter is a helper for writing events to an io.Writer stream.
type EventWriter struct {
	w http.ResponseWriter
}

// NewEventWriter returns a new EventWriter.
func NewEventWriter(w http.ResponseWriter) *EventWriter {
	w.Header().Set(contentTypeHeader, textEventStream)
	return &EventWriter{w: w}
}

// Write takes an Event, serializes the Event to bytes, and writes them out to the http stream.  If
// the stream is also an http.Flusher, then it will also flush the bytes out so the event
// will be sent over the wire immediately. It automatically sets the Content-Type header to
// text/event-stream.
func (evw *EventWriter) Write(e Event) error {

	if _, err := evw.w.Write(e.Bytes()); err != nil {
		return err
	}
	if flusher, ok := evw.w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}
