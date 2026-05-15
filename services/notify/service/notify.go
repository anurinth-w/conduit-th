package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/anurinth-w/conduit-th/services/notify/line"
	"github.com/anurinth-w/conduit-th/services/notify/repository"
	"github.com/google/uuid"
)

var ErrLineNotConfigured = errors.New("LINE not configured for this company")

// Event types
const (
	EventJobCreated    = "job_created"     // งานสร้างใหม่ → แจ้ง Manager
	EventJobAssigned   = "job_assigned"    // assign งาน → ส่งไปกลุ่มช่าง
	EventJobDone       = "job_done"        // ช่างทำเสร็จ → แจ้ง Manager
	EventJobConfirmed  = "job_confirmed"   // Manager confirm → ส่งกลับกลุ่ม source
)

type NotifyInput struct {
	Event     string
	JobID     uuid.UUID
	CompanyID uuid.UUID
	Message   string
}

type NotifyService struct {
	repo       *repository.NotifyRepository
	lineClient *line.Client
}

func NewNotifyService(repo *repository.NotifyRepository, lineClient *line.Client) *NotifyService {
	return &NotifyService{repo: repo, lineClient: lineClient}
}

func (s *NotifyService) Notify(ctx context.Context, input NotifyInput) error {
	// ดึง LINE token ของบริษัท
	cfg, err := s.repo.GetCompanyLineConfig(ctx, input.CompanyID)
	if err != nil {
		return fmt.Errorf("get line config: %w", err)
	}
	if cfg == nil {
		return ErrLineNotConfigured
	}

	switch input.Event {

	case EventJobCreated:
		// แจ้ง Manager ทุกคนในบริษัทโดยตรง (1:1)
		return s.notifyManagers(ctx, cfg.ChannelAccessToken, input.CompanyID, input.Message)

	case EventJobAssigned:
		// ส่งไปกลุ่ม destination (กลุ่มช่าง)
		return s.notifyGroups(ctx, cfg.ChannelAccessToken, input.JobID, "destination", input.Message)

	case EventJobDone:
		// แจ้ง Manager ทุกคนในบริษัทโดยตรง (1:1)
		return s.notifyManagers(ctx, cfg.ChannelAccessToken, input.CompanyID, input.Message)

	case EventJobConfirmed:
		// ส่งกลับกลุ่ม source (กลุ่มลูกค้าต้นทาง)
		return s.notifyGroups(ctx, cfg.ChannelAccessToken, input.JobID, "source", input.Message)

	default:
		return fmt.Errorf("unknown event: %s", input.Event)
	}
}

// notifyManagers ส่งหา Manager ทุกคนแบบ 1:1
func (s *NotifyService) notifyManagers(ctx context.Context, token string, companyID uuid.UUID, message string) error {
	managers, err := s.repo.GetManagersByCompany(ctx, companyID)
	if err != nil {
		return fmt.Errorf("get managers: %w", err)
	}

	for _, m := range managers {
		if err := s.lineClient.PushText(ctx, token, m.LineUserID, message); err != nil {
			// log แต่ไม่ stop — คนอื่นยังส่งได้
			log.Printf("warn: failed to notify manager %s: %v", m.UserID, err)
		}
	}
	return nil
}

// notifyGroups ส่งไปกลุ่ม LINE ตาม role (source/destination)
func (s *NotifyService) notifyGroups(ctx context.Context, token string, jobID uuid.UUID, role, message string) error {
	groupIDs, err := s.repo.GetLineGroupsByJob(ctx, jobID, role)
	if err != nil {
		return fmt.Errorf("get line groups: %w", err)
	}

	for _, gid := range groupIDs {
		if err := s.lineClient.PushText(ctx, token, gid, message); err != nil {
			log.Printf("warn: failed to notify group %s: %v", gid, err)
		}
	}
	return nil
}
