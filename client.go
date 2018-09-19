package sse

import (
	"fmt"
	"net/http"
	"time"
)

// note that this header key does not conform to the canonical HTTP form, which would have "Id" at
// the end instead of "ID".  It DOES, however, conform to the SSE spec.
const (
	lastEventIDHeader  = "Last-Event-ID"
	cacheControlHeader = "Cache-Control"
	noCache            = "no-cache"
	contentTypeHeader  = "Content-Type"
	textEventStream    = "text/event-stream"
)

// Client works like a regular http.Client, with an additional Subscribe method for connecting to
// Server Sent Event streams.
type Client struct {
	http.Client
	RetryDuration time.Duration
	LastID        string
}

// Subscribe performs the provided request, then passes the response into the SSE parser, calling
// the provided callback with each Event parsed from the stream.
func (c *Client) Subscribe(req *http.Request, f func(Event), options ...ClientOption) error {
	// apply the options
	co := &clientOptions{}
	for _, o := range options {
		o(co)
	}

	req.Header.Set(cacheControlHeader, noCache)

	if co.lastID != "" {
		req.Header.Set(lastEventIDHeader, co.lastID)
	}

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

	err := c.doSSERequest(req, callback)
	if err != nil {
		return err
	}

	// if we get here, then the stream disconnected.  We'll try to reconnect if we were configured to
	// do so.
	if co.autoRetry {
		for {
			time.Sleep(c.RetryDuration)
			req.Header.Set(lastEventIDHeader, c.LastID)
			// TODO: double check the spec.  Is it right to return here if we can't connect at all, or
			// should we keep retrying forever?
			err := c.doSSERequest(req, callback)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) doSSERequest(req *http.Request, callback func(Event)) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.Header.Get(contentTypeHeader) != textEventStream {
		return fmt.Errorf("response does not have %s Content-Type", textEventStream)
	}
	Parse(resp.Body, callback)
	return nil
}

type clientOptions struct {
	autoRetry bool
	lastID    string
}

// A ClientOption is a function that can be passed as an argument to Client.Subscribe to alter its
// behavior.
type ClientOption func(*clientOptions)

// AutoRetry takes a bool that flags whether the client should automatically reconnect on errors.
// The ClientOption that you get back should be passed into Client.Subscribe.
func AutoRetry(retry bool) ClientOption {
	return func(co *clientOptions) {
		co.autoRetry = retry
	}
}

func LastEventID(id string) ClientOption {
	return func(co *clientOptions) {
		co.lastID = id
	}
}
