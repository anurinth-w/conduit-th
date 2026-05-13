package service

import (
"context"
"errors"
"fmt"
"time"

"github.com/anurinth-w/conduit-th/services/job/repository"
"github.com/google/uuid"
)

var (
ErrJobNotFound       = errors.New("job not found")
ErrForbidden         = errors.New("forbidden")
ErrNoFormat          = errors.New("job code format not configured for this job type")
ErrInvalidTransition = errors.New("invalid status transition")
)

type CreateJobInput struct {
CompanyID          uuid.UUID
CreatedBy          uuid.UUID
JobType            string
JobCodeFormat      string
RefCode            string
ReportNumber       string
WaterUserCode      string
Cause              string
LocationText       string
Subdistrict        string
District           string
Province           string
JobSource          string
ContactTechnician  string
ContactCoordinator string
NotifiedAt         *time.Time
}

type AssignJobInput struct {
JobID          uuid.UUID
TechnicianID   uuid.UUID
AssignedBy     uuid.UUID
AssignmentType string
}

type JobService struct {
repo *repository.JobRepository
}

func NewJobService(repo *repository.JobRepository) *JobService {
return &JobService{repo: repo}
}

func (s *JobService) Create(ctx context.Context, in CreateJobInput) (*repository.Job, error) {
if in.JobCodeFormat == "" {
return nil, ErrNoFormat
}

jobCode, err := s.GenerateJobCode(ctx, in.CompanyID, in.JobType, in.JobCodeFormat)
if err != nil {
return nil, fmt.Errorf("generate job code: %w", err)
}

return s.repo.Create(ctx, repository.CreateJobParams{
CompanyID:          in.CompanyID,
CreatedBy:          in.CreatedBy,
JobCode:            jobCode,
RefCode:            in.RefCode,
ReportNumber:       in.ReportNumber,
WaterUserCode:      in.WaterUserCode,
JobType:            in.JobType,
Cause:              in.Cause,
LocationText:       in.LocationText,
Subdistrict:        in.Subdistrict,
District:           in.District,
Province:           in.Province,
JobSource:          in.JobSource,
ContactTechnician:  in.ContactTechnician,
ContactCoordinator: in.ContactCoordinator,
NotifiedAt:         in.NotifiedAt,
})
}

func (s *JobService) GetByID(ctx context.Context, id uuid.UUID) (*repository.Job, error) {
j, err := s.repo.FindByID(ctx, id)
if err != nil {
return nil, err
}
if j == nil {
return nil, ErrJobNotFound
}
return j, nil
}

func (s *JobService) ListByCompany(ctx context.Context, companyID uuid.UUID, status, jobType string) ([]repository.Job, error) {
return s.repo.ListByCompany(ctx, companyID, status, jobType)
}

func (s *JobService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string) error {
job, err := s.repo.FindByID(ctx, id)
if err != nil {
return err
}
if job == nil {
return ErrJobNotFound
}

if !canTransition(job.Status, newStatus) {
return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, job.Status, newStatus)
}

return s.repo.UpdateStatus(ctx, id, newStatus)
}

func (s *JobService) Assign(ctx context.Context, in AssignJobInput) (*repository.Assignment, error) {
job, err := s.repo.FindByID(ctx, in.JobID)
if err != nil {
return nil, err
}
if job == nil {
return nil, ErrJobNotFound
}

assignmentType := in.AssignmentType
if assignmentType == "" {
assignmentType = "main"
}

assignment, err := s.repo.CreateAssignment(ctx, in.JobID, in.TechnicianID, in.AssignedBy, assignmentType)
if err != nil {
return nil, err
}

if job.Status == "open" {
_ = s.repo.UpdateStatus(ctx, in.JobID, "assigned")
}

return assignment, nil
}

func (s *JobService) GetAssignments(ctx context.Context, jobID uuid.UUID) ([]repository.Assignment, error) {
return s.repo.GetAssignments(ctx, jobID)
}
