package main

import (
	"context"
	"log"
	"os"

	_ "github.com/Vla8islav/gophprofile/docs" // generated OpenAPI spec (swag init)
	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/repository"
	"github.com/Vla8islav/gophprofile/internal/run"
	"go.uber.org/zap"
)

// @title           GophProfile API
// @version         1.0
// @description     Profile picture manager service.
// @BasePath        /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Paste "Bearer <token>" - the token returned by /api/user/login or /api/user/register.
func main() {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize lg: %v", err)
	}
	defer lg.Sync() // flushes buffer, if any

	currentConfig, err := config.ReadFlagsServer(os.Args[1:], lg)
	if err != nil {
		lg.Fatal("failed to read config", zap.Error(err))
		return
	}
	lg.Info("starting server ", zap.String("Server addr", currentConfig.ServerAddress.Value))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := repository.WrapPostgres(currentConfig)
	if err != nil {
		lg.Fatal("init db: ", zap.Error(err))
	}

	err = run.Run(ctx, db, currentConfig, lg)
	if err != nil {
		lg.Fatal("failed to start server", zap.Error(err))
		return
	}

}
