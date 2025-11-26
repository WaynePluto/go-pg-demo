package permission_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
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

// setupTestPermission 在数据库中创建一个权限用于测试，并注册一个清理函数以便在测试结束后删除它
func setupTestPermission(t *testing.T, name string) map[string]any {
	t.Helper()

	method := "get"
	path := "/test/path"
	metadata := map[string]any{
		"method": &method,
		"path":   &path,
	}

	// 将 metadata 序列化为 JSON 字符串，因为 PostgreSQL 需要 JSON 类型
	metadataJSON, err := json.Marshal(metadata)
	assert.NoError(t, err, "序列化 metadata 不应出错")

	entity := map[string]any{
		"name":     name,
		"type":     "api",
		"metadata": string(metadataJSON), // 存储为 JSON 字符串
	}

	// 定义一个结构体来接收返回的数据
	type Result struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	var result Result
	// 直接在数据库中创建实体
	query := `INSERT INTO iacc_permission (name, type, metadata) VALUES ($1, $2, $3::json) RETURNING id, created_at, updated_at`
	err = testDB.QueryRowContext(context.Background(), query, entity["name"], entity["type"], entity["metadata"]).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)
	assert.NoError(t, err, "创建测试权限不应出错")

	// 将结果合并到 entity map 中
	entity["id"] = result.ID
	entity["created_at"] = result.CreatedAt
	entity["updated_at"] = result.UpdatedAt

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.ExecContext(context.Background(), "DELETE FROM iacc_permission WHERE id = $1", result.ID)
		if err != nil {
			t.Errorf("清理测试权限失败: %v", err)
		}
	})

	return entity
}

// createSortTestPermission 创建用于排序测试的权限
func createSortTestPermission(t *testing.T, name string, pType string) map[string]any {
	t.Helper()

	method := "get"
	path := "/test/path"
	metadata := map[string]any{
		"method": &method,
		"path":   &path,
	}

	// 将 metadata 序列化为 JSON 字符串，因为 PostgreSQL 需要 JSON 类型
	metadataJSON, err := json.Marshal(metadata)
	assert.NoError(t, err, "序列化 metadata 不应出错")

	entity := map[string]any{
		"name":     name,
		"type":     pType,
		"metadata": string(metadataJSON), // 存储为 JSON 字符串
	}

	// 定义一个结构体来接收返回的数据
	type Result struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	var result Result
	// 直接在数据库中创建实体
	query := `INSERT INTO iacc_permission (name, type, metadata) VALUES ($1, $2, $3::json) RETURNING id, created_at, updated_at`
	err = testDB.QueryRowContext(context.Background(), query, entity["name"], entity["type"], entity["metadata"]).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)
	assert.NoError(t, err, "创建测试权限不应出错")

	// 将结果合并到 entity map 中
	entity["id"] = result.ID
	entity["created_at"] = result.CreatedAt
	entity["updated_at"] = result.UpdatedAt

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.ExecContext(context.Background(), "DELETE FROM iacc_permission WHERE id = $1", result.ID)
		if err != nil {
			t.Errorf("清理测试权限失败: %v", err)
		}
	})

	return entity
}

// getAuthToken 获取认证令牌
func getAuthToken(t *testing.T, permissions []string) string {
	t.Helper()
	// 创建 TestUtil 实例
	testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	// 获取token
	return testUtil.GetAccessUserToken(permissions)
}

// stringPtr 添加辅助函数用于创建字符串指针
func stringPtr(s string) *string {
	return &s
}

// compareTimestamps 比较两个时间戳字符串，只比较到秒级精度
func compareTimestamps(t1, t2 string) bool {
	// 解析时间戳
	parsedT1, err1 := time.Parse(time.RFC3339Nano, t1)
	parsedT2, err2 := time.Parse(time.RFC3339Nano, t2)

	if err1 != nil || err2 != nil {
		// 如果解析失败，尝试其他格式
		parsedT1, err1 = time.Parse("2006-01-02 15:04:05.999999", t1)
		parsedT2, err2 = time.Parse("2006-01-02 15:04:05.999999", t2)

		if err1 != nil || err2 != nil {
			// 如果还是失败，直接比较字符串（去掉微秒部分）
			t1Sec := strings.Split(t1, ".")[0]
			t2Sec := strings.Split(t2, ".")[0]
			return t1Sec == t2Sec
		}
	}

	// 比较到秒级精度
	return parsedT1.Truncate(time.Second).Equal(parsedT2.Truncate(time.Second))
}
