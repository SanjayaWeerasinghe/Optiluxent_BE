package role

import (
	"strconv"
	"strings"

	domainperm "erp-system/internal/domain/permission"
	domainrole "erp-system/internal/domain/role"
	httputil "erp-system/pkg/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	roleRepo domainrole.Repository
	permRepo domainperm.Repository
	validate *validator.Validate
}

func NewHandler(roleRepo domainrole.Repository, permRepo domainperm.Repository) *Handler {
	return &Handler{
		roleRepo: roleRepo,
		permRepo: permRepo,
		validate: validator.New(),
	}
}

// GET /api/v1/roles
func (h *Handler) List(c *fiber.Ctx) error {
	tenantID := tenantIDFromCtx(c)
	roles, err := h.roleRepo.List(c.UserContext(), tenantID)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list roles")
	}
	resp := make([]RoleResponse, len(roles))
	for i, r := range roles {
		resp[i] = toRoleResponse(r)
	}
	return httputil.Success(c, "Roles retrieved", resp)
}

// POST /api/v1/roles
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	tenantID := tenantIDFromCtx(c)
	existing, _ := h.roleRepo.GetByName(c.UserContext(), tenantID, req.Name)
	if existing != nil {
		return httputil.Conflict(c, "A role with this name already exists")
	}

	role := &domainrole.Role{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := h.roleRepo.Create(c.UserContext(), role); err != nil {
		return httputil.InternalServerError(c, "Failed to create role")
	}

	return httputil.Created(c, "Role created", toRoleResponse(role))
}

// GET /api/v1/roles/:id
func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return httputil.BadRequest(c, "Invalid role ID")
	}

	role, err := h.roleRepo.GetByID(c.UserContext(), id)
	if err != nil {
		return httputil.InternalServerError(c, "Failed to get role")
	}
	if role == nil {
		return httputil.NotFound(c, "Role not found")
	}

	return httputil.Success(c, "Role retrieved", toRoleResponse(role))
}

// PUT /api/v1/roles/:id
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return httputil.BadRequest(c, "Invalid role ID")
	}

	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	role, err := h.roleRepo.GetByID(c.UserContext(), id)
	if err != nil || role == nil {
		return httputil.NotFound(c, "Role not found")
	}
	if role.IsSystem {
		return httputil.BadRequest(c, "Cannot modify system roles")
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := h.roleRepo.Update(c.UserContext(), role); err != nil {
		return httputil.InternalServerError(c, "Failed to update role")
	}

	return httputil.Success(c, "Role updated", toRoleResponse(role))
}

// DELETE /api/v1/roles/:id
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return httputil.BadRequest(c, "Invalid role ID")
	}

	role, err := h.roleRepo.GetByID(c.UserContext(), id)
	if err != nil || role == nil {
		return httputil.NotFound(c, "Role not found")
	}
	if role.IsSystem {
		return httputil.BadRequest(c, "Cannot delete system roles")
	}

	if err := h.roleRepo.Delete(c.UserContext(), id); err != nil {
		return httputil.InternalServerError(c, "Failed to delete role")
	}

	return httputil.Success(c, "Role deleted", nil)
}

// POST /api/v1/roles/:id/permissions
func (h *Handler) AssignPermissions(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return httputil.BadRequest(c, "Invalid role ID")
	}

	var req AssignPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}
	if fields := validateStruct(h.validate, req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	if err := h.roleRepo.AssignPermissions(c.UserContext(), id, req.PermissionIDs); err != nil {
		return httputil.InternalServerError(c, "Failed to assign permissions")
	}

	return httputil.Success(c, "Permissions assigned", nil)
}

// DELETE /api/v1/roles/:id/permissions/:perm_id
func (h *Handler) RemovePermission(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return httputil.BadRequest(c, "Invalid role ID")
	}
	permID, err := strconv.ParseUint(c.Params("perm_id"), 10, 64)
	if err != nil {
		return httputil.BadRequest(c, "Invalid permission ID")
	}

	if err := h.roleRepo.RemovePermission(c.UserContext(), id, uint(permID)); err != nil {
		return httputil.InternalServerError(c, "Failed to remove permission")
	}

	return httputil.Success(c, "Permission removed", nil)
}

// GET /api/v1/permissions
func (h *Handler) ListPermissions(c *fiber.Ctx) error {
	perms, err := h.permRepo.List(c.UserContext())
	if err != nil {
		return httputil.InternalServerError(c, "Failed to list permissions")
	}
	resp := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		resp[i] = PermissionResponse{
			ID:          p.ID,
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
		}
	}
	return httputil.Success(c, "Permissions retrieved", resp)
}

// helpers

func parseID(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	return uint(id), err
}

func tenantIDFromCtx(c *fiber.Ctx) *uint {
	if tid, ok := c.Locals("tenant_id").(uint); ok && tid != 0 {
		return &tid
	}
	return nil
}

func toRoleResponse(r *domainrole.Role) RoleResponse {
	resp := RoleResponse{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
	}
	for _, p := range r.Permissions {
		resp.Permissions = append(resp.Permissions, PermissionResponse{
			ID:          p.ID,
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
		})
	}
	return resp
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
