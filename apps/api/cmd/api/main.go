package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"museum-of-recollections/apps/api/internal/config"
	"museum-of-recollections/apps/api/internal/dependencies"
	"museum-of-recollections/apps/api/internal/health"
	"museum-of-recollections/apps/api/internal/httpapi"
	"museum-of-recollections/apps/api/internal/server"
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
		log.Error("API stopped", zap.String("reason", err.Error()))
	}
	_ = log.Sync()
	if err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, get func(string) string, log *zap.Logger) error {
	cfg, err := config.Load(get)
	if err != nil {
		return err
	}
	db, err := sqlx.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return errors.New("database configuration invalid")
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	storage, err := minio.New(cfg.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""), Secure: cfg.UseSSL, Region: "us-east-1"})
	if err != nil {
		return errors.New("storage configuration invalid")
	}
	readiness := health.New(2*time.Second, dependencies.Database{DB: db}, dependencies.Storage{Client: storage, Bucket: cfg.Bucket})
	gin.SetMode(gin.ReleaseMode)
	srv := &http.Server{Handler: httpapi.New(readiness, log), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10}
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		return errors.New("HTTP listener unavailable")
	}
	log.Info("API listening", zap.Int("port", cfg.Port))
	if err := server.Serve(ctx, srv, ln, cfg.ShutdownTimeout); err != nil {
		return errors.New("HTTP server stopped unexpectedly")
	}
	return nil
}
