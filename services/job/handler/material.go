package handler

import (
"net/http"

"github.com/anurinth-w/conduit-th/services/job/repository"
"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

type addMaterialRequest struct {
MaterialID string  `json:"material_id"`
Code       string  `json:"code"`
Name       string  `json:"name"       binding:"required"`
Unit       string  `json:"unit"       binding:"required"`
Quantity   float64 `json:"quantity"   binding:"required"`
UnitPrice  float64 `json:"unit_price"`
LaborCost  float64 `json:"labor_cost"`
}

type MaterialHandler struct {
repo *repository.MaterialRepository
}

func NewMaterialHandler(repo *repository.MaterialRepository) *MaterialHandler {
return &MaterialHandler{repo: repo}
}

func (h *MaterialHandler) Add(c *gin.Context) {
jobID, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
return
}

var req addMaterialRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var matID *uuid.UUID
if req.MaterialID != "" {
id, err := uuid.Parse(req.MaterialID)
if err == nil {
matID = &id
}
}

m, err := h.repo.Add(c.Request.Context(), repository.AddMaterialParams{
JobID:      jobID,
MaterialID: matID,
Code:       req.Code,
Name:       req.Name,
Unit:       req.Unit,
Quantity:   req.Quantity,
UnitPrice:  req.UnitPrice,
LaborCost:  req.LaborCost,
})
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusCreated, m)
}

func (h *MaterialHandler) List(c *gin.Context) {
jobID, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
return
}

materials, err := h.repo.ListByJob(c.Request.Context(), jobID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, materials)
}

func (h *MaterialHandler) Delete(c *gin.Context) {
id, err := uuid.Parse(c.Param("material_id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid material id"})
return
}

if err := h.repo.Delete(c.Request.Context(), id); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
