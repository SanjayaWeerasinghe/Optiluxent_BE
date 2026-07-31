package hr

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	httputil "erp-system/pkg/http"
)

// UploadDir — root folder on the API container's filesystem where
// employee CVs (and any future HR docs) are stored. Mounted as a Docker
// volume in dev so restarts don't wipe uploads. Served back to the FE
// via Fiber's Static handler mounted at /uploads.
var UploadDir = "/app/uploads"

// EnsureUploadDir returns UploadDir, creating the folder tree if needed.
// Called from module.Initialize once per boot.
func EnsureUploadDir() (string, error) {
	if err := os.MkdirAll(UploadDir, 0o755); err != nil {
		return "", err
	}
	return UploadDir, nil
}

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

func tenantFromCtx(c *fiber.Ctx) uint {
	v, _ := c.Locals("tenant_id").(uint)
	return v
}

func userFromCtx(c *fiber.Ctx) uint {
	v, _ := c.Locals("user_id").(uint)
	return v
}

func validateStruct(v *validator.Validate, s interface{}) map[string]interface{} {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	errs := map[string]interface{}{}
	for _, e := range err.(validator.ValidationErrors) {
		errs[e.Field()] = e.Tag()
	}
	return errs
}

func parseID(c *fiber.Ctx, name string) (uint, error) {
	id, err := strconv.ParseUint(c.Params(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ── Family ──────────────────────────────────────────────────────────────────

func (h *Handler) ListFamily(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	rows, err := h.svc.ListFamily(c.Context(), tenantFromCtx(c), employeeID)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "family retrieved", rows)
}

func (h *Handler) AddFamily(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	var req UpsertFamilyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	f, err := h.svc.AddFamily(c.Context(), tenantFromCtx(c), employeeID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "family member added", f)
}

func (h *Handler) UpdateFamily(c *fiber.Ctx) error {
	id, err := parseID(c, "familyId")
	if err != nil {
		return httputil.BadRequest(c, "invalid family id")
	}
	var req UpsertFamilyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	f, err := h.svc.UpdateFamily(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "family member updated", f)
}

func (h *Handler) DeleteFamily(c *fiber.Ctx) error {
	id, err := parseID(c, "familyId")
	if err != nil {
		return httputil.BadRequest(c, "invalid family id")
	}
	if err := h.svc.DeleteFamily(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Emergency ───────────────────────────────────────────────────────────────

func (h *Handler) ListEmergency(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	rows, err := h.svc.ListEmergency(c.Context(), tenantFromCtx(c), employeeID)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "emergency contacts retrieved", rows)
}

func (h *Handler) AddEmergency(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	var req UpsertEmergencyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	e, err := h.svc.AddEmergency(c.Context(), tenantFromCtx(c), employeeID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "emergency contact added", e)
}

func (h *Handler) UpdateEmergency(c *fiber.Ctx) error {
	id, err := parseID(c, "contactId")
	if err != nil {
		return httputil.BadRequest(c, "invalid contact id")
	}
	var req UpsertEmergencyRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	e, err := h.svc.UpdateEmergency(c.Context(), tenantFromCtx(c), id, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Success(c, "emergency contact updated", e)
}

func (h *Handler) DeleteEmergency(c *fiber.Ctx) error {
	id, err := parseID(c, "contactId")
	if err != nil {
		return httputil.BadRequest(c, "invalid contact id")
	}
	if err := h.svc.DeleteEmergency(c.Context(), tenantFromCtx(c), id); err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.NoContent(c)
}

// ── Attendance ──────────────────────────────────────────────────────────────

func (h *Handler) ListAttendance(c *fiber.Ctx) error {
	tenantID := tenantFromCtx(c)
	employeeID, _ := strconv.ParseUint(c.Query("employee_id", "0"), 10, 64)
	from := c.Query("from")
	to := c.Query("to")
	page, perPage, limit, offset := httputil.ParsePage(c)
	total, err := h.svc.CountAttendance(c.Context(), tenantID, uint(employeeID), from, to)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	rows, err := h.svc.ListAttendance(c.Context(), tenantID, uint(employeeID), from, to, limit, offset)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.SuccessWithMeta(c, "attendance retrieved", rows, httputil.Paginate(page, perPage, int(total)))
}

// Per-employee tab — unbounded (fits inside the modal). No pager here.
func (h *Handler) ListAttendanceForEmployee(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	from := c.Query("from")
	to := c.Query("to")
	rows, err := h.svc.ListAttendance(c.Context(), tenantFromCtx(c), employeeID, from, to, 0, 0)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "attendance retrieved", rows)
}

func (h *Handler) UpsertAttendance(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	var req UpsertAttendanceRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	a, err := h.svc.UpsertAttendance(c.Context(), tenantFromCtx(c), employeeID, &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "attendance recorded", a)
}

// ── Salary revisions ────────────────────────────────────────────────────────

func (h *Handler) ListSalaryHistory(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	rows, err := h.svc.ListSalaryHistory(c.Context(), tenantFromCtx(c), employeeID)
	if err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "salary history retrieved", rows)
}

func (h *Handler) AddSalaryRevision(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	var req CreateSalaryRevisionRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "invalid request body")
	}
	if errs := validateStruct(h.validate, req); errs != nil {
		return httputil.ValidationError(c, "validation failed", errs)
	}
	sh, err := h.svc.AddSalaryRevision(c.Context(), tenantFromCtx(c), employeeID, userFromCtx(c), &req)
	if err != nil {
		return httputil.BadRequest(c, err.Error())
	}
	return httputil.Created(c, "salary revision recorded", sh)
}

// ── CV upload ───────────────────────────────────────────────────────────────

// UploadCV stores the uploaded file at
// UploadDir/hr/emp-<id>/cv<ext> and updates employees.cv_url to the
// public path (/uploads/hr/emp-<id>/cv<ext>). Overwrites on re-upload.
func (h *Handler) UploadCV(c *fiber.Ctx) error {
	employeeID, err := parseID(c, "id")
	if err != nil {
		return httputil.BadRequest(c, "invalid employee id")
	}
	file, err := c.FormFile("file")
	if err != nil {
		return httputil.BadRequest(c, "expected a 'file' form field")
	}
	// Guard file size — 5MB is plenty for a CV. Fiber's default body limit
	// still applies at the server level too.
	const maxBytes = 5 * 1024 * 1024
	if file.Size > maxBytes {
		return httputil.BadRequest(c, fmt.Sprintf("file too large (%.2f MB max)", float64(maxBytes)/1024/1024))
	}
	// Keep the original extension so browsers know how to open it.
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".pdf"
	}
	empDir := filepath.Join(UploadDir, "hr", fmt.Sprintf("emp-%d", employeeID))
	if err := os.MkdirAll(empDir, 0o755); err != nil {
		return httputil.InternalServerError(c, fmt.Sprintf("failed to create upload dir: %s", err.Error()))
	}
	dst := filepath.Join(empDir, "cv"+ext)
	if err := c.SaveFile(file, dst); err != nil {
		return httputil.InternalServerError(c, fmt.Sprintf("save failed: %s", err.Error()))
	}
	publicPath := fmt.Sprintf("/uploads/hr/emp-%d/cv%s", employeeID, ext)
	if err := h.svc.SetEmployeeCVUrl(c.Context(), tenantFromCtx(c), employeeID, publicPath); err != nil {
		return httputil.InternalServerError(c, err.Error())
	}
	return httputil.Success(c, "CV uploaded", fiber.Map{"cv_url": publicPath})
}
