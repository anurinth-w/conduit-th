package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobPhoto struct {
	ID         uuid.UUID
	JobID      uuid.UUID
	UploadedBy uuid.UUID
	S3Key      string
	URL        string
	Stage      string
	IsSelected bool
	Caption    string
	UploadedAt time.Time
}

type CreatePhotoParams struct {
	JobID      uuid.UUID
	UploadedBy uuid.UUID
	S3Key      string
	URL        string
	Stage      string
	Caption    string
}

type PhotoRepository struct {
	db *pgxpool.Pool
}

func NewPhotoRepository(db *pgxpool.Pool) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) Create(ctx context.Context, p CreatePhotoParams) (*JobPhoto, error) {
	ph := &JobPhoto{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO job_photos (job_id, uploaded_by, s3_key, url, stage, caption)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, job_id, uploaded_by, s3_key, url, stage,
		           is_selected, coalesce(caption,''), uploaded_at`,
		p.JobID, p.UploadedBy, p.S3Key, p.URL, p.Stage, p.Caption,
	).Scan(
		&ph.ID, &ph.JobID, &ph.UploadedBy, &ph.S3Key, &ph.URL,
		&ph.Stage, &ph.IsSelected, &ph.Caption, &ph.UploadedAt,
	)
	return ph, err
}

func (r *PhotoRepository) FindByID(ctx context.Context, id uuid.UUID) (*JobPhoto, error) {
	ph := &JobPhoto{}
	err := r.db.QueryRow(ctx,
		`SELECT id, job_id, uploaded_by, s3_key, url, stage,
		        is_selected, coalesce(caption,''), uploaded_at
		 FROM job_photos WHERE id = $1`,
		id,
	).Scan(
		&ph.ID, &ph.JobID, &ph.UploadedBy, &ph.S3Key, &ph.URL,
		&ph.Stage, &ph.IsSelected, &ph.Caption, &ph.UploadedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return ph, err
}

func (r *PhotoRepository) ListByJob(ctx context.Context, jobID uuid.UUID, stage string) ([]JobPhoto, error) {
	query := `SELECT id, job_id, uploaded_by, s3_key, url, stage,
		             is_selected, coalesce(caption,''), uploaded_at
		      FROM job_photos WHERE job_id = $1`
	args := []interface{}{jobID}

	if stage != "" {
		args = append(args, stage)
		query += ` AND stage = $2`
	}
	query += ` ORDER BY uploaded_at ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []JobPhoto
	for rows.Next() {
		var ph JobPhoto
		if err := rows.Scan(
			&ph.ID, &ph.JobID, &ph.UploadedBy, &ph.S3Key, &ph.URL,
			&ph.Stage, &ph.IsSelected, &ph.Caption, &ph.UploadedAt,
		); err != nil {
			return nil, err
		}
		photos = append(photos, ph)
	}
	return photos, nil
}

func (r *PhotoRepository) UpdateURL(ctx context.Context, id uuid.UUID, url string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE job_photos SET url=$1 WHERE id=$2`,
		url, id,
	)
	return err
}

func (r *PhotoRepository) SetSelected(ctx context.Context, id uuid.UUID, selected bool) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE job_photos SET is_selected=$1 WHERE id=$2`,
		selected, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PhotoRepository) Delete(ctx context.Context, id uuid.UUID) (string, error) {
	var s3Key string
	err := r.db.QueryRow(ctx,
		`DELETE FROM job_photos WHERE id=$1 RETURNING s3_key`,
		id,
	).Scan(&s3Key)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return s3Key, err
}
