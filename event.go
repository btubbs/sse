package sse

import (
	"bytes"
	"fmt"
	"time"
)

const (
	newLine            = "\n"
	defaultMessageType = "message"
	idPrefix           = "id"
	eventPrefix        = "event"
	retryPrefix        = "retry"
	dataPrefix         = "data"
	colonSpace         = ": " // spaces are optional, but nice when viewing raw stream data.
)

var newLineBytes = []byte(newLine)

type Event struct {
	ID    string
	Event string
	Retry time.Duration
	Data  []byte
}

// Bytes serializes the Event to a byte slice suitable for writing out to the wire as a valid SSE
// event.
func (e Event) Bytes() []byte {
	var out []byte
	// if we have an id, put that first
	if e.ID != "" {
		out = append(out, concatField(idPrefix, e.ID)...)
	}
	// if we have an event other than "message" or "", put that next.
	if e.Event != "" && e.Event != defaultMessageType {
		out = append(out, concatField(eventPrefix, e.Event)...)
	}
	// if we have a retry, put that next
	if e.Retry > 0 {
		out = append(out, concatField(retryPrefix, fmt.Sprintf("%d", milliseconds(e.Retry)))...)
	}
	// append the data.  Split it on newlines and output each one with a "data: " prefix.
	lines := bytes.Split(e.Data, newLineBytes)
	for _, l := range lines {
		out = append(out, []byte(dataPrefix)...)
		out = append(out, []byte(colonSpace)...)
		out = append(out, l...)
		out = append(out, newLineBytes...)
	}
	// add a final newline
	out = append(out, newLineBytes...)
	return out
}

func concatField(prefix, data string) []byte {
	return []byte(prefix + colonSpace + data + newLine)
}

func milliseconds(d time.Duration) int {
	// there are a million nanoseconds per millisecond.  Since a time.Duration is just a number of
	// nanoseconds, divide by a million to get a number of milliseconds.
	return int(d / time.Millisecond)
}
