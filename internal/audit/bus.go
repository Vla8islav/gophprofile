// Package audit
package audit

import (
	"context"
)

// Event describes an audited request against the secrets API.
type Event struct {
	Time       UnixTime `json:"ts"`
	Operation  string   `json:"operation"`
	UserID     int64    `json:"user_id"`
	SecretID   string   `json:"secret_id,omitempty"`
	RemoteAddr string   `json:"remote_addr"`
	Status     int      `json:"status"`
}

// Sink writes audit events to a destination
type Sink interface {
	Write(ctx context.Context, event Event) error
}

// Publisher sends audit events to one or more sinks
type Publisher struct {
	sinks []Sink
}

// NewPublisher creates a Publisher that writes events to sinks
func NewPublisher(sinks ...Sink) *Publisher {
	return &Publisher{sinks: sinks}
}

// Publish writes event to each configured sink
func (p *Publisher) Publish(ctx context.Context, event Event) error {
	for _, sink := range p.sinks {
		if err := sink.Write(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
