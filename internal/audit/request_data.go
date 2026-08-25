package audit

import "context"

type requestDataKey struct{}

// RequestData stores audit metadata
type RequestData struct {
	Operation string
	UserID    int64
	SecretID  string
}

func NewRequestData() *RequestData {
	return &RequestData{}
}

func WithRequestData(ctx context.Context, data *RequestData) context.Context {
	return context.WithValue(ctx, requestDataKey{}, data)
}

// FromContext returns audit request metadata stored in ctx, if any.
func FromContext(ctx context.Context) *RequestData {
	data, _ := ctx.Value(requestDataKey{}).(*RequestData)
	return data
}

// SetOperation records the operation name in the RequestData stored in ctx.
func SetOperation(ctx context.Context, operation string) {
	if data := FromContext(ctx); data != nil {
		data.Operation = operation
	}
}

// SetUserID records the authenticated user id in the RequestData stored in ctx.
func SetUserID(ctx context.Context, userID int64) {
	if data := FromContext(ctx); data != nil {
		data.UserID = userID
	}
}

// SetSecretID records the target secret id in the RequestData stored in ctx.
func SetSecretID(ctx context.Context, id string) {
	if data := FromContext(ctx); data != nil {
		data.SecretID = id
	}
}
