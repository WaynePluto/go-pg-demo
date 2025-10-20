package template

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	v1 "go-pg-demo/api/v1"
	"go-pg-demo/internal/middlewares"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testDB        *sqlx.DB
	testLogger    *zap.Logger
	testValidator *pkgs.RequestValidator
	testRouter    *v1.Router
)

func TestMain(m *testing.M) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化日志记录器
	testLogger, _ = zap.NewDevelopment()
	defer testLogger.Sync()

	// 使用 pkgs.NewConfig 函数加载配置
	testConfig, err := pkgs.NewConfig()
	if err != nil {
		testLogger.Fatal("加载配置失败", zap.Error(err))
	}

	// 数据库连接
	testDB, err = pkgs.NewConnection(testConfig)
	if err != nil {
		testLogger.Fatal("连接数据库失败", zap.Error(err))
	}
	defer testDB.Close()

	// 创建验证器实例
	testValidator = pkgs.NewRequestValidator()

	// 创建中间件实例
	authMiddleware := middlewares.NewAuthMiddleware(testConfig, testLogger)

	// 创建处理器实例
	testHandler := NewTemplateHandler(testDB, testLogger, testValidator)

	// 设置路由
	engine := gin.New()
	testRouter = &v1.Router{
		Engine:          engine,
		TemplateHandler: testHandler,
	}
	// 注册全局中间件
	testRouter.Engine.Use(gin.HandlerFunc(authMiddleware))
	// 注册路由组
	testRouter.RouterGroup = testRouter.Engine.Group("/v1")
	testRouter.RegisterTemplate()

	// 运行测试
	exitCode := m.Run()

	// 退出
	os.Exit(exitCode)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	testRouter.Engine.ServeHTTP(rr, req)
	return rr
}

// setupTestTemplate 在数据库中创建一个模板用于测试，并注册一个清理函数以便在测试结束后删除它
func setupTestTemplate(t *testing.T) TemplateEntity {
	t.Helper()

	num := 100
	template := TemplateEntity{
		Name: "Test Template",
		Num:  &num,
	}

	// 直接在数据库中创建实体
	query := `INSERT INTO template (name, num) VALUES ($1, $2) RETURNING id`
	err := testDB.QueryRow(query, template.Name, *template.Num).Scan(&template.ID)
	assert.NoError(t, err)

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.Exec("DELETE FROM template WHERE id = $1", template.ID)
		if err != nil {
			t.Errorf("清理测试模板失败: %v", err)
		}
	})

	return template
}

func TestCreateTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		num := 100
		createReqBody := CreateTemplateRequest{
			Name: "新的测试模板",
			Num:  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 执行
		w := executeRequest(req)

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
		createReqBody := CreateTemplateRequest{
			Num: &num, // 缺少名称
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 执行
		w := executeRequest(req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err, "解析错误响应体不应出错")
		assert.Equal(t, http.StatusBadRequest, errResp.Code, "响应码应该是 400")
		assert.Contains(t, errResp.Msg, "Key: 'CreateTemplateRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag", "错误消息应包含名称必填的验证错误")
	})
}

func TestGetTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		template := setupTestTemplate(t)

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+template.ID, nil)
		w := executeRequest(req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		type TemplateResponse struct {
			Code int            `json:"code"`
			Data TemplateEntity `json:"data"`
		}
		var resp TemplateResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")
		assert.Equal(t, template.Name, resp.Data.Name, "获取到的模板名称应与创建时一致")
		assert.Equal(t, *template.Num, *resp.Data.Num, "获取到的模板数量应与创建时一致")
	})

	t.Run("未找到", func(t *testing.T) {
		// 准备
		nonExistentID := "a-b-c-d-e"

		// 执行
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+nonExistentID, nil)
		w := executeRequest(req)

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
		template := setupTestTemplate(t)
		updateName := "更新后的名称"
		updateReqBody := UpdateTemplateRequest{
			Name: &updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+template.ID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 执行
		w := executeRequest(req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var updateResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, updateResp.Code, "响应码应该是 200")

		// 验证更新
		var updatedTemplate TemplateEntity
		err = testDB.Get(&updatedTemplate, "SELECT * FROM template WHERE id = $1", template.ID)
		assert.NoError(t, err, "从数据库获取更新后的模板不应出错")
		assert.Equal(t, updateName, updatedTemplate.Name, "模板名称应已更新")
		assert.Equal(t, *template.Num, *updatedTemplate.Num, "模板数量不应改变")
	})
}

func TestDeleteTemplate(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		template := setupTestTemplate(t)

		// 执行
		req, _ := http.NewRequest(http.MethodDelete, "/v1/template/"+template.ID, nil)
		w := executeRequest(req)

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
		err = testDB.Get(&count, "SELECT COUNT(*) FROM template WHERE id = $1", template.ID)
		assert.NoError(t, err, "查询已删除模板的计数不应出错")
		assert.Equal(t, 0, count, "删除后模板在数据库中应不存在")
	})
}

func TestBatchCreateTemplates(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		num1, num2 := 200, 300
		batchCreateReq := CreateTemplatesRequest{
			Templates: []CreateTemplateRequest{
				{Name: "批量模板 1", Num: &num1},
				{Name: "批量模板 2", Num: &num2},
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 执行
		w := executeRequest(req)

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

func TestQueryTemplateList(t *testing.T) {
	// 准备
	template1 := setupTestTemplate(t)
	template2 := setupTestTemplate(t)

	// 执行
	req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?page=1&pageSize=10", nil)
	w := executeRequest(req)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")

	type QueryResponseData struct {
		List  []TemplateEntity `json:"list"`
		Total int64            `json:"total"`
	}
	type QueryResponse struct {
		Code int               `json:"code"`
		Data QueryResponseData `json:"data"`
	}

	var queryResp QueryResponse
	err := json.Unmarshal(w.Body.Bytes(), &queryResp)
	assert.NoError(t, err, "解析响应体不应出错")
	assert.Equal(t, http.StatusOK, queryResp.Code, "响应码应该是 200")
	assert.True(t, queryResp.Data.Total >= 2, "总数应至少为 2")

	retrievedIDs := make(map[string]bool)
	for _, item := range queryResp.Data.List {
		retrievedIDs[item.ID] = true
	}
	assert.True(t, retrievedIDs[template1.ID], "结果中应包含第一个测试模板")
	assert.True(t, retrievedIDs[template2.ID], "结果中应包含第二个测试模板")
}

func TestBatchDeleteTemplates(t *testing.T) {
	// 准备
	template1 := setupTestTemplate(t)
	template2 := setupTestTemplate(t)
	idsToDelete := []string{template1.ID, template2.ID}

	deleteReqBody := DeleteTemplatesRequest{
		IDs: idsToDelete,
	}
	bodyBytes, _ := json.Marshal(deleteReqBody)
	req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-delete", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 执行
	w := executeRequest(req)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
	var deleteResp pkgs.Response
	err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
	assert.NoError(t, err, "解析响应体不应出错")
	assert.Equal(t, http.StatusOK, deleteResp.Code, "响应码应该是 200")

	// 验证删除
	var count int
	query, args, err := sqlx.In("SELECT COUNT(*) FROM template WHERE id IN (?)", idsToDelete)
	assert.NoError(t, err, "创建 IN 查询不应出错")
	query = testDB.Rebind(query)
	err = testDB.Get(&count, query, args...)
	assert.NoError(t, err, "查询已删除模板的计数不应出错")
	assert.Equal(t, 0, count, "批量删除后模板在数据库中应不存在")
}
