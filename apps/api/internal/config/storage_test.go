package config

import "testing"

func TestStorageOnlyConfig(t *testing.T) {
	env := environment()
	delete(env, "DATABASE_URL")
	env["API_PORT"] = "invalid"
	c, err := LoadStorage(func(k string) string { return env[k] })
	if err != nil || c.Bucket != "recollections" {
		t.Fatalf("storage-only config failed: %v", err)
	}
}
