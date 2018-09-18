package sse

import (
	"net/http"
	"time"
)

// so we want the following:
// provide a URL, get a stream of events
// provide a prepared request, get a stream of events.
// I can implement the URL nice one on top of the prepared request one.

type Client struct {
	http.Client
	RetryDuration time.Duration
	LastID        string
}

func (c *Client) Subscribe(req *http.Request, f func(Event)) error {
	// wrap the provided callback in our own that will do the right thing with the ID and Retry
	// fields.
	callback := func(e Event) {
		if e.Retry > 0 {
			c.RetryDuration = e.Retry
		}

		if e.ID != "" {
			c.LastID = e.ID
		}
		f(e)
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	Parse(resp.Body, callback)
	return nil
}
