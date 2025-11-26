package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

// 在数据库中创建一个用户用于测试，并注册一个清理函数以便在测试结束后删除它
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

func TestCreateUser(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		username := "testuser1_" + uuid.NewString()[:8]
		phone := fmt.Sprintf("138%s", uuid.NewString()[:8])
		password := "password123"
		createReqBody := map[string]any{
			"username": username,
			"phone":    phone,
			"password": password,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(bodyBytes))
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
		assert.Equal(t, 200, createResp.Code, "响应码应该是 200")

		createdID, ok := createResp.Data.(string)
		assert.True(t, ok, "响应数据应该是字符串类型的 ID")

		// 验证用户确实被创建
		type UserEntity struct {
			ID        string `db:"id"`
			Phone     string `db:"phone"`
			CreatedAt string `db:"created_at"`
			UpdatedAt string `db:"updated_at"`
		}
		var entity UserEntity
		query := `SELECT id, phone, created_at, updated_at FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &entity, query, createdID)
		assert.NoError(t, err, "应该能在数据库中找到创建的用户")
		assert.Equal(t, phone, entity.Phone, "用户手机号应该匹配")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.NamedExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = :id`, map[string]any{"id": createdID})
			assert.NoError(t, err, "清理创建的用户不应出错")
		})
	})

	t.Run("无效输入 - 缺少手机号", func(t *testing.T) {
		// 准备
		username := "testuser2_" + uuid.NewString()[:8]
		password := "password123"
		createReqBody := map[string]any{
			"username": username,
			"password": password, // 缺少手机号
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(bodyBytes))
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
	})

	t.Run("无效输入 - 缺少密码", func(t *testing.T) {
		// 准备
		username := "testuser3_" + uuid.NewString()[:8]
		phone := fmt.Sprintf("138%s", uuid.NewString()[:8])
		createReqBody := map[string]any{
			"username": username,
			"phone":    phone, // 缺少密码
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(bodyBytes))
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
	})
}

