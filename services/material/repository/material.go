package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Material struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Code      string
	Name      string
	Unit      string
	UnitPrice float64
	LaborCost float64
	KFactor   string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PriceHistory struct {
	ID         uuid.UUID
	MaterialID uuid.UUID
	OldPrice   float64
	NewPrice   float64
	ChangedBy  uuid.UUID
	ChangedAt  time.Time
}

type CreateMaterialParams struct {
	CompanyID uuid.UUID
	Code      string
	Name      string
	Unit      string
	UnitPrice float64
	LaborCost float64
	KFactor   string
}

type UpdateMaterialParams struct {
	Name      string
	Unit      string
	LaborCost float64
	KFactor   string
}

type MaterialRepository struct {
	db *pgxpool.Pool
}

func NewMaterialRepository(db *pgxpool.Pool) *MaterialRepository {
	return &MaterialRepository{db: db}
}

func (r *MaterialRepository) Create(ctx context.Context, p CreateMaterialParams) (*Material, error) {
	m := &Material{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO materials (company_id, code, name, unit, unit_price, labor_cost, k_factor)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, company_id, code, name, unit, unit_price, labor_cost,
		           coalesce(k_factor,''), is_active, created_at, updated_at`,
		p.CompanyID, p.Code, p.Name, p.Unit, p.UnitPrice, p.LaborCost, p.KFactor,
	).Scan(
		&m.ID, &m.CompanyID, &m.Code, &m.Name, &m.Unit,
		&m.UnitPrice, &m.LaborCost, &m.KFactor, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

func (r *MaterialRepository) FindByID(ctx context.Context, id uuid.UUID) (*Material, error) {
	m := &Material{}
	err := r.db.QueryRow(ctx,
		`SELECT id, company_id, code, name, unit, unit_price, labor_cost,
		        coalesce(k_factor,''), is_active, created_at, updated_at
		 FROM materials WHERE id = $1 AND is_active = TRUE`,
		id,
	).Scan(
		&m.ID, &m.CompanyID, &m.Code, &m.Name, &m.Unit,
		&m.UnitPrice, &m.LaborCost, &m.KFactor, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *MaterialRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]Material, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, company_id, code, name, unit, unit_price, labor_cost,
		        coalesce(k_factor,''), is_active, created_at, updated_at
		 FROM materials
		 WHERE company_id = $1 AND is_active = TRUE
		 ORDER BY code ASC`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var m Material
		if err := rows.Scan(
			&m.ID, &m.CompanyID, &m.Code, &m.Name, &m.Unit,
			&m.UnitPrice, &m.LaborCost, &m.KFactor, &m.IsActive,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		materials = append(materials, m)
	}
	return materials, nil
}

func (r *MaterialRepository) Update(ctx context.Context, id uuid.UUID, p UpdateMaterialParams) (*Material, error) {
	m := &Material{}
	err := r.db.QueryRow(ctx,
		`UPDATE materials
		 SET name=$1, unit=$2, labor_cost=$3, k_factor=$4, updated_at=NOW()
		 WHERE id=$5 AND is_active=TRUE
		 RETURNING id, company_id, code, name, unit, unit_price, labor_cost,
		           coalesce(k_factor,''), is_active, created_at, updated_at`,
		p.Name, p.Unit, p.LaborCost, p.KFactor, id,
	).Scan(
		&m.ID, &m.CompanyID, &m.Code, &m.Name, &m.Unit,
		&m.UnitPrice, &m.LaborCost, &m.KFactor, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *MaterialRepository) UpdatePrice(ctx context.Context, id uuid.UUID, newPrice float64, changedBy uuid.UUID) (*Material, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// ดึง old price
	var oldPrice float64
	err = tx.QueryRow(ctx,
		`SELECT unit_price FROM materials WHERE id=$1 AND is_active=TRUE`, id,
	).Scan(&oldPrice)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// บันทึก price history
	_, err = tx.Exec(ctx,
		`INSERT INTO material_price_history (material_id, old_price, new_price, changed_by)
		 VALUES ($1, $2, $3, $4)`,
		id, oldPrice, newPrice, changedBy,
	)
	if err != nil {
		return nil, err
	}

	// อัปเดต price
	m := &Material{}
	err = tx.QueryRow(ctx,
		`UPDATE materials SET unit_price=$1, updated_at=NOW()
		 WHERE id=$2
		 RETURNING id, company_id, code, name, unit, unit_price, labor_cost,
		           coalesce(k_factor,''), is_active, created_at, updated_at`,
		newPrice, id,
	).Scan(
		&m.ID, &m.CompanyID, &m.Code, &m.Name, &m.Unit,
		&m.UnitPrice, &m.LaborCost, &m.KFactor, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return m, tx.Commit(ctx)
}

func (r *MaterialRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE materials SET is_active=FALSE, updated_at=NOW() WHERE id=$1 AND is_active=TRUE`,
		id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *MaterialRepository) GetPriceHistory(ctx context.Context, materialID uuid.UUID) ([]PriceHistory, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, material_id, old_price, new_price, changed_by, changed_at
		 FROM material_price_history
		 WHERE material_id = $1
		 ORDER BY changed_at DESC`,
		materialID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []PriceHistory
	for rows.Next() {
		var h PriceHistory
		if err := rows.Scan(
			&h.ID, &h.MaterialID, &h.OldPrice, &h.NewPrice,
			&h.ChangedBy, &h.ChangedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}
