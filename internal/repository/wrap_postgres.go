package repository

import (
	"fmt"

	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/domain"
)

func WrapPostgres(currentConfig *config.OptionsServer) (domain.GophprofileRepository, error) {
	var db domain.GophprofileRepository
	var err error

	// Case 1
	if currentConfig.DatabaseURI.BeenSet {
		db, err = NewPostgresStorage(currentConfig, currentConfig.MigrationsFolder.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize gophprofile repository: %w", err)
		}
		return db, nil
	}

	return nil, fmt.Errorf("something strange happened: " +
		"DB DSN isn't configured")
}
