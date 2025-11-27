package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAssignRoles 测试用户角色分配功能
// 包含两个子测试：成功分配、无效ID
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

// TestGetRoles 测试获取用户角色列表功能
// 包含四个子测试：成功获取、用户不存在、用户无角色、无效用户ID
func TestGetRoles(t *testing.T) {
	t.Run("成功获取用户角色列表", func(t *testing.T) {
		// Arrange - 准备测试数据
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}

		// 创建测试用户
		testUser := testUtil.SetupTestUser()

		// 创建两个测试角色
		role1 := testUtil.SetupTestRole()
		role2 := testUtil.SetupTestRole()

		// 为用户分配角色
		testUtil.AssignRoleToUser(testUser.ID, role1.ID)
		testUtil.AssignRoleToUser(testUser.ID, role2.ID)

		// Act - 执行请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+testUser.ID+"/roles", nil)
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert - 断言结果
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 验证响应数据结构
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是map类型")

		list, ok := data["list"].([]any)
		assert.True(t, ok, "list字段应该是数组类型")
		assert.Equal(t, 2, len(list), "应该返回2个角色")

		total, ok := data["total"].(float64)
		assert.True(t, ok, "total字段应该是数字类型")
		assert.Equal(t, float64(2), total, "总数应该是2")

		// 验证角色信息
		for _, role := range list {
			roleMap, ok := role.(map[string]any)
			assert.True(t, ok, "角色信息应该是map类型")

			// 验证必要字段存在
			assert.Contains(t, roleMap, "id", "角色应该有id字段")
			assert.Contains(t, roleMap, "name", "角色应该有name字段")
			assert.Contains(t, roleMap, "created_at", "角色应该有created_at字段")
			assert.Contains(t, roleMap, "updated_at", "角色应该有updated_at字段")
		}
	})

	t.Run("用户不存在", func(t *testing.T) {
		// Arrange - 准备测试数据
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		nonExistentUserID := uuid.NewString()

		// Act - 执行请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+nonExistentUserID+"/roles", nil)
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert - 断言结果
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 验证响应数据结构
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是map类型")

		list, ok := data["list"].([]any)
		assert.True(t, ok, "list字段应该是数组类型")
		assert.Equal(t, 0, len(list), "应该返回空列表")

		total, ok := data["total"].(float64)
		assert.True(t, ok, "total字段应该是数字类型")
		assert.Equal(t, float64(0), total, "总数应该是0")
	})

	t.Run("用户无角色", func(t *testing.T) {
		// Arrange - 准备测试数据
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		testUser := testUtil.SetupTestUser()

		// Act - 执行请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+testUser.ID+"/roles", nil)
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert - 断言结果
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 验证响应数据结构
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是map类型")

		list, ok := data["list"].([]any)
		assert.True(t, ok, "list字段应该是数组类型")
		assert.Equal(t, 0, len(list), "应该返回空列表")

		total, ok := data["total"].(float64)
		assert.True(t, ok, "total字段应该是数字类型")
		assert.Equal(t, float64(0), total, "总数应该是0")
	})

	t.Run("无效用户ID", func(t *testing.T) {
		// Arrange - 准备测试数据
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		invalidUserID := "invalid-id"

		// Act - 执行请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+invalidUserID+"/roles", nil)
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert - 断言结果
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应码应该是 400")
	})
}
