package service

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// PermissionService 权限服务
type PermissionService struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(db *sqlx.DB, logger *zap.Logger) *PermissionService {
	return &PermissionService{
		db:     db,
		logger: logger,
	}
}

// CalculateEffectivePermissionsForUser 计算用户的有效权限集
// 工作逻辑：加载用户所有 assignedRoleIds 对应的 Role 聚合，将其 permissions 集合合并、去重后返回
func (s *PermissionService) CalculateEffectivePermissionsForUser(userID string) ([]string, error) {
	// 查询用户拥有的所有角色ID
	var roleIDs []string
	query := `SELECT role_id FROM iacc_user_role WHERE user_id = $1`
	err := s.db.Select(&roleIDs, query, userID)
	if err != nil {
		s.logger.Error("Failed to query user roles", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// 如果用户没有任何角色，返回空权限集
	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	// 查询这些角色对应的权限集
	query = `SELECT permissions FROM iacc_role WHERE id = ANY($1) AND permissions IS NOT NULL`
	rows, err := s.db.Query(query, roleIDs)
	if err != nil {
		s.logger.Error("Failed to query roles permissions", zap.Strings("role_ids", roleIDs), zap.Error(err))
		return nil, fmt.Errorf("failed to query roles permissions: %w", err)
	}
	defer rows.Close()

	// 合并所有权限并去重
	permissionSet := make(map[string]struct{})
	for rows.Next() {
		var permissionsJSON sql.NullString
		if err := rows.Scan(&permissionsJSON); err != nil {
			s.logger.Error("Failed to scan role permissions", zap.Error(err))
			return nil, fmt.Errorf("failed to scan role permissions: %w", err)
		}

		// 解析JSON格式的权限数组
		if permissionsJSON.Valid {
			var permissions []string
			if err := json.Unmarshal([]byte(permissionsJSON.String), &permissions); err != nil {
				s.logger.Error("Failed to unmarshal role permissions", zap.String("permissions_json", permissionsJSON.String), zap.Error(err))
				return nil, fmt.Errorf("failed to unmarshal role permissions: %w", err)
			}

			// 添加到权限集合中
			for _, perm := range permissions {
				if perm != "" {
					permissionSet[perm] = struct{}{}
				}
			}
		}
	}

	// 检查迭代过程中是否有错误
	if err = rows.Err(); err != nil {
		s.logger.Error("Error iterating role permissions", zap.Error(err))
		return nil, fmt.Errorf("error iterating role permissions: %w", err)
	}

	// 将权限集合转换为切片
	effectivePermissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		effectivePermissions = append(effectivePermissions, perm)
	}

	return effectivePermissions, nil
}