package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Models ---

type DocumentTemplate struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Name        string
	DocType     string
	HTMLContent string
	Version     int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DocumentBundle struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	JobType   string
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DocumentBundlePage struct {
	ID         uuid.UUID
	BundleID   uuid.UUID
	TemplateID uuid.UUID
	PageOrder  int
}

type BundleWithPages struct {
	Bundle DocumentBundle
	Pages  []DocumentBundlePage
}

type Document struct {
	ID          uuid.UUID
	JobID       uuid.UUID
	GeneratedBy uuid.UUID
	DocType     string
	S3Key       string
	URL         string
	Status      string
	ErrorMsg    string
	GeneratedAt time.Time
}

// --- Params ---

type CreateTemplateParams struct {
	CompanyID   uuid.UUID
	Name        string
	DocType     string
	HTMLContent string
}

type CreateBundleParams struct {
	CompanyID uuid.UUID
	JobType   string
	Name      string
}

type CreateDocumentParams struct {
	JobID       uuid.UUID
	GeneratedBy uuid.UUID
	DocType     string
}

// --- Repository ---

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// --- Templates ---

func (r *DocumentRepository) CreateTemplate(ctx context.Context, p CreateTemplateParams) (*DocumentTemplate, error) {
	// หา version ล่าสุดของ doc_type นี้
	var lastVersion int
	err := r.db.QueryRow(ctx,
		`SELECT coalesce(max(version), 0) FROM document_templates
		 WHERE company_id=$1 AND doc_type=$2`,
		p.CompanyID, p.DocType,
	).Scan(&lastVersion)
	if err != nil {
		return nil, err
	}

	t := &DocumentTemplate{}
	err = r.db.QueryRow(ctx,
		`INSERT INTO document_templates (company_id, name, doc_type, html_content, version)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, company_id, name, doc_type, html_content, version,
		           is_active, created_at, updated_at`,
		p.CompanyID, p.Name, p.DocType, p.HTMLContent, lastVersion+1,
	).Scan(
		&t.ID, &t.CompanyID, &t.Name, &t.DocType, &t.HTMLContent,
		&t.Version, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

func (r *DocumentRepository) FindTemplateByID(ctx context.Context, id uuid.UUID) (*DocumentTemplate, error) {
	t := &DocumentTemplate{}
	err := r.db.QueryRow(ctx,
		`SELECT id, company_id, name, doc_type, html_content, version,
		        is_active, created_at, updated_at
		 FROM document_templates WHERE id=$1`,
		id,
	).Scan(
		&t.ID, &t.CompanyID, &t.Name, &t.DocType, &t.HTMLContent,
		&t.Version, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (r *DocumentRepository) ListTemplatesByCompany(ctx context.Context, companyID uuid.UUID) ([]DocumentTemplate, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, company_id, name, doc_type, html_content, version,
		        is_active, created_at, updated_at
		 FROM document_templates
		 WHERE company_id=$1 AND is_active=TRUE
		 ORDER BY doc_type, version DESC`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []DocumentTemplate
	for rows.Next() {
		var t DocumentTemplate
		if err := rows.Scan(
			&t.ID, &t.CompanyID, &t.Name, &t.DocType, &t.HTMLContent,
			&t.Version, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

// --- Bundles ---

func (r *DocumentRepository) CreateBundle(ctx context.Context, p CreateBundleParams) (*DocumentBundle, error) {
	b := &DocumentBundle{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO document_bundles (company_id, job_type, name)
		 VALUES ($1, $2, $3)
		 RETURNING id, company_id, job_type, name, is_active, created_at, updated_at`,
		p.CompanyID, p.JobType, p.Name,
	).Scan(
		&b.ID, &b.CompanyID, &b.JobType, &b.Name,
		&b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	return b, err
}

func (r *DocumentRepository) AddPageToBundle(ctx context.Context, bundleID, templateID uuid.UUID, pageOrder int) (*DocumentBundlePage, error) {
	p := &DocumentBundlePage{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO document_bundle_pages (bundle_id, template_id, page_order)
		 VALUES ($1, $2, $3)
		 RETURNING id, bundle_id, template_id, page_order`,
		bundleID, templateID, pageOrder,
	).Scan(&p.ID, &p.BundleID, &p.TemplateID, &p.PageOrder)
	return p, err
}

func (r *DocumentRepository) FindBundleByCompanyAndJobType(ctx context.Context, companyID uuid.UUID, jobType string) (*BundleWithPages, error) {
	b := &DocumentBundle{}
	err := r.db.QueryRow(ctx,
		`SELECT id, company_id, job_type, name, is_active, created_at, updated_at
		 FROM document_bundles
		 WHERE company_id=$1 AND job_type=$2 AND is_active=TRUE`,
		companyID, jobType,
	).Scan(
		&b.ID, &b.CompanyID, &b.JobType, &b.Name,
		&b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// ดึง pages เรียงตาม page_order
	rows, err := r.db.Query(ctx,
		`SELECT id, bundle_id, template_id, page_order
		 FROM document_bundle_pages
		 WHERE bundle_id=$1
		 ORDER BY page_order ASC`,
		b.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []DocumentBundlePage
	for rows.Next() {
		var p DocumentBundlePage
		if err := rows.Scan(&p.ID, &p.BundleID, &p.TemplateID, &p.PageOrder); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	return &BundleWithPages{Bundle: *b, Pages: pages}, nil
}

// --- Documents (output) ---

func (r *DocumentRepository) CreateDocument(ctx context.Context, p CreateDocumentParams) (*Document, error) {
	d := &Document{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO documents (job_id, generated_by, doc_type, status)
		 VALUES ($1, $2, $3, 'generating')
		 RETURNING id, job_id, generated_by, doc_type,
		           coalesce(s3_key,''), coalesce(url,''),
		           status, coalesce(error_msg,''), generated_at`,
		p.JobID, p.GeneratedBy, p.DocType,
	).Scan(
		&d.ID, &d.JobID, &d.GeneratedBy, &d.DocType,
		&d.S3Key, &d.URL, &d.Status, &d.ErrorMsg, &d.GeneratedAt,
	)
	return d, err
}

func (r *DocumentRepository) UpdateDocumentDone(ctx context.Context, id uuid.UUID, s3Key, url string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE documents SET status='done', s3_key=$1, url=$2 WHERE id=$3`,
		s3Key, url, id,
	)
	return err
}

func (r *DocumentRepository) UpdateDocumentFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE documents SET status='failed', error_msg=$1 WHERE id=$2`,
		errMsg, id,
	)
	return err
}

func (r *DocumentRepository) ListDocumentsByJob(ctx context.Context, jobID uuid.UUID) ([]Document, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, job_id, generated_by, doc_type,
		        coalesce(s3_key,''), coalesce(url,''),
		        status, coalesce(error_msg,''), generated_at
		 FROM documents WHERE job_id=$1
		 ORDER BY generated_at DESC`,
		jobID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(
			&d.ID, &d.JobID, &d.GeneratedBy, &d.DocType,
			&d.S3Key, &d.URL, &d.Status, &d.ErrorMsg, &d.GeneratedAt,
		); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}
