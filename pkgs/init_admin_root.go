package pkgs

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// InitAdminRoot 初始化 admin 用户和 root 角色
// 1. 检索是否存在username为admin的用户，如果没有则创建
// 2. 检索是否存在name为root的角色，如果没有则创建
// 3. 确保admin用户拥有root角色
// 4. 确保root角色拥有所有的权限
func InitAdminRoot(db *sqlx.DB, logger *zap.Logger) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			logger.Debug("回滚事务")
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Error("提交事务失败", zap.Error(err))
			}
		}
	}()

	// 1. 检索或创建 administrator 用户，密码是md5(123456)
	var adminID string
	err = tx.Get(&adminID, `SELECT id FROM iacc_user WHERE username = 'administrator'`)
	if err == sql.ErrNoRows {
		logger.Info("admin 用户不存在，正在创建...")
		err = tx.Get(&adminID, `INSERT INTO iacc_user (username, password) VALUES ('administrator', 'e10adc3949ba59abbe56e057f20f883e') RETURNING id`)
		if err != nil {
			return fmt.Errorf("创建 admin 用户失败: %w", err)
		}
		logger.Info("admin 用户创建成功")
	} else if err != nil {
		return fmt.Errorf("检索 admin 用户失败: %w", err)
	} else {
		logger.Info("admin 用户已存在")
	}

	// 2. 检索或创建 root 角色
	var rootID string
	err = tx.Get(&rootID, `SELECT id FROM iacc_role WHERE name = 'root'`)
	if err == sql.ErrNoRows {
		logger.Info("root 角色不存在，正在创建...")
		err = tx.Get(&rootID, `INSERT INTO iacc_role (name) VALUES ('root') RETURNING id`)
		if err != nil {
			return fmt.Errorf("创建 root 角色失败: %w", err)
		}
		logger.Info("root 角色创建成功")
	} else if err != nil {
		return fmt.Errorf("检索 root 角色失败: %w", err)
	} else {
		logger.Info("root 角色已存在")
	}

	// 3. 确保 admin 用户拥有 root 角色
	var count int
	err = tx.Get(&count, `SELECT count(*) FROM iacc_user_role WHERE user_id = $1 AND role_id = $2`, adminID, rootID)
	if err != nil {
		return fmt.Errorf("检查 admin 用户的 root 角色失败: %w", err)
	}
	if count == 0 {
		logger.Info("为 admin 用户分配 root 角色...")
		_, err = tx.Exec(`INSERT INTO iacc_user_role (user_id, role_id) VALUES ($1, $2)`, adminID, rootID)
		if err != nil {
			return fmt.Errorf("为 admin 用户分配 root 角色失败: %w", err)
		}
		logger.Info("为 admin 用户分配 root 角色成功")
	} else {
		logger.Info("admin 用户已拥有 root 角色")
	}

	// 4. 确保 root 角色拥有所有权限
	var permissionIDs []string
	err = tx.Select(&permissionIDs, `SELECT id FROM iacc_permission`)
	if err != nil {
		return fmt.Errorf("获取所有权限失败: %w", err)
	}

	if len(permissionIDs) > 0 {
		logger.Info(fmt.Sprintf("为 root 角色分配 %d 个权限...", len(permissionIDs)))
		query := `INSERT INTO iacc_role_permission (role_id, permission_id) VALUES `
		args := []interface{}{}
		for _, pid := range permissionIDs {
			query += fmt.Sprintf("($%d, $%d),", len(args)+1, len(args)+2)
			args = append(args, rootID, pid)
		}
		query = query[:len(query)-1] // 移除末尾的逗号
		query += " ON CONFLICT (role_id, permission_id) DO NOTHING"

		_, err = tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("为 root 角色分配权限失败: %w", err)
		}
		logger.Info("为 root 角色分配权限成功")
	} else {
		logger.Info("没有需要分配的权限")
	}

	return nil
}
