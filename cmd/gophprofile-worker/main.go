// consumes avatar events from Kafka
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Vla8islav/gophprofile/internal/broker"
	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/filestorage"
	"github.com/Vla8islav/gophprofile/internal/repository"
	"github.com/Vla8islav/gophprofile/internal/tracing"
	"github.com/Vla8islav/gophprofile/internal/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const consumerGroupID = "gophprofile-worker"
const metricsAddress = ":9091"

func main() {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer lg.Sync()

	// same flag pools
	currentConfig, err := config.ReadFlagsServer(os.Args[1:], lg)
	if err != nil {
		lg.Fatal("failed to read config", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := tracing.Init(ctx, "gophprofile-worker")
	if err != nil {
		lg.Fatal("init tracing", zap.Error(err))
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(flushCtx); err != nil {
			lg.Warn("tracing shutdown", zap.Error(err))
		}
	}()

	db, err := repository.NewPostgresStorage(currentConfig, "")
	if err != nil {
		lg.Fatal("init db", zap.Error(err))
	}

	fileStorage, err := filestorage.NewMinioStorage(ctx,
		currentConfig.S3Endpoint.Value,
		currentConfig.S3AccessKey.Value,
		currentConfig.S3SecretKey.Value,
		currentConfig.S3Bucket.Value,
		currentConfig.S3UseSSL.Value,
	)
	if err != nil {
		lg.Fatal("init file storage", zap.Error(err))
	}

	consumer := broker.NewKafkaConsumer(
		strings.Split(currentConfig.KafkaBrokers.Value, ","),
		currentConfig.KafkaTopic.Value,
		consumerGroupID,
		lg,
	)
	avatarWorker := worker.New(db, fileStorage)

	// metrics
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	metricsSrv := &http.Server{Addr: metricsAddress, Handler: metricsMux}
	go func() {
		lg.Info("metrics server listening", zap.String("addr", metricsAddress))
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Error("metrics server failed", zap.Error(err))
		}
	}()

	lg.Info("worker starting",
		zap.String("topic", currentConfig.KafkaTopic.Value),
		zap.String("group", consumerGroupID),
	)
	if err := consumer.Run(ctx, avatarWorker.HandleEvent); err != nil {
		lg.Fatal("consumer stopped", zap.Error(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		lg.Warn("metrics server shutdown", zap.Error(err))
	}

	lg.Info("worker stopped gracefully")
}
