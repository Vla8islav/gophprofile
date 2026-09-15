package servicemetrics

import (
	"context"
	"io"
	"time"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	serviceOps = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gophprofile_service_operations_total",
		Help: "Domain service operations, by method and result.",
	}, []string{"method", "result"})

	serviceDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gophprofile_service_operation_duration_seconds",
		Help:    "Domain service operation latency, by method.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})
)

type withMetrics struct {
	next domain.GophprofileService
}

func (m *withMetrics) Ping(ctx context.Context) error {
	start := time.Now()
	err := m.next.Ping(ctx)
	observe("Ping", start, err)
	return err
}

func (m *withMetrics) FileStoragePing(ctx context.Context) error {
	start := time.Now()
	err := m.next.FileStoragePing(ctx)
	observe("FileStoragePing", start, err)
	return err
}

func (m *withMetrics) BrokerPing(ctx context.Context) error {
	start := time.Now()
	err := m.next.BrokerPing(ctx)
	observe("BrokerPing", start, err)
	return err
}

func (m *withMetrics) CreateUser(ctx context.Context, request domain.UserRegisterRequest) (*domain.AuthResult, error) {
	start := time.Now()
	r, err := m.next.CreateUser(ctx, request)
	observe("CreateUser", start, err)
	return r, err

}

func (m *withMetrics) LoginUser(ctx context.Context, request domain.UserLoginRequest) (*domain.AuthResult, error) {
	start := time.Now()
	r, err := m.next.LoginUser(ctx, request)
	observe("LoginUser", start, err)
	return r, err

}

func (m *withMetrics) GetAvatarContent(ctx context.Context, avatarID string, sizeVariant string) (*domain.Avatar, io.ReadCloser, error) {
	start := time.Now()
	a, rc, err := m.next.GetAvatarContent(ctx, avatarID, sizeVariant)
	observe("GetAvatarContent", start, err)
	return a, rc, err
}

func (m *withMetrics) GetUserAvatarContent(ctx context.Context, userID int64, sizeVariant string) (*domain.Avatar, io.ReadCloser, error) {
	start := time.Now()
	a, rc, err := m.next.GetUserAvatarContent(ctx, userID, sizeVariant)
	observe("GetUserAvatarContent", start, err)
	return a, rc, err
}

func (m *withMetrics) GetAvatarMetadata(ctx context.Context, avatarID string) (*domain.Avatar, error) {
	start := time.Now()
	r, err := m.next.GetAvatarMetadata(ctx, avatarID)
	observe("GetAvatarMetadata", start, err)
	return r, err
}

func (m *withMetrics) ListUserAvatars(ctx context.Context, userID int64) ([]domain.Avatar, error) {
	start := time.Now()
	r, err := m.next.ListUserAvatars(ctx, userID)
	observe("ListUserAvatars", start, err)
	return r, err
}

func (m *withMetrics) DeleteAvatar(ctx context.Context, avatarID string, requesterID int64) error {
	start := time.Now()
	err := m.next.DeleteAvatar(ctx, avatarID, requesterID)
	observe("DeleteAvatar", start, err)
	return err
}

func (m *withMetrics) DeleteUserAvatar(ctx context.Context, userID int64, requesterID int64) error {
	start := time.Now()
	err := m.next.DeleteUserAvatar(ctx, userID, requesterID)
	observe("DeleteUserAvatar", start, err)
	return err
}

// compile-time proof we implement the full interface —
// if a method is missing or a signature drifts, this line fails the build
var _ domain.GophprofileService = (*withMetrics)(nil)

func Wrap(next domain.GophprofileService) domain.GophprofileService {
	return &withMetrics{next: next}
}

func observe(method string, start time.Time, err error) {
	result := "ok"
	if err != nil {
		result = "error"
	}
	serviceOps.WithLabelValues(method, result).Inc()
	serviceDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}

func (m *withMetrics) UploadAvatar(ctx context.Context, userID int64, fileName, mimeType string, size int64, content io.Reader) (*domain.Avatar, error) {
	start := time.Now()
	a, err := m.next.UploadAvatar(ctx, userID, fileName, mimeType, size, content)
	observe("UploadAvatar", start, err)
	return a, err
}
