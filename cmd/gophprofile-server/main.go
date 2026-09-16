package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/Vla8islav/gophprofile/docs" // generated OpenAPI spec (swag init)
	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/gophprofile_server"
	"github.com/Vla8islav/gophprofile/internal/repository"
	"github.com/Vla8islav/gophprofile/internal/tracing"
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
		log.Fatalf("failed to initialize logger: %v", err)
	}

	err = run(lg)
	if err != nil {
		lg.Error("server failed", zap.Error(err))
	}
	_ = lg.Sync() // explicit, BEFORE exit — os.Exit skips main's defers too
	if err != nil {
		os.Exit(1)
	}

}

func run(lg *zap.Logger) error {

	currentConfig, err := config.ReadFlagsServer(os.Args[1:], lg)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	lg.Info("starting server ", zap.String("Server addr", currentConfig.ServerAddress.Value))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracing, err := tracing.Init(ctx, "gophprofile-server")
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(flushCtx); err != nil {
			lg.Warn("tracing shutdown", zap.Error(err))
		}
	}()

	db, err := repository.WrapPostgres(currentConfig)
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	err = gophprofile_server.Run(ctx, db, currentConfig, lg)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
