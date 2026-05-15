package handler

import (
	"errors"
	"net/http"

	"github.com/anurinth-w/conduit-th/services/media/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MediaHandler struct {
	svc *service.MediaService
}

func NewMediaHandler(svc *service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Upload POST /v1/jobs/:job_id/photos
func (h *MediaHandler) Upload(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	uploadedBy, err := uuid.Parse(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid X-User-ID header"})
		return
	}

	stage := c.PostForm("stage")
	caption := c.PostForm("caption")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	photo, err := h.svc.Upload(c.Request.Context(), service.UploadPhotoInput{
		JobID:      jobID,
		UploadedBy: uploadedBy,
		Stage:      stage,
		Caption:    caption,
		File:       file,
		FileHeader: header,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStage):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidFileType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		}
		return
	}

	c.JSON(http.StatusCreated, photo)
}

// ListByJob GET /v1/jobs/:job_id/photos
func (h *MediaHandler) ListByJob(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	stage := c.Query("stage") // optional filter

	photos, err := h.svc.ListByJob(c.Request.Context(), jobID, stage)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStage) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, photos)
}

// GetPresignURL GET /v1/photos/:id/url
func (h *MediaHandler) GetPresignURL(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo id"})
		return
	}

	url, err := h.svc.GetPresignURL(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPhotoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "photo not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// SetSelected PATCH /v1/photos/:id/select
func (h *MediaHandler) SetSelected(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo id"})
		return
	}

	var req struct {
		Selected bool `json:"selected"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.SetSelected(c.Request.Context(), id, req.Selected); err != nil {
		if errors.Is(err, service.ErrPhotoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "photo not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Delete DELETE /v1/photos/:id
func (h *MediaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrPhotoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "photo not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "photo deleted"})
}

// RefreshURLs POST /v1/jobs/:job_id/photos/refresh-urls
func (h *MediaHandler) RefreshURLs(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	photos, err := h.svc.RefreshURLs(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, photos)
}
