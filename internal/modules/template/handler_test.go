package template

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-pg-demo/internal/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testHandler *Handler
	testDB      *sqlx.DB
	testLogger  *zap.Logger
)

func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize logger
	testLogger, _ = zap.NewDevelopment()
	defer testLogger.Sync()

	// Load config using the pkgs.NewConfig function
	config, err := pkgs.NewConfig()
	if err != nil {
		testLogger.Fatal("Failed to load config", zap.Error(err))
	}

	// Database connection
	testDB, err = pkgs.NewConnection(config)
	if err != nil {
		testLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer testDB.Close()

	// Create handler instance
	testHandler = NewTemplateHandler(testDB, testLogger)

	// Run tests
	exitCode := m.Run()

	// Exit
	os.Exit(exitCode)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router := gin.Default()
	// 使用 routes.go 中的方法注册路由
	apiGroup := router.Group("/v1")
	testHandler.RegisterRoutesV1(apiGroup)
	router.ServeHTTP(rr, req)
	return rr
}

// TestTemplateHandler_HandleOne tests the creation, retrieval, updating, and deletion of a template
func TestTemplateHandler_HandleOne(t *testing.T) {
	var createdID string
	var ok bool
	// --- Create Template ---
	{
		num := 100
		createReqBody := CreateTemplateRequest{
			Name: "Test Template",
			Num:  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)

		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, createResp.Code)
		// Extract the ID from the response data
		createdID, ok = createResp.Data.(string)
		assert.True(t, ok, "Created ID should be a string")
	}

	// --- Get Template By ID ---
	{
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+createdID, nil)
		w := executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)
		// 为了更安全、清晰地访问嵌套字段，我们可以定义一个包含具体类型的响应结构
		type TemplateResponse struct {
			Code int            `json:"code"`
			Msg  string         `json:"msg"`
			Data TemplateEntity `json:"data"`
		}
		var getOneResp TemplateResponse

		err := json.Unmarshal(w.Body.Bytes(), &getOneResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, getOneResp.Code)
		assert.Equal(t, "Test Template", getOneResp.Data.Name)
		assert.Equal(t, 100, *getOneResp.Data.Num)

	}

	// --- update Template By ID ---
	{
		updateName := "Updated Test Template"
		updateReqBody := UpdateTemplateRequest{
			Name: &updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+createdID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)

		var updateResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, updateResp.Code)

		// Verify update
		req, _ = http.NewRequest(http.MethodGet, "/v1/template/"+createdID, nil)
		w = executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)

		type TemplateResponse struct {
			Code int            `json:"code"`
			Msg  string         `json:"msg"`
			Data TemplateEntity `json:"data"`
		}
		var getOneResp TemplateResponse

		err = json.Unmarshal(w.Body.Bytes(), &getOneResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, getOneResp.Code)
		assert.Equal(t, updateName, getOneResp.Data.Name)
		assert.Equal(t, 100, *getOneResp.Data.Num) // Num should not be changed
	}

	// -- Delete Template By ID --
	{
		req, _ := http.NewRequest(http.MethodDelete, "/v1/template/"+createdID, nil)
		w := executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)

		var deleteResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, deleteResp.Code)
		// 断言影响的行数, 转换为整数
		affectedRows := int(math.Round(deleteResp.Data.(float64)))
		assert.Equal(t, 1, affectedRows)

		// Verify deletion
		req, _ = http.NewRequest(http.MethodGet, "/v1/template/"+createdID, nil)
		w = executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestTemplateHandler_HandleBatch(t *testing.T) {
	var createdIDs []string

	// --- Batch Create Templates ---
	{
		num1, num2 := 200, 300
		batchCreateReq := CreateTemplatesRequest{
			Templates: []CreateTemplateRequest{
				{Name: "Batch Template 1", Num: &num1},
				{Name: "Batch Template 2", Num: &num2},
			},
		}
		bodyBytes, _ := json.Marshal(batchCreateReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-create", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)

		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, createResp.Code)

		// Extract the IDs from the response data
		ids, ok := createResp.Data.([]interface{})
		assert.True(t, ok, "Created IDs should be a slice of strings")
		for _, id := range ids {
			createdIDs = append(createdIDs, id.(string))
		}
		assert.Len(t, createdIDs, 2)
	}

	// --- Query Template List ---
	{
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?page=1&pageSize=10", nil)
		w := executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)

		type QueryResponseData struct {
			List  []TemplateEntity `json:"list"`
			Total int64            `json:"total"`
		}
		type QueryResponse struct {
			Code int               `json:"code"`
			Msg  string            `json:"msg"`
			Data QueryResponseData `json:"data"`
		}

		var queryResp QueryResponse
		err := json.Unmarshal(w.Body.Bytes(), &queryResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, queryResp.Code)
		// Since other tests might run, we check if at least the 2 created items are there.
		assert.True(t, queryResp.Data.Total >= 2)
		assert.True(t, len(queryResp.Data.List) >= 2)

		// Verify that the retrieved IDs match the created IDs
		retrievedIDs := make(map[string]bool)
		for _, item := range queryResp.Data.List {
			retrievedIDs[item.ID] = true
		}

		for _, id := range createdIDs {
			assert.True(t, retrievedIDs[id], "Created ID %s should be in the list", id)
		}
	}

	// --- Batch Delete Templates ---
	{
		deleteReqBody := DeleteTemplatesRequest{
			IDs: createdIDs,
		}
		bodyBytes, _ := json.Marshal(deleteReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-delete", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code)

		var deleteResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, deleteResp.Code)

		// Verify deletion by trying to get one of the deleted templates
		req, _ = http.NewRequest(http.MethodGet, "/v1/template/"+createdIDs[0], nil)
		w = executeRequest(req)
		assert.Equal(t, http.StatusOK, w.Code) // The handler returns 200 even for not found, but the response body indicates the error
		var getResp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &getResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, getResp.Code)
		assert.Equal(t, "Template not found", getResp.Msg)
	}
}
