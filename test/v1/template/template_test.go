package template_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-pg-demo/internal/app"
	"go-pg-demo/internal/modules/template"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testDB     *sqlx.DB
	testLogger *zap.Logger
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建应用实例
	testApp, _, err := app.InitializeApp()
	if err != nil {
		// 如果应用初始化失败，直接退出
		os.Exit(1)
	}

	testDB = testApp.DB
	testLogger = testApp.Logger
	testRouter = testApp.Server

	// 运行测试
	exitCode := m.Run()

	// 退出
	os.Exit(exitCode)
}

// 在数据库中创建一个模板用于测试，并注册一个清理函数以便在测试结束后删除它
func setupTestTemplate(t *testing.T) template.TemplateEntity {
	t.Helper()

	num := 100
	entity := template.TemplateEntity{
		Name: "Test Template",
		Num:  &num,
	}

	// 直接在数据库中创建实体
	query := `INSERT INTO template (name, num) VALUES (:name, :num) RETURNING id, created_at, updated_at`
	stmt, err := testDB.PrepareNamedContext(context.Background(), query)
	assert.NoError(t, err)
	defer stmt.Close()

	err = stmt.GetContext(context.Background(), &entity, entity)
	assert.NoError(t, err)

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM template WHERE id = :id", map[string]interface{}{"id": entity.ID})
		if err != nil {
			t.Errorf("清理测试模板失败: %v", err)
		}
	})

	return entity
}

func TestCreateTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		num := 100
		createReqBody := template.CreateOneReq{
			Name: "新的测试模板",
			Num:  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
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
		assert.Equal(t, 200, createResp.Code, "响应码应该是 200")

		createdID, ok := createResp.Data.(string)
		assert.True(t, ok, "响应数据应该是字符串类型的 ID")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM template WHERE id = :id", map[string]interface{}{"id": createdID})
			assert.NoError(t, err, "清理创建的模板不应出错")
		})
	})

	t.Run("无效输入 - 缺少名称", func(t *testing.T) {
		// 准备
		num := 100
		createReqBody := template.CreateOneReq{
			Num: &num, // 缺少名称
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
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
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, errResp.Code, "响应码应该是 400")
		assert.Contains(t, errResp.Msg, "模板名称为必填字段", "错误消息应包含名称必填的验证错误")
	})

	t.Run("无效输入 - Num 超出范围", func(t *testing.T) {
		// 准备
		num := 0 // 无效的 Num
		createReqBody := template.CreateOneReq{
			Name: "无效Num模板",
			Num:  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
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
		assert.Contains(t, errResp.Msg, "模板数量最小只能为1")
	})
}

func TestGetByIdTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t)

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+entity.ID, nil)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 使用 map 来灵活处理响应数据
		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "响应数据应该是一个 map")

		assert.Equal(t, entity.Name, data["name"], "获取到的模板名称应与创建时一致")

		// Num 可能为 null，需要小心处理
		if num, ok := data["num"]; ok && num != nil {
			assert.Equal(t, float64(*entity.Num), num, "获取到的模板数量应与创建时一致")
		} else {
			assert.Nil(t, entity.Num, "如果响应中没有 num，则原始数据也应为 nil")
		}
	})

	t.Run("错误的ID", func(t *testing.T) {
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

	t.Run("未找到", func(t *testing.T) {
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
}

func TestUpdateByIdTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t)
		updateName := "更新后的名称"
		updateReqBody := template.UpdateOneReq{
			Name: &updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+entity.ID, bytes.NewBuffer(bodyBytes))
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
		var updateResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, updateResp.Code, "响应码应该是 200")

		// 验证更新
		var updatedTemplate template.TemplateEntity
		err = testDB.GetContext(context.Background(), &updatedTemplate, "SELECT * FROM template WHERE id = $1", entity.ID)
		assert.NoError(t, err, "从数据库获取更新后的模板不应出错")
		assert.Equal(t, updateName, updatedTemplate.Name, "模板名称应已更新")
		assert.Equal(t, *entity.Num, *updatedTemplate.Num, "模板数量不应改变")
	})

	t.Run("无效输入 - Num 超出范围", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t)
		invalidNum := 0
		updateReqBody := template.UpdateOneReq{
			Num: &invalidNum,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+entity.ID, bytes.NewBuffer(bodyBytes))
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
		assert.Contains(t, errResp.Msg, "模板数量最小只能为1")
	})
}

func TestDeleteByIdTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t)

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/template/"+entity.ID, nil)

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
		err = testDB.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM template WHERE id = $1", entity.ID)
		assert.NoError(t, err, "查询已删除模板的计数不应出错")
		assert.Equal(t, 0, count, "删除后模板在数据库中应不存在")
	})
}

func TestBatchCreateTemplates(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		num1, num2 := 200, 300
		batchCreateReq := template.BatchCreateReq{
			Templates: []template.CreateOneReq{
				{Name: "批量模板 1", Num: &num1},
				{Name: "批量模板 2", Num: &num2},
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
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

		ids, ok := createResp.Data.([]interface{})
		assert.True(t, ok, "响应数据应该是 ID 数组")
		assert.Len(t, ids, 2, "应创建 2 个模板")

		// 清理
		t.Cleanup(func() {
			for _, id := range ids {
				_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM template WHERE id = :id", map[string]interface{}{"id": id.(string)})
				assert.NoError(t, err, "清理批量创建的模板不应出错")
			}
		})
	})

	t.Run("无效输入 - 空列表", func(t *testing.T) {
		// 准备
		batchCreateReq := template.BatchCreateReq{
			Templates: []template.CreateOneReq{},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
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
		assert.Contains(t, errResp.Msg, "模板列表必须至少包含1项")
	})

	t.Run("无效输入 - 列表中有无效数据", func(t *testing.T) {
		// 准备
		num := 100
		batchCreateReq := template.BatchCreateReq{
			Templates: []template.CreateOneReq{
				{Name: "有效模板", Num: &num},
				{Num: &num}, // 无效模板，缺少 Name
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
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
		assert.Contains(t, errResp.Msg, "模板名称为必填字段")
	})
}

func TestBatchDeleteTemplates(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity1 := setupTestTemplate(t)
		entity2 := setupTestTemplate(t)
		deleteReq := template.DeleteTemplatesReq{
			IDs: []string{entity1.ID, entity2.ID},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-delete", bytes.NewBuffer(bodyBytes))
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
		query, args, err := sqlx.In("SELECT COUNT(*) FROM template WHERE id IN (?)", []string{entity1.ID, entity2.ID})
		assert.NoError(t, err)
		query = testDB.Rebind(query)
		var count int
		err = testDB.GetContext(context.Background(), &count, query, args...)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "模板应已被删除")
	})

	t.Run("无效输入 - 空ID列表", func(t *testing.T) {
		// 准备
		deleteReq := template.DeleteTemplatesReq{
			IDs: []string{},
		}
		bodyBytes, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-delete", bytes.NewBuffer(bodyBytes))
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
		assert.Contains(t, errResp.Msg, "模板ID列表必须至少包含1项")
	})
}

func TestQueryListTemplates(t *testing.T) {
	t.Run("成功 - 无参数", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t) // 确保至少有一个模板存在

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Greater(t, int(data["total"].(float64)), 0)
		list, ok := data["list"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, list)
		// 确保返回的第一个元素是我们刚创建的
		firstItem := list[0].(map[string]interface{})
		assert.Equal(t, entity.ID, firstItem["id"])
	})

	t.Run("成功 - 带名称搜索", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t)

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=Test Template", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(1), data["total"])
		list, ok := data["list"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, list, 1)
		assert.Equal(t, entity.ID, list[0].(map[string]interface{})["id"])
	})

	t.Run("成功 - 结果为空", func(t *testing.T) {
		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?name=nonexistent", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(0), data["total"])
		list, ok := data["list"].([]interface{})
		assert.True(t, ok)
		assert.Empty(t, list)
	})

	t.Run("无效输入 - pageSize 过大", func(t *testing.T) {
		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?pageSize=200", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errResp.Code)
		assert.Contains(t, errResp.Msg, "每页大小必须小于或等于100")
	})
}
