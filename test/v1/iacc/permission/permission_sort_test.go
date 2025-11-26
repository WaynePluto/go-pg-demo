package permission_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestQueryListPermissions_Sort 测试权限列表排序功能
func TestQueryListPermissions_Sort(t *testing.T) {
	// 准备数据
	id1 := createSortTestPermission(t, "SortTest_A", "api")
	id2 := createSortTestPermission(t, "SortTest_B", "menu")
	id3 := createSortTestPermission(t, "SortTest_C", "button")

	t.Run("按 ID ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=id&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		// 应该只有三个结果
		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, id1["id"], list[0].(map[string]any)["id"])
			assert.Equal(t, id2["id"], list[1].(map[string]any)["id"])
			assert.Equal(t, id3["id"], list[2].(map[string]any)["id"])
		}
	})

	t.Run("按 ID DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=id&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, id3["id"], list[0].(map[string]any)["id"])
			assert.Equal(t, id2["id"], list[1].(map[string]any)["id"])
			assert.Equal(t, id1["id"], list[2].(map[string]any)["id"])
		}
	})

	t.Run("按 Name ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=name&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, "SortTest_A", list[0].(map[string]any)["name"])
			assert.Equal(t, "SortTest_B", list[1].(map[string]any)["name"])
			assert.Equal(t, "SortTest_C", list[2].(map[string]any)["name"])
		}
	})

	t.Run("按 Name DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=name&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, "SortTest_C", list[0].(map[string]any)["name"])
			assert.Equal(t, "SortTest_B", list[1].(map[string]any)["name"])
			assert.Equal(t, "SortTest_A", list[2].(map[string]any)["name"])
		}
	})

	t.Run("按 Type ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=type&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, "api", list[0].(map[string]any)["type"])
			assert.Equal(t, "button", list[1].(map[string]any)["type"])
			assert.Equal(t, "menu", list[2].(map[string]any)["type"])
		}
	})

	t.Run("按 Type DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=type&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序
			assert.Equal(t, "menu", list[0].(map[string]any)["type"])
			assert.Equal(t, "button", list[1].(map[string]any)["type"])
			assert.Equal(t, "api", list[2].(map[string]any)["type"])
		}
	})

	t.Run("按 CreatedAt ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=created_at&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id1["created_at"].(string), list[0].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id2["created_at"].(string), list[1].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id3["created_at"].(string), list[2].(map[string]any)["created_at"].(string)))
		}
	})

	t.Run("按 CreatedAt DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=created_at&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id3["created_at"].(string), list[0].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id2["created_at"].(string), list[1].(map[string]any)["created_at"].(string)))
			assert.True(t, compareTimestamps(id1["created_at"].(string), list[2].(map[string]any)["created_at"].(string)))
		}
	})

	t.Run("按 UpdatedAt ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=updated_at&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id1["updated_at"].(string), list[0].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id2["updated_at"].(string), list[1].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id3["updated_at"].(string), list[2].(map[string]any)["updated_at"].(string)))
		}
	})

	t.Run("按 UpdatedAt DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?name=SortTest_&orderBy=updated_at&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

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

		assert.Equal(t, 3, len(list))
		if len(list) >= 3 {
			// 验证排序顺序 - 使用时间戳比较函数
			assert.True(t, compareTimestamps(id3["updated_at"].(string), list[0].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id2["updated_at"].(string), list[1].(map[string]any)["updated_at"].(string)))
			assert.True(t, compareTimestamps(id1["updated_at"].(string), list[2].(map[string]any)["updated_at"].(string)))
		}
	})

	t.Run("无效 OrderBy", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?orderBy=invalid_field", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "排序字段不存在", resp.Msg)
	})

	t.Run("无效 Order", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list?orderBy=name&order=invalid_order", nil)
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "排序顺序参数错误", resp.Msg)
	})
}
