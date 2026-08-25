package helpers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeDoRequester struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (f *fakeDoRequester) Do(req *http.Request) (*http.Response, error) {
	return f.doFunc(req)
}

func newResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestHTTPRetryClient_Do_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				return newResponse(http.StatusOK, "ok"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 1, calls)
}

func TestHTTPRetryClient_Do_RetriesOnRetriableStatusThenSucceeds(t *testing.T) {
	calls := 0

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return newResponse(http.StatusServiceUnavailable, "retry"), nil
				}
				return newResponse(http.StatusOK, "ok"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 2, calls)
}

func TestHTTPRetryClient_Do_ReturnsLastRetriableStatusOnLastAttempt(t *testing.T) {
	calls := 0

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				return newResponse(http.StatusServiceUnavailable, "still failing"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.Equal(t, 3, calls)
}

func TestHTTPRetryClientDoesNotRetryOnNonRetriableStatus(t *testing.T) {
	calls := 0

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				return newResponse(http.StatusBadRequest, "bad request"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, 1, calls)
}

func TestHTTPRetryClientRetriesOnErrorThenSucceeds(t *testing.T) {
	calls := 0
	wantErr := errors.New("temporary transport error")

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return nil, wantErr
				}
				return newResponse(http.StatusOK, "ok"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 2, calls)
}

func TestHTTPRetryClientReturnsErrorOnLastAttempt(t *testing.T) {
	calls := 0
	wantErr := errors.New("permanent transport error")

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				return nil, wantErr
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.ErrorIs(t, err, wantErr)
	require.Nil(t, resp)
	require.Equal(t, 3, calls)
}

func TestHTTPRetryClientUsesDefaultRetryStatusWhenNil(t *testing.T) {
	calls := 0

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return newResponse(http.StatusGatewayTimeout, "retry"), nil
				}
				return newResponse(http.StatusOK, "ok"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: nil,
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 2, calls)
}

func TestHTTPRetryClientRecreatesBodyBetweenAttempts(t *testing.T) {
	calls := 0
	var seenBodies []string

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++

				data, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				seenBodies = append(seenBodies, string(data))

				if calls == 1 {
					return newResponse(http.StatusServiceUnavailable, "retry"), nil
				}
				return newResponse(http.StatusOK, "ok"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	bodyBytes := []byte("payload")
	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(bodyBytes))
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 2, calls)
	require.Len(t, seenBodies, 2)
	require.Equal(t, "payload", seenBodies[0])
	require.Equal(t, "payload", seenBodies[1])
}

func TestHTTPRetryClientReturnsGetBodyError(t *testing.T) {
	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				t.Fatal("client.Do must not be called when GetBody fails")
				return nil, nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        0,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("payload")))
	require.NoError(t, err)

	wantErr := errors.New("get body failed")
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, wantErr
	}

	resp, err := client.Do(req)
	require.ErrorIs(t, err, wantErr)
	require.Nil(t, resp)
}

func TestHTTPRetryClientStopsOnContextCancellationDuringDelay(t *testing.T) {
	calls := 0

	client := &HTTPRetryClient{
		client: &fakeDoRequester{
			doFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				return newResponse(http.StatusServiceUnavailable, "retry"), nil
			},
		},
		maxAttempts:         3,
		attemptDelay:        time.Second,
		shouldRetryOnStatus: DefaultShouldRetryStatus,
	}

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	done := make(chan struct{})
	var resp *http.Response
	var doErr error

	go func() {
		defer close(done)
		resp, doErr = client.Do(req)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	require.ErrorIs(t, doErr, context.Canceled)
	require.Nil(t, resp)
	require.Equal(t, 1, calls)
}
