package dependencies

import (
	"context"

	"github.com/minio/minio-go/v7"
	"museum-of-recollections/apps/api/internal/storageinit"
)

type Provisioner struct{ Client *minio.Client }

func (p Provisioner) Exists(ctx context.Context, bucket string) (bool, error) {
	return p.Client.BucketExists(ctx, bucket)
}
func (p Provisioner) Create(ctx context.Context, bucket string) error {
	err := p.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"})
	if minio.ToErrorResponse(err).Code == "BucketAlreadyOwnedByYou" {
		return storageinit.ErrAlreadyOwned
	}
	return err
}
