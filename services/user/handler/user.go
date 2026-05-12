package handler

import (
"net/http"

"github.com/anurinth-w/conduit-th/services/user/service"
"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

type UserHandler struct {
svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
return &UserHandler{svc: svc}
}

type createUserRequest struct {
Email    string `json:"email"    binding:"required,email"`
Password string `json:"password" binding:"required,min=6"`
Name     string `json:"name"     binding:"required"`
Phone    string `json:"phone"`
}

type updateUserRequest struct {
Name  string `json:"name"  binding:"required"`
Phone string `json:"phone"`
}

type addMembershipRequest struct {
UserID    string   `json:"user_id"    binding:"required"`
CompanyID string   `json:"company_id" binding:"required"`
Role      string   `json:"role"       binding:"required,oneof=admin manager office technician"`
Scope     []string `json:"scope"`
}

func (h *UserHandler) Create(c *gin.Context) {
var req createUserRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

user, err := h.svc.Create(c.Request.Context(), service.CreateUserInput{
Email:    req.Email,
Password: req.Password,
Name:     req.Name,
Phone:    req.Phone,
})
if err != nil {
switch err {
case service.ErrEmailTaken:
c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
default:
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
return
}

c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetByID(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
return
}

user, err := h.svc.GetByID(c.Request.Context(), id)
if err != nil {
switch err {
case service.ErrUserNotFound:
c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
default:
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
return
}

c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Update(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
return
}

var req updateUserRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

user, err := h.svc.Update(c.Request.Context(), id, req.Name, req.Phone)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Deactivate(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
return
}

if err := h.svc.Deactivate(c.Request.Context(), id); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "user deactivated"})
}

func (h *UserHandler) AddMembership(c *gin.Context) {
var req addMembershipRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

userID, err := uuid.Parse(req.UserID)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
return
}

companyID, err := uuid.Parse(req.CompanyID)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
return
}

m, err := h.svc.AddMembership(c.Request.Context(), service.AddMembershipInput{
UserID:    userID,
CompanyID: companyID,
Role:      req.Role,
Scope:     req.Scope,
})
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusCreated, m)
}

func (h *UserHandler) GetMemberships(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
return
}

memberships, err := h.svc.GetMemberships(c.Request.Context(), id)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusOK, memberships)
}

func (h *UserHandler) ListByCompany(c *gin.Context) {
companyID, err := uuid.Parse(c.Param("company_id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
return
}

users, err := h.svc.ListByCompany(c.Request.Context(), companyID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
return
}

c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Health(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
