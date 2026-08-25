package run

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Vla8islav/gophprofile/internal/audit"
	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/domain"
	"github.com/Vla8islav/gophprofile/internal/handler"
	"github.com/Vla8islav/gophprofile/internal/middlewares"
	"github.com/Vla8islav/gophprofile/internal/service"
	"go.uber.org/zap"
)

func Run(ctx context.Context, db domain.GophprofileRepository, cfg *config.OptionsServer, logger *zap.Logger) error {

	srvApp := service.NewMetricsService(db,
		cfg.AuthTokenSecret.Value)

	h := handler.NewHandler(srvApp, logger)
	r := handler.NewRouter(h, cfg)

	// Middleware chain - first arg to ChainMiddlewares is the outermost wrapper
	mws := []middlewares.Middleware{middlewares.WithLogging(logger)}
	if cfg.AuditLogPath.Value != "" {
		sink := audit.NewFileSink(cfg.AuditLogPath.Value)
		defer func() { _ = sink.Close() }()
		publisher := audit.NewPublisher(sink)
		mws = append([]middlewares.Middleware{middlewares.WithAudit(publisher)}, mws...)
		logger.Info("audit log enabled", zap.String("path", cfg.AuditLogPath.Value))
	}
	handlerWithMW := middlewares.ChainMiddlewares(r, mws...)

	srv := &http.Server{
		Addr:         cfg.ServerAddress.Value,
		Handler:      handlerWithMW,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	// TLS when both cert and key are configured (local mkcert dev); otherwise plain
	// HTTP, for running behind a TLS-terminating reverse proxy, caddy, in prod
	var err error
	if cfg.PublicCertKey.Value != "" && cfg.PrivateKey.Value != "" {
		err = srv.ListenAndServeTLS(cfg.PublicCertKey.Value, cfg.PrivateKey.Value)
	} else {
		logger.Info("serving plain HTTP; no cert configured; expecting a TLS-terminating proxy in front")
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
