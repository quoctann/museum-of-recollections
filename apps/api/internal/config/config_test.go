package config

import (
	"strings"
	"testing"
)

func environment() map[string]string {
	return map[string]string{"DATABASE_URL": "postgres://user:pass@localhost:5432/museum", "S3_ENDPOINT": "localhost:9000", "S3_ACCESS_KEY": "test-key", "S3_SECRET_KEY": "test-secret"}
}
func TestLoadDefaults(t *testing.T) {
	env := environment()
	c, err := Load(func(k string) string { return env[k] })
	if err != nil || c.Port != 8080 || c.Bucket != "recollections" || c.UseSSL || c.ShutdownTimeout.Seconds() != 10 {
		t.Fatalf("defaults invalid: %v", err)
	}
}
func TestLoadValidOverrides(t *testing.T) {
	env := environment()
	env["API_PORT"] = "9090"
	env["S3_USE_SSL"] = "true"
	env["S3_BUCKET"] = "my-bucket"
	env["SHUTDOWN_TIMEOUT"] = "2s"
	c, err := Load(func(k string) string { return env[k] })
	if err != nil || c.Port != 9090 || !c.UseSSL || c.Bucket != "my-bucket" || c.ShutdownTimeout.Seconds() != 2 {
		t.Fatalf("invalid override: %v", err)
	}
}
func TestLoadRejectsInvalid(t *testing.T) {
	for _, tc := range []struct{ k, v string }{{"DATABASE_URL", ""}, {"S3_ENDPOINT", ""}, {"S3_ACCESS_KEY", ""}, {"S3_SECRET_KEY", ""}, {"DATABASE_URL", "https://secret.invalid"}, {"DATABASE_URL", "postgres://"}, {"API_PORT", "0"}, {"API_PORT", "65536"}, {"API_PORT", "oops"}, {"S3_ENDPOINT", "http://localhost:9000"}, {"S3_ENDPOINT", "localhost"}, {"S3_ENDPOINT", "localhost:0"}, {"S3_ENDPOINT", "localhost:99999"}, {"S3_USE_SSL", "maybe"}, {"S3_BUCKET", "Bad Bucket"}, {"SHUTDOWN_TIMEOUT", "0s"}, {"SHUTDOWN_TIMEOUT", "oops"}, {"SHUTDOWN_TIMEOUT", "2h"}} {
		t.Run(tc.k+tc.v, func(t *testing.T) {
			env := environment()
			env[tc.k] = tc.v
			_, err := Load(func(k string) string { return env[k] })
			if err == nil {
				t.Fatal("expected validation failure")
			}
			if strings.Contains(err.Error(), "secret.invalid") {
				t.Fatal("input leaked")
			}
		})
	}
}
