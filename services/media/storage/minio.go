package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client        *minio.Client
	bucket        string
	presignExpiry time.Duration
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool, presignExpirySec int) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	return &MinIOStorage{
		client:        client,
		bucket:        bucket,
		presignExpiry: time.Duration(presignExpirySec) * time.Second,
	}, nil
}

// Upload อัปโหลดไฟล์ไปยัง MinIO คืน s3Key
func (s *MinIOStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// PresignURL สร้าง URL ชั่วคราวสำหรับดูรูป
func (s *MinIOStorage) PresignURL(ctx context.Context, key string) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, s.presignExpiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Delete ลบไฟล์จาก MinIO
func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
