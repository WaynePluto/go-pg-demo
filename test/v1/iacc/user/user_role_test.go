package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAssignRoles 测试用户角色分配功能
// 包含两个子测试：成功分配、无效ID
func TestAssignRoles(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestUser(t)

		// 先创建一个角色用于测试
		roleID := uuid.NewString()
		_, err := testDB.ExecContext(context.Background(),
			`INSERT INTO "iacc_role" (id, name, description) VALUES ($1, $2, $3)`,
			roleID, "测试角色", "用于测试的角色")
		assert.NoError(t, err)

		// 测试完成后清理角色
		t.Cleanup(func() {
			_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_role" WHERE id = $1`, roleID)
			if err != nil {
				t.Errorf("清理测试角色失败: %v", err)
			}
		})

		roleIDs := []string{roleID}
		assignReq := map[string]any{
			"role_ids": roleIDs,
		}
		bodyBytes, _ := json.Marshal(assignReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/"+entity["id"].(string)+"/role", bytes.NewBuffer(bodyBytes))
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
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		affectedRows := int(math.Round(resp.Data.(float64)))
		assert.Equal(t, 1, affectedRows, "应影响 1 行")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 准备
		roleIDs := []string{uuid.NewString()}
		assignReq := map[string]any{
			"role_ids": roleIDs,
		}
		bodyBytes, _ := json.Marshal(assignReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user/invalid-id/role", bytes.NewBuffer(bodyBytes))
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
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应码应该是 400")
	})
}
