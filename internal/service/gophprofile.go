package service

import (
	"github.com/Vla8islav/gophprofile/internal/domain"
)

type gophprofileService struct {
	repository domain.GophprofileRepository
	authSecret []byte
}

func NewGophprofileService(
	repo domain.GophprofileRepository,
	authSecret string,
) domain.GophprofileService {
	return gophprofileService{
		repository: repo,
		authSecret: []byte(authSecret),
	}
}
