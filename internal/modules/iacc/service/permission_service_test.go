package service

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go-pg-demo/pkgs"
)

// setupTestDBAndService 初始化测试数据库和服务
func setupTestDBAndService(t *testing.T) (*PermissionService, *sqlx.DB) {
	// 加载配置
	cfg, err := pkgs.NewConfig()
	require.NoError(t, err)

	// 初始化日志
	logger, err := pkgs.NewLogger(cfg)
	require.NoError(t, err)

	// 初始化数据库
	db, err := pkgs.NewConnection(cfg)
	require.NoError(t, err)

	// 创建服务实例
	permissionService := NewPermissionService(db, logger)

	return permissionService, db
}

func TestPermissionService_CalculateEffectivePermissionsForUser(t *testing.T) {
	s, sqlxDB := setupTestDBAndService(t)

	// 定义一个辅助函数来清理数据
	cleanup := func(userID string, roleIDs []string) {
		if userID != "" {
			_, err := sqlxDB.Exec("DELETE FROM iacc_user_role WHERE user_id = $1", userID)
			if err != nil {
				zap.L().Error("failed to cleanup user_role", zap.Error(err))
			}
			_, err = sqlxDB.Exec("DELETE FROM iacc_user WHERE id = $1", userID)
			if err != nil {
				zap.L().Error("failed to cleanup user", zap.Error(err))
			}
		}
		if len(roleIDs) > 0 {
			query, args, err := sqlx.In("DELETE FROM iacc_role WHERE id IN (?)", roleIDs)
			if err != nil {
				zap.L().Error("failed to create IN query for role cleanup", zap.Error(err))
				return
			}
			query = sqlxDB.Rebind(query)
			_, err = sqlxDB.Exec(query, args...)
			if err != nil {
				zap.L().Error("failed to cleanup role", zap.Error(err))
			}
		}
	}

	t.Run("正常用例：用户有多个角色，权限合并去重", func(t *testing.T) {
		// Arrange
		// 1. 创建用户
		userID := uuid.New().String()
		_, err := sqlxDB.Exec("INSERT INTO iacc_user (id, username, password) VALUES ($1, $2, $3)", userID, "testuser_perm", "password")
		require.NoError(t, err)

		// 2. 创建角色和权限
		role1ID := uuid.New().String()
		perms1, _ := json.Marshal([]string{"perm:read", "perm:write"})
		_, err = sqlxDB.Exec("INSERT INTO iacc_role (id, name, permissions) VALUES ($1, $2, $3)", role1ID, "RoleA", perms1)
		require.NoError(t, err)

		role2ID := uuid.New().String()
		perms2, _ := json.Marshal([]string{"perm:write", "perm:delete"})
		_, err = sqlxDB.Exec("INSERT INTO iacc_role (id, name, permissions) VALUES ($1, $2, $3)", role2ID, "RoleB", perms2)
		require.NoError(t, err)

		// 3. 关联用户和角色
		_, err = sqlxDB.Exec("INSERT INTO iacc_user_role (user_id, role_id) VALUES ($1, $2), ($1, $3)", userID, role1ID, role2ID)
		require.NoError(t, err)

		// 注册清理函数
		t.Cleanup(func() {
			cleanup(userID, []string{role1ID, role2ID})
		})

		// Act
		effectivePermissions, err := s.CalculateEffectivePermissionsForUser(userID)
		require.NoError(t, err)

		// Assert
		expectedPermissions := []string{"perm:read", "perm:write", "perm:delete"}
		sort.Strings(effectivePermissions)
		sort.Strings(expectedPermissions)
		assert.Equal(t, expectedPermissions, effectivePermissions)
	})

	t.Run("边缘用例：用户没有角色", func(t *testing.T) {
		// Arrange
		userID := uuid.New().String()
		_, err := sqlxDB.Exec("INSERT INTO iacc_user (id, username, password) VALUES ($1, $2, $3)", userID, "testuser_norole", "password")
		require.NoError(t, err)
		t.Cleanup(func() {
			cleanup(userID, nil)
		})

		// Act
		permissions, err := s.CalculateEffectivePermissionsForUser(userID)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, permissions)
	})

	t.Run("边缘用例：用户角色没有权限", func(t *testing.T) {
		// Arrange
		userID := uuid.New().String()
		_, err := sqlxDB.Exec("INSERT INTO iacc_user (id, username, password) VALUES ($1, $2, $3)", userID, "testuser_nopermrole", "password")
		require.NoError(t, err)

		roleID := uuid.New().String()
		_, err = sqlxDB.Exec("INSERT INTO iacc_role (id, name, permissions) VALUES ($1, $2, $3)", roleID, "RoleC_NoPerms", "[]")
		require.NoError(t, err)

		_, err = sqlxDB.Exec("INSERT INTO iacc_user_role (user_id, role_id) VALUES ($1, $2)", userID, roleID)
		require.NoError(t, err)

		t.Cleanup(func() {
			cleanup(userID, []string{roleID})
		})

		// Act
		permissions, err := s.CalculateEffectivePermissionsForUser(userID)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, permissions)
	})

	t.Run("异常用例：用户ID不存在", func(t *testing.T) {
		// Arrange
		nonExistentUserID := uuid.New().String()

		// Act
		permissions, err := s.CalculateEffectivePermissionsForUser(nonExistentUserID)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, permissions)
	})
}
