package s3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

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
		return nil, fmt.Errorf("initializing minio: %w", err)
	}

	exists, err := client.BucketExists(cfg.Bucket)
	switch {
	case err != nil:
		return nil, fmt.Errorf("checking bucket: %w", err)
	case !exists:
		err = client.MakeBucket(cfg.Bucket, cfg.Location)
		if err != nil {
			return nil, fmt.Errorf("making bucket: %w", err)
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
) (r io.ReadCloser, err error) {
	const errCodeNotFound = "NoSuchKey"

	obj, err := s.client.GetObject(s.cfg.Bucket, id, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting object: %w", err)
	}

	_, err = obj.Stat()
	if err != nil {
		var errResp minio.ErrorResponse
		if errors.As(err, &errResp) && errResp.Code == errCodeNotFound {
			return nil, imerrors.NotFoundError{Err: err}
		}

		return nil, fmt.Errorf("getting stat: %w", err)
	}

	return obj, nil
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
		return fmt.Errorf("putting object: %w", err)
	}

	return nil
}

func (s S3Storage) Delete(ctx context.Context, id string) (err error) {
	err = s.client.RemoveObject(s.cfg.Bucket, id)
	if err != nil {
		return fmt.Errorf("removing object: %w", err)
	}

	return nil
}
