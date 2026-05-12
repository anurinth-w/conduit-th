package service

import (
"context"
"crypto/sha256"
"encoding/hex"
"errors"
"fmt"
"time"

"github.com/anurinth-w/conduit-th/services/auth/config"
"github.com/anurinth-w/conduit-th/services/auth/repository"
"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"
"golang.org/x/crypto/bcrypt"
)

var (
ErrInvalidCredentials = errors.New("invalid email or password")
ErrUserInactive       = errors.New("user is inactive")
ErrInvalidToken       = errors.New("invalid token")
)

type Claims struct {
jwt.RegisteredClaims
UserID string `json:"uid"`
Email  string `json:"email"`
}

type LoginResponse struct {
AccessToken  string        `json:"access_token"`
RefreshToken string        `json:"refresh_token"`
ExpiresIn    time.Duration `json:"expires_in"`
}

type AuthService struct {
repo *repository.UserRepository
cfg  *config.Config
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
user, err := s.repo.FindByEmail(ctx, email)
if err != nil {
return nil, fmt.Errorf("find user: %w", err)
}
if user == nil {
return nil, ErrInvalidCredentials
}
if !user.IsActive {
return nil, ErrUserInactive
}

if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
return nil, ErrInvalidCredentials
}

accessToken, err := s.generateAccessToken(user)
if err != nil {
return nil, fmt.Errorf("generate access token: %w", err)
}

refreshToken := uuid.New().String()
tokenHash := hashToken(refreshToken)
expiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)

if err := s.repo.SaveRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
return nil, fmt.Errorf("save refresh token: %w", err)
}

return &LoginResponse{
AccessToken:  accessToken,
RefreshToken: refreshToken,
ExpiresIn:    s.cfg.AccessTokenTTL,
}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResponse, error) {
tokenHash := hashToken(refreshToken)

userID, err := s.repo.FindRefreshToken(ctx, tokenHash)
if err != nil {
return nil, fmt.Errorf("find refresh token: %w", err)
}
if userID == nil {
return nil, ErrInvalidToken
}

user, err := s.repo.FindByID(ctx, *userID)
if err != nil || user == nil {
return nil, ErrInvalidToken
}

if err := s.repo.DeleteRefreshToken(ctx, tokenHash); err != nil {
return nil, fmt.Errorf("delete old token: %w", err)
}

accessToken, err := s.generateAccessToken(user)
if err != nil {
return nil, fmt.Errorf("generate access token: %w", err)
}

newRefreshToken := uuid.New().String()
newTokenHash := hashToken(newRefreshToken)
expiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)

if err := s.repo.SaveRefreshToken(ctx, user.ID, newTokenHash, expiresAt); err != nil {
return nil, fmt.Errorf("save new refresh token: %w", err)
}

return &LoginResponse{
AccessToken:  accessToken,
RefreshToken: newRefreshToken,
ExpiresIn:    s.cfg.AccessTokenTTL,
}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
tokenHash := hashToken(refreshToken)
return s.repo.DeleteRefreshToken(ctx, tokenHash)
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*Claims, error) {
token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
}
return []byte(s.cfg.JWTSecret), nil
})
if err != nil || !token.Valid {
return nil, ErrInvalidToken
}

claims, ok := token.Claims.(*Claims)
if !ok {
return nil, ErrInvalidToken
}
return claims, nil
}

func (s *AuthService) generateAccessToken(user *repository.User) (string, error) {
claims := Claims{
RegisteredClaims: jwt.RegisteredClaims{
ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AccessTokenTTL)),
IssuedAt:  jwt.NewNumericDate(time.Now()),
Subject:   user.ID.String(),
},
UserID: user.ID.String(),
Email:  user.Email,
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return token.SignedString([]byte(s.cfg.JWTSecret))
}

func hashToken(token string) string {
h := sha256.Sum256([]byte(token))
return hex.EncodeToString(h[:])
}
