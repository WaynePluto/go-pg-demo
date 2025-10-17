package pkgs

// Permission 定义权限项
type Permission struct {
	Key         string
	Description string
}

// Permissions 所有权限常量定义
var Permissions = struct {
	// 用户管理权限
	UserCreate Permission
	UserUpdate Permission
	UserDelete Permission
	UserView   Permission
	UserList   Permission
	
	// 角色管理权限
	RoleCreate Permission
	RoleUpdate Permission
	RoleDelete Permission
	RoleView   Permission
	RoleList   Permission
	
	// 权限分配权限
	RoleAssignPermission   Permission
	RoleRevokePermission   Permission
	UserAssignRole         Permission
	UserRemoveRole         Permission
	
	// 用户自身权限
	UserChangePassword Permission
	UserUpdateProfile  Permission
	
	// 权限查看权限
	UserViewPermissions Permission
}{
	UserCreate: Permission{
		Key:         "user:create",
		Description: "创建用户",
	},
	UserUpdate: Permission{
		Key:         "user:update",
		Description: "更新用户",
	},
	UserDelete: Permission{
		Key:         "user:delete",
		Description: "删除用户",
	},
	UserView: Permission{
		Key:         "user:view",
		Description: "查看用户详情",
	},
	UserList: Permission{
		Key:         "user:list",
		Description: "查看用户列表",
	},
	RoleCreate: Permission{
		Key:         "role:create",
		Description: "创建角色",
	},
	RoleUpdate: Permission{
		Key:         "role:update",
		Description: "更新角色",
	},
	RoleDelete: Permission{
		Key:         "role:delete",
		Description: "删除角色",
	},
	RoleView: Permission{
		Key:         "role:view",
		Description: "查看角色详情",
	},
	RoleList: Permission{
		Key:         "role:list",
		Description: "查看角色列表",
	},
	RoleAssignPermission: Permission{
		Key:         "role:assign_permission",
		Description: "为角色分配权限",
	},
	RoleRevokePermission: Permission{
		Key:         "role:revoke_permission",
		Description: "移除角色权限",
	},
	UserAssignRole: Permission{
		Key:         "user:assign_role",
		Description: "为用户分配角色",
	},
	UserRemoveRole: Permission{
		Key:         "user:remove_role",
		Description: "移除用户角色",
	},
	UserChangePassword: Permission{
		Key:         "user:change_password",
		Description: "修改用户密码",
	},
	UserUpdateProfile: Permission{
		Key:         "user:update_profile",
		Description: "更新用户资料",
	},
	UserViewPermissions: Permission{
		Key:         "user:view_permissions",
		Description: "查看用户权限",
	},
}

// GetAllPermissions 获取所有权限列表
func GetAllPermissions() []Permission {
	return []Permission{
		Permissions.UserCreate,
		Permissions.UserUpdate,
		Permissions.UserDelete,
		Permissions.UserView,
		Permissions.UserList,
		Permissions.RoleCreate,
		Permissions.RoleUpdate,
		Permissions.RoleDelete,
		Permissions.RoleView,
		Permissions.RoleList,
		Permissions.RoleAssignPermission,
		Permissions.RoleRevokePermission,
		Permissions.UserAssignRole,
		Permissions.UserRemoveRole,
		Permissions.UserChangePassword,
		Permissions.UserUpdateProfile,
		Permissions.UserViewPermissions,
	}
}