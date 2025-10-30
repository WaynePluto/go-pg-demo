package template_test

import (
	"bytes"
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
	query := `INSERT INTO template (name, num) VALUES ($1, $2) RETURNING id`
	err := testDB.QueryRow(query, entity.Name, *entity.Num).Scan(&entity.ID)
	assert.NoError(t, err)

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.Exec("DELETE FROM template WHERE id = $1", entity.ID)
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
		createReqBody := template.CreateTemplateReq{
			Name: "新的测试模板",
			Num:  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

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
			_, err := testDB.Exec("DELETE FROM template WHERE id = $1", createdID)
			assert.NoError(t, err, "清理创建的模板不应出错")
		})
	})

	t.Run("无效输入 - 缺少名称", func(t *testing.T) {
		// 准备
		num := 100
		createReqBody := template.CreateTemplateReq{
			Num: &num, // 缺少名称
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, errResp.Code, "响应码应该是 400")
		assert.Contains(t, errResp.Msg, "Key: 'CreateTemplateReq.Name' Error:Field validation for 'Name' failed on the 'required' tag", "错误消息应包含名称必填的验证错误")
	})
}

func TestGetTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t)

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+entity.ID, nil)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		type TemplateResponse struct {
			Code int                     `json:"code"`
			Data template.TemplateEntity `json:"data"`
		}
		var resp TemplateResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")
		assert.Equal(t, entity.Name, resp.Data.Name, "获取到的模板名称应与创建时一致")
		assert.Equal(t, *entity.Num, *resp.Data.Num, "获取到的模板数量应与创建时一致")
	})

	t.Run("未找到", func(t *testing.T) {
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
		assert.Equal(t, http.StatusNotFound, resp.Code, "响应码应该是 404")
		assert.Equal(t, "Template not found", resp.Msg, "错误消息应为 'Template not found'")
	})
}

func TestUpdateTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		entity := setupTestTemplate(t)
		updateName := "更新后的名称"
		updateReqBody := template.UpdateTemplateReq{
			Name: &updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+entity.ID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

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
		err = testDB.Get(&updatedTemplate, "SELECT * FROM template WHERE id = $1", entity.ID)
		assert.NoError(t, err, "从数据库获取更新后的模板不应出错")
		assert.Equal(t, updateName, updatedTemplate.Name, "模板名称应已更新")
		assert.Equal(t, *entity.Num, *updatedTemplate.Num, "模板数量不应改变")
	})
}

func TestDeleteTemplate(t *testing.T) {
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
		err = testDB.Get(&count, "SELECT COUNT(*) FROM template WHERE id = $1", entity.ID)
		assert.NoError(t, err, "查询已删除模板的计数不应出错")
		assert.Equal(t, 0, count, "删除后模板在数据库中应不存在")
	})
}

func TestBatchCreateTemplates(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		num1, num2 := 200, 300
		batchCreateReq := template.CreateTemplatesReq{
			Templates: []template.CreateTemplateReq{
				{Name: "批量模板 1", Num: &num1},
				{Name: "批量模板 2", Num: &num2},
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

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
				_, err := testDB.Exec("DELETE FROM template WHERE id = $1", id.(string))
				assert.NoError(t, err, "清理批量创建的模板不应出错")
			}
		})
	})
}
