package service

import (
"context"
"errors"

"github.com/anurinth-w/conduit-th/services/user/repository"
"github.com/google/uuid"
"golang.org/x/crypto/bcrypt"
)

var (
ErrEmailTaken   = errors.New("email already taken")
ErrUserNotFound = errors.New("user not found")
ErrForbidden    = errors.New("forbidden")
)

type CreateUserInput struct {
Email    string
Password string
Name     string
Phone    string
}

type AddMembershipInput struct {
UserID    uuid.UUID
CompanyID uuid.UUID
Role      string
Scope     []string
}

type UserService struct {
repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*repository.User, error) {
existing, err := s.repo.FindByEmail(ctx, in.Email)
if err != nil {
return nil, err
}
if existing != nil {
return nil, ErrEmailTaken
}

hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
if err != nil {
return nil, err
}

return s.repo.Create(ctx, repository.CreateUserParams{
Email:    in.Email,
Password: string(hash),
Name:     in.Name,
Phone:    in.Phone,
})
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*repository.User, error) {
u, err := s.repo.FindByID(ctx, id)
if err != nil {
return nil, err
}
if u == nil {
return nil, ErrUserNotFound
}
return u, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, name, phone string) (*repository.User, error) {
u, err := s.repo.FindByID(ctx, id)
if err != nil {
return nil, err
}
if u == nil {
return nil, ErrUserNotFound
}
return s.repo.Update(ctx, id, repository.UpdateUserParams{
Name:  name,
Phone: phone,
})
}

func (s *UserService) Deactivate(ctx context.Context, id uuid.UUID) error {
return s.repo.SetActive(ctx, id, false)
}

func (s *UserService) AddMembership(ctx context.Context, in AddMembershipInput) (*repository.Membership, error) {
return s.repo.AddMembership(ctx, in.UserID, in.CompanyID, in.Role, in.Scope)
}

func (s *UserService) GetMemberships(ctx context.Context, userID uuid.UUID) ([]repository.Membership, error) {
return s.repo.GetMemberships(ctx, userID)
}

func (s *UserService) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]repository.User, error) {
return s.repo.ListByCompany(ctx, companyID)
}
