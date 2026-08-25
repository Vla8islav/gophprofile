package mocks

//go:generate mockgen -destination=domain_mock.go -package=mocks github.com/Vla8islav/gophprofile/internal/domain GophprofileRepository,GophprofileService
