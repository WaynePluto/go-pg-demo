package role_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestQueryListRoles_Sort 测试角色列表排序功能
// 包含按ID、Name、Description、CreatedAt、UpdatedAt的升序和降序排序测试
func TestQueryListRoles_Sort(t *testing.T) {
	// 准备数据
	desc1 := "描述A"
	desc2 := "描述B"
	desc3 := "描述C"
	id1 := createTestRole(t, "SortTest_A", &desc1)
	id2 := createTestRole(t, "SortTest_B", &desc2)
	id3 := createTestRole(t, "SortTest_C", &desc3)

	t.Run("按 ID ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=id&order=asc", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=id&order=desc", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=name&order=asc", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=name&order=desc", nil)
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

	t.Run("按 Description ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=description&order=asc", nil)
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
			assert.Equal(t, "描述A", list[0].(map[string]any)["description"])
			assert.Equal(t, "描述B", list[1].(map[string]any)["description"])
			assert.Equal(t, "描述C", list[2].(map[string]any)["description"])
		}
	})

	t.Run("按 Description DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=description&order=desc", nil)
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
			assert.Equal(t, "描述C", list[0].(map[string]any)["description"])
			assert.Equal(t, "描述B", list[1].(map[string]any)["description"])
			assert.Equal(t, "描述A", list[2].(map[string]any)["description"])
		}
	})

	t.Run("按 CreatedAt ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=created_at&order=asc", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=created_at&order=desc", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=updated_at&order=asc", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=SortTest_&orderBy=updated_at&order=desc", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?orderBy=invalid_field", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?orderBy=name&order=invalid_order", nil)
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

// compareTimestamps 比较两个时间戳字符串，只比较到秒级精度
func compareTimestamps(t1, t2 string) bool {
	// 解析时间戳
	parsedT1, err1 := time.Parse(time.RFC3339Nano, t1)
	parsedT2, err2 := time.Parse(time.RFC3339Nano, t2)

	if err1 != nil || err2 != nil {
		// 如果解析失败，尝试其他格式
		parsedT1, err1 = time.Parse("2006-01-02 15:04:05.999999", t1)
		parsedT2, err2 = time.Parse("2006-01-02 15:04:05.999999", t2)

		if err1 != nil || err2 != nil {
			// 如果还是失败，直接比较字符串（去掉微秒部分）
			t1Sec := strings.Split(t1, ".")[0]
			t2Sec := strings.Split(t2, ".")[0]
			return t1Sec == t2Sec
		}
	}

	// 比较到秒级精度
	return parsedT1.Truncate(time.Second).Equal(parsedT2.Truncate(time.Second))
}
