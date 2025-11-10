package permission_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-pg-demo/internal/app"
	"go-pg-demo/internal/modules/iacc/permission"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testDB     *sqlx.DB
	testLogger *zap.Logger
	testRouter *gin.Engine
)

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

// 在数据库中创建一个权限用于测试，并注册一个清理函数以便在测试结束后删除它
func setupTestPermission(t *testing.T, name string) permission.PermissionEntity {
	t.Helper()

	metadata := permission.Metadata{
		Method: stringPtr("get"),
		Path:   stringPtr("/test/path"),
	}
	entity := permission.PermissionEntity{
		Name:     name,
		Type:     "api",
		Metadata: metadata,
	}

	// 直接在数据库中创建实体
	query := `INSERT INTO iacc_permission (name, type, metadata) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := testDB.QueryRowContext(context.Background(), query, entity.Name, entity.Type, entity.Metadata).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	assert.NoError(t, err, "创建测试权限不应出错")

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.ExecContext(context.Background(), "DELETE FROM iacc_permission WHERE id = $1", entity.ID)
		if err != nil {
			t.Errorf("清理测试权限失败: %v", err)
		}
	})

	return entity
}

func TestCreatePermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		metadata := permission.Metadata{
			Path:   stringPtr("/test/path"),
			Method: stringPtr("post"),
		}
		createReqBody := permission.CreatePermissionReq{
			Name:     "新的测试权限",
			Type:     "api",
			Metadata: metadata,
		}

		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, createResp.Code, "响应码应该是 200")

		createdID, ok := createResp.Data.(string)
		assert.True(t, ok, "响应数据应该是字符串类型的 ID")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.ExecContext(context.Background(), "DELETE FROM iacc_permission WHERE id = $1", createdID)
			assert.NoError(t, err, "清理创建的权限不应出错")
		})
	})

	t.Run("无效输入 - 缺少名称", func(t *testing.T) {
		// 准备
		metadata := permission.Metadata{
			Path:   stringPtr("/test/path"),
			Method: stringPtr("post"),
		}
		createReqBody := permission.CreatePermissionReq{
			Type:     "api", // 缺少名称
			Metadata: metadata,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, errResp.Code, "响应码应该是 400")
		assert.Contains(t, errResp.Msg, "权限名称为必填字段", "错误消息应包含名称必填的验证错误")
	})
}

func TestGetPermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestPermission(t, "测试权限-Get")

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/"+entity.ID, nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 使用 map 来灵活处理响应数据
		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "响应数据应该是一个 map")
		assert.Equal(t, entity.Name, data["name"], "获取到的权限名称应与创建时一致")
		assert.Equal(t, entity.Type, data["type"], "获取到的权限类型应与创建时一致")
	})

	t.Run("未找到", func(t *testing.T) {
		// 准备
		nonExistentID := "123e4567-e89b-12d3-a456-426614174000" // 使用有效的UUID格式但数据库中不存在的ID

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/"+nonExistentID, nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "处理器应返回 200 状态码，但在响应体中包含错误码")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusNotFound, resp.Code, "响应码应该是 404")
		assert.Equal(t, "权限不存在", resp.Msg, "错误消息应为 '权限不存在'")
	})
}

func TestUpdatePermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestPermission(t, "测试权限-Update")
		updateName := "更新后的权限名称"
		updateReqBody := map[string]any{
			"name": updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/permission/"+entity.ID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var updateResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, updateResp.Code, "响应码应该是 200")
		assert.Equal(t, int64(1), int64(updateResp.Data.(float64)), "应影响 1 行")

		// 验证更新
		var updatedPermission permission.PermissionEntity
		err = testDB.GetContext(context.Background(), &updatedPermission, "SELECT * FROM iacc_permission WHERE id = $1", entity.ID)
		assert.NoError(t, err, "从数据库获取更新后的权限不应出错")
		assert.Equal(t, updateName, updatedPermission.Name, "权限名称应已更新")
		assert.Equal(t, entity.Type, updatedPermission.Type, "权限类型不应改变")
	})
}

func TestDeletePermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestPermission(t, "测试权限-Delete")

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/permission/"+entity.ID, nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var deleteResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, deleteResp.Code, "响应码应该是 200")
		assert.Equal(t, int64(1), int64(deleteResp.Data.(float64)), "应影响 1 行")

		// 验证删除
		var count int
		err = testDB.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM iacc_permission WHERE id = $1", entity.ID)
		assert.NoError(t, err, "查询已删除权限的计数不应出错")
		assert.Equal(t, 0, count, "删除后权限在数据库中应不存在")
	})
}

func TestQueryPermissionList(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备 - 创建多个测试权限
		setupTestPermission(t, "测试权限-Query1")
		setupTestPermission(t, "测试权限-Query2")

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?page=1&pageSize=10", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")
		
		// 使用 map 来灵活处理响应数据
		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "响应数据应该是一个 map")
		total, ok := data["total"]
		assert.True(t, ok, "响应数据应该包含 total 字段")
		assert.GreaterOrEqual(t, int64(total.(float64)), int64(2), "总权限数应至少为 2")
	})
}

// 添加辅助函数用于创建字符串指针
func stringPtr(s string) *string {
	return &s
}