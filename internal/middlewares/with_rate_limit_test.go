package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimit_ServicePathsExempt(t *testing.T) {
	limited := WithRateLimit(t.Context(), 1, 1)( // rps=1, burst=1 almost no budget
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

	for path := range ServicePaths {
		for i := 0; i < 20; i++ {
			rec := httptest.NewRecorder()
			limited.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("%s request %d: got %d, want 200 — service path hit the limiter",
					path, i+1, rec.Code)
			}
		}
	}
}

func TestRateLimit_RegularPathsLimited(t *testing.T) {
	limited := WithRateLimit(t.Context(), 1, 1)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

	got429 := false
	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		limited.ServeHTTP(rec, httptest.NewRequest(http.MethodGet,
			"/api/v1/avatars/x", nil))
		if rec.Code == http.StatusTooManyRequests {
			got429 = true
		}
	}
	if !got429 {
		t.Fatal("20 rapid requests never hit 429 — limiter inactive, " +
			"exemption test would pass vacuously")
	}
}
