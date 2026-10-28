package main

import (
	"context"
	"go.uber.org/zap"
	"net"
	"strconv"
	"testing"
)

func TestRunInvalidConfiguration(t *testing.T) {
	if err := run(context.Background(), func(string) string { return "" }, zap.NewNop()); err == nil {
		t.Fatal("expected invalid config")
	}
}
func TestRunCanceled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	env := map[string]string{"API_PORT": strconv.Itoa(port), "DATABASE_URL": "postgres://user:pass@localhost:5432/museum", "S3_ENDPOINT": "localhost:9000", "S3_ACCESS_KEY": "test", "S3_SECRET_KEY": "test"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := run(ctx, func(k string) string { return env[k] }, zap.NewNop()); err != nil {
		t.Fatal(err)
	}
}
func TestRunBindFailure(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	env := map[string]string{"API_PORT": strconv.Itoa(ln.Addr().(*net.TCPAddr).Port), "DATABASE_URL": "postgres://user:pass@localhost:5432/museum", "S3_ENDPOINT": "localhost:9000", "S3_ACCESS_KEY": "test", "S3_SECRET_KEY": "test"}
	if err := run(context.Background(), func(k string) string { return env[k] }, zap.NewNop()); err == nil {
		t.Fatal("expected bind error")
	}
}
