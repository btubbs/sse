package sse

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventBytes(t *testing.T) {
	tt := []struct {
		ev       Event
		expected []byte
	}{
		{
			ev: Event{
				Data: []byte("foo"),
			},
			expected: []byte("data: foo\n\n"),
		},
		{
			ev: Event{
				ID:   "bar",
				Data: []byte("foo"),
			},
			expected: []byte("id: bar\ndata: foo\n\n"),
		},
		{
			ev: Event{
				ID:    "bar",
				Event: "blah",
				Data:  []byte("foo"),
			},
			expected: []byte("id: bar\nevent: blah\ndata: foo\n\n"),
		},
		{
			ev: Event{
				ID:    "bar",
				Event: "blah",
				Data:  []byte("foo"),
				Retry: time.Millisecond * 500,
			},
			expected: []byte("id: bar\nevent: blah\nretry: 500\ndata: foo\n\n"),
		},
		{
			ev: Event{
				ID:    "bar",
				Event: "blah",
				Data:  []byte("foo\nbaz"),
				Retry: time.Millisecond * 500,
			},
			expected: []byte("id: bar\nevent: blah\nretry: 500\ndata: foo\ndata: baz\n\n"),
		},
	}

	for _, tc := range tt {
		result := tc.ev.Bytes()
		fmt.Println(string(result))
		assert.Equal(t, tc.expected, result)
	}
}
