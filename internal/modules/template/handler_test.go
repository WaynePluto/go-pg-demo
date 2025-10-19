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
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testHandler   *Handler
	testDB        *sqlx.DB
	testLogger    *zap.Logger
	testValidator *pkgs.RequestValidator
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

	// Create validator instance
	testValidator = pkgs.NewRequestValidator()

	// Create handler instance
	testHandler = NewTemplateHandler(testDB, testLogger, testValidator)

	// Run tests
	exitCode := m.Run()

	// Exit
	os.Exit(exitCode)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	engine := gin.New()
	v1Router := v1.Router{
		Engine:          engine,
		RouterGroup:     engine.Group("/v1"),
		TemplateHandler: testHandler,
	}
	v1Router.RegisterTemplate()
	engine.ServeHTTP(rr, req)
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
			t.Errorf("Failed to clean up test template: %v", err)
		}
	})

	return template
}

func TestCreateTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		num := 100
		createReqBody := CreateTemplateRequest{
			Name: "New Test Template",
			Num:  &num,
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err)
		assert.Equal(t, 200, createResp.Code)

		createdID, ok := createResp.Data.(string)
		assert.True(t, ok)

		// Cleanup
		t.Cleanup(func() {
			_, err := testDB.Exec("DELETE FROM template WHERE id = $1", createdID)
			assert.NoError(t, err)
		})
	})

	t.Run("Invalid Input - Missing Name", func(t *testing.T) {
		// Arrange
		num := 100
		createReqBody := CreateTemplateRequest{
			Num: &num, // Name is missing
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/template", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var errResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errResp.Code)
		assert.Contains(t, errResp.Msg, "Key: 'CreateTemplateRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag")
	})
}

func TestGetTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		template := setupTestTemplate(t)

		// Act
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+template.ID, nil)
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		type TemplateResponse struct {
			Code int            `json:"code"`
			Data TemplateEntity `json:"data"`
		}
		var resp TemplateResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, template.Name, resp.Data.Name)
		assert.Equal(t, *template.Num, *resp.Data.Num)
	})

	t.Run("Not Found", func(t *testing.T) {
		// Arrange
		nonExistentID := "a-b-c-d-e"

		// Act
		req, _ := http.NewRequest(http.MethodGet, "/v1/template/"+nonExistentID, nil)
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code) // Handler returns 200 but with error code in body
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, "Template not found", resp.Msg)
	})
}

func TestUpdateTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		template := setupTestTemplate(t)
		updateName := "Updated Name"
		updateReqBody := UpdateTemplateRequest{
			Name: &updateName,
		}
		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/template/"+template.ID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var updateResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, updateResp.Code)

		// Verify update
		var updatedTemplate TemplateEntity
		err = testDB.Get(&updatedTemplate, "SELECT * FROM template WHERE id = $1", template.ID)
		assert.NoError(t, err)
		assert.Equal(t, updateName, updatedTemplate.Name)
		assert.Equal(t, *template.Num, *updatedTemplate.Num) // Num should not change
	})
}

func TestDeleteTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		template := setupTestTemplate(t)

		// Act
		req, _ := http.NewRequest(http.MethodDelete, "/v1/template/"+template.ID, nil)
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var deleteResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, deleteResp.Code)
		affectedRows := int(math.Round(deleteResp.Data.(float64)))
		assert.Equal(t, 1, affectedRows)

		// Verify deletion
		var count int
		err = testDB.Get(&count, "SELECT COUNT(*) FROM template WHERE id = $1", template.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestBatchCreateTemplates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
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

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, createResp.Code)

		ids, ok := createResp.Data.([]interface{})
		assert.True(t, ok)
		assert.Len(t, ids, 2)

		// Cleanup
		t.Cleanup(func() {
			for _, id := range ids {
				_, err := testDB.Exec("DELETE FROM template WHERE id = $1", id.(string))
				assert.NoError(t, err)
			}
		})
	})
}

func TestQueryTemplateList(t *testing.T) {
	// Arrange
	template1 := setupTestTemplate(t)
	template2 := setupTestTemplate(t)

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/v1/template/list?page=1&pageSize=10", nil)
	w := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

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
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, queryResp.Code)
	assert.True(t, queryResp.Data.Total >= 2)

	retrievedIDs := make(map[string]bool)
	for _, item := range queryResp.Data.List {
		retrievedIDs[item.ID] = true
	}
	assert.True(t, retrievedIDs[template1.ID])
	assert.True(t, retrievedIDs[template2.ID])
}

func TestBatchDeleteTemplates(t *testing.T) {
	// Arrange
	template1 := setupTestTemplate(t)
	template2 := setupTestTemplate(t)
	idsToDelete := []string{template1.ID, template2.ID}

	deleteReqBody := DeleteTemplatesRequest{
		IDs: idsToDelete,
	}
	bodyBytes, _ := json.Marshal(deleteReqBody)
	req, _ := http.NewRequest(http.MethodPost, "/v1/template/batch-delete", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := executeRequest(req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var deleteResp pkgs.Response
	err := json.Unmarshal(w.Body.Bytes(), &deleteResp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, deleteResp.Code)

	// Verify deletion
	var count int
	query, args, err := sqlx.In("SELECT COUNT(*) FROM template WHERE id IN (?)", idsToDelete)
	assert.NoError(t, err)
	query = testDB.Rebind(query)
	err = testDB.Get(&count, query, args...)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
