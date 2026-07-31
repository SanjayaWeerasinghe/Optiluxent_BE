package hr

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	// Family
	ListFamily(ctx context.Context, tenantID, employeeID uint) ([]Family, error)
	AddFamily(ctx context.Context, f *Family) error
	UpdateFamily(ctx context.Context, f *Family) error
	DeleteFamily(ctx context.Context, tenantID, id uint) error
	GetFamily(ctx context.Context, tenantID, id uint) (*Family, error)

	// Emergency
	ListEmergency(ctx context.Context, tenantID, employeeID uint) ([]EmergencyContact, error)
	AddEmergency(ctx context.Context, e *EmergencyContact) error
	UpdateEmergency(ctx context.Context, e *EmergencyContact) error
	DeleteEmergency(ctx context.Context, tenantID, id uint) error
	GetEmergency(ctx context.Context, tenantID, id uint) (*EmergencyContact, error)

	// Attendance — one row per (employee, date); Upsert handles conflict.
	// `limit=0` = unbounded. Per-employee tabs still fetch everything for
	// the small volumes they show; the cross-employee list uses pagination.
	ListAttendance(ctx context.Context, tenantID uint, employeeID uint, from, to string, limit, offset int) ([]Attendance, error)
	CountAttendance(ctx context.Context, tenantID uint, employeeID uint, from, to string) (int64, error)
	UpsertAttendance(ctx context.Context, a *Attendance) error
	DeleteAttendance(ctx context.Context, tenantID, id uint) error

	// Salary history
	ListSalaryHistory(ctx context.Context, tenantID, employeeID uint) ([]SalaryHistory, error)
	AddSalaryRevision(ctx context.Context, s *SalaryHistory) error

	// Employee — thin passthrough for the salary revision (which
	// bumps employees.basic_salary as a side-effect).
	UpdateEmployeeSalary(ctx context.Context, tenantID, employeeID uint, salary float64) error
	SetEmployeeCVUrl(ctx context.Context, tenantID, employeeID uint, url string) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

// ── Family ──────────────────────────────────────────────────────────────────

func (r *dbRepository) ListFamily(ctx context.Context, tenantID, employeeID uint) ([]Family, error) {
	var rows []Family
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND employee_id = ?", tenantID, employeeID).
		Order("id").Find(&rows).Error
}

func (r *dbRepository) GetFamily(ctx context.Context, tenantID, id uint) (*Family, error) {
	var f Family
	return &f, r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&f).Error
}

func (r *dbRepository) AddFamily(ctx context.Context, f *Family) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *dbRepository) UpdateFamily(ctx context.Context, f *Family) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *dbRepository) DeleteFamily(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Family{}).Error
}

// ── Emergency ───────────────────────────────────────────────────────────────

func (r *dbRepository) ListEmergency(ctx context.Context, tenantID, employeeID uint) ([]EmergencyContact, error) {
	var rows []EmergencyContact
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND employee_id = ?", tenantID, employeeID).
		Order("is_primary DESC, id").Find(&rows).Error
}

func (r *dbRepository) GetEmergency(ctx context.Context, tenantID, id uint) (*EmergencyContact, error) {
	var e EmergencyContact
	return &e, r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&e).Error
}

func (r *dbRepository) AddEmergency(ctx context.Context, e *EmergencyContact) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *dbRepository) UpdateEmergency(ctx context.Context, e *EmergencyContact) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *dbRepository) DeleteEmergency(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&EmergencyContact{}).Error
}

// ── Attendance ──────────────────────────────────────────────────────────────

// attendanceFilterWhere — shared filter body for List + Count.
func attendanceFilterWhere(q *gorm.DB, tenantID, employeeID uint, from, to string) *gorm.DB {
	q = q.Where("tenant_id = ?", tenantID)
	if employeeID > 0 {
		q = q.Where("employee_id = ?", employeeID)
	}
	if from != "" {
		q = q.Where("attend_date >= ?", from)
	}
	if to != "" {
		q = q.Where("attend_date <= ?", to)
	}
	return q
}

func (r *dbRepository) ListAttendance(ctx context.Context, tenantID, employeeID uint, from, to string, limit, offset int) ([]Attendance, error) {
	q := attendanceFilterWhere(r.db.WithContext(ctx), tenantID, employeeID, from, to).
		Order("attend_date DESC, employee_id")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []Attendance
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountAttendance(ctx context.Context, tenantID, employeeID uint, from, to string) (int64, error) {
	var n int64
	return n, attendanceFilterWhere(r.db.WithContext(ctx).Model(&Attendance{}), tenantID, employeeID, from, to).Count(&n).Error
}

// UpsertAttendance — one row per (employee, date). Uses the DB unique
// constraint uidx_attendance_emp_date so callers can freely POST for the
// same day (edit vs new) and get "last write wins" semantics.
func (r *dbRepository) UpsertAttendance(ctx context.Context, a *Attendance) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO attendance (tenant_id, employee_id, attend_date, check_in, check_out, status, hours_worked, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT ON CONSTRAINT uidx_attendance_emp_date DO UPDATE SET
			check_in     = EXCLUDED.check_in,
			check_out    = EXCLUDED.check_out,
			status       = EXCLUDED.status,
			hours_worked = EXCLUDED.hours_worked,
			notes        = EXCLUDED.notes,
			updated_at   = NOW()`,
		a.TenantID, a.EmployeeID, a.AttendDate, a.CheckIn, a.CheckOut,
		a.Status, a.HoursWorked, a.Notes,
	).Error
}

func (r *dbRepository) DeleteAttendance(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Attendance{}).Error
}

// ── Salary history ──────────────────────────────────────────────────────────

func (r *dbRepository) ListSalaryHistory(ctx context.Context, tenantID, employeeID uint) ([]SalaryHistory, error) {
	var rows []SalaryHistory
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND employee_id = ?", tenantID, employeeID).
		Order("effective_date DESC, id DESC").Find(&rows).Error
}

func (r *dbRepository) AddSalaryRevision(ctx context.Context, s *SalaryHistory) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// ── Employee shims (bypass the masterdata module to avoid circular deps) ───

func (r *dbRepository) UpdateEmployeeSalary(ctx context.Context, tenantID, employeeID uint, salary float64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE employees SET basic_salary = ?, updated_at = NOW()
		WHERE tenant_id = ? AND id = ?`,
		salary, tenantID, employeeID,
	).Error
}

func (r *dbRepository) SetEmployeeCVUrl(ctx context.Context, tenantID, employeeID uint, url string) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE employees SET cv_url = ?, updated_at = NOW()
		WHERE tenant_id = ? AND id = ?`,
		url, tenantID, employeeID,
	).Error
}
