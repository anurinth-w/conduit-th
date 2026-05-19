package repository

import (
"context"
"encoding/json"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
ID       uuid.UUID
Email    string
Name     string
Phone    string
IsActive bool
Role     string 
}

type Membership struct {
ID           uuid.UUID
UserID       uuid.UUID
CompanyID    uuid.UUID
CompanyName  string
Role         string
JobTypeScope []string
IsActive     bool
}

type CreateUserParams struct {
Email    string
Password string
Name     string
Phone    string
}

type UpdateUserParams struct {
Name  string
Phone string
}

type UserRepository struct {
db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, p CreateUserParams) (*User, error) {
u := &User{}
err := r.db.QueryRow(ctx,
`INSERT INTO users (email, password, name, phone)
 VALUES ($1, $2, $3, $4)
 RETURNING id, email, name, phone, is_active`,
p.Email, p.Password, p.Name, p.Phone,
).Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.IsActive)
return u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
u := &User{}
err := r.db.QueryRow(ctx,
`SELECT id, email, name, phone, is_active
 FROM users WHERE id = $1`,
id,
).Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.IsActive)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return u, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
u := &User{}
err := r.db.QueryRow(ctx,
`SELECT id, email, name, phone, is_active
 FROM users WHERE email = $1`,
email,
).Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.IsActive)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return u, err
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, p UpdateUserParams) (*User, error) {
u := &User{}
err := r.db.QueryRow(ctx,
`UPDATE users SET name=$1, phone=$2, updated_at=NOW()
 WHERE id=$3
 RETURNING id, email, name, phone, is_active`,
p.Name, p.Phone, id,
).Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.IsActive)
return u, err
}

func (r *UserRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
_, err := r.db.Exec(ctx,
`UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2`,
active, id,
)
return err
}

func (r *UserRepository) GetMemberships(ctx context.Context, userID uuid.UUID) ([]Membership, error) {
    rows, err := r.db.Query(ctx,
        `SELECT m.id, m.user_id, m.company_id, c.name, m.role, m.job_type_scope, m.is_active
         FROM user_company_memberships m
         JOIN companies c ON c.id = m.company_id
         WHERE m.user_id = $1 AND m.is_active = true`,
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
        if err := rows.Scan(&m.ID, &m.UserID, &m.CompanyID, &m.CompanyName,
                            &m.Role, &scope, &m.IsActive); err != nil {
            return nil, err
        }
        if len(scope) > 0 {
            _ = json.Unmarshal(scope, &m.JobTypeScope)
        }
        memberships = append(memberships, m)
    }
    return memberships, nil
}

func (r *UserRepository) AddMembership(ctx context.Context, userID, companyID uuid.UUID, role string, scope []string) (*Membership, error) {
m := &Membership{}
var scopeBytes []byte
err := r.db.QueryRow(ctx,
`INSERT INTO user_company_memberships (user_id, company_id, role, job_type_scope)
 VALUES ($1, $2, $3, $4)
 ON CONFLICT (user_id, company_id) DO UPDATE
 SET role=$3, job_type_scope=$4, is_active=true, updated_at=NOW()
 RETURNING id, user_id, company_id, role, job_type_scope, is_active`,
userID, companyID, role, scope,
).Scan(&m.ID, &m.UserID, &m.CompanyID, &m.Role, &scopeBytes, &m.IsActive)
return m, err
}

func (r *UserRepository) ListByCompany(ctx context.Context, companyID uuid.UUID) ([]User, error) {
    rows, err := r.db.Query(ctx,
        `SELECT u.id, u.email, u.name, u.phone, u.is_active, m.role
         FROM users u
         JOIN user_company_memberships m ON m.user_id = u.id
         WHERE m.company_id = $1 AND m.is_active = true`,
        companyID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.IsActive, &u.Role); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, nil
}
