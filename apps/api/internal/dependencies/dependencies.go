package dependencies

import (
	"context"
	"errors"

	"github.com/minio/minio-go/v7"
)

type Pinger interface{ PingContext(context.Context) error }
type Database struct{ DB Pinger }

func (d Database) Check(ctx context.Context) error { return d.DB.PingContext(ctx) }

type BucketClient interface {
	BucketExists(context.Context, string) (bool, error)
}

// Storage checks the explicitly provisioned bucket; it never creates infrastructure.
type Storage struct {
	Client BucketClient
	Bucket string
}

func (s Storage) Check(ctx context.Context) error {
	exists, err := s.Client.BucketExists(ctx, s.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("storage unavailable")
	}
	return nil
}

var _ BucketClient = (*minio.Client)(nil)
