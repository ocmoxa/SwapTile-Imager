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

// Storage implements storage.FileStorage for S3.
type Storage struct {
	client *minio.Client
	cfg    config.S3
}

// NewStorage create new S3 file storage that implements
// storage.FileStorage interface.
func NewStorage(cfg config.S3) (*Storage, error) {
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

	return &Storage{
		client: client,
		cfg:    cfg,
	}, nil
}

// Health checks that bucket found.
func (s Storage) Health(ctx context.Context) (err error) {
	ok, err := s.client.BucketExists(s.cfg.Bucket)
	switch {
	case err != nil:
		return fmt.Errorf("s3: checking bucket: %w", err)
	case !ok:
		return imerrors.NewNotFoundError(imerrors.Error("s3: bucket not found"))
	default:
		return nil
	}
}

// Get an image by ID from S3.
func (s Storage) Get(
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

// Upload an image to S3.
func (s Storage) Upload(ctx context.Context, im imager.ImageMeta, r io.Reader) (err error) {
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

// Delete S3 image by ID.
func (s Storage) Delete(ctx context.Context, id string) (err error) {
	err = s.client.RemoveObject(s.cfg.Bucket, id)
	if err != nil {
		return fmt.Errorf("s3 removing object: %w", err)
	}

	return nil
}
