package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestCreateUser 测试创建用户功能
// 包含三个子测试：成功创建、无效输入-缺少手机号、无效输入-缺少密码
func TestCreateUser(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		username := "testuser1_" + uuid.NewString()[:8]
		phone := "138" + uuid.NewString()[:8]
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
		phone := "138" + uuid.NewString()[:8]
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

// TestGetUserByID 测试根据ID获取用户功能
// 包含三个子测试：成功获取、无效ID、用户不存在
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

// TestUpdateUser 测试根据ID更新用户功能
// 包含两个子测试：成功更新、无效ID
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

// TestDeleteUser 测试根据ID删除用户功能
// 包含两个子测试：成功删除、无效ID
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

		affectedRows := int(resp.Data.(float64))
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
