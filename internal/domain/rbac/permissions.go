package rbac

// Permission constants follow the format "resource:action".
const (
	// User permissions
	PermUsersCreate = "users:create"
	PermUsersRead   = "users:read"
	PermUsersUpdate = "users:update"
	PermUsersDelete = "users:delete"

	// Role permissions
	PermRolesCreate = "roles:create"
	PermRolesRead   = "roles:read"
	PermRolesUpdate = "roles:update"
	PermRolesDelete = "roles:delete"

	// Permission management
	PermPermissionsRead = "permissions:read"

	// Tenant permissions
	PermTenantsCreate = "tenants:create"
	PermTenantsRead   = "tenants:read"
	PermTenantsUpdate = "tenants:update"
	PermTenantsDelete = "tenants:delete"

	// Module permissions
	PermModulesRead   = "modules:read"
	PermModulesManage = "modules:manage"

	// Audit permissions
	PermAuditRead = "audit:read"

	// Feature flag permissions
	PermFeatureFlagsRead   = "feature_flags:read"
	PermFeatureFlagsManage = "feature_flags:manage"

	// Master data permissions
	PermMasterDataRead   = "masterdata:read"
	PermMasterDataWrite  = "masterdata:write"
	PermMasterDataDelete = "masterdata:delete"

	// Procurement permissions
	PermProcurementRead    = "procurement:read"
	PermProcurementWrite   = "procurement:write"
	PermProcurementDelete  = "procurement:delete"
	PermProcurementApprove = "procurement:approve"

	// Inventory permissions
	PermInventoryRead    = "inventory:read"
	PermInventoryWrite   = "inventory:write"
	PermInventoryDelete  = "inventory:delete"
	PermInventoryApprove = "inventory:approve"
)

// AllPermissions lists every defined permission.
// Used by the seeder to populate the permissions table.
var AllPermissions = []string{
	PermUsersCreate, PermUsersRead, PermUsersUpdate, PermUsersDelete,
	PermRolesCreate, PermRolesRead, PermRolesUpdate, PermRolesDelete,
	PermPermissionsRead,
	PermTenantsCreate, PermTenantsRead, PermTenantsUpdate, PermTenantsDelete,
	PermModulesRead, PermModulesManage,
	PermAuditRead,
	PermFeatureFlagsRead, PermFeatureFlagsManage,
	PermMasterDataRead, PermMasterDataWrite, PermMasterDataDelete,
	PermProcurementRead, PermProcurementWrite, PermProcurementDelete, PermProcurementApprove,
	PermInventoryRead, PermInventoryWrite, PermInventoryDelete, PermInventoryApprove,
}
