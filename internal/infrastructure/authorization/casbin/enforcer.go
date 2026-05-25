package casbin

import (
	"context"
	"fmt"

	casbinv2 "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

// CasbinEnforcer wraps the casbin Enforcer and implements domain/rbac.Enforcer.
type CasbinEnforcer struct {
	e  *casbinv2.Enforcer
	db *gorm.DB
}

// New creates a Casbin enforcer backed by the provided GORM *DB.
// Uses a custom DB adapter built on our own schema, avoiding gorm-adapter
// version conflicts between casbin/v2 and casbin/v3.
func New(db *gorm.DB) (*CasbinEnforcer, error) {
	m, err := model.NewModelFromString(RBACModel)
	if err != nil {
		return nil, fmt.Errorf("casbin: parse model: %w", err)
	}

	adapter := newDBAdapter(db)

	e, err := casbinv2.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("casbin: create enforcer: %w", err)
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("casbin: load policy: %w", err)
	}

	return &CasbinEnforcer{e: e, db: db}, nil
}

func (c *CasbinEnforcer) Enforce(_ context.Context, subject, resource, action string) (bool, error) {
	return c.e.Enforce(subject, resource, action)
}

func (c *CasbinEnforcer) AddPolicy(role, resource, action string) error {
	_, err := c.e.AddPolicy(role, resource, action)
	return err
}

func (c *CasbinEnforcer) RemovePolicy(role, resource, action string) error {
	_, err := c.e.RemovePolicy(role, resource, action)
	return err
}

func (c *CasbinEnforcer) AddRoleForUser(user, role string) error {
	_, err := c.e.AddRoleForUser(user, role)
	return err
}

func (c *CasbinEnforcer) RemoveRoleForUser(user, role string) error {
	_, err := c.e.DeleteRoleForUser(user, role)
	return err
}

func (c *CasbinEnforcer) GetRolesForUser(user string) ([]string, error) {
	return c.e.GetRolesForUser(user)
}

func (c *CasbinEnforcer) LoadPolicy() error {
	return c.e.LoadPolicy()
}

// ─────────────────────────────────────────────────────────────────────────────
// dbAdapter implements casbin/v2 persist.Adapter backed by our schema.
// It loads role→permission policies from role_permissions + user→role mappings
// from user_roles. The Casbin "casbin_rule" table is NOT used.
// ─────────────────────────────────────────────────────────────────────────────

type dbAdapter struct {
	db *gorm.DB
}

func newDBAdapter(db *gorm.DB) persist.Adapter {
	return &dbAdapter{db: db}
}

type rolePermRow struct {
	RoleName string
	Resource string
	Action   string
}

type userRoleRow struct {
	UserID uint
	RoleName string
}

func (a *dbAdapter) LoadPolicy(m model.Model) error {
	// Load role → permission policies (p, role, resource, action)
	var rps []rolePermRow
	err := a.db.Raw(`
		SELECT r.name AS role_name, p.resource, p.action
		FROM role_permissions rp
		JOIN roles r       ON r.id = rp.role_id       AND r.deleted_at IS NULL
		JOIN permissions p ON p.id = rp.permission_id
	`).Scan(&rps).Error
	if err != nil {
		return fmt.Errorf("casbin adapter: load role permissions: %w", err)
	}

	for _, row := range rps {
		persist.LoadPolicyLine(
			fmt.Sprintf("p, %s, %s, %s", row.RoleName, row.Resource, row.Action),
			m,
		)
	}

	// Load user → role groupings (g, user_id, role_name)
	var urs []userRoleRow
	err = a.db.Raw(`
		SELECT ur.user_id, r.name AS role_name
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
	`).Scan(&urs).Error
	if err != nil {
		return fmt.Errorf("casbin adapter: load user roles: %w", err)
	}

	for _, row := range urs {
		persist.LoadPolicyLine(
			fmt.Sprintf("g, %d, %s", row.UserID, row.RoleName),
			m,
		)
	}

	return nil
}

// SavePolicy is not used; we persist via our own role/permission repos.
func (a *dbAdapter) SavePolicy(_ model.Model) error { return nil }

// AddPolicy / RemovePolicy persist via enforcer.AddPolicy but we persist
// changes directly through the role handler, so these are no-ops here.
func (a *dbAdapter) AddPolicy(_ string, _ string, _ []string) error { return nil }
func (a *dbAdapter) RemovePolicy(_ string, _ string, _ []string) error { return nil }
func (a *dbAdapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return nil
}
