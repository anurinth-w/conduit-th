package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyLineConfig struct {
	ChannelAccessToken string
	ChannelSecret      string
}

type ManagerLineID struct {
	UserID     uuid.UUID
	LineUserID string
}

type NotifyRepository struct {
	db *pgxpool.Pool
}

func NewNotifyRepository(db *pgxpool.Pool) *NotifyRepository {
	return &NotifyRepository{db: db}
}

// GetCompanyLineConfig ดึง LINE token ของบริษัท
func (r *NotifyRepository) GetCompanyLineConfig(ctx context.Context, companyID uuid.UUID) (*CompanyLineConfig, error) {
	cfg := &CompanyLineConfig{}
	err := r.db.QueryRow(ctx,
		`SELECT coalesce(line_channel_access_token,''), coalesce(line_channel_secret,'')
		 FROM companies WHERE id=$1 AND is_active=TRUE`,
		companyID,
	).Scan(&cfg.ChannelAccessToken, &cfg.ChannelSecret)
	if err != nil {
		return nil, err
	}
	if cfg.ChannelAccessToken == "" {
		return nil, nil // บริษัทนี้ยังไม่ได้ตั้งค่า LINE
	}
	return cfg, nil
}

// GetManagersByCompany ดึง line_user_id ของ manager ทุกคนในบริษัท
func (r *NotifyRepository) GetManagersByCompany(ctx context.Context, companyID uuid.UUID) ([]ManagerLineID, error) {
	rows, err := r.db.Query(ctx,
		`SELECT u.id, coalesce(u.line_user_id,'')
		 FROM users u
		 JOIN user_company_memberships m ON m.user_id = u.id
		 WHERE m.company_id=$1
		   AND m.role IN ('manager','admin')
		   AND m.is_active=TRUE
		   AND u.is_active=TRUE
		   AND u.line_user_id IS NOT NULL`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var managers []ManagerLineID
	for rows.Next() {
		var m ManagerLineID
		if err := rows.Scan(&m.UserID, &m.LineUserID); err != nil {
			return nil, err
		}
		managers = append(managers, m)
	}
	return managers, nil
}

// GetLineGroupsByJob ดึงกลุ่ม LINE ที่เชื่อมกับงาน
func (r *NotifyRepository) GetLineGroupsByJob(ctx context.Context, jobID uuid.UUID, role string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT lg.line_group_id
		 FROM job_line_groups jlg
		 JOIN line_groups lg ON lg.id = jlg.line_group_id
		 WHERE jlg.job_id=$1 AND jlg.role=$2 AND lg.is_active=TRUE`,
		jobID, role,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var gid string
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, gid)
	}
	return groupIDs, nil
}

// GetJobCompanyID ดึง company_id จาก job
func (r *NotifyRepository) GetJobCompanyID(ctx context.Context, jobID uuid.UUID) (uuid.UUID, error) {
	var companyID uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT company_id FROM jobs WHERE id=$1`, jobID,
	).Scan(&companyID)
	return companyID, err
}
