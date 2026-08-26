package service

import (
	"github.com/Vla8islav/gophprofile/internal/domain"
)

type gophprofileService struct {
	repository  domain.GophprofileRepository
	fileStorage domain.FileStorage
	authSecret  []byte
}

func NewGophprofileService(
	repo domain.GophprofileRepository,
	fileStorage domain.FileStorage,
	authSecret string,
) domain.GophprofileService {
	return gophprofileService{
		repository:  repo,
		fileStorage: fileStorage,
		authSecret:  []byte(authSecret),
	}
}
