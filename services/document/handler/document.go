package handler

import (
	"errors"
	"net/http"

	"github.com/anurinth-w/conduit-th/services/document/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	svc *service.DocumentService
}

func NewDocumentHandler(svc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func (h *DocumentHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// --- Templates ---

type createTemplateRequest struct {
	Name        string `json:"name"         binding:"required"`
	DocType     string `json:"doc_type"     binding:"required"`
	HTMLContent string `json:"html_content" binding:"required"`
}

func (h *DocumentHandler) CreateTemplate(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	var req createTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.svc.CreateTemplate(c.Request.Context(), service.CreateTemplateInput{
		CompanyID:   companyID,
		Name:        req.Name,
		DocType:     req.DocType,
		HTMLContent: req.HTMLContent,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

func (h *DocumentHandler) ListTemplates(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	templates, err := h.svc.ListTemplates(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (h *DocumentHandler) GetTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	t, err := h.svc.GetTemplate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, t)
}

// --- Bundles ---

type createBundleRequest struct {
	JobType string      `json:"job_type" binding:"required"`
	Name    string      `json:"name"     binding:"required"`
	PageIDs []uuid.UUID `json:"page_ids" binding:"required,min=1"` // template IDs เรียงตามลำดับ
}

func (h *DocumentHandler) CreateBundle(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	var req createBundleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bundle, err := h.svc.CreateBundle(c.Request.Context(), service.CreateBundleInput{
		CompanyID: companyID,
		JobType:   req.JobType,
		Name:      req.Name,
		PageIDs:   req.PageIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bundle)
}

// --- Generate ---

type generateRequest struct {
	CompanyID string              `json:"company_id" binding:"required"`
	JobType   string              `json:"job_type"   binding:"required"`
	JobData   service.JobData     `json:"job_data"   binding:"required"`
}

func (h *DocumentHandler) Generate(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	generatedBy, err := uuid.Parse(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid X-User-ID header"})
		return
	}

	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	doc, err := h.svc.Generate(c.Request.Context(), service.GenerateInput{
		JobID:       jobID,
		CompanyID:   companyID,
		JobType:     req.JobType,
		GeneratedBy: generatedBy,
		JobData:     req.JobData,
	})
	if err != nil {
		if errors.Is(err, service.ErrBundleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) ListByJob(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	docs, err := h.svc.ListDocumentsByJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, docs)
}
