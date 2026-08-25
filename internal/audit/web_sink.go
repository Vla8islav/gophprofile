package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Vla8islav/gophprofile/internal/helpers"
)

// WebSink writes audit events to an HTTP endpoint
// generate:reset
type WebSink struct {
	client   *helpers.HTTPRetryClient
	auditURL string
}

// NewWebSink creates a WebSink that posts audit events to auditURL
func NewWebSink(auditURL string) (*WebSink, error) {
	retryClient := helpers.NewHTTPRetryClient(helpers.DefaultShouldRetryStatus,
		5*time.Second, 2)
	err := validateURL(auditURL)
	if err != nil {
		return nil, err
	}
	return &WebSink{auditURL: auditURL, client: retryClient}, nil
}

func validateURL(auditURL string) error {
	parsedURL, err := url.ParseRequestURI(auditURL)

	if err != nil {
		return fmt.Errorf("couldn't parse audit url: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("audit url must use http or https")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("audit url must include host")
	}

	return nil
}

// Write posts e to the configured audit endpoint as a gzipped JSON payload
func (s *WebSink) Write(ctx context.Context, e Event) error {

	payloadBytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	payloadBytesCompressed, err := helpers.GzipCompress(payloadBytes)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.auditURL, bytes.NewReader(payloadBytesCompressed))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil

}
