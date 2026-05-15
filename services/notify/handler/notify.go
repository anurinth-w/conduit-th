package handler

import (
	"errors"
	"net/http"

	"github.com/anurinth-w/conduit-th/services/notify/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotifyHandler struct {
	svc *service.NotifyService
}

func NewNotifyHandler(svc *service.NotifyService) *NotifyHandler {
	return &NotifyHandler{svc: svc}
}

func (h *NotifyHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type notifyRequest struct {
	Event     string `json:"event"      binding:"required"`
	JobID     string `json:"job_id"     binding:"required"`
	CompanyID string `json:"company_id" binding:"required"`
	Message   string `json:"message"    binding:"required"`
}

// Send POST /v1/notify
func (h *NotifyHandler) Send(c *gin.Context) {
	var req notifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job_id"})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	err = h.svc.Notify(c.Request.Context(), service.NotifyInput{
		Event:     req.Event,
		JobID:     jobID,
		CompanyID: companyID,
		Message:   req.Message,
	})
	if err != nil {
		if errors.Is(err, service.ErrLineNotConfigured) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "LINE not configured for this company"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notified"})
}
