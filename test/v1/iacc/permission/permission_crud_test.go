package permission_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestCreatePermission 测试创建权限功能
func TestCreatePermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		metadata := map[string]any{
			"path":   stringPtr("/test/path"),
			"method": stringPtr("post"),
		}
		createReqBody := map[string]any{
			"name":     "新的测试权限",
			"type":     "api",
			"metadata": metadata,
		}

		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
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
		metadata := map[string]any{
			"path":   stringPtr("/test/path"),
			"method": stringPtr("post"),
		}
		createReqBody := map[string]any{
			"type":     "api", // 缺少名称
			"metadata": metadata,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
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

// TestGetPermission 测试根据ID获取权限功能
func TestGetPermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestPermission(t, "测试权限-Get")

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/"+entity["id"].(string), nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
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
		assert.Equal(t, entity["name"], data["name"], "获取到的权限名称应与创建时一致")
		assert.Equal(t, entity["type"], data["type"], "获取到的权限类型应与创建时一致")
	})

	t.Run("未找到", func(t *testing.T) {
		// 准备
		nonExistentID := "123e4567-e89b-12d3-a456-426614174000" // 使用有效的UUID格式但数据库中不存在的ID

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/"+nonExistentID, nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
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

// TestUpdatePermission 测试根据ID更新权限功能
func TestUpdatePermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestPermission(t, "测试权限-Update")
		updateName := "更新后的权限名称"
		updateReqBody := map[string]any{
			"name": updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/permission/"+entity["id"].(string), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
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
		type UpdatedPermission struct {
			Name string `db:"name"`
			Type string `db:"type"`
		}
		var updatedPermission UpdatedPermission
		err = testDB.GetContext(context.Background(), &updatedPermission, "SELECT name, type FROM iacc_permission WHERE id = $1", entity["id"])
		assert.NoError(t, err, "从数据库获取更新后的权限不应出错")
		assert.Equal(t, updateName, updatedPermission.Name, "权限名称应已更新")
		assert.Equal(t, entity["type"], updatedPermission.Type, "权限类型不应改变")
	})
}

// TestDeletePermission 测试根据ID删除权限功能
func TestDeletePermission(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestPermission(t, "测试权限-Delete")

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/permission/"+entity["id"].(string), nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
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
		err = testDB.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM iacc_permission WHERE id = $1", entity["id"])
		assert.NoError(t, err, "查询已删除权限的计数不应出错")
		assert.Equal(t, 0, count, "删除后权限在数据库中应不存在")
	})
}
