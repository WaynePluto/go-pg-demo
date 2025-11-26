package template_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/pkgs"

	"github.com/stretchr/testify/assert"
)

// TestCreateTemplate 测试创建模板功能
// 包含三个子测试：成功创建、无效数据、已存在名称
func TestCreateTemplate(t *testing.T) {
	t.Run("成功创建", func(t *testing.T) {
		// 准备
		num := 100
		createReqBody := map[string]any{
			"name": "新的测试模板",
			"num":  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
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
		assert.Equal(t, 200, createResp.Code, "响应码应该是 200")

		createdID, ok := createResp.Data.(string)
		assert.True(t, ok, "响应数据应该是字符串类型的 ID")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM template WHERE id = :id", map[string]any{"id": createdID})
			assert.NoError(t, err, "清理创建的模板不应出错")
		})
	})

	t.Run("无效数据", func(t *testing.T) {
		// 准备
		num := 100
		createReqBody := map[string]any{
			"num": &num, // 缺少名称
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, errResp.Code, "响应码应该是 400")
		assert.Contains(t, errResp.Msg, "模板名称为必填字段", "错误消息应包含名称必填的验证错误")
	})

	t.Run("已存在名称", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil)

		num := 100
		createReqBody := map[string]any{
			"name": entity["name"], // 使用已存在的名称
			"num":  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言 - API允许创建同名模板
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, 200, createResp.Code, "响应码应该是 200")

		createdID, ok := createResp.Data.(string)
		assert.True(t, ok, "响应数据应该是字符串类型的 ID")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM template WHERE id = :id", map[string]any{"id": createdID})
			assert.NoError(t, err, "清理创建的模板不应出错")
		})
	})
}

// TestGetByIdTemplate 测试根据ID获取模板功能
// 包含三个子测试：成功获取、不存在的ID、无效ID格式
func TestGetByIdTemplate(t *testing.T) {
	t.Run("成功获取", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil)

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+entity["id"].(string), nil)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 使用 map 来灵活处理响应数据
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是一个 map")

		assert.Equal(t, entity["name"], data["name"], "获取到的模板名称应与创建时一致")

		// Num 可能为 null，需要小心处理
		if num, ok := data["num"]; ok && num != nil {
			assert.Equal(t, float64(*entity["num"].(*int)), num, "获取到的模板数量应与创建时一致")
		} else {
			assert.Nil(t, entity["num"], "如果响应中没有 num，则原始数据也应为 nil")
		}
	})

	t.Run("不存在的ID", func(t *testing.T) {
		// 准备
		nonExistentID := "123e4567-e89b-12d3-a456-426614174000"
		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+nonExistentID, nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "处理器应返回 200 状态码，但在响应体中包含错误码")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusNotFound, resp.Code, "响应码应该是 404")
		assert.Equal(t, "模板不存在", resp.Msg, "错误消息应为 '模板不存在'")
	})

	t.Run("无效ID格式", func(t *testing.T) {
		// 准备
		nonExistentID := "a-b-c-d-e"

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+nonExistentID, nil)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "处理器应返回 200 状态码，但在响应体中包含错误码")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应码应该是 400")
		assert.Equal(t, "模板ID必须是一个有效的UUID", resp.Msg, "错误消息应为 '模板ID必须是一个有效的UUID'")
	})
}

// TestUpdateByIdTemplate 测试根据ID更新模板功能
// 包含两个子测试：成功更新、不存在的ID
func TestUpdateByIdTemplate(t *testing.T) {
	t.Run("成功更新", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil)
		updateName := "更新后的名称"
		updateReqBody := map[string]any{
			"name": updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+entity["id"].(string), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var updateResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, updateResp.Code, "响应码应该是 200")

		// 验证更新
		type UpdatedTemplate struct {
			Name string `db:"name"`
			Num  *int   `db:"num"`
		}
		var updatedTemplate UpdatedTemplate
		err = testDB.GetContext(context.Background(), &updatedTemplate, "SELECT name, num FROM template WHERE id = $1", entity["id"])
		assert.NoError(t, err, "从数据库获取更新后的模板不应出错")
		assert.Equal(t, updateName, updatedTemplate.Name, "模板名称应已更新")
		assert.Equal(t, *entity["num"].(*int), *updatedTemplate.Num, "模板数量不应改变")
	})

	t.Run("不存在的ID", func(t *testing.T) {
		// 准备
		nonExistentID := "123e4567-e89b-12d3-a456-426614174000"
		updateName := "更新后的名称"
		updateReqBody := map[string]any{
			"name": updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+nonExistentID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// 获取token
		token := getAuthToken(t, []string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言 - API返回成功，但影响行数为0
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var updateResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, 200, updateResp.Code, "响应码应该是 200")

		// 验证影响行数为0（因为记录不存在）
		affectedRows := int(updateResp.Data.(float64))
		assert.Equal(t, 0, affectedRows, "影响行数应为0，因为记录不存在")
	})
}

// TestDeleteByIdTemplate 测试根据ID删除模板功能
func TestDeleteByIdTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := createTestTemplate(t, "", nil)

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/template/"+entity["id"].(string), nil)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var deleteResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, deleteResp.Code, "响应码应该是 200")
		affectedRows := int(math.Round(deleteResp.Data.(float64)))
		assert.Equal(t, 1, affectedRows, "应影响 1 行")

		// 验证删除
		var count int
		err = testDB.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM template WHERE id = $1", entity["id"])
		assert.NoError(t, err, "查询已删除模板的计数不应出错")
		assert.Equal(t, 0, count, "删除后模板在数据库中应不存在")
	})

	t.Run("不存在的ID", func(t *testing.T) {
		// 准备
		nonExistentID := "123e4567-e89b-12d3-a456-426614174000"

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/template/"+nonExistentID, nil)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言 - API返回成功，但影响行数为0
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var deleteResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, 200, deleteResp.Code, "响应码应该是 200")

		// 验证影响行数为0（因为记录不存在）
		affectedRows := int(deleteResp.Data.(float64))
		assert.Equal(t, 0, affectedRows, "影响行数应为0，因为记录不存在")
	})
}
