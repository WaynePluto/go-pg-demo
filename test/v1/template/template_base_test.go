package template_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go-pg-demo/internal/app"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// 全局测试变量
var (
	testDB     *sqlx.DB    // 测试数据库连接
	testLogger *zap.Logger // 测试日志记录器
	testRouter *gin.Engine // 测试路由器
)

// TestMain 初始化测试环境
// 在所有测试运行前执行，用于设置测试所需的全局变量和环境
func TestMain(m *testing.M) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建应用实例
	testApp, _, err := app.InitializeApp()
	if err != nil {
		// 如果应用初始化失败，直接退出
		os.Exit(1)
	}

	testDB = testApp.DB
	testLogger = testApp.Logger
	testRouter = testApp.Server

	// 运行测试
	exitCode := m.Run()

	// 退出
	os.Exit(exitCode)
}

// createTestTemplate 在数据库中创建一个模板用于测试，并注册一个清理函数以便在测试结束后删除它
// 整合了原来的 setupTestTemplate 和 createSortTestTemplate 函数
// 参数:
//   - t: 测试实例
//   - name: 模板名称，如果为空则使用自动生成的唯一名称
//   - num: 模板数量，如果为nil则使用默认值100
//
// 返回值:
//   - map[string]any: 包含创建的模板信息的map
func createTestTemplate(t *testing.T, name string, num *int) map[string]any {
	t.Helper()

	// 如果没有提供名称，使用时间戳生成唯一的名称，避免测试之间的干扰
	if name == "" {
		name = "Test Template_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// 如果没有提供数量，使用默认值100
	defaultNum := 100
	if num == nil {
		num = &defaultNum
	}

	entity := map[string]any{
		"name": name,
		"num":  num,
	}

	// 定义一个结构体来接收返回的数据
	type Result struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	var result Result
	// 直接在数据库中创建实体
	query := `INSERT INTO template (name, num) VALUES (:name, :num) RETURNING id, created_at, updated_at`
	stmt, err := testDB.PrepareNamedContext(context.Background(), query)
	assert.NoError(t, err)
	defer stmt.Close()

	err = stmt.GetContext(context.Background(), &result, entity)
	assert.NoError(t, err)

	// 将结果合并到 entity map 中
	entity["id"] = result.ID
	entity["created_at"] = result.CreatedAt
	entity["updated_at"] = result.UpdatedAt

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM template WHERE id = :id", map[string]any{"id": result.ID})
		if err != nil {
			t.Errorf("清理测试模板失败: %v", err)
		}
	})

	return entity
}

// getAuthToken 获取认证令牌
// 参数:
//   - t: 测试实例
//   - permissions: 权限列表，空数组表示无权限访问令牌
//
// 返回值:
//   - string: 认证令牌
func getAuthToken(t *testing.T, permissions []string) string {
	t.Helper()
	// 创建 TestUtil 实例
	testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	// 获取token
	return testUtil.GetAccessUserToken(permissions)
}
