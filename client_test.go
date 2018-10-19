package sse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	tt := []struct {
		desc              string
		handlers          []http.HandlerFunc
		expectedEvents    []Event
		expectedErrRegexp interface{}
		extraAssertions   func(Client)
		clientOptions     []ClientOption
	}{
		{
			desc: "just data",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					// assert that the required request headers are present
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					assert.Equal(t, noCache, r.Header.Get(cacheControlHeader))
					writeEvent(w, Event{Data: []byte("msg1")})
					writeEvent(w, Event{Data: []byte("msg2")})
				}},
			expectedEvents: []Event{
				{Data: []byte("msg1")},
				{Data: []byte("msg2")},
			},
		},
		{
			desc: "cache control header",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					assert.Equal(t, noCache, r.Header.Get(cacheControlHeader))
				}},
			expectedEvents: []Event{},
		},
		{
			desc: "last ID option",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					assert.Equal(t, "foo", r.Header.Get(lastEventIDHeader))
				}},
			expectedEvents: []Event{},
			clientOptions:  []ClientOption{LastEventID("foo")},
		},
		{
			desc: "bad content type",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, "foo/foo")
				}},
			expectedEvents:    []Event{},
			expectedErrRegexp: "response does not have text/event-stream Content-Type",
		},
		{
			desc: "missing content type",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, "foo/foo")
				}},
			expectedEvents:    []Event{},
			expectedErrRegexp: "response does not have text/event-stream Content-Type",
		},
		{
			desc: "IDs",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					writeEvent(w, Event{Data: []byte("msg1"), ID: "foo"})
					writeEvent(w, Event{Data: []byte("msg2"), ID: "bar"})
				},
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
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					writeEvent(w, Event{Data: []byte("msg1"), Retry: 500 * time.Millisecond})
					writeEvent(w, Event{Data: []byte("msg2"), Retry: 1200 * time.Millisecond})
				},
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
			desc: "server error on first request",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					panic("boom")
				},
			},
			expectedErrRegexp: `Get http://127.0.0.1:\d+: EOF`,
		},
		{
			desc: "auto retry on disconnect",
			handlers: []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					writeEvent(w, Event{Data: []byte("msg1"), Retry: 500 * time.Millisecond})
				},
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					if _, err := w.Write([]byte(":\n")); err != nil {
						panic("i hate you, linter")
					}
				},
				func(w http.ResponseWriter, r *http.Request) {
					h := w.Header()
					h.Set(contentTypeHeader, textEventStream)
					writeEvent(w, Event{Data: []byte("msg2")})
				},
				func(w http.ResponseWriter, r *http.Request) {
					panic("boom") // force breaking out of the loop
				},
			},
			clientOptions: []ClientOption{AutoRetry(true)},
			expectedEvents: []Event{
				{Data: []byte("msg1"), Retry: 500 * time.Millisecond},
				{Data: []byte("msg2")},
			},
			expectedErrRegexp: `Get http://127.0.0.1:\d+: EOF`,
		},
	}

	for _, tc := range tt {
		server := httptest.NewServer(buildTestServer(tc.handlers))

		client := Client{}
		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		idx := 0
		err := client.Subscribe(req, func(e Event) {
			fmt.Println("event", e)
			assert.Equal(t, tc.expectedEvents[idx], e, tc.desc)
			idx++
		},
			tc.clientOptions...,
		)
		if err != nil || tc.expectedErrRegexp != nil {
			assert.Regexp(t, tc.expectedErrRegexp, err.Error())
		}
		assert.Equal(t, len(tc.expectedEvents), idx, tc.desc)
		if tc.extraAssertions != nil {
			tc.extraAssertions(client)
		}
	}
}

func TestSubscribe(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set(contentTypeHeader, textEventStream)
		writeEvent(w, Event{Data: []byte("msg1"), Retry: 500 * time.Millisecond})
	}

	server := httptest.NewServer(http.HandlerFunc(f))
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	err := Subscribe(req, func(e Event) {
		assert.Equal(t, "msg1", string(e.Data))
	})
	assert.Nil(t, err)
}

func buildTestServer(handlers []http.HandlerFunc) http.HandlerFunc {
	i := 0
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("i", i)
		if i < len(handlers) {
			handlers[i](w, r)
			i++
		}
	}
}

func writeEvent(w http.ResponseWriter, e Event) {
	if _, err := w.Write(e.Bytes()); err != nil {
		panic(err.Error())
	}
	flusher := w.(http.Flusher)
	flusher.Flush()
}
