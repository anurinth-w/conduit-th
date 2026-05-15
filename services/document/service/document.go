package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/anurinth-w/conduit-th/services/document/pdf"
	"github.com/anurinth-w/conduit-th/services/document/repository"
	"github.com/anurinth-w/conduit-th/services/document/storage"
	"github.com/google/uuid"
)

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrBundleNotFound   = errors.New("no document bundle configured for this job type")
)

// JobData — ข้อมูลงานที่ใช้ render template
type JobData struct {
	JobCode            string
	RefCode            string
	ReportNumber       string
	JobType            string
	Cause              string
	LocationText       string
	Subdistrict        string
	District           string
	Province           string
	PipeType           string
	PipeSizeMM         int
	SurfaceCondition   string
	SurfaceAreaSqm     float64
	WorkMethod         string
	ContactTechnician  string
	ContactCoordinator string
	CostMain           float64
	CostSurface        float64
	NotifiedAt         *time.Time
	StartedAt          *time.Time
	EndedAt            *time.Time
	CompanyName        string
	GeneratedAt        time.Time
}

type GenerateInput struct {
	JobID       uuid.UUID
	CompanyID   uuid.UUID
	JobType     string
	GeneratedBy uuid.UUID
	JobData     JobData
}

type CreateTemplateInput struct {
	CompanyID   uuid.UUID
	Name        string
	DocType     string
	HTMLContent string
}

type CreateBundleInput struct {
	CompanyID uuid.UUID
	JobType   string
	Name      string
	PageIDs   []uuid.UUID // template IDs เรียงตามลำดับหน้า
}

type DocumentService struct {
	repo       *repository.DocumentRepository
	pdfClient  *pdf.GotenbergClient
	store      *storage.MinIOStorage
}

func NewDocumentService(
	repo *repository.DocumentRepository,
	pdfClient *pdf.GotenbergClient,
	store *storage.MinIOStorage,
) *DocumentService {
	return &DocumentService{repo: repo, pdfClient: pdfClient, store: store}
}

// CreateTemplate — Dev upload template ใหม่
func (s *DocumentService) CreateTemplate(ctx context.Context, input CreateTemplateInput) (*repository.DocumentTemplate, error) {
	return s.repo.CreateTemplate(ctx, repository.CreateTemplateParams{
		CompanyID:   input.CompanyID,
		Name:        input.Name,
		DocType:     input.DocType,
		HTMLContent: input.HTMLContent,
	})
}

func (s *DocumentService) GetTemplate(ctx context.Context, id uuid.UUID) (*repository.DocumentTemplate, error) {
	t, err := s.repo.FindTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTemplateNotFound
	}
	return t, nil
}

func (s *DocumentService) ListTemplates(ctx context.Context, companyID uuid.UUID) ([]repository.DocumentTemplate, error) {
	return s.repo.ListTemplatesByCompany(ctx, companyID)
}

// CreateBundle — กำหนดชุดเอกสารต่อบริษัท × ประเภทงาน
func (s *DocumentService) CreateBundle(ctx context.Context, input CreateBundleInput) (*repository.DocumentBundle, error) {
	bundle, err := s.repo.CreateBundle(ctx, repository.CreateBundleParams{
		CompanyID: input.CompanyID,
		JobType:   input.JobType,
		Name:      input.Name,
	})
	if err != nil {
		return nil, err
	}

	// เพิ่มหน้าตามลำดับ
	for i, templateID := range input.PageIDs {
		if _, err := s.repo.AddPageToBundle(ctx, bundle.ID, templateID, i+1); err != nil {
			return nil, fmt.Errorf("add page %d: %w", i+1, err)
		}
	}

	return bundle, nil
}

func (s *DocumentService) ListDocumentsByJob(ctx context.Context, jobID uuid.UUID) ([]repository.Document, error) {
	return s.repo.ListDocumentsByJob(ctx, jobID)
}

// Generate — generate PDF จาก bundle ของบริษัท × job_type
func (s *DocumentService) Generate(ctx context.Context, input GenerateInput) (*repository.Document, error) {
	// 1. หา bundle
	bundle, err := s.repo.FindBundleByCompanyAndJobType(ctx, input.CompanyID, input.JobType)
	if err != nil {
		return nil, err
	}
	if bundle == nil {
		return nil, ErrBundleNotFound
	}

	// 2. สร้าง document record (status: generating)
	doc, err := s.repo.CreateDocument(ctx, repository.CreateDocumentParams{
		JobID:       input.JobID,
		GeneratedBy: input.GeneratedBy,
		DocType:     "repair_form_1", // ใช้ doc_type แรกของ bundle เป็น label
	})
	if err != nil {
		return nil, err
	}

	// 3. render HTML แต่ละหน้าแล้วรวมกัน
	var combinedHTML string
	input.JobData.GeneratedAt = time.Now()

	for _, page := range bundle.Pages {
		tmpl, err := s.repo.FindTemplateByID(ctx, page.TemplateID)
		if err != nil || tmpl == nil {
			_ = s.repo.UpdateDocumentFailed(ctx, doc.ID, fmt.Sprintf("template %s not found", page.TemplateID))
			return nil, fmt.Errorf("template not found for page %d", page.PageOrder)
		}

		rendered, err := renderHTML(tmpl.HTMLContent, input.JobData)
		if err != nil {
			_ = s.repo.UpdateDocumentFailed(ctx, doc.ID, err.Error())
			return nil, fmt.Errorf("render page %d: %w", page.PageOrder, err)
		}

		// เพิ่ม page break ระหว่างหน้า
		if combinedHTML != "" {
			combinedHTML += `<div style="page-break-after: always;"></div>`
		}
		combinedHTML += rendered
	}

	// 4. ส่งไป Gotenberg แปลงเป็น PDF
	pdfBytes, err := s.pdfClient.HTMLToPDF(ctx, combinedHTML)
	if err != nil {
		_ = s.repo.UpdateDocumentFailed(ctx, doc.ID, err.Error())
		return nil, fmt.Errorf("generate pdf: %w", err)
	}

	// 5. เก็บ PDF ใน MinIO
	s3Key := fmt.Sprintf("documents/%s/%s.pdf",
		input.JobID.String(),
		doc.ID.String(),
	)

	if err := s.store.Upload(ctx, s3Key, bytes.NewReader(pdfBytes), int64(len(pdfBytes)), "application/pdf"); err != nil {
		_ = s.repo.UpdateDocumentFailed(ctx, doc.ID, err.Error())
		return nil, fmt.Errorf("upload pdf: %w", err)
	}

	// 6. สร้าง presigned URL
	url, err := s.store.PresignURL(ctx, s3Key)
	if err != nil {
		_ = s.repo.UpdateDocumentFailed(ctx, doc.ID, err.Error())
		return nil, fmt.Errorf("presign url: %w", err)
	}

	// 7. อัปเดต document เป็น done
	if err := s.repo.UpdateDocumentDone(ctx, doc.ID, s3Key, url); err != nil {
		return nil, err
	}

	doc.S3Key = s3Key
	doc.URL = url
	doc.Status = "done"

	return doc, nil
}

// renderHTML — render HTML template ด้วย Go template engine
func renderHTML(htmlTemplate string, data JobData) (string, error) {
	tmpl, err := template.New("doc").Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
