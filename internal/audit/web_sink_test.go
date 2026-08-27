package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebSink_WritePostsEvent(t *testing.T) {
	var got bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		got = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sink, err := NewWebSink(srv.URL)
	require.NoError(t, err)
	require.NoError(t, sink.Write(context.Background(), Event{Operation: "secret.create", Status: 201}))
	require.True(t, got)
}

func TestWebSink_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	sink, err := NewWebSink(srv.URL)
	require.NoError(t, err)
	require.Error(t, sink.Write(context.Background(), Event{Operation: "x"}))
}

func TestNewWebSink_BadURL(t *testing.T) {
	_, err := NewWebSink("://not-a-url")
	require.Error(t, err)
}
