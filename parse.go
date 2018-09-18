package sse

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"time"
)

var (
	bom        = []byte("\ufeff")
	space      = []byte(" ")
	colon      = []byte(":")
	newline    = []byte("\n")
	fieldEvent = []byte("event")
	fieldData  = []byte("data")
	fieldID    = []byte("id")
	fieldRetry = []byte("retry")
)

// Parse takes a reader and a callback function.  It reads the stream and parses out SSE event
// payloads, builds an Event struct for each, and passes it into the provided callback.
func Parse(r io.Reader, f func(Event)) {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanSSELines)

	checkBOM := true
	ev := Event{}
	for scanner.Scan() {
		b := scanner.Bytes()
		// discard the leading BOM from first line, if present.
		if checkBOM {
			if len(b) >= 3 && bytes.Equal(b[:3], bom) {
				b = b[3:]
			}
			checkBOM = false
		}

		// if blank, dispatch the event
		if len(b) == 0 {
			if len(ev.Data) > 0 {
				f(ev)
			}
			ev = Event{}
			continue
		}

		// if starts with colon, then skip
		if bytes.HasPrefix(b, colon) {
			// lines starting with colons are comments.  Skip.
			continue
		}

		field, value := splitLine(b)

		if bytes.Equal(field, fieldEvent) {
			ev.Event = string(value)
			continue
		}

		if bytes.Equal(field, fieldData) {
			if len(ev.Data) > 0 {
				ev.Data = append(ev.Data, newLine...)
			}
			ev.Data = append(ev.Data, value...)
			continue
		}

		if bytes.Equal(field, fieldID) {
			ev.ID = string(value)
			continue
		}

		if bytes.Equal(field, fieldRetry) {
			retry, err := parseRetry(value)
			if err != nil {
				// malformed retry values should be skipped
				continue
			}
			ev.Retry = retry
			continue
		}
	}
}

func splitLine(b []byte) ([]byte, []byte) {
	idx := bytes.Index(b, colon)
	if idx == -1 {
		return b, []byte{}
	}
	name := b[:idx]
	data := b[idx+1:]
	// discard space, if there was one right after the colon.
	if len(data) > 0 && bytes.HasPrefix(data, space) {
		data = data[1:]
	}
	return name, data
}

func parseRetry(b []byte) (time.Duration, error) {
	millis, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}
	return time.Millisecond * time.Duration(millis), nil
}

// scanSSELines is a split function for a Scanner that returns each line of text, stripped of any
// trailing end-of-line marker. The returned line may be empty.  This function treats \r as a valid
// line separator in addition to the \r\n and \n separators supported by bufio.ScanLines.
func scanSSELines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' {
			if data[i+1] == '\n' {
				return i + 2, data[0:i], nil
			}
			return i + 1, data[0:i], nil
		}

		if data[i] == '\n' {
			return i + 1, data[0:i], nil
		}
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
