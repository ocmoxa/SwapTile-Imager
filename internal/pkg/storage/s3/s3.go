package s3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage"

	"github.com/minio/minio-go"
)

// S3Storage implements storage.FileStorage for S3.
type S3Storage struct {
	client *minio.Client
	cfg    config.S3
}

// NewS3Storage create new S3 file storage that implements
// storage.FileStorage interface.
func NewS3Storage(cfg config.S3) (*S3Storage, error) {
	client, err := minio.New(
		cfg.Endpoint,
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		cfg.Secure,
	)
	if err != nil {
		return nil, fmt.Errorf("s3 initializing minio: %w", err)
	}

	exists, err := client.BucketExists(cfg.Bucket)
	switch {
	case err != nil:
		return nil, fmt.Errorf("s3 checking bucket: %w", err)
	case !exists:
		err = client.MakeBucket(cfg.Bucket, cfg.Location)
		if err != nil {
			return nil, fmt.Errorf("s3 making bucket: %w", err)
		}
	}

	return &S3Storage{
		client: client,
		cfg:    cfg,
	}, nil
}

func (s S3Storage) Get(
	ctx context.Context,
	id string,
) (f storage.File, err error) {
	const errCodeNotFound = "NoSuchKey"

	obj, err := s.client.GetObject(s.cfg.Bucket, id, minio.GetObjectOptions{})
	if err != nil {
		return storage.File{}, fmt.Errorf("s3 getting object: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		var errResp minio.ErrorResponse
		if errors.As(err, &errResp) && errResp.Code == errCodeNotFound {
			return storage.File{}, imerrors.NewNotFoundError(err)
		}

		return storage.File{}, fmt.Errorf("s3 getting stat: %w", err)
	}

	return storage.File{
		ReadCloser:  obj,
		ContentType: stat.ContentType,
	}, nil
}

func (s S3Storage) Upload(ctx context.Context, im imager.ImageMeta, r io.Reader) (err error) {
	opts := minio.PutObjectOptions{
		ContentType: im.MIMEType,
	}

	_, err = s.client.PutObject(
		s.cfg.Bucket,
		im.ID,
		r,
		im.Size,
		opts,
	)
	if err != nil {
		return fmt.Errorf("s3 putting object: %w", err)
	}

	return nil
}

func (s S3Storage) Delete(ctx context.Context, id string) (err error) {
	err = s.client.RemoveObject(s.cfg.Bucket, id)
	if err != nil {
		return fmt.Errorf("s3 removing object: %w", err)
	}

	return nil
}
