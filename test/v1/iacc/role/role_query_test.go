package role_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestQueryRoleList 测试角色列表查询功能
// 包含两个子测试：成功查询角色列表、带名称筛选查询角色列表
func TestQueryRoleList(t *testing.T) {
	t.Run("成功查询角色列表", func(t *testing.T) {
		// 准备测试数据
		description1 := "测试角色1描述"
		description2 := "测试角色2描述"
		createTestRole(t, "列表测试角色1", &description1)
		createTestRole(t, "列表测试角色2", &description2)

		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?page=1&pageSize=10", nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.NotNil(t, resp.Data, "返回数据不应该为空")

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "数据应该是对象类型")
		assert.Contains(t, data, "list", "数据应该包含list字段")
		assert.Contains(t, data, "total", "数据应该包含total字段")
	})

	t.Run("带名称筛选查询角色列表", func(t *testing.T) {
		// 准备测试数据
		description := "筛选测试描述"
		createTestRole(t, "筛选测试角色ABC", &description)
		createTestRole(t, "另一个测试角色", &description)

		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=ABC", nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")

		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "数据应该是对象类型")
		list := data["list"].([]any)
		assert.Equal(t, 1, len(list), "应该只返回一条匹配的记录")
	})
}
