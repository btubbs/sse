package sse

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	tt := []struct {
		desc              string
		server            http.HandlerFunc
		doAssertions      func()
		expectedEvents    []Event
		expectedErrRegexp interface{}
		extraAssertions   func(Client)
	}{
		{
			desc: "just data",
			server: func(w http.ResponseWriter, r *http.Request) {
				w.Write(Event{Data: []byte("msg1")}.Bytes())
				w.Write(Event{Data: []byte("msg2")}.Bytes())
			},
			expectedEvents: []Event{
				{Data: []byte("msg1")},
				{Data: []byte("msg2")},
			},
		},
		{
			desc: "IDs",
			server: func(w http.ResponseWriter, r *http.Request) {
				w.Write(Event{Data: []byte("msg1"), ID: "foo"}.Bytes())
				w.Write(Event{Data: []byte("msg2"), ID: "bar"}.Bytes())
			},
			expectedEvents: []Event{
				{Data: []byte("msg1"), ID: "foo"},
				{Data: []byte("msg2"), ID: "bar"},
			},
			extraAssertions: func(c Client) {
				assert.Equal(t, "bar", c.LastID)
			},
		},
		{
			desc: "retries",
			server: func(w http.ResponseWriter, r *http.Request) {
				w.Write(Event{Data: []byte("msg1"), Retry: 500 * time.Millisecond}.Bytes())
				w.Write(Event{Data: []byte("msg2"), Retry: 1200 * time.Millisecond}.Bytes())
			},
			expectedEvents: []Event{
				{Data: []byte("msg1"), Retry: 500 * time.Millisecond},
				{Data: []byte("msg2"), Retry: 1200 * time.Millisecond},
			},
			extraAssertions: func(c Client) {
				assert.Equal(t, 1200*time.Millisecond, c.RetryDuration)
			},
		},
		{
			desc: "server error",
			server: func(w http.ResponseWriter, r *http.Request) {
				panic("boom")
			},
			expectedErrRegexp: `Get http://127.0.0.1:\d+: EOF`,
		},
	}

	for _, tc := range tt {
		server := httptest.NewServer(tc.server)

		client := Client{}
		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		idx := 0
		err := client.Subscribe(req, func(e Event) {
			assert.Equal(t, tc.expectedEvents[idx], e, tc.desc)
			idx++
		})
		assert.Equal(t, len(tc.expectedEvents), idx, tc.desc)
		if tc.extraAssertions != nil {
			tc.extraAssertions(client)
		}
		if tc.expectedErrRegexp != nil {
			assert.Regexp(t, tc.expectedErrRegexp, err.Error())
		}
	}
}
