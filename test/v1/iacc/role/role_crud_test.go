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

// TestCreateRole 测试创建角色功能
// 包含两个子测试：成功创建、缺少必填字段
func TestCreateRole(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		description := "测试角色描述"
		createReqBody := map[string]any{
			"name":        "新的测试角色",
			"description": &description,
		}

		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.NotEmpty(t, resp.Data, "返回数据不应该为空")

		// 清理测试数据
		t.Cleanup(func() {
			if resp.Data != nil {
				if roleID, ok := resp.Data.(string); ok {
					_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM iacc_role WHERE id = :id", map[string]any{"id": roleID})
					if err != nil {
						t.Errorf("清理测试角色失败: %v", err)
					}
				}
			}
		})
	})

	t.Run("缺少必填字段", func(t *testing.T) {
		// 准备
		createReqBody := map[string]any{
			// 故意不提供Name字段
		}

		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(bodyBytes))
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
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应中的Code应该是400")
		assert.Equal(t, "角色名称为必填字段", resp.Msg, "应该返回字段验证错误信息")
	})
}

// TestGetRole 测试获取角色功能
// 包含三个子测试：成功获取角色、角色不存在、无效ID
func TestGetRole(t *testing.T) {
	t.Run("成功获取角色", func(t *testing.T) {
		// 准备
		description := "测试角色描述"
		entity := createTestRole(t, "获取测试角色", &description)

		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+entity["id"].(string), nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.NotNil(t, resp.Data, "返回数据不应该为空")

		roleData, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "数据应该是对象类型")
		assert.Equal(t, entity["id"], roleData["id"], "返回的角色ID应该匹配")
		assert.Equal(t, entity["name"], roleData["name"], "返回的角色名称应该匹配")
	})

	t.Run("角色不存在", func(t *testing.T) {
		// 准备
		fakeID := "123e4567-e89b-12d3-a456-426614174000"
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+fakeID, nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		// 根据handler实现，不存在的角色会返回空数据而不是404错误
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Nil(t, resp.Data, "返回数据应该为空")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 准备
		invalidID := "invalid-id"
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+invalidID, nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		// 根据handler实现，无效ID会返回空数据而不是404错误
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Nil(t, resp.Data, "返回数据应该为空")
	})
}

// TestUpdateRole 测试更新角色功能
// 包含两个子测试：成功更新角色、更新不存在的角色
func TestUpdateRole(t *testing.T) {
	t.Run("成功更新角色", func(t *testing.T) {
		// 准备
		description := "原始描述"
		entity := createTestRole(t, "待更新角色", &description)

		newDescription := "更新后的描述"
		updateReqBody := map[string]any{
			"description": &newDescription,
		}

		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/role/"+entity["id"].(string), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "更新角色请求应该返回200状态码")

		// 验证数据库中的值已更新
		type UpdatedRole struct {
			Description *string `db:"description"`
		}
		var updatedRole UpdatedRole
		query := `SELECT description FROM iacc_role WHERE id = $1`
		err := testDB.GetContext(context.Background(), &updatedRole, query, entity["id"])
		assert.NoError(t, err, "应该能够查询到更新后的角色")
		assert.Equal(t, newDescription, *updatedRole.Description, "描述应该已被更新")
	})

	t.Run("更新不存在的角色", func(t *testing.T) {
		// 准备
		fakeID := "123e4567-e89b-12d3-a456-426614174000"
		newDescription := "更新后的描述"
		updateReqBody := map[string]any{
			"description": &newDescription,
		}

		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/role/"+fakeID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		// 根据handler实现，更新不存在的角色不会报错，只是影响0行
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, float64(0), resp.Data, "更新不存在的角色应该影响零行数据")
	})
}

// TestDeleteRole 测试删除角色功能
// 包含两个子测试：成功删除角色、删除不存在的角色
func TestDeleteRole(t *testing.T) {
	t.Run("成功删除角色", func(t *testing.T) {
		// 准备
		description := "待删除角色描述"
		entity := createTestRole(t, "待删除角色", &description)

		req, _ := http.NewRequest(http.MethodDelete, "/v1/role/"+entity["id"].(string), nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, float64(1), resp.Data, "应该影响一行数据")

		// 验证角色确实已被删除
		var count int64
		query := `SELECT COUNT(*) FROM iacc_role WHERE id = $1`
		err = testDB.GetContext(context.Background(), &count, query, entity["id"])
		assert.NoError(t, err, "应该能够执行计数查询")
		assert.Equal(t, int64(0), count, "角色应该已被删除")
	})

	t.Run("删除不存在的角色", func(t *testing.T) {
		// 准备
		fakeID := "123e4567-e89b-12d3-a456-426614174000"
		req, _ := http.NewRequest(http.MethodDelete, "/v1/role/"+fakeID, nil)

		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, float64(0), resp.Data, "应该影响零行数据")
	})
}
