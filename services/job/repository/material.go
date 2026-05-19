package repository

import (
"context"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5/pgxpool"
)

type JobMaterial struct {
ID         uuid.UUID
JobID      uuid.UUID
MaterialID *uuid.UUID
Code       string
Name       string
Unit       string
Quantity   float64
UnitPrice  float64
LaborCost  float64
Total      float64
CreatedAt  time.Time
}

type AddMaterialParams struct {
JobID      uuid.UUID
MaterialID *uuid.UUID
Code       string
Name       string
Unit       string
Quantity   float64
UnitPrice  float64
LaborCost  float64
}

type MaterialRepository struct {
db *pgxpool.Pool
}

func NewMaterialRepository(db *pgxpool.Pool) *MaterialRepository {
return &MaterialRepository{db: db}
}

func (r *MaterialRepository) Add(ctx context.Context, p AddMaterialParams) (*JobMaterial, error) {
m := &JobMaterial{}
err := r.db.QueryRow(ctx,
`INSERT INTO job_materials (job_id, material_id, code, name, unit, quantity, unit_price, labor_cost)
 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
 RETURNING id, job_id, material_id, code, name, unit, quantity, unit_price, labor_cost, total, created_at`,
p.JobID, p.MaterialID, p.Code, p.Name, p.Unit, p.Quantity, p.UnitPrice, p.LaborCost,
).Scan(&m.ID, &m.JobID, &m.MaterialID, &m.Code, &m.Name, &m.Unit,
&m.Quantity, &m.UnitPrice, &m.LaborCost, &m.Total, &m.CreatedAt)
return m, err
}

func (r *MaterialRepository) ListByJob(ctx context.Context, jobID uuid.UUID) ([]JobMaterial, error) {
rows, err := r.db.Query(ctx,
`SELECT id, job_id, material_id, code, name, unit, quantity, unit_price, labor_cost, total, created_at
 FROM job_materials WHERE job_id = $1 ORDER BY created_at ASC`,
jobID,
)
if err != nil {
return nil, err
}
defer rows.Close()

var materials []JobMaterial
for rows.Next() {
var m JobMaterial
if err := rows.Scan(&m.ID, &m.JobID, &m.MaterialID, &m.Code, &m.Name, &m.Unit,
&m.Quantity, &m.UnitPrice, &m.LaborCost, &m.Total, &m.CreatedAt); err != nil {
return nil, err
}
materials = append(materials, m)
}
return materials, nil
}

func (r *MaterialRepository) Delete(ctx context.Context, id uuid.UUID) error {
_, err := r.db.Exec(ctx, `DELETE FROM job_materials WHERE id = $1`, id)
return err
}
