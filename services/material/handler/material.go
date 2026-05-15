package handler

import (
	"errors"
	"net/http"

	"github.com/anurinth-w/conduit-th/services/material/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MaterialHandler struct {
	svc *service.MaterialService
}

func NewMaterialHandler(svc *service.MaterialService) *MaterialHandler {
	return &MaterialHandler{svc: svc}
}

type createMaterialRequest struct {
	Code      string  `json:"code"       binding:"required"`
	Name      string  `json:"name"       binding:"required"`
	Unit      string  `json:"unit"       binding:"required"`
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0"`
	LaborCost float64 `json:"labor_cost"`
	KFactor   string  `json:"k_factor"`
}

type updateMaterialRequest struct {
	Name      string  `json:"name"       binding:"required"`
	Unit      string  `json:"unit"       binding:"required"`
	LaborCost float64 `json:"labor_cost"`
	KFactor   string  `json:"k_factor"`
}

type updatePriceRequest struct {
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0"`
}

func (h *MaterialHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *MaterialHandler) Create(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	createdBy, err := uuid.Parse(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid X-User-ID header"})
		return
	}
	_ = createdBy // materials ไม่มี created_by column แต่เก็บไว้สำหรับ audit log

	var req createMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.svc.Create(c.Request.Context(), service.CreateMaterialInput{
		CompanyID: companyID,
		Code:      req.Code,
		Name:      req.Name,
		Unit:      req.Unit,
		UnitPrice: req.UnitPrice,
		LaborCost: req.LaborCost,
		KFactor:   req.KFactor,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

func (h *MaterialHandler) ListByCompany(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	materials, err := h.svc.ListByCompany(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, materials)
}

func (h *MaterialHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid material id"})
		return
	}

	m, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrMaterialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "material not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, m)
}

func (h *MaterialHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid material id"})
		return
	}

	var req updateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.svc.Update(c.Request.Context(), id, service.UpdateMaterialInput{
		Name:      req.Name,
		Unit:      req.Unit,
		LaborCost: req.LaborCost,
		KFactor:   req.KFactor,
	})
	if err != nil {
		if errors.Is(err, service.ErrMaterialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "material not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, m)
}

func (h *MaterialHandler) UpdatePrice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid material id"})
		return
	}

	changedBy, err := uuid.Parse(c.GetHeader("X-User-ID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid X-User-ID header"})
		return
	}

	var req updatePriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.svc.UpdatePrice(c.Request.Context(), id, service.UpdatePriceInput{
		NewPrice:  req.UnitPrice,
		ChangedBy: changedBy,
	})
	if err != nil {
		if errors.Is(err, service.ErrMaterialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "material not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, m)
}

func (h *MaterialHandler) GetPriceHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid material id"})
		return
	}

	history, err := h.svc.GetPriceHistory(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrMaterialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "material not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *MaterialHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid material id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrMaterialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "material not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "material deleted"})
}
