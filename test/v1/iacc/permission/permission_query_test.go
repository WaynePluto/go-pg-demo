package permission_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestQueryPermissionList 测试权限列表查询功能
func TestQueryPermissionList(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备 - 创建多个测试权限
		setupTestPermission(t, "测试权限-Query1")
		setupTestPermission(t, "测试权限-Query2")

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?page=1&pageSize=10", nil)
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
		total, ok := data["total"]
		assert.True(t, ok, "响应数据应该包含 total 字段")
		assert.GreaterOrEqual(t, int64(total.(float64)), int64(2), "总权限数应至少为 2")
	})
}
