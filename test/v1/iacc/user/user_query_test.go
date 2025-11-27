package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestQueryUserList 测试用户列表查询功能
// 包含三个子测试：成功查询、空结果、成功-带用户名搜索
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
