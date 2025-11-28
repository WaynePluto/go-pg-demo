package user_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestQueryListWithNullPhone 测试查询包含 NULL phone 值的用户列表
func TestQueryListWithNullPhone(t *testing.T) {
	// 直接在数据库中创建一个有 phone 的用户
	_ = createTestUser(t, "test_null_phone", "13800138000", "password123")

	// 直接在数据库中创建一个无 phone 的用户（phone 为 NULL）
	query := `INSERT INTO "iacc_user" (username, password) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	var result struct {
		ID        string `db:"id"`
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}
	err := testDB.QueryRowContext(context.Background(), query, "test_no_phone", "password123").Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)
	assert.NoError(t, err, "创建无 phone 用户不应出错")

	// 注册清理函数
	t.Cleanup(func() {
		_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = $1`, result.ID)
		if err != nil {
			t.Errorf("清理无 phone 用户失败: %v", err)
		}
	})

	// 创建 TestUtil 实例
	testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	// 获取token
	token := testUtil.GetAccessUserToken([]string{})

	// 使用 HTTP 请求查询用户列表
	req, _ := http.NewRequest("GET", "/v1/user/list?page=1&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "查询用户列表应成功")

	// 解析响应
	var resp pkgs.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "解析响应不应出错")
	assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

	// 使用 map 来灵活处理响应数据
	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok, "响应数据应该是一个 map")

	total := int(data["total"].(float64))
	assert.True(t, total >= 2, fmt.Sprintf("用户数量不足，期望至少2个，实际: %d", total))

	list := data["list"].([]any)
	assert.True(t, len(list) >= 2, "列表长度应该至少为2")

	// 查找我们创建的用户
	var foundUserWithPhone, foundUserWithoutPhone bool
	for _, item := range list {
		user := item.(map[string]any)
		username := user["username"].(string)
		phone := user["phone"].(string)

		if username == "test_null_phone" {
			foundUserWithPhone = true
			assert.Equal(t, "13800138000", phone, "有 phone 用户的 phone 值应正确")
		}
		if username == "test_no_phone" {
			foundUserWithoutPhone = true
			assert.Equal(t, "", phone, "无 phone 用户的 phone 值应为空字符串")
		}
	}

	assert.True(t, foundUserWithPhone, "应找到有 phone 的测试用户")
	assert.True(t, foundUserWithoutPhone, "应找到无 phone 的测试用户")

	t.Log("成功处理包含 NULL phone 值的用户查询")
}
