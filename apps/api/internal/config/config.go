package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	StorageConfig
	Port            int
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

type StorageConfig struct {
	Endpoint, AccessKey, SecretKey, Bucket string
	UseSSL                                 bool
}

// Load validates configuration without ever including input values in errors.
func Load(get func(string) string) (Config, error) {
	storage, err := LoadStorage(get)
	if err != nil {
		return Config{}, err
	}
	c := Config{Port: 8080, StorageConfig: storage, ShutdownTimeout: 10 * time.Second}
	c.DatabaseURL = get("DATABASE_URL")
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	u, err := url.Parse(c.DatabaseURL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return Config{}, fmt.Errorf("DATABASE_URL must be a PostgreSQL URL")
	}
	if v := get("API_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 || p > 65535 {
			return Config{}, fmt.Errorf("API_PORT is invalid")
		}
		c.Port = p
	}
	if v := get("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 || d > time.Minute {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT must be between 0 and 1m")
		}
		c.ShutdownTimeout = d
	}
	return c, nil
}

// LoadStorage validates only S3 configuration; provisioning does not need a database.
func LoadStorage(get func(string) string) (StorageConfig, error) {
	c := StorageConfig{Endpoint: get("S3_ENDPOINT"), AccessKey: get("S3_ACCESS_KEY"), SecretKey: get("S3_SECRET_KEY"), Bucket: "recollections"}
	for _, key := range []string{"S3_ENDPOINT", "S3_ACCESS_KEY", "S3_SECRET_KEY"} {
		if strings.TrimSpace(get(key)) == "" {
			return StorageConfig{}, fmt.Errorf("%s is required", key)
		}
	}
	host, port, err := net.SplitHostPort(c.Endpoint)
	p, e := strconv.Atoi(port)
	if err != nil || e != nil || host == "" || strings.ContainsAny(host, "/?#@ \t\r\n") || p < 1 || p > 65535 {
		return StorageConfig{}, fmt.Errorf("S3_ENDPOINT must be host:port")
	}
	if v := get("S3_USE_SSL"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return StorageConfig{}, fmt.Errorf("S3_USE_SSL is invalid")
		}
		c.UseSSL = b
	}
	if v := get("S3_BUCKET"); v != "" {
		c.Bucket = v
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`).MatchString(c.Bucket) || strings.Contains(c.Bucket, "..") || net.ParseIP(c.Bucket) != nil {
		return StorageConfig{}, fmt.Errorf("S3_BUCKET is invalid")
	}
	return c, nil
}
