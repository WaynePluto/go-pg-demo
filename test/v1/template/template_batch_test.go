package template_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// TestBatchCreateTemplates 批量创建模板测试
func TestBatchCreateTemplates(t *testing.T) {
	t.Run("成功批量创建", func(t *testing.T) {
		// 准备
		num1, num2 := 200, 300
		batchCreateReq := map[string]any{
			"templates": []map[string]any{
				{"name": "批量模板 1", "num": &num1},
				{"name": "批量模板 2", "num": &num2},
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, createResp.Code, "响应码应该是 200")

		ids, ok := createResp.Data.([]any)
		assert.True(t, ok, "响应数据应该是 ID 数组")
		assert.Len(t, ids, 2, "应创建 2 个模板")

		// 清理
		t.Cleanup(func() {
			for _, id := range ids {
				_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM template WHERE id = :id", map[string]any{"id": id.(string)})
				assert.NoError(t, err, "清理批量创建的模板不应出错")
			}
		})
	})

	t.Run("部分数据无效", func(t *testing.T) {
		// 准备
		num := 100
		batchCreateReq := map[string]any{
			"templates": []map[string]any{
				{"name": "有效模板", "num": &num},
				{"num": &num}, // 无效模板，缺少 Name
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errResp.Code)
		assert.Contains(t, errResp.Msg, "模板名称为必填字段")
	})

	t.Run("空数组", func(t *testing.T) {
		// 准备
		batchCreateReq := map[string]any{
			"templates": []map[string]any{},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errResp.Code)
		assert.Contains(t, errResp.Msg, "模板列表必须至少包含1项")
	})
}

// TestBatchDeleteTemplates 批量删除模板测试
func TestBatchDeleteTemplates(t *testing.T) {
	t.Run("成功批量删除", func(t *testing.T) {
		// 准备
		entity1 := createTestTemplate(t, "", nil)
		entity2 := createTestTemplate(t, "", nil)
		deleteReq := map[string]any{
			"ids": []string{entity1["id"].(string), entity2["id"].(string)},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-delete", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		// 验证
		query, args, err := sqlx.In("SELECT COUNT(*) FROM template WHERE id IN (?)", []string{entity1["id"].(string), entity2["id"].(string)})
		assert.NoError(t, err)
		query = testDB.Rebind(query)
		var count int
		err = testDB.GetContext(context.Background(), &count, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "模板应已被删除")
	})

	t.Run("部分ID不存在", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil)
		nonExistentID := "non-existent-id"
		deleteReq := map[string]any{
			"ids": []string{entity["id"].(string), nonExistentID},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-delete", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code, "当部分ID不存在时，API应返回400错误")

		// 验证存在的模板仍然存在（因为整个操作失败）
		query, args, err := sqlx.In("SELECT COUNT(*) FROM template WHERE id = ?", entity["id"].(string))
		assert.NoError(t, err)
		query = testDB.Rebind(query)
		var count int
		err = testDB.GetContext(context.Background(), &count, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "由于操作失败，存在的模板应仍然存在")
	})
}
