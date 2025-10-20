package role

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	v1 "go-pg-demo/api/v1"
	"go-pg-demo/internal/middlewares"
	"go-pg-demo/internal/modules/iacc/auth"
	"go-pg-demo/internal/modules/iacc/service"
	"go-pg-demo/internal/modules/iacc/user"
	"go-pg-demo/internal/utils"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	testRoleHandler *Handler
	testAuthHandler *auth.Handler
	testDB          *sqlx.DB
	testLogger      *zap.Logger
	testValidator   *pkgs.RequestValidator
	testConfig      *pkgs.Config
	testRouter      *v1.Router
)

// TestMain 设置测试环境
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	var err error
	testLogger, _ = zap.NewDevelopment()
	defer testLogger.Sync()

	testConfig, err = pkgs.NewConfig()
	if err != nil {
		testLogger.Fatal("无法加载配置", zap.Error(err))
	}

	testDB, err = pkgs.NewConnection(testConfig)
	if err != nil {
		testLogger.Fatal("无法连接到数据库", zap.Error(err))
	}
	defer testDB.Close()

	testValidator = pkgs.NewRequestValidator()
	permissionService := service.NewPermissionService(testDB, testLogger)
	permissionMiddleware := middlewares.NewPermissionMiddleware(permissionService, testLogger)
	authMiddleware := middlewares.NewAuthMiddleware(testConfig, testLogger)

	testRoleHandler = NewRoleHandler(testDB, testLogger, testValidator)
	testAuthHandler = auth.NewAuthHandler(testDB, testLogger, testValidator, testConfig)

	engine := gin.New()
	testRouter = &v1.Router{
		Engine:               engine,
		AuthHandler:          testAuthHandler,
		RoleHandler:          testRoleHandler,
		PermissionMiddleware: permissionMiddleware,
	}
	// 注册全局中间件
	testRouter.Engine.Use(gin.HandlerFunc(authMiddleware))
	// 注册路由组
	testRouter.RouterGroup = testRouter.Engine.Group("/v1")
	// 注册登录路由
	testRouter.RegisterIACCAuth()
	// 注册role的测试路由
	testRouter.RegisterIACCRole()

	exitCode := m.Run()
	os.Exit(exitCode)
}

// executeRequest 辅助函数，用于执行HTTP请求并返回响应记录器
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	testRouter.Engine.ServeHTTP(rr, req)
	return rr
}

// 创建一个用于测试的角色
func setupTestRole(t *testing.T, permissions []string) Role {
	t.Helper()
	var permsStr *string
	if permissions != nil {
		permsJSON, err := json.Marshal(permissions)
		require.NoError(t, err)
		s := string(permsJSON)
		permsStr = &s
	}

	r := Role{
		Name:        "testrole_" + uuid.NewString()[:8],
		Description: "A role for testing",
		Permissions: permsStr,
	}

	// 插入角色，然后返回角色ID
	query := `INSERT INTO iacc_role ( name, description, permissions) VALUES ($1, $2, $3) RETURNING id`
	err := testDB.QueryRowContext(context.Background(), query, r.Name, r.Description, r.Permissions).Scan(&r.ID)
	require.NoError(t, err, "创建测试角色失败")

	t.Cleanup(func() {
		_, err := testDB.Exec("DELETE FROM iacc_role WHERE id = $1", r.ID)
		assert.NoError(t, err, "清理测试角色失败")
	})

	return r
}

func TestCreateRole(t *testing.T) {
	t.Run("成功创建角色", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleCreate.Key})
		reqBody := CreateRoleRequest{
			Name:        "new_role_" + uuid.NewString()[:8],
			Description: "A new role for test",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		roleID := resp.Data.(string)
		assert.NotEmpty(t, roleID)

		// Cleanup
		_, err = testDB.Exec("DELETE FROM iacc_role WHERE id = $1", roleID)
		assert.NoError(t, err)
	})

	t.Run("失败 - 缺少角色名", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleCreate.Key})
		reqBody := CreateRoleRequest{Description: "A role without name"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Msg, "Name", "错误消息应提示角色名字段问题")
	})
}

func TestGetRole(t *testing.T) {
	t.Run("成功获取角色", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleView.Key})
		testRole := setupTestRole(t, []string{"perm:read"})
		url := fmt.Sprintf("/v1/role/%s", testRole.ID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var roleResp RoleResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &roleResp)
		require.NoError(t, err)

		assert.Equal(t, testRole.ID, roleResp.ID)
		assert.Equal(t, testRole.Name, roleResp.Name)
		assert.JSONEq(t, *testRole.Permissions, *roleResp.Permissions)
	})

	t.Run("失败 - 角色不存在", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleCreate.Key})
		nonExistentID := uuid.NewString()
		url := fmt.Sprintf("/v1/role/%s", nonExistentID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, "Role not found", resp.Msg)
	})
}

func TestUpdateRole(t *testing.T) {
	t.Run("成功更新角色", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleCreate.Key})
		testRole := setupTestRole(t, nil)
		updatedName := "updated_" + testRole.Name
		updatedDesc := "Updated description"
		reqBody := UpdateRoleRequest{
			Name:        &updatedName,
			Description: &updatedDesc,
		}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/role/%s", testRole.ID)
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var roleResp RoleResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &roleResp)
		require.NoError(t, err)
		assert.Equal(t, updatedName, roleResp.Name)
		assert.Equal(t, updatedDesc, roleResp.Description)
	})
}

func TestDeleteRole(t *testing.T) {
	t.Run("成功删除角色", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleCreate.Key})
		testRole := setupTestRole(t, nil)
		url := fmt.Sprintf("/v1/role/%s", testRole.ID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		// Verify deletion
		var count int
		err = testDB.Get(&count, "SELECT COUNT(*) FROM iacc_role WHERE id = $1", testRole.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("失败 - 角色仍被用户使用", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleCreate.Key})
		// 1. Create a user
		u := user.User{ID: uuid.NewString(), Username: "user_with_role", Password: "pw"}
		_, err := testDB.Exec("INSERT INTO iacc_user (id, username, password) VALUES ($1, $2, $3)", u.ID, u.Username, u.Password)
		require.NoError(t, err)
		defer testDB.Exec("DELETE FROM iacc_user WHERE id = $1", u.ID)

		// 2. Create a role
		r := setupTestRole(t, nil)

		// 3. Assign role to user
		_, err = testDB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3)", uuid.NewString(), u.ID, r.ID)
		require.NoError(t, err)
		defer testDB.Exec("DELETE FROM iacc_user_role WHERE role_id = $1", r.ID)

		// 4. Attempt to delete the role
		url := fmt.Sprintf("/v1/role/%s", r.ID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.Code)
		assert.Equal(t, "该角色正在被使用，无法删除", resp.Msg)
	})
}

func TestListRoles(t *testing.T) {
	t.Run("成功列出角色", func(t *testing.T) {
		// Arrange
		token := utils.SetupAccessUserToken(t, testRouter.Engine, testDB, []string{pkgs.Permissions.RoleCreate.Key})
		_ = setupTestRole(t, nil) // Ensure at least one role exists
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var listResp ListRolesResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &listResp)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, listResp.TotalCount, int64(1))
		assert.NotEmpty(t, listResp.Roles)
	})
}
