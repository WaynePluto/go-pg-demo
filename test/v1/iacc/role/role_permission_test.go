package role_test

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

// TestAssignPermission 测试权限分配功能
// 包含两个子测试：成功分配权限给角色、清空角色权限
func TestAssignPermission(t *testing.T) {
	t.Run("成功分配权限给角色", func(t *testing.T) {
		// 准备
		// 先创建角色
		description := "分配权限测试角色"
		entity := createTestRole(t, "权限分配测试角色", &description)

		// 创建一些测试权限
		perm1 := createTestPermission(t, "权限1")
		perm2 := createTestPermission(t, "权限2")

		assignReqBody := map[string]any{
			"permission_ids": []string{perm1["id"].(string), perm2["id"].(string)},
		}

		bodyBytes, _ := json.Marshal(assignReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role/"+entity["id"].(string)+"/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		// 验证权限已分配
		var count int64
		query := `SELECT COUNT(*) FROM iacc_role_permission WHERE role_id = $1`
		err := testDB.GetContext(context.Background(), &count, query, entity["id"])
		assert.NoError(t, err, "应该能够执行计数查询")
		assert.Equal(t, int64(2), count, "应该有两条权限关联记录")
	})

	t.Run("清空角色权限", func(t *testing.T) {
		// 准备
		// 先创建角色
		description := "清空权限测试角色"
		entity := createTestRole(t, "清空权限测试角色", &description)

		// 创建测试权限并关联到角色
		perm := createTestPermission(t, "权限A")

		// 手动插入关联记录
		_, err := testDB.ExecContext(context.Background(),
			"INSERT INTO iacc_role_permission (role_id, permission_id) VALUES ($1, $2)",
			entity["id"], perm["id"])
		assert.NoError(t, err, "应该能成功插入权限关联")

		// 发送空权限列表请求应该失败，因为验证规则要求至少一个权限
		assignReqBody := map[string]any{
			"permission_ids": []string{}, // 空列表
		}

		bodyBytes, _ := json.Marshal(assignReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role/"+entity["id"].(string)+"/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200（统一错误响应格式）")
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应中的Code应该是400")
	})
}

// TestRoleHandler_GetPermissions 测试查询角色权限列表功能
// 包含三个子测试：成功查询角色权限列表、查询不存在角色的权限列表、查询没有权限的角色的权限列表
func TestRoleHandler_GetPermissions(t *testing.T) {
	t.Run("成功查询角色权限列表", func(t *testing.T) {
		// 准备
		// 创建测试角色
		description := "权限查询测试角色"
		role := createTestRole(t, "权限查询测试角色", &description)

		// 创建测试权限
		perm1 := createTestPermission(t, "查询权限1")
		perm2 := createTestPermission(t, "查询权限2")

		// 手动关联权限到角色
		_, err := testDB.ExecContext(context.Background(),
			"INSERT INTO iacc_role_permission (role_id, permission_id) VALUES ($1, $2), ($1, $3)",
			role["id"], perm1["id"], perm2["id"])
		assert.NoError(t, err, "应该能成功插入权限关联")

		// 创建请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+role["id"].(string)+"/permission", nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		// 解析响应
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")

		// 验证响应数据
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是map类型")

		// 验证权限列表
		list, ok := data["list"].([]any)
		assert.True(t, ok, "list字段应该是数组类型")
		assert.Equal(t, 2, len(list), "应该返回两个权限")

		// 验证总数
		total, ok := data["total"].(float64) // JSON数字默认解析为float64
		assert.True(t, ok, "total字段应该是数字类型")
		assert.Equal(t, float64(2), total, "总数应该是2")
	})

	t.Run("查询不存在角色的权限列表", func(t *testing.T) {
		// 准备
		nonExistentRoleID := "00000000-0000-0000-0000-000000000000" // 不存在的UUID

		// 创建请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+nonExistentRoleID+"/permission", nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200（统一错误响应格式）")

		// 解析响应
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, http.StatusNotFound, resp.Code, "响应中的Code应该是404")
	})

	t.Run("查询没有权限的角色的权限列表", func(t *testing.T) {
		// 准备
		// 创建测试角色（不关联任何权限）
		description := "无权限角色"
		role := createTestRole(t, "无权限角色", &description)

		// 创建请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+role["id"].(string)+"/permission", nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		// 解析响应
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")

		// 验证响应数据
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是map类型")

		// 验证权限列表为空或nil
		if data["list"] == nil {
			// 如果list字段为nil，这也是可接受的
			assert.Nil(t, data["list"], "list字段可以为nil")
		} else {
			// 如果list字段不为nil，则应该是数组类型且为空
			list, ok := data["list"].([]any)
			assert.True(t, ok, "list字段应该是数组类型")
			assert.Equal(t, 0, len(list), "应该返回空列表")
		}

		// 验证总数为0
		total, ok := data["total"].(float64) // JSON数字默认解析为float64
		assert.True(t, ok, "total字段应该是数字类型")
		assert.Equal(t, float64(0), total, "总数应该是0")
	})
}
