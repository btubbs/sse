package sse

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tt := []struct {
		desc   string
		source []byte
		events []Event
	}{
		{
			desc:   "empty",
			source: []byte(""),
			events: []Event{},
		},
		{
			desc:   "no colon",
			source: []byte("blah"),
			events: []Event{},
		},
		{
			desc:   "comment",
			source: []byte(":"),
			events: []Event{},
		},
		{
			desc:   "simple.  no newline at end",
			source: []byte("data:foo"),
			events: []Event{{Data: []byte("foo")}},
		},
		{
			desc:   "space after colon",
			source: []byte("data: foo"),
			events: []Event{{Data: []byte("foo")}},
		},
		{
			desc:   "multi lines in one event",
			source: []byte("data:foo\ndata:bar"),
			events: []Event{{Data: []byte("foo\nbar")}},
		},
		{
			desc:   "two events",
			source: []byte("data:foo\n\ndata:bar"),
			events: []Event{
				{Data: []byte("foo")},
				{Data: []byte("bar")},
			},
		},
		{
			desc:   "leading bom",
			source: []byte("\ufeffdata:foo"),
			events: []Event{{Data: []byte("foo")}},
		},
		{
			desc:   "named event",
			source: []byte("data:foo\nevent:bar"),
			events: []Event{{Event: "bar", Data: []byte("foo")}},
		},
		{
			desc:   "an id",
			source: []byte("data:foo\nid:bar"),
			events: []Event{{ID: "bar", Data: []byte("foo")}},
		},
		{
			desc:   "a retry",
			source: []byte("data:foo\nid:bar\nretry:345"),
			events: []Event{
				{
					ID:    "bar",
					Data:  []byte("foo"),
					Retry: time.Millisecond * 345,
				},
			},
		},
		{
			desc:   "a malformed retry",
			source: []byte("data:foo\nid:bar\nretry:potato"),
			events: []Event{
				{
					ID:   "bar",
					Data: []byte("foo"),
				},
			},
		},
	}

	for _, tc := range tt {
		idx := 0
		Parse(bytes.NewBuffer(tc.source), func(e Event) {
			assert.Equal(t, tc.events[idx], e, tc.desc)
			idx++
		})
	}
}

func TestScanSSELines(t *testing.T) {
	input := "foo\nbar\rbaz\r\nquux"
	lines := []string{"foo", "bar", "baz", "quux"}

	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(scanSSELines)

	idx := 0
	for scanner.Scan() {
		assert.Equal(t, lines[idx], scanner.Text())
		idx++
	}
}
