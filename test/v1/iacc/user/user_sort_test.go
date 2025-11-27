package user_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestQueryListUsers_Sort 测试用户列表排序功能
// 包含按ID、用户名、手机号、创建时间、更新时间的升序和降序排序测试
func TestQueryListUsers_Sort(t *testing.T) {
	// 准备数据
	phone1 := fmt.Sprintf("138%s", uuid.NewString()[:8])
	phone2 := fmt.Sprintf("138%s", uuid.NewString()[:8])
	phone3 := fmt.Sprintf("138%s", uuid.NewString()[:8])
	id1 := createSortTestUser(t, "SortTest_A", phone1)
	id2 := createSortTestUser(t, "SortTest_B", phone2)
	id3 := createSortTestUser(t, "SortTest_C", phone3)

	t.Run("按 ID ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=id&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=id&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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

	t.Run("按 Username ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=username&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
			assert.Equal(t, "SortTest_A", list[0].(map[string]any)["username"])
			assert.Equal(t, "SortTest_B", list[1].(map[string]any)["username"])
			assert.Equal(t, "SortTest_C", list[2].(map[string]any)["username"])
		}
	})

	t.Run("按 Username DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=username&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
			assert.Equal(t, "SortTest_C", list[0].(map[string]any)["username"])
			assert.Equal(t, "SortTest_B", list[1].(map[string]any)["username"])
			assert.Equal(t, "SortTest_A", list[2].(map[string]any)["username"])
		}
	})

	t.Run("按 Phone ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=phone&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
			phones := []string{phone1, phone2, phone3}
			// 对手机号进行排序以便比较
			for i := 0; i < len(phones)-1; i++ {
				for j := i + 1; j < len(phones); j++ {
					if phones[i] > phones[j] {
						phones[i], phones[j] = phones[j], phones[i]
					}
				}
			}
			assert.Equal(t, phones[0], list[0].(map[string]any)["phone"])
			assert.Equal(t, phones[1], list[1].(map[string]any)["phone"])
			assert.Equal(t, phones[2], list[2].(map[string]any)["phone"])
		}
	})

	t.Run("按 Phone DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=phone&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
			phones := []string{phone1, phone2, phone3}
			// 对手机号进行降序排序以便比较
			for i := 0; i < len(phones)-1; i++ {
				for j := i + 1; j < len(phones); j++ {
					if phones[i] < phones[j] {
						phones[i], phones[j] = phones[j], phones[i]
					}
				}
			}
			assert.Equal(t, phones[0], list[0].(map[string]any)["phone"])
			assert.Equal(t, phones[1], list[1].(map[string]any)["phone"])
			assert.Equal(t, phones[2], list[2].(map[string]any)["phone"])
		}
	})

	t.Run("按 CreatedAt ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=created_at&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=created_at&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=updated_at&order=asc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?username=SortTest_&orderBy=updated_at&order=desc", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?orderBy=invalid_field", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list?orderBy=username&order=invalid_order", nil)
		req.Header.Set("Content-Type", "application/json")
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