func TestBatchCreateUsers(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		batchCreateReq := map[string]any{
			"users": []map[string]any{
				{"username": "批量用户1_" + uuid.NewString()[:8], "phone": fmt.Sprintf("138%s", uuid.NewString()[:8]), "password": "password123"},
				{"username": "批量用户2_" + uuid.NewString()[:8], "phone": fmt.Sprintf("138%s", uuid.NewString()[:8]), "password": "password123"},
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-create", bytes.NewBuffer(bodyBytes))
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

		ids, ok := createResp.Data.([]any)
		assert.True(t, ok, "响应数据应该是 ID 数组")
		assert.Len(t, ids, 2, "应创建 2 个用户")

		// 清理
		t.Cleanup(func() {
			for _, id := range ids {
				_, err := testDB.NamedExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = :id`, map[string]any{"id": id.(string)})
				assert.NoError(t, err, "清理批量创建的用户不应出错")
			}
		})
	})

	t.Run("无效输入 - 空列表", func(t *testing.T) {
		// 准备
		batchCreateReq := map[string]any{
			"users": []map[string]any{},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-create", bytes.NewBuffer(bodyBytes))
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
		assert.Equal(t, http.StatusOK, w.Code)
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errResp.Code)
		assert.Contains(t, errResp.Msg, "用户列表必须至少包含1项")
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestUser(t)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+entity["id"].(string), nil)
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
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是一个 map")

		assert.Equal(t, entity["phone"], data["phone"], "获取到的用户手机号应与创建时一致")
		assert.Equal(t, entity["username"], data["username"], "获取到的用户名应与创建时一致")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/invalid-id", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应码应该是 400")
	})

	t.Run("用户不存在", func(t *testing.T) {
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		nonExistentID := "123e4567-e89b-12d3-a456-426614174000"

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+nonExistentID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusNotFound, resp.Code, "响应码应该是 404")
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestUser(t)
		newPassword := "newpassword123"
		updateReqBody := map[string]any{
			"password": &newPassword,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodPut, "/v1/user/"+entity["id"].(string), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 验证用户确实被更新
		type UpdatedUser struct {
			Password string `db:"password"`
		}
		var updatedUser UpdatedUser
		query := `SELECT password FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &updatedUser, query, entity["id"])
		assert.NoError(t, err, "应该能在数据库中找到更新的用户")
		assert.Equal(t, newPassword, updatedUser.Password, "用户密码应该已被更新")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 准备
		newPassword := "newpassword123"
		updateReqBody := map[string]any{
			"password": &newPassword,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodPut, "/v1/user/invalid-id", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应码应该是 400")
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestUser(t)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/user/"+entity["id"].(string), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		affectedRows := int(math.Round(resp.Data.(float64)))
		assert.Equal(t, 1, affectedRows, "应影响 1 行")

		// 验证用户确实被删除
		type DeletedUser struct {
			ID        string `db:"id"`
			Phone     string `db:"phone"`
			CreatedAt string `db:"created_at"`
			UpdatedAt string `db:"updated_at"`
		}
		var deletedUser DeletedUser
		query := `SELECT id, phone, created_at, updated_at FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &deletedUser, query, entity["id"])
		assert.Error(t, err, "应该无法在数据库中找到已删除的用户")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/user/invalid-id", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应码应该是 400")
	})
}

func TestBatchDeleteUsers(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity1 := setupTestUser(t)
		entity2 := setupTestUser(t)
		deleteReq := map[string]any{
			"ids": []string{entity1["id"].(string), entity2["id"].(string)},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-delete", bytes.NewBuffer(bodyBytes))
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
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		// 验证
		query, args, err := sqlx.In("SELECT COUNT(*) FROM \"iacc_user\" WHERE id IN (?)", []string{entity1["id"].(string), entity2["id"].(string)})
		assert.NoError(t, err)
		query = testDB.Rebind(query)
		var count int
		err = testDB.GetContext(context.Background(), &count, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "用户应已被删除")
	})

	t.Run("无效输入 - 空ID列表", func(t *testing.T) {
		// 准备
		deleteReq := map[string]any{
			"ids": []string{},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-delete", bytes.NewBuffer(bodyBytes))
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
		assert.Equal(t, http.StatusOK, w.Code)
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errResp.Code)
		assert.Contains(t, errResp.Msg, "用户ID列表必须至少包含1项")
	})
}

func TestQueryUserList(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备数据
		entity := setupTestUser(t)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?phone="+entity["phone"].(string)+"&page=1&pageSize=10", nil)
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
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是一个 map")

		total := int(data["total"].(float64))
		assert.GreaterOrEqual(t, total, 1, "总数量应该至少为1")

		list := data["list"].([]any)
		assert.GreaterOrEqual(t, len(list), 1, "列表长度应该至少为1")
	})

	t.Run("空结果", func(t *testing.T) {
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?phone=nonexistent123456789&page=1&pageSize=10", nil)
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
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是一个 map")

		total := int(data["total"].(float64))
		assert.Equal(t, 0, total, "总数量应该为0")

		list := data["list"].([]any)
		assert.Equal(t, 0, len(list), "列表长度应该为0")
	})

	t.Run("成功 - 带用户名搜索", func(t *testing.T) {
		// 准备
		entity := setupTestUser(t)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username="+entity["username"].(string), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		assert.Greater(t, int(data["total"].(float64)), 0)
		list, ok := data["list"].([]any)
		assert.True(t, ok)
		assert.NotEmpty(t, list)
		// 确保返回的第一个元素是我们刚创建的
		firstItem := list[0].(map[string]any)
		assert.Equal(t, entity["id"], firstItem["id"])
	})
}

func TestAssignRoles(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestUser(t)

		// 先创建一个角色用于测试
		roleID := uuid.NewString()
		_, err := testDB.ExecContext(context.Background(),
			`INSERT INTO "iacc_role" (id, name, description) VALUES ($1, $2, $3)`,
			roleID, "测试角色", "用于测试的角色")
		assert.NoError(t, err)

		// 测试完成后清理角色
		t.Cleanup(func() {
			_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_role" WHERE id = $1`, roleID)
			if err != nil {
				t.Errorf("清理测试角色失败: %v", err)
			}
		})

		roleIDs := []string{roleID}
		assignReq := map[string]any{
			"role_ids": roleIDs,
		}
		bodyBytes, _ := json.Marshal(assignReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/"+entity["id"].(string)+"/role", bytes.NewBuffer(bodyBytes))
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
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		affectedRows := int(math.Round(resp.Data.(float64)))
		assert.Equal(t, 1, affectedRows, "应影响 1 行")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 准备
		roleIDs := []string{uuid.NewString()}
		assignReq := map[string]any{
			"role_ids": roleIDs,
		}
		bodyBytes, _ := json.Marshal(assignReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/invalid-id/role", bytes.NewBuffer(bodyBytes))
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
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应码应该是 400")
	})
}

// 创建用于排序测试的用户
func createSortTestUser(t *testing.T, username string, phone string) map[string]any {
	t.Helper()

	password := "password123"
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

func TestQueryListUsers_Sort(t *testing.T) {
	// 准备数据
	phone1 := fmt.Sprintf("138%s", uuid.NewString()[:8])
	phone2 := fmt.Sprintf("138%s", uuid.NewString()[:8])
	phone3 := fmt.Sprintf("138%s", uuid.NewString()[:8])
	id1 := createSortTestUser(t, "SortTest_A", phone1)
	id2 := createSortTestUser(t, "SortTest_B", phone2)
	id3 := createSortTestUser(t, "SortTest_C", phone3)

	t.Run("按 ID ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=id&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		// 应该只有三个结果
		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, id1["id"], list[0].(map[string]any)["id"])
			assert.Equal(t, id2["id"], list[1].(map[string]any)["id"])
			assert.Equal(t, id3["id"], list[2].(map[string]any)["id"])
		}
	})

	t.Run("按 ID DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=id&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, id3["id"], list[0].(map[string]any)["id"])
			assert.Equal(t, id2["id"], list[1].(map[string]any)["id"])
			assert.Equal(t, id1["id"], list[2].(map[string]any)["id"])
		}
	})

	t.Run("按 Username ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=username&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, "SortTest_A", list[0].(map[string]any)["username"])
			assert.Equal(t, "SortTest_B", list[1].(map[string]any)["username"])
			assert.Equal(t, "SortTest_C", list[2].(map[string]any)["username"])
		}
	})

	t.Run("按 Username DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=username&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, "SortTest_C", list[0].(map[string]any)["username"])
			assert.Equal(t, "SortTest_B", list[1].(map[string]any)["username"])
			assert.Equal(t, "SortTest_A", list[2].(map[string]any)["username"])
		}
	})

	t.Run("按 Phone ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=phone&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			phones := []string{phone1, phone2, phone3}
			// 对手机号进行排序以便比较
			for i := 0; i < len(phones)-1; i++ {
				for j := i + 1; j < len(phones); j++ {
					if phones[i] > phones[j] {
						phones[i], phones[j] = phones[j], phones[i]
					}
				}
			}
			assert.Equal(t, phones[0], list[0].(map[string]any)["phone"])
			assert.Equal(t, phones[1], list[1].(map[string]any)["phone"])
			assert.Equal(t, phones[2], list[2].(map[string]any)["phone"])
		}
	})

	t.Run("按 Phone DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=phone&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			phones := []string{phone1, phone2, phone3}
			// 对手机号进行降序排序以便比较
			for i := 0; i < len(phones)-1; i++ {
				for j := i + 1; j < len(phones); j++ {
					if phones[i] < phones[j] {
						phones[i], phones[j] = phones[j], phones[i]
					}
				}
			}
			assert.Equal(t, phones[0], list[0].(map[string]any)["phone"])
			assert.Equal(t, phones[1], list[1].(map[string]any)["phone"])
			assert.Equal(t, phones[2], list[2].(map[string]any)["phone"])
		}
	})

	t.Run("按 CreatedAt ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=created_at&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id1["created_at"].(string), list[0].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id2["created_at"].(string), list[1].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id3["created_at"].(string), list[2].(map[string]any)["created_at"].(string)))
		}
	})

	t.Run("按 CreatedAt DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=created_at&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id3["created_at"].(string), list[0].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id2["created_at"].(string), list[1].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id1["created_at"].(string), list[2].(map[string]any)["created_at"].(string)))
		}
	})

	t.Run("按 UpdatedAt ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=updated_at&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id1["updated_at"].(string), list[0].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id2["updated_at"].(string), list[1].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id3["updated_at"].(string), list[2].(map[string]any)["updated_at"].(string)))
		}
	})

	t.Run("按 UpdatedAt DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=updated_at&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok)
		list, ok := data["list"].([]any)
		assert.True(t, ok)

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id3["updated_at"].(string), list[0].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id2["updated_at"].(string), list[1].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id1["updated_at"].(string), list[2].(map[string]any)["updated_at"].(string)))
		}
	})

	t.Run("无效 OrderBy", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?orderBy=invalid_field", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "排序字段不存在", resp.Msg)
	})

	t.Run("无效 Order", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?orderBy=username&order=invalid_order", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "排序顺序参数错误", resp.Msg)
	})
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
