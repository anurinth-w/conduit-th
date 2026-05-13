package handler

import (
"errors"
"net/http"

"github.com/anurinth-w/conduit-th/services/job/service"
"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

type JobHandler struct {
svc *service.JobService
}

func NewJobHandler(svc *service.JobService) *JobHandler {
return &JobHandler{svc: svc}
}

type createJobRequest struct {
CompanyID          string `json:"company_id"          binding:"required"`
JobType            string `json:"job_type"            binding:"required"`
JobCodeFormat      string `json:"job_code_format"     binding:"required"`
RefCode            string `json:"ref_code"`
ReportNumber       string `json:"report_number"`
WaterUserCode      string `json:"water_user_code"`
Cause              string `json:"cause"`
LocationText       string `json:"location_text"`
Subdistrict        string `json:"subdistrict"`
District           string `json:"district"`
Province           string `json:"province"`
JobSource          string `json:"job_source"`
ContactTechnician  string `json:"contact_technician"`
ContactCoordinator string `json:"contact_coordinator"`
}

type updateStatusRequest struct {
Status string `json:"status" binding:"required"`
}

type assignRequest struct {
TechnicianID   string `json:"technician_id"   binding:"required"`
AssignmentType string `json:"assignment_type"`
}

func (h *JobHandler) Create(c *gin.Context) {
var req createJobRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

companyID, err := uuid.Parse(req.CompanyID)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
return
}

createdBy, err := uuid.Parse(c.GetHeader("X-User-ID"))
if err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid X-User-ID header"})
return
}

job, err := h.svc.Create(c.Request.Context(), service.CreateJobInput{
CompanyID:          companyID,
CreatedBy:          createdBy,
JobType:            req.JobType,
JobCodeFormat:      req.JobCodeFormat,
RefCode:            req.RefCode,
ReportNumber:       req.ReportNumber,
WaterUserCode:      req.WaterUserCode,
Cause:              req.Cause,
LocationText:       req.LocationText,
Subdistrict:        req.Subdistrict,
District:           req.District,
Province:           req.Province,
JobSource:          req.JobSource,
ContactTechnician:  req.ContactTechnician,
ContactCoordinator: req.ContactCoordinator,
})
if err != nil {
if errors.Is(err, service.ErrNoFormat) {
c.JSON(http.StatusBadRequest, gin.H{"error": "job code format not configured"})
} else {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
return
}

c.JSON(http.StatusCreated, job)
}

func (h *JobHandler) GetByID(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
return
}

job, err := h.svc.GetByID(c.Request.Context(), id)
if err != nil {
if errors.Is(err, service.ErrJobNotFound) {
c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
} else {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
return
}

c.JSON(http.StatusOK, job)
}

func (h *JobHandler) ListByCompany(c *gin.Context) {
companyID, err := uuid.Parse(c.Param("company_id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
return
}

status := c.Query("status")
jobType := c.Query("job_type")

jobs, err := h.svc.ListByCompany(c.Request.Context(), companyID, status, jobType)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusOK, jobs)
}

func (h *JobHandler) UpdateStatus(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
return
}

var req updateStatusRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
if errors.Is(err, service.ErrJobNotFound) {
c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
} else if errors.Is(err, service.ErrInvalidTransition) {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
} else {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
return
}

c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *JobHandler) Assign(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
return
}

var req assignRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

techID, err := uuid.Parse(req.TechnicianID)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid technician_id"})
return
}

assignedBy, err := uuid.Parse(c.GetHeader("X-User-ID"))
if err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid X-User-ID header"})
return
}

assignment, err := h.svc.Assign(c.Request.Context(), service.AssignJobInput{
JobID:          id,
TechnicianID:   techID,
AssignedBy:     assignedBy,
AssignmentType: req.AssignmentType,
})
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusCreated, assignment)
}

func (h *JobHandler) GetAssignments(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
return
}

assignments, err := h.svc.GetAssignments(c.Request.Context(), id)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusOK, assignments)
}

func (h *JobHandler) Health(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
