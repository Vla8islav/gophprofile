package domain

import (
	"context"
)

type GophprofileService interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, request UserRegisterRequest) (*AuthResult, error)
	LoginUser(ctx context.Context, request UserLoginRequest) (*AuthResult, error)
}
