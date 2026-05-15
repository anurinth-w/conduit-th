package service

import (
	"context"
	"errors"

	"github.com/anurinth-w/conduit-th/services/material/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrMaterialNotFound = errors.New("material not found")
)

type CreateMaterialInput struct {
	CompanyID uuid.UUID
	Code      string
	Name      string
	Unit      string
	UnitPrice float64
	LaborCost float64
	KFactor   string
}

type UpdateMaterialInput struct {
	Name      string
	Unit      string
	LaborCost float64
	KFactor   string
}

type UpdatePriceInput struct {
	NewPrice  float64
	ChangedBy uuid.UUID
}

type MaterialService struct {
	repo *repository.MaterialRepository
}

func NewMaterialService(repo *repository.MaterialRepository) *MaterialService {
	return &MaterialService{repo: repo}
}

func (s *MaterialService) Create(ctx context.Context, input CreateMaterialInput) (*repository.Material, error) {
	return s.repo.Create(ctx, repository.CreateMaterialParams{
		CompanyID: input.CompanyID,
		Code:      input.Code,
		Name:      input.Name,
		Unit:      input.Unit,
		UnitPrice: input.UnitPrice,
		LaborCost: input.LaborCost,
		KFactor:   input.KFactor,
	})
}

func (s *MaterialService) GetByID(ctx context.Context, id uuid.UUID) (*repository.Material, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMaterialNotFound
	}
	return m, nil
}

func (s *MaterialService) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]repository.Material, error) {
	return s.repo.ListByCompany(ctx, companyID)
}

func (s *MaterialService) Update(ctx context.Context, id uuid.UUID, input UpdateMaterialInput) (*repository.Material, error) {
	m, err := s.repo.Update(ctx, id, repository.UpdateMaterialParams{
		Name:      input.Name,
		Unit:      input.Unit,
		LaborCost: input.LaborCost,
		KFactor:   input.KFactor,
	})
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMaterialNotFound
	}
	return m, nil
}

func (s *MaterialService) UpdatePrice(ctx context.Context, id uuid.UUID, input UpdatePriceInput) (*repository.Material, error) {
	m, err := s.repo.UpdatePrice(ctx, id, input.NewPrice, input.ChangedBy)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMaterialNotFound
	}
	return m, nil
}

func (s *MaterialService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.SoftDelete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrMaterialNotFound
	}
	return err
}

func (s *MaterialService) GetPriceHistory(ctx context.Context, materialID uuid.UUID) ([]repository.PriceHistory, error) {
	m, err := s.repo.FindByID(ctx, materialID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMaterialNotFound
	}
	return s.repo.GetPriceHistory(ctx, materialID)
}
