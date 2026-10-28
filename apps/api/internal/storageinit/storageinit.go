// Package storageinit provisions a private bucket independently of API startup.
package storageinit

import (
	"context"
	"errors"
	"time"
)

var ErrAlreadyOwned = errors.New("bucket already owned")

type Buckets interface {
	Exists(context.Context, string) (bool, error)
	Create(context.Context, string) error
}

// Ensure never sets a bucket policy. Existing buckets are left unchanged.
func Ensure(ctx context.Context, buckets Buckets, bucket string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	exists, err := buckets.Exists(ctx, bucket)
	if err != nil {
		return errors.New("storage initialization check failed")
	}
	if exists {
		return nil
	}
	if err := buckets.Create(ctx, bucket); err != nil && !errors.Is(err, ErrAlreadyOwned) {
		return errors.New("storage initialization failed")
	}
	return nil
}
