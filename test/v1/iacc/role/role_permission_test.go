package role_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestAssignPermission 测试权限分配功能
// 包含两个子测试：成功分配权限给角色、清空角色权限
func TestAssignPermission(t *testing.T) {
	t.Run("成功分配权限给角色", func(t *testing.T) {
		// 准备
		// 先创建角色
		description := "分配权限测试角色"
		entity := createTestRole(t, "权限分配测试角色", &description)

		// 创建一些测试权限
		perm1 := createTestPermission(t, "权限1")
		perm2 := createTestPermission(t, "权限2")

		assignReqBody := map[string]any{
			"permission_ids": []string{perm1["id"].(string), perm2["id"].(string)},
		}

		bodyBytes, _ := json.Marshal(assignReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role/"+entity["id"].(string)+"/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		// 验证权限已分配
		var count int64
		query := `SELECT COUNT(*) FROM iacc_role_permission WHERE role_id = $1`
		err := testDB.GetContext(context.Background(), &count, query, entity["id"])
		assert.NoError(t, err, "应该能够执行计数查询")
		assert.Equal(t, int64(2), count, "应该有两条权限关联记录")
	})

	t.Run("清空角色权限", func(t *testing.T) {
		// 准备
		// 先创建角色
		description := "清空权限测试角色"
		entity := createTestRole(t, "清空权限测试角色", &description)

		// 创建测试权限并关联到角色
		perm := createTestPermission(t, "权限A")

		// 手动插入关联记录
		_, err := testDB.ExecContext(context.Background(),
			"INSERT INTO iacc_role_permission (role_id, permission_id) VALUES ($1, $2)",
			entity["id"], perm["id"])
		assert.NoError(t, err, "应该能成功插入权限关联")

		// 发送空权限列表请求应该失败，因为验证规则要求至少一个权限
		assignReqBody := map[string]any{
			"permission_ids": []string{}, // 空列表
		}

		bodyBytes, _ := json.Marshal(assignReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role/"+entity["id"].(string)+"/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200（统一错误响应格式）")
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应中的Code应该是400")
	})
}
