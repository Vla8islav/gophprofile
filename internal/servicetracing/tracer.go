package servicetracing

import (
	"context"
	"io"

	"github.com/Vla8islav/gophprofile/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("gophprofile/service")
var _ domain.GophprofileService = (*withTracing)(nil) // compile guard

type withTracing struct {
	next domain.GophprofileService
}

func recordSpanResult(span trace.Span, err error) error {
	if err != nil && !domain.IsBusinessErr(err) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (t *withTracing) Ping(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "service.Ping")
	defer span.End()

	err := t.next.Ping(ctx)
	return recordSpanResult(span, err)
}

func (t *withTracing) FileStoragePing(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "service.FileStoragePing")
	defer span.End()

	err := t.next.FileStoragePing(ctx)
	return recordSpanResult(span, err)
}

func (t *withTracing) BrokerPing(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "service.BrokerPing")
	defer span.End()

	err := t.next.BrokerPing(ctx)
	return recordSpanResult(span, err)
}

func (t *withTracing) CreateUser(ctx context.Context, request domain.UserRegisterRequest) (*domain.AuthResult, error) {
	ctx, span := tracer.Start(ctx, "service.CreateUser")
	defer span.End()

	a, err := t.next.CreateUser(ctx, request)
	return a, recordSpanResult(span, err)
}

func (t *withTracing) LoginUser(ctx context.Context, request domain.UserLoginRequest) (*domain.AuthResult, error) {
	ctx, span := tracer.Start(ctx, "service.LoginUser")
	defer span.End()

	a, err := t.next.LoginUser(ctx, request)
	return a, recordSpanResult(span, err)
}

func (t *withTracing) GetAvatarContent(ctx context.Context, avatarID string, sizeVariant string) (*domain.Avatar, io.ReadCloser, error) {
	ctx, span := tracer.Start(ctx, "service.GetAvatarContent")
	defer span.End()

	a, r, err := t.next.GetAvatarContent(ctx, avatarID, sizeVariant)
	return a, r, recordSpanResult(span, err)
}

func (t *withTracing) GetUserAvatarContent(ctx context.Context, userID int64, sizeVariant string) (*domain.Avatar, io.ReadCloser, error) {
	ctx, span := tracer.Start(ctx, "service.GetUserAvatarContent")
	defer span.End()

	a, r, err := t.next.GetUserAvatarContent(ctx, userID, sizeVariant)
	return a, r, recordSpanResult(span, err)
}

func (t *withTracing) GetAvatarMetadata(ctx context.Context, avatarID string) (*domain.Avatar, error) {
	ctx, span := tracer.Start(ctx, "service.GetAvatarMetadata")
	defer span.End()

	a, err := t.next.GetAvatarMetadata(ctx, avatarID)
	return a, recordSpanResult(span, err)
}

func (t *withTracing) ListUserAvatars(ctx context.Context, userID int64) ([]domain.Avatar, error) {
	ctx, span := tracer.Start(ctx, "service.ListUserAvatars")
	defer span.End()

	a, err := t.next.ListUserAvatars(ctx, userID)
	return a, recordSpanResult(span, err)

}

func (t *withTracing) DeleteAvatar(ctx context.Context, avatarID string, requesterID int64) error {
	ctx, span := tracer.Start(ctx, "service.DeleteAvatar")
	defer span.End()

	err := t.next.DeleteAvatar(ctx, avatarID, requesterID)
	return recordSpanResult(span, err)
}

func (t *withTracing) DeleteUserAvatar(ctx context.Context, userID int64, requesterID int64) error {
	ctx, span := tracer.Start(ctx, "service.DeleteUserAvatar")
	defer span.End()

	err := t.next.DeleteUserAvatar(ctx, userID, requesterID)
	return recordSpanResult(span, err)
}

func Wrap(next domain.GophprofileService) domain.GophprofileService {
	return &withTracing{next: next}
}

func (t *withTracing) UploadAvatar(ctx context.Context, userID int64, fileName, mimeType string, size int64, content io.Reader) (*domain.Avatar, error) {
	ctx, span := tracer.Start(ctx, "service.UploadAvatar")
	defer span.End()

	a, err := t.next.UploadAvatar(ctx, userID, fileName, mimeType, size, content)
	return a, recordSpanResult(span, err)
}
