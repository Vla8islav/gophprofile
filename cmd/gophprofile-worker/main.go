// consumes avatar events from Kafka
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Vla8islav/gophprofile/internal/broker"
	"github.com/Vla8islav/gophprofile/internal/config"
	"github.com/Vla8islav/gophprofile/internal/filestorage"
	"github.com/Vla8islav/gophprofile/internal/repository"
	"github.com/Vla8islav/gophprofile/internal/worker"
	"go.uber.org/zap"
)

const consumerGroupID = "gophprofile-worker"

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
	avatarWorker := worker.New(db, fileStorage, lg)

	lg.Info("worker starting",
		zap.String("topic", currentConfig.KafkaTopic.Value),
		zap.String("group", consumerGroupID),
	)
	if err := consumer.Run(ctx, avatarWorker.HandleEvent); err != nil {
		lg.Fatal("consumer stopped", zap.Error(err))
	}
	lg.Info("worker stopped gracefully")
}
