package user

import (
	"strconv"
	"strings"

	domainuser "erp-system/internal/domain/user"
	httputil "erp-system/pkg/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	userRepo domainuser.Repository
	validate *validator.Validate
}

func NewHandler(userRepo domainuser.Repository) *Handler {
	return &Handler{
		userRepo: userRepo,
		validate: validator.New(),
	}
}

// GET /api/v1/users
func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit > 100 {
		limit = 100
	}

	users, total, err := h.userRepo.List(c.UserContext(), limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list users")
	}

	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}

	return httputil.Success(c, "Users retrieved", ListUsersResponse{
		Users:  resp,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// GET /api/v1/users/:id
func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid user ID")
	}

	user, err := h.userRepo.GetByID(c.UserContext(), uint(id))
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get user")
	}
	if user == nil {
		return httputil.NotFound(c, "User not found")
	}

	return httputil.Success(c, "User retrieved", toUserResponse(user))
}

// PUT /api/v1/users/:id
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid user ID")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	user, err := h.userRepo.GetByID(c.UserContext(), uint(id))
	if err != nil || user == nil {
		return httputil.NotFound(c, "User not found")
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := h.userRepo.Update(c.UserContext(), user); err != nil {
		return httputil.InternalServerError(c, "Failed to update user")
	}

	return httputil.Success(c, "User updated", toUserResponse(user))
}

// DELETE /api/v1/users/:id
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid user ID")
	}

	callerID, _ := c.Locals("user_id").(uint)
	if callerID == uint(id) {
		return httputil.BadRequest(c, "Cannot delete your own account")
	}

	user, err := h.userRepo.GetByID(c.UserContext(), uint(id))
	if err != nil || user == nil {
		return httputil.NotFound(c, "User not found")
	}

	if err := h.userRepo.Delete(c.UserContext(), uint(id)); err != nil {
		return httputil.InternalServerError(c, "Failed to delete user")
	}

	return httputil.Success(c, "User deleted", nil)
}

// PUT /api/v1/profile  (own profile)
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	callerID, _ := c.Locals("user_id").(uint)

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	user, err := h.userRepo.GetByID(c.UserContext(), callerID)
	if err != nil || user == nil {
		return httputil.InternalServerError(c, "Failed to get user")
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := h.userRepo.Update(c.UserContext(), user); err != nil {
		return httputil.InternalServerError(c, "Failed to update profile")
	}

	return httputil.Success(c, "Profile updated", toUserResponse(user))
}

// PUT /api/v1/password  (own password)
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	callerID, _ := c.Locals("user_id").(uint)

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	user, err := h.userRepo.GetByID(c.UserContext(), callerID)
	if err != nil || user == nil {
		return httputil.InternalServerError(c, "Failed to get user")
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return httputil.BadRequest(c, "Current password is incorrect")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to process password")
	}

	if err := h.userRepo.ChangePassword(c.UserContext(), callerID, string(hashed)); err != nil {
		return httputil.InternalServerError(c, "Failed to change password")
	}

	return httputil.Success(c, "Password changed successfully", nil)
}

func toUserResponse(u *domainuser.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		Status:    u.Status,
	}
}

func validateStruct(v *validator.Validate, s interface{}) map[string]interface{} {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	fields := make(map[string]interface{})
	for _, e := range err.(validator.ValidationErrors) {
		fields[strings.ToLower(e.Field())] = e.Tag()
	}
	return fields
}
