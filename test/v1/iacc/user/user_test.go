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
	"testing"

	"go-pg-demo/internal/app"
	"go-pg-demo/internal/modules/iacc/user"
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
func setupTestUser(t *testing.T) user.UserEntity {
	t.Helper()

	// 使用 TestUtil 创建测试用户
	testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	testUser := testUtil.SetupTestUser()

	// 转换为 UserEntity
	entity := user.UserEntity{
		ID:       testUser.ID,
		Username: testUser.Username,
		Phone:    testUser.Phone,
		Password: testUser.Password,
	}

	return entity
}

func TestCreateUser(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		username := "testuser1_" + uuid.NewString()[:8]
		phone := fmt.Sprintf("138%s", uuid.NewString()[:8])
		password := "password123"
		createReqBody := user.CreateReq{
			Username: username,
			Phone:    phone,
			Password: password,
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
		var entity user.UserEntity
		query := `SELECT id, phone, profile, created_at, updated_at FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &entity, query, createdID)
		assert.NoError(t, err, "应该能在数据库中找到创建的用户")
		assert.Equal(t, phone, entity.Phone, "用户手机号应该匹配")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.NamedExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = :id`, map[string]interface{}{"id": createdID})
			assert.NoError(t, err, "清理创建的用户不应出错")
		})
	})

	t.Run("无效输入 - 缺少手机号", func(t *testing.T) {
		// 准备
		username := "testuser2_" + uuid.NewString()[:8]
		password := "password123"
		createReqBody := user.CreateReq{
			Username: username,
			Password: password, // 缺少手机号
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
		createReqBody := user.CreateReq{
			Username: username,
			Phone:    phone, // 缺少密码
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
		batchCreateReq := user.BatchCreateReq{
			Users: []user.CreateReq{
				{Username: "批量用户1_" + uuid.NewString()[:8], Phone: fmt.Sprintf("138%s", uuid.NewString()[:8]), Password: "password123"},
				{Username: "批量用户2_" + uuid.NewString()[:8], Phone: fmt.Sprintf("138%s", uuid.NewString()[:8]), Password: "password123"},
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

		ids, ok := createResp.Data.([]interface{})
		assert.True(t, ok, "响应数据应该是 ID 数组")
		assert.Len(t, ids, 2, "应创建 2 个用户")

		// 清理
		t.Cleanup(func() {
			for _, id := range ids {
				_, err := testDB.NamedExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = :id`, map[string]interface{}{"id": id.(string)})
				assert.NoError(t, err, "清理批量创建的用户不应出错")
			}
		})
	})

	t.Run("无效输入 - 空列表", func(t *testing.T) {
		// 准备
		batchCreateReq := user.BatchCreateReq{
			Users: []user.CreateReq{},
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+entity.ID, nil)
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

		assert.Equal(t, entity.Phone, data["phone"], "获取到的用户手机号应与创建时一致")
		assert.Equal(t, entity.Username, data["username"], "获取到的用户名应与创建时一致")
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
		updateReqBody := user.UpdateByIDReq{
			ID:       entity.ID,
			Password: &newPassword,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})

		// 执行
		req, _ := http.NewRequest(http.MethodPut, "/v1/user/"+entity.ID, bytes.NewBuffer(bodyBytes))
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
		var updatedEntity user.UserEntity
		query := `SELECT id, phone, password, profile, created_at, updated_at FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &updatedEntity, query, entity.ID)
		assert.NoError(t, err, "应该能在数据库中找到更新的用户")
		assert.Equal(t, newPassword, updatedEntity.Password, "用户密码应该已被更新")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 准备
		newPassword := "newpassword123"
		updateReqBody := user.UpdateByIDReq{
			ID:       "invalid-id",
			Password: &newPassword,
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
		req, _ := http.NewRequest(http.MethodDelete, "/v1/user/"+entity.ID, nil)
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
		var deletedEntity user.UserEntity
		query := `SELECT id, phone, profile, created_at, updated_at FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &deletedEntity, query, entity.ID)
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
		deleteReq := user.DeleteUsersReq{
			IDs: []string{entity1.ID, entity2.ID},
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
		query, args, err := sqlx.In("SELECT COUNT(*) FROM \"iacc_user\" WHERE id IN (?)", []string{entity1.ID, entity2.ID})
		assert.NoError(t, err)
		query = testDB.Rebind(query)
		var count int
		err = testDB.GetContext(context.Background(), &count, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "用户应已被删除")
	})

	t.Run("无效输入 - 空ID列表", func(t *testing.T) {
		// 准备
		deleteReq := user.DeleteUsersReq{
			IDs: []string{},
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?phone="+entity.Phone+"&page=1&pageSize=10", nil)
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

		total := int(data["total"].(float64))
		assert.GreaterOrEqual(t, total, 1, "总数量应该至少为1")

		list := data["list"].([]interface{})
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
		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "响应数据应该是一个 map")

		total := int(data["total"].(float64))
		assert.Equal(t, 0, total, "总数量应该为0")

		list := data["list"].([]interface{})
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username="+entity.Username, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Greater(t, int(data["total"].(float64)), 0)
		list, ok := data["list"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, list)
		// 确保返回的第一个元素是我们刚创建的
		firstItem := list[0].(map[string]interface{})
		assert.Equal(t, entity.ID, firstItem["id"])
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
		assignReq := user.AssignRolesReq{
			ID:      entity.ID,
			RoleIDs: roleIDs,
		}
		bodyBytes, _ := json.Marshal(assignReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/"+entity.ID+"/role", bytes.NewBuffer(bodyBytes))
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
		assignReq := user.AssignRolesReq{
			ID:      "invalid-id",
			RoleIDs: roleIDs,
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
