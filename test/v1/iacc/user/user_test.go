package user_test

import (
	"bytes"
	"context"
	"encoding/json"
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

	username := "testuser"
	phone := "13800138000"  // 移除+86前缀以匹配查询测试
	password := "password123"
	entity := user.UserEntity{
		Username: username,
		Phone:    phone,
		Password: password,
	}

	// 直接在数据库中创建实体
	query := `INSERT INTO "iacc_user" (username, phone, password) VALUES (:username, :phone, :password) RETURNING id, created_at, updated_at`
	stmt, err := testDB.PrepareNamedContext(context.Background(), query)
	assert.NoError(t, err)
	defer stmt.Close()

	err = stmt.GetContext(context.Background(), &entity, entity)
	assert.NoError(t, err)

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.NamedExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = :id`, map[string]interface{}{"id": entity.ID})
		if err != nil {
			t.Errorf("清理测试用户失败: %v", err)
		}
	})

	return entity
}

func TestCreateUser(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		username := "testuser1"
		phone := "13900139001"
		password := "password123"
		createReqBody := user.CreateUserReq{
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
		username := "testuser2"
		password := "password123"
		createReqBody := user.CreateUserReq{
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
		username := "testuser3"
		phone := "+8613900139002"
		createReqBody := user.CreateUserReq{
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

		nonExistentID := uuid.New().String()

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
		updateReqBody := user.UpdateUserReq{
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
		updateReqBody := user.UpdateUserReq{
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
}
