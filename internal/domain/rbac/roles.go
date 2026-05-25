package rbac

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleManager    = "manager"
	RoleUser       = "user"
	RoleGuest      = "guest"
)

// DefaultRolePermissions maps each role to its default permissions.
// super_admin uses wildcard "*" — the seeder assigns all permissions.
var DefaultRolePermissions = map[string][]string{
	RoleSuperAdmin: {"*"},
	RoleAdmin: {
		PermUsersCreate, PermUsersRead, PermUsersUpdate, PermUsersDelete,
		PermRolesCreate, PermRolesRead, PermRolesUpdate, PermRolesDelete,
		PermPermissionsRead,
		PermModulesRead, PermModulesManage,
		PermAuditRead,
		PermFeatureFlagsRead, PermFeatureFlagsManage,
		PermMasterDataRead, PermMasterDataWrite, PermMasterDataDelete,
		PermProcurementRead, PermProcurementWrite, PermProcurementDelete, PermProcurementApprove,
	},
	RoleManager: {
		PermUsersRead, PermUsersUpdate,
		PermRolesRead,
		PermPermissionsRead,
		PermModulesRead,
		PermFeatureFlagsRead,
		PermMasterDataRead, PermMasterDataWrite,
		PermProcurementRead, PermProcurementWrite, PermProcurementApprove,
	},
	RoleUser: {
		PermUsersRead,
		PermMasterDataRead,
		PermProcurementRead, PermProcurementWrite,
	},
	RoleGuest: {
		PermMasterDataRead,
		PermProcurementRead,
	},
}

func IsValidRole(role string) bool {
	switch role {
	case RoleSuperAdmin, RoleAdmin, RoleManager, RoleUser, RoleGuest:
		return true
	}
	return false
}
