package repository

import (
"context"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
ID       uuid.UUID
Email    string
Password string
Name     string
Phone    string
IsActive bool
}

type Membership struct {
CompanyID    uuid.UUID
Role         string
JobTypeScope []string
IsActive     bool
}

type UserRepository struct {
db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
u := &User{}
err := r.db.QueryRow(ctx,
`SELECT id, email, password, name, phone, is_active
 FROM users WHERE email = $1`,
email,
).Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.Phone, &u.IsActive)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
u := &User{}
err := r.db.QueryRow(ctx,
`SELECT id, email, password, name, phone, is_active
 FROM users WHERE id = $1`,
id,
).Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.Phone, &u.IsActive)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return u, err
}

func (r *UserRepository) GetMemberships(ctx context.Context, userID uuid.UUID) ([]Membership, error) {
rows, err := r.db.Query(ctx,
`SELECT company_id, role, job_type_scope, is_active
 FROM user_company_memberships
 WHERE user_id = $1 AND is_active = true`,
userID,
)
if err != nil {
return nil, err
}
defer rows.Close()

var memberships []Membership
for rows.Next() {
var m Membership
var scope []byte
if err := rows.Scan(&m.CompanyID, &m.Role, &scope, &m.IsActive); err != nil {
return nil, err
}
memberships = append(memberships, m)
}
return memberships, nil
}

func (r *UserRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt interface{}) error {
_, err := r.db.Exec(ctx,
`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
 VALUES ($1, $2, $3)`,
userID, tokenHash, expiresAt,
)
return err
}

func (r *UserRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*uuid.UUID, error) {
var userID uuid.UUID
err := r.db.QueryRow(ctx,
`SELECT user_id FROM refresh_tokens
 WHERE token_hash = $1 AND expires_at > NOW()`,
tokenHash,
).Scan(&userID)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return &userID, err
}

func (r *UserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
_, err := r.db.Exec(ctx,
`DELETE FROM refresh_tokens WHERE token_hash = $1`,
tokenHash,
)
return err
}

func (r *UserRepository) DeleteAllRefreshTokens(ctx context.Context, userID uuid.UUID) error {
_, err := r.db.Exec(ctx,
`DELETE FROM refresh_tokens WHERE user_id = $1`,
userID,
)
return err
}
