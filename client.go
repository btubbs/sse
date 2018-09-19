package sse

import (
	"net/http"
	"time"
)

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
