package rbac

import "context"

// Enforcer abstracts the Casbin enforcement logic so it can be mocked in tests.
type Enforcer interface {
	// Enforce returns true if the subject (role) has permission to perform action on resource.
	Enforce(ctx context.Context, subject, resource, action string) (bool, error)

	// AddPolicy adds a role-permission policy.
	AddPolicy(role, resource, action string) error

	// RemovePolicy removes a role-permission policy.
	RemovePolicy(role, resource, action string) error

	// AddRoleForUser assigns a role to a user (subject).
	AddRoleForUser(user, role string) error

	// RemoveRoleForUser removes a role from a user.
	RemoveRoleForUser(user, role string) error

	// GetRolesForUser returns all roles assigned to a user.
	GetRolesForUser(user string) ([]string, error)

	// LoadPolicy reloads all policies from storage.
	LoadPolicy() error
}
