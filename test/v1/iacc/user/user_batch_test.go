package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// TestBatchCreateUsers 测试批量创建用户功能
// 包含两个子测试：成功批量创建、无效输入-空列表
func TestBatchCreateUsers(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		batchCreateReq := map[string]any{
			"users": []map[string]any{
				{"username": "批量用户1_" + uuid.NewString()[:8], "phone": "138" + uuid.NewString()[:8], "password": "password123"},
				{"username": "批量用户2_" + uuid.NewString()[:8], "phone": "138" + uuid.NewString()[:8], "password": "password123"},
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		assert.Len(t, ids, 2, "应创建 2 个用户")

		// 清理
		t.Cleanup(func() {
			for _, id := range ids {
				_, err := testDB.NamedExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = :id`, map[string]any{"id": id.(string)})
				assert.NoError(t, err, "清理批量创建的用户不应出错")
			}
		})
	})

	t.Run("无效输入 - 空列表", func(t *testing.T) {
		// 准备
		batchCreateReq := map[string]any{
			"users": []map[string]any{},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		assert.Contains(t, errResp.Msg, "用户列表必须至少包含1项")
	})
}

// TestBatchDeleteUsers 测试批量删除用户功能
// 包含两个子测试：成功批量删除、无效输入-空ID列表
func TestBatchDeleteUsers(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity1 := setupTestUser(t)
		entity2 := setupTestUser(t)
		deleteReq := map[string]any{
			"ids": []string{entity1["id"].(string), entity2["id"].(string)},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-delete", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		query, args, err := sqlx.In("SELECT COUNT(*) FROM \"iacc_user\" WHERE id IN (?)", []string{entity1["id"].(string), entity2["id"].(string)})
		assert.NoError(t, err)
		query = testDB.Rebind(query)
		var count int
		err = testDB.GetContext(context.Background(), &count, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "用户应已被删除")
	})

	t.Run("无效输入 - 空ID列表", func(t *testing.T) {
		// 准备
		deleteReq := map[string]any{
			"ids": []string{},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/batch-delete", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
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
		assert.Contains(t, errResp.Msg, "用户ID列表必须至少包含1项")
	})
}
