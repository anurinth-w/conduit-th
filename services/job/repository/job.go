package repository

import (
"context"
"errors"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
)

type Job struct {
ID                 uuid.UUID
CompanyID          uuid.UUID
CreatedBy          uuid.UUID
JobCode            string
RefCode            string
ReportNumber       string
WaterUserCode      string
JobType            string
Status             string
Cause              string
LocationText       string
Subdistrict        string
District           string
Province           string
Lat                float64
Lng                float64
PipeType           string
PipeSizeMM         int
SurfaceCondition   string
SurfaceAreaSqm     float64
WorkMethod         string
JobSource          string
ContactTechnician  string
ContactCoordinator string
CostMain           float64
CostSurface        float64
NotifiedAt         *time.Time
StartedAt          *time.Time
EndedAt            *time.Time
DuplicateRef       string
DuplicateNote      string
CreatedAt          time.Time
UpdatedAt          time.Time
}

type Assignment struct {
ID             uuid.UUID
JobID          uuid.UUID
TechnicianID   uuid.UUID
AssignedBy     uuid.UUID
AssignmentType string
Status         string
CostShare      float64
Note           string
AssignedAt     time.Time
}

type CreateJobParams struct {
CompanyID          uuid.UUID
CreatedBy          uuid.UUID
JobCode            string
RefCode            string
ReportNumber       string
WaterUserCode      string
JobType            string
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

type UpdateJobParams struct {
RefCode            string
ReportNumber       string
WaterUserCode      string
Cause              string
LocationText       string
Subdistrict        string
District           string
Province           string
Lat                float64
Lng                float64
PipeType           string
PipeSizeMM         int
SurfaceCondition   string
SurfaceAreaSqm     float64
WorkMethod         string
JobSource          string
ContactTechnician  string
ContactCoordinator string
StartedAt          *time.Time
EndedAt            *time.Time
}

type JobRepository struct {
db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, p CreateJobParams) (*Job, error) {
j := &Job{}
err := r.db.QueryRow(ctx,
`INSERT INTO jobs (
company_id, created_by, job_code, ref_code, report_number,
water_user_code, job_type, cause, location_text,
subdistrict, district, province, job_source,
contact_technician, contact_coordinator, notified_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
RETURNING id, company_id, created_by, job_code, ref_code,
report_number, water_user_code, job_type, status,
cause, location_text, subdistrict, district, province,
coalesce(lat,0), coalesce(lng,0),
coalesce(pipe_type,''), coalesce(pipe_size_mm,0),
coalesce(surface_condition,''), coalesce(surface_area_sqm,0),
coalesce(work_method,''), coalesce(job_source,''),
coalesce(contact_technician,''), coalesce(contact_coordinator,''),
coalesce(cost_main,0), coalesce(cost_surface,0),
notified_at, started_at, ended_at,
coalesce(duplicate_ref,''), coalesce(duplicate_note,''),
created_at, updated_at`,
p.CompanyID, p.CreatedBy, p.JobCode, p.RefCode, p.ReportNumber,
p.WaterUserCode, p.JobType, p.Cause, p.LocationText,
p.Subdistrict, p.District, p.Province, p.JobSource,
p.ContactTechnician, p.ContactCoordinator, p.NotifiedAt,
).Scan(
&j.ID, &j.CompanyID, &j.CreatedBy, &j.JobCode, &j.RefCode,
&j.ReportNumber, &j.WaterUserCode, &j.JobType, &j.Status,
&j.Cause, &j.LocationText, &j.Subdistrict, &j.District, &j.Province,
&j.Lat, &j.Lng, &j.PipeType, &j.PipeSizeMM,
&j.SurfaceCondition, &j.SurfaceAreaSqm, &j.WorkMethod, &j.JobSource,
&j.ContactTechnician, &j.ContactCoordinator,
&j.CostMain, &j.CostSurface,
&j.NotifiedAt, &j.StartedAt, &j.EndedAt,
&j.DuplicateRef, &j.DuplicateNote,
&j.CreatedAt, &j.UpdatedAt,
)
return j, err
}

func (r *JobRepository) FindByID(ctx context.Context, id uuid.UUID) (*Job, error) {
j := &Job{}
err := r.db.QueryRow(ctx,
`SELECT id, company_id, created_by, job_code, ref_code,
report_number, water_user_code, job_type, status,
cause, location_text, subdistrict, district, province,
coalesce(lat,0), coalesce(lng,0),
coalesce(pipe_type,''), coalesce(pipe_size_mm,0),
coalesce(surface_condition,''), coalesce(surface_area_sqm,0),
coalesce(work_method,''), coalesce(job_source,''),
coalesce(contact_technician,''), coalesce(contact_coordinator,''),
coalesce(cost_main,0), coalesce(cost_surface,0),
notified_at, started_at, ended_at,
coalesce(duplicate_ref,''), coalesce(duplicate_note,''),
created_at, updated_at
FROM jobs WHERE id = $1`, id,
).Scan(
&j.ID, &j.CompanyID, &j.CreatedBy, &j.JobCode, &j.RefCode,
&j.ReportNumber, &j.WaterUserCode, &j.JobType, &j.Status,
&j.Cause, &j.LocationText, &j.Subdistrict, &j.District, &j.Province,
&j.Lat, &j.Lng, &j.PipeType, &j.PipeSizeMM,
&j.SurfaceCondition, &j.SurfaceAreaSqm, &j.WorkMethod, &j.JobSource,
&j.ContactTechnician, &j.ContactCoordinator,
&j.CostMain, &j.CostSurface,
&j.NotifiedAt, &j.StartedAt, &j.EndedAt,
&j.DuplicateRef, &j.DuplicateNote,
&j.CreatedAt, &j.UpdatedAt,
)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return j, err
}

func (r *JobRepository) ListByCompany(ctx context.Context, companyID uuid.UUID, status, jobType string) ([]Job, error) {
query := `SELECT id, company_id, created_by, job_code, ref_code,
report_number, water_user_code, job_type, status,
cause, location_text, subdistrict, district, province,
coalesce(lat,0), coalesce(lng,0),
coalesce(pipe_type,''), coalesce(pipe_size_mm,0),
coalesce(surface_condition,''), coalesce(surface_area_sqm,0),
coalesce(work_method,''), coalesce(job_source,''),
coalesce(contact_technician,''), coalesce(contact_coordinator,''),
coalesce(cost_main,0), coalesce(cost_surface,0),
notified_at, started_at, ended_at,
coalesce(duplicate_ref,''), coalesce(duplicate_note,''),
created_at, updated_at
FROM jobs WHERE company_id = $1`

args := []interface{}{companyID}
if status != "" {
args = append(args, status)
query += ` AND status = $` + string(rune('0'+len(args)))
}
if jobType != "" {
args = append(args, jobType)
query += ` AND job_type = $` + string(rune('0'+len(args)))
}
query += ` ORDER BY created_at DESC`

rows, err := r.db.Query(ctx, query, args...)
if err != nil {
return nil, err
}
defer rows.Close()

var jobs []Job
for rows.Next() {
var j Job
if err := rows.Scan(
&j.ID, &j.CompanyID, &j.CreatedBy, &j.JobCode, &j.RefCode,
&j.ReportNumber, &j.WaterUserCode, &j.JobType, &j.Status,
&j.Cause, &j.LocationText, &j.Subdistrict, &j.District, &j.Province,
&j.Lat, &j.Lng, &j.PipeType, &j.PipeSizeMM,
&j.SurfaceCondition, &j.SurfaceAreaSqm, &j.WorkMethod, &j.JobSource,
&j.ContactTechnician, &j.ContactCoordinator,
&j.CostMain, &j.CostSurface,
&j.NotifiedAt, &j.StartedAt, &j.EndedAt,
&j.DuplicateRef, &j.DuplicateNote,
&j.CreatedAt, &j.UpdatedAt,
); err != nil {
return nil, err
}
jobs = append(jobs, j)
}
return jobs, nil
}

func (r *JobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
_, err := r.db.Exec(ctx,
`UPDATE jobs SET status=$1, updated_at=NOW() WHERE id=$2`,
status, id,
)
return err
}

func (r *JobRepository) Update(ctx context.Context, id uuid.UUID, p UpdateJobParams) (*Job, error) {
j := &Job{}
err := r.db.QueryRow(ctx,
`UPDATE jobs SET
ref_code=$1, report_number=$2, water_user_code=$3,
cause=$4, location_text=$5, subdistrict=$6, district=$7, province=$8,
lat=$9, lng=$10, pipe_type=$11, pipe_size_mm=$12,
surface_condition=$13, surface_area_sqm=$14, work_method=$15,
job_source=$16, contact_technician=$17, contact_coordinator=$18,
started_at=$19, ended_at=$20, updated_at=NOW()
WHERE id=$21
RETURNING id, job_code, job_type, status, updated_at`,
p.RefCode, p.ReportNumber, p.WaterUserCode,
p.Cause, p.LocationText, p.Subdistrict, p.District, p.Province,
p.Lat, p.Lng, p.PipeType, p.PipeSizeMM,
p.SurfaceCondition, p.SurfaceAreaSqm, p.WorkMethod,
p.JobSource, p.ContactTechnician, p.ContactCoordinator,
p.StartedAt, p.EndedAt, id,
).Scan(&j.ID, &j.JobCode, &j.JobType, &j.Status, &j.UpdatedAt)
return j, err
}

func (r *JobRepository) NextSequence(ctx context.Context, companyID uuid.UUID, jobType, period string) (int, error) {
var seq int
err := r.db.QueryRow(ctx,
`INSERT INTO job_code_sequences (company_id, job_type, period, last_seq)
 VALUES ($1, $2, $3, 1)
 ON CONFLICT (company_id, job_type, period)
 DO UPDATE SET last_seq = job_code_sequences.last_seq + 1, updated_at = NOW()
 RETURNING last_seq`,
companyID, jobType, period,
).Scan(&seq)
return seq, err
}

func (r *JobRepository) CreateAssignment(ctx context.Context, jobID, technicianID, assignedBy uuid.UUID, assignmentType string) (*Assignment, error) {
a := &Assignment{}
err := r.db.QueryRow(ctx,
`INSERT INTO job_assignments (job_id, technician_id, assigned_by, assignment_type)
 VALUES ($1, $2, $3, $4)
 RETURNING id, job_id, technician_id, assigned_by, assignment_type, status, coalesce(cost_share,0), coalesce(note,''), assigned_at`,
jobID, technicianID, assignedBy, assignmentType,
).Scan(&a.ID, &a.JobID, &a.TechnicianID, &a.AssignedBy, &a.AssignmentType, &a.Status, &a.CostShare, &a.Note, &a.AssignedAt)
return a, err
}

func (r *JobRepository) GetAssignments(ctx context.Context, jobID uuid.UUID) ([]Assignment, error) {
rows, err := r.db.Query(ctx,
`SELECT id, job_id, technician_id, assigned_by, assignment_type, status, coalesce(cost_share,0), coalesce(note,''), assigned_at
 FROM job_assignments WHERE job_id = $1`,
jobID,
)
if err != nil {
return nil, err
}
defer rows.Close()

var assignments []Assignment
for rows.Next() {
var a Assignment
if err := rows.Scan(&a.ID, &a.JobID, &a.TechnicianID, &a.AssignedBy, &a.AssignmentType, &a.Status, &a.CostShare, &a.Note, &a.AssignedAt); err != nil {
return nil, err
}
assignments = append(assignments, a)
}
return assignments, nil
}
