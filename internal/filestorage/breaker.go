package filestorage

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
)

var breakerState = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "gophprofile_storage_breaker_state",
	Help: "Circuit breaker state for file storage: 0=closed, 1=half-open, 2=open",
})

type BreakerStorage struct {
	next domain.FileStorage
	cb   *gobreaker.CircuitBreaker[any]
}

func NewBreakerStorage(next domain.FileStorage, logger *zap.Logger) *BreakerStorage {
	cb := gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
		Name:        "filestorage",
		MaxRequests: 1,                // trial calls allowed in half-open
		Timeout:     10 * time.Second, // open → half-open after this
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			breakerState.Set(float64(to))
			logger.Warn("storage circuit breaker state change",
				zap.String("from", from.String()), zap.String("to", to.String()))
		},
	})
	return &BreakerStorage{next: next, cb: cb}
}

func (b *BreakerStorage) Upload(ctx context.Context, key, contentType string, size int64, content io.Reader) error {
	_, err := b.cb.Execute(func() (any, error) {
		return nil, b.next.Upload(ctx, key, contentType, size, content)
	})
	return mapBreakerErr(err)
}

func (b *BreakerStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	v, err := b.cb.Execute(func() (any, error) {
		return b.next.Download(ctx, key)
	})
	if err != nil {
		return nil, mapBreakerErr(err)
	}
	return v.(io.ReadCloser), nil
}

func (b *BreakerStorage) Delete(ctx context.Context, key string) error {
	_, err := b.cb.Execute(func() (any, error) {
		return nil, b.next.Delete(ctx, key)
	})
	if err != nil {
		return mapBreakerErr(err)
	}
	return mapBreakerErr(err)
}

// Ping bypasses the breaker deliberately: /health should report minio's
// REAL state, and its successes/failures shouldn't vote in the breaker.
func (b *BreakerStorage) Ping(ctx context.Context) error { return b.next.Ping(ctx) }

func mapBreakerErr(err error) error {
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return domain.ErrStorageUnavailable
	}
	return err
}
