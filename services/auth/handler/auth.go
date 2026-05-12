package handler

import (
"net/http"

"github.com/anurinth-w/conduit-th/services/auth/service"
"github.com/gin-gonic/gin"
)

type AuthHandler struct {
svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
return &AuthHandler{svc: svc}
}

type loginRequest struct {
Email    string `json:"email"    binding:"required,email"`
Password string `json:"password" binding:"required,min=6"`
}

type refreshRequest struct {
RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
var req loginRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

resp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
if err != nil {
switch err {
case service.ErrInvalidCredentials:
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
case service.ErrUserInactive:
c.JSON(http.StatusForbidden, gin.H{"error": "account is inactive"})
default:
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
return
}

c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
var req refreshRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

resp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
if err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
return
}

c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
var req logoutRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Health(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
