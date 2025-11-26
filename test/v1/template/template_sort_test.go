package template_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

func createSortTestTemplate(t *testing.T, name string, num int) string {
	t.Helper()

	query := `INSERT INTO template (name, num) VALUES (:name, :num) RETURNING id`
	params := map[string]any{
		"name": name,
		"num":  num,
	}

	// 使用 NamedQueryContext 获取 id
	rows, err := testDB.NamedQueryContext(context.Background(), query, params)
	assert.NoError(t, err)
	defer rows.Close()

	var id string
	if rows.Next() {
		err = rows.Scan(&id)
		assert.NoError(t, err)
	}

	t.Cleanup(func() {
		testDB.Exec("DELETE FROM template WHERE id = $1", id)
	})

	return id
}

func TestQueryListTemplates_Sort(t *testing.T) {
	// 准备数据
	id1 := createSortTestTemplate(t, "SortTest_A", 10)
	id2 := createSortTestTemplate(t, "SortTest_B", 20)

	t.Run("按 Num ASC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=SortTest_&orderBy=num&order=asc", nil)
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

	t.Run("按 Num DESC 排序", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=SortTest_&orderBy=num&order=desc", nil)
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

	t.Run("无效 OrderBy", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?orderBy=invalid_field", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?orderBy=num&order=invalid_order", nil)
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
