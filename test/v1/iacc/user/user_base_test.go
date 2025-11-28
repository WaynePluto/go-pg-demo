package user_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go-pg-demo/internal/app"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// setupTestUser 在数据库中创建一个用户用于测试，并注册一个清理函数以便在测试结束后删除它
func setupTestUser(t *testing.T) map[string]any {
	t.Helper()

	// 使用 TestUtil 创建测试用户
	testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	testUser := testUtil.SetupTestUser()

	// 转换为 map
	entity := map[string]any{
		"id":       testUser.ID,
		"username": testUser.Username,
		"phone":    testUser.Phone,
		"password": testUser.Password,
	}

	return entity
}

// createTestUser 在数据库中创建一个用户用于测试，并注册一个清理函数以便在测试结束后删除它
// 参数:
//   - t: 测试实例
//   - username: 用户名，如果为空则使用自动生成的唯一名称
//   - phone: 手机号，如果为空则使用自动生成的唯一手机号
//   - password: 密码，如果为空则使用默认密码
//
// 返回值:
//   - map[string]any: 包含创建的用户信息的map
func createTestUser(t *testing.T, username, phone, password string) map[string]any {
	t.Helper()

	// 如果没有提供用户名，使用时间戳生成唯一的名称，避免测试之间的干扰
	if username == "" {
		username = "TestUser_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// 如果没有提供手机号，生成一个唯一的手机号
	if phone == "" {
		phone = fmt.Sprintf("138%s", uuid.NewString()[:8])
	}

	// 如果没有提供密码，使用默认密码
	if password == "" {
		password = "password123"
	}

	entity := map[string]any{
		"username": username,
		"phone":    phone,
		"password": password,
	}

	// 定义一个结构体来接收返回的数据
	type Result struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	var result Result
	// 直接在数据库中创建实体
	query := `INSERT INTO "iacc_user" (username, phone, password) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := testDB.QueryRowContext(context.Background(), query, entity["username"], entity["phone"], entity["password"]).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)
	assert.NoError(t, err, "创建测试用户不应出错")

	// 将结果合并到 entity map 中
	entity["id"] = result.ID
	entity["created_at"] = result.CreatedAt
	entity["updated_at"] = result.UpdatedAt

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = $1`, result.ID)
		if err != nil {
			t.Errorf("清理测试用户失败: %v", err)
		}
	})

	return entity
}

// createSortTestUser 创建用于排序测试的用户
func createSortTestUser(t *testing.T, username string, phone string) map[string]any {
	t.Helper()
	if phone == "" {
		phone = fmt.Sprintf("138%s", uuid.NewString()[:8])
	}
	return createTestUser(t, username, phone, "password123")
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
			t1Sec := t1[:19] // 取前19个字符，即 "2006-01-02 15:04:05"
			t2Sec := t2[:19]
			return t1Sec == t2Sec
		}
	}

	// 比较到秒级精度
	return parsedT1.Truncate(time.Second).Equal(parsedT2.Truncate(time.Second))
}
