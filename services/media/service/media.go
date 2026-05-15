package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/anurinth-w/conduit-th/services/media/repository"
	"github.com/anurinth-w/conduit-th/services/media/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrPhotoNotFound    = errors.New("photo not found")
	ErrInvalidStage     = errors.New("invalid stage: must be before, during, or after")
	ErrInvalidFileType  = errors.New("invalid file type: only jpg, jpeg, png, webp allowed")
)

var allowedExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

var validStages = map[string]bool{
	"before": true,
	"during": true,
	"after":  true,
}

type UploadPhotoInput struct {
	JobID      uuid.UUID
	UploadedBy uuid.UUID
	Stage      string
	Caption    string
	File       multipart.File
	FileHeader *multipart.FileHeader
}

type MediaService struct {
	repo    *repository.PhotoRepository
	storage *storage.MinIOStorage
}

func NewMediaService(repo *repository.PhotoRepository, storage *storage.MinIOStorage) *MediaService {
	return &MediaService{repo: repo, storage: storage}
}

func (s *MediaService) Upload(ctx context.Context, input UploadPhotoInput) (*repository.JobPhoto, error) {
	// validate stage
	if !validStages[input.Stage] {
		return nil, ErrInvalidStage
	}

	// validate file type
	ext := strings.ToLower(filepath.Ext(input.FileHeader.Filename))
	contentType, ok := allowedExtensions[ext]
	if !ok {
		return nil, ErrInvalidFileType
	}

	// สร้าง unique key: photos/{job_id}/{timestamp}_{uuid}{ext}
	s3Key := fmt.Sprintf("photos/%s/%d_%s%s",
		input.JobID.String(),
		time.Now().UnixMilli(),
		uuid.New().String()[:8],
		ext,
	)

	// อัปโหลดไปยัง MinIO
	if err := s.storage.Upload(ctx, s3Key, input.File, input.FileHeader.Size, contentType); err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	// สร้าง presigned URL
	presignURL, err := s.storage.PresignURL(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("generate presign url: %w", err)
	}

	// บันทึกลง database
	return s.repo.Create(ctx, repository.CreatePhotoParams{
		JobID:      input.JobID,
		UploadedBy: input.UploadedBy,
		S3Key:      s3Key,
		URL:        presignURL,
		Stage:      input.Stage,
		Caption:    input.Caption,
	})
}

func (s *MediaService) ListByJob(ctx context.Context, jobID uuid.UUID, stage string) ([]repository.JobPhoto, error) {
	if stage != "" && !validStages[stage] {
		return nil, ErrInvalidStage
	}
	return s.repo.ListByJob(ctx, jobID, stage)
}

func (s *MediaService) GetPresignURL(ctx context.Context, id uuid.UUID) (string, error) {
	ph, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if ph == nil {
		return "", ErrPhotoNotFound
	}

	// สร้าง presigned URL ใหม่ (ของเก่าอาจหมดอายุ)
	url, err := s.storage.PresignURL(ctx, ph.S3Key)
	if err != nil {
		return "", fmt.Errorf("generate presign url: %w", err)
	}

	// อัปเดต URL ใน DB
	_ = s.repo.UpdateURL(ctx, id, url)

	return url, nil
}

func (s *MediaService) SetSelected(ctx context.Context, id uuid.UUID, selected bool) error {
	err := s.repo.SetSelected(ctx, id, selected)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPhotoNotFound
	}
	return err
}

func (s *MediaService) Delete(ctx context.Context, id uuid.UUID) error {
	s3Key, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if s3Key == "" {
		return ErrPhotoNotFound
	}

	// ลบออกจาก MinIO ด้วย
	return s.storage.Delete(ctx, s3Key)
}

// RefreshURLs รีเฟรช presigned URL ทั้งหมดของงาน (เรียกก่อน generate PDF)
func (s *MediaService) RefreshURLs(ctx context.Context, jobID uuid.UUID) ([]repository.JobPhoto, error) {
	photos, err := s.repo.ListByJob(ctx, jobID, "")
	if err != nil {
		return nil, err
	}

	for _, ph := range photos {
		url, err := s.storage.PresignURL(ctx, ph.S3Key)
		if err != nil {
			continue
		}
		_ = s.repo.UpdateURL(ctx, ph.ID, url)
		ph.URL = url
	}

	return photos, nil
}

// ensure io.Reader is used (suppress unused import)
var _ io.Reader = nil
