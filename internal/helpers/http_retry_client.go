package helpers

import (
	"io"
	"net/http"
	"time"
)

// HTTPDoRequester describes the subset of http.Client used by HTTPRetryClient
type HTTPDoRequester interface {
	// Do sends an HTTP request and returns the response.
	Do(req *http.Request) (*http.Response, error)
}

// HTTPRetryClient wraps an HTTP client with retry behavior for transient failures
type HTTPRetryClient struct {
	client HTTPDoRequester

	maxAttempts  int
	attemptDelay time.Duration
	timeout      time.Duration

	shouldRetryOnStatus func(statusCode int) bool
}

// NewHTTPRetryClient creates an HTTPRetryClient via predicate, timeout, and attempt limit
func NewHTTPRetryClient(shouldRetryOnStatus func(int) bool,
	timeout time.Duration, maxAttempts int) *HTTPRetryClient {
	if shouldRetryOnStatus == nil {
		shouldRetryOnStatus = DefaultShouldRetryStatus
	}
	return &HTTPRetryClient{client: &http.Client{Timeout: timeout}, maxAttempts: maxAttempts,
		shouldRetryOnStatus: shouldRetryOnStatus,
		timeout:             timeout}
}

// DefaultShouldRetryStatus reports whether statusCode represents a retryable HTTP response
func DefaultShouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// Do sends req and retries while the response status is retryable or the request fails
//
// If req has a GetBody function, Do resets the request body before each attempt
func (c *HTTPRetryClient) Do(req *http.Request) (*http.Response, error) {

	var resp *http.Response
	var doErr error

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		// reset the request body
		if req.GetBody != nil {
			body, getBodyErr := req.GetBody()
			if getBodyErr != nil {
				return nil, getBodyErr
			}
			req.Body = body

		}
		resp, doErr = c.client.Do(req)
		if doErr == nil {
			shouldRetryStatus := c.shouldRetryOnStatus
			if shouldRetryStatus == nil {
				shouldRetryStatus = DefaultShouldRetryStatus
			}

			if !shouldRetryStatus(resp.StatusCode) || attempt >= c.maxAttempts {
				return resp, nil
			}

			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

		} else if attempt >= c.maxAttempts {
			return resp, doErr
		}

		timer := time.NewTimer(c.attemptDelay)
		select {
		case <-req.Context().Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, req.Context().Err()
		case <-timer.C:
		}

	}

	return resp, doErr
}
