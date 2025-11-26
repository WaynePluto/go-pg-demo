package template_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestQueryListTemplates 测试模板列表查询功能
// 包含基本查询、分页查询、搜索查询和无效参数测试
func TestQueryListTemplates(t *testing.T) {
	t.Run("成功查询", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil) // 确保至少有一个模板存在

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list", nil)
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

	t.Run("分页查询", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil)

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?page=1&pageSize=10", nil)
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

	t.Run("搜索查询", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil)

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name="+entity["name"].(string), nil)
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
		assert.Equal(t, float64(1), data["total"])
		list, ok := data["list"].([]any)
		assert.True(t, ok)
		assert.Len(t, list, 1)
		assert.Equal(t, entity["id"], list[0].(map[string]any)["id"])
	})

	t.Run("无效参数", func(t *testing.T) {
		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?pageSize=200", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errResp.Code)
		assert.Contains(t, errResp.Msg, "每页大小必须小于或等于100")
	})
}

// TestQueryListTemplates_Sort 测试模板列表排序功能
// 包含按名称升序、按名称降序、按创建时间升序、按创建时间降序排序测试
func TestQueryListTemplates_Sort(t *testing.T) {
	// 准备数据
	id1 := createSortTestTemplate(t, "SortTest_A", 10)
	id2 := createSortTestTemplate(t, "SortTest_B", 20)

	t.Run("按名称升序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=SortTest_&orderBy=name&order=asc", nil)
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

		// 应该只有两个结果
		assert.Equal(t, 2, len(list))
		if len(list) >= 2 {
			assert.Equal(t, id1, list[0].(map[string]any)["id"])
			assert.Equal(t, id2, list[1].(map[string]any)["id"])
		}
	})

	t.Run("按名称降序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=SortTest_&orderBy=name&order=desc", nil)
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

		assert.Equal(t, 2, len(list))
		if len(list) >= 2 {
			assert.Equal(t, id2, list[0].(map[string]any)["id"])
			assert.Equal(t, id1, list[1].(map[string]any)["id"])
		}
	})

	t.Run("按创建时间升序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=SortTest_&orderBy=created_at&order=asc", nil)
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

		assert.Equal(t, 2, len(list))
		if len(list) >= 2 {
			assert.Equal(t, id1, list[0].(map[string]any)["id"])
			assert.Equal(t, id2, list[1].(map[string]any)["id"])
		}
	})

	t.Run("按创建时间降序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=SortTest_&orderBy=created_at&order=desc", nil)
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

		assert.Equal(t, 2, len(list))
		if len(list) >= 2 {
			assert.Equal(t, id2, list[0].(map[string]any)["id"])
			assert.Equal(t, id1, list[1].(map[string]any)["id"])
		}
	})
}
