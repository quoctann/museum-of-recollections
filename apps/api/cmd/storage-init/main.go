package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"museum-of-recollections/apps/api/internal/config"
	"museum-of-recollections/apps/api/internal/dependencies"
	"museum-of-recollections/apps/api/internal/storageinit"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err = run(ctx, os.Getenv, log)
	stop()
	if err != nil {
		log.Error("storage initialization stopped", zap.String("reason", err.Error()))
	}
	_ = log.Sync()
	if err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, get func(string) string, log *zap.Logger) error {
	cfg, err := config.LoadStorage(get)
	if err != nil {
		return err
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""), Secure: cfg.UseSSL, Region: "us-east-1"})
	if err != nil {
		return errors.New("storage configuration invalid")
	}
	if err := storageinit.Ensure(ctx, dependencies.Provisioner{Client: client}, cfg.Bucket, 30*time.Second); err != nil {
		return err
	}
	log.Info("storage initialized")
	return nil
}
