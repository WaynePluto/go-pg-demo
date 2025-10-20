package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	v1 "go-pg-demo/api/v1"
	"go-pg-demo/internal/middlewares"
	"go-pg-demo/internal/modules/iacc/auth"
	"go-pg-demo/internal/modules/iacc/role"
	"go-pg-demo/internal/modules/iacc/service"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	testUserHandler *Handler
	testRoleHandler *role.Handler
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

	testAuthHandler = auth.NewAuthHandler(testDB, testLogger, testValidator, testConfig)
	testUserHandler = NewUserHandler(testDB, testLogger, testValidator, permissionService)
	testRoleHandler = role.NewRoleHandler(testDB, testLogger, testValidator)

	engine := gin.New()
	testRouter = &v1.Router{
		Engine:               engine,
		UserHandler:          testUserHandler,
		RoleHandler:          testRoleHandler,
		AuthHandler:          testAuthHandler,
		PermissionMiddleware: permissionMiddleware,
	}
	// 注册全局中间件
	testRouter.Engine.Use(gin.HandlerFunc(authMiddleware))
	// 注册路由组
	testRouter.RouterGroup = testRouter.Engine.Group("/v1")
	// 注册IACC路由
	testRouter.RegisterIACCAuth()
	testRouter.RegisterIACCUser()
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

// setupTestUser 创建一个用于测试的用户
func setupTestUser(t *testing.T) User {
	t.Helper()
	u := User{
		ID:       uuid.NewString(),
		Username: "testuser_" + uuid.NewString()[:8],
		Phone:    "1234567890",
		Password: "password",
	}

	query := `INSERT INTO iacc_user (id, username, phone, password) VALUES ($1, $2, $3, $4)`
	_, err := testDB.Exec(query, u.ID, u.Username, u.Phone, u.Password)
	require.NoError(t, err, "创建测试用户失败")

	t.Cleanup(func() {
		_, err := testDB.Exec("DELETE FROM iacc_user WHERE id = $1", u.ID)
		assert.NoError(t, err, "清理测试用户失败")
	})

	return u
}

// setupTestRole 创建一个用于测试的角色
func setupTestRole(t *testing.T, permissions []string) role.Role {
	t.Helper()
	permsJSON, err := json.Marshal(permissions)
	require.NoError(t, err)
	permsStr := string(permsJSON)

	r := role.Role{
		ID:          uuid.NewString(),
		Name:        "testrole_" + uuid.NewString()[:8],
		Description: "A role for testing",
		Permissions: &permsStr,
	}

	query := `INSERT INTO iacc_role (id, name, description, permissions) VALUES ($1, $2, $3, $4)`
	_, err = testDB.Exec(query, r.ID, r.Name, r.Description, r.Permissions)
	require.NoError(t, err, "创建测试角色失败")

	t.Cleanup(func() {
		_, err := testDB.Exec("DELETE FROM iacc_role WHERE id = $1", r.ID)
		assert.NoError(t, err, "清理测试角色失败")
	})

	return r
}

// setupUserAndLogin 创建一个用户，为其分配一个具有指定权限的角色，然后登录以获取token
func setupUserAndLogin(t *testing.T, permissions []string) (User, string) {
	t.Helper()

	// 1. 创建一个具有指定权限的角色
	testRole := setupTestRole(t, permissions)

	// 2. 创建一个用户
	testUser := setupTestUser(t)

	// 3. 将角色分配给用户
	_, err := testDB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3)",
		uuid.NewString(), testUser.ID, testRole.ID)
	require.NoError(t, err, "为测试用户分配角色失败")

	// 4. 用户登录以获取token
	loginReqBody := auth.LoginRequest{
		Username: testUser.Username,
		Password: "password", // "password" 是 setupTestUser 中设置的默认密码
	}
	body, _ := json.Marshal(loginReqBody)
	req, _ := http.NewRequest(http.MethodPost, "/v1/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := executeRequest(req)
	require.Equal(t, http.StatusOK, w.Code, "登录请求失败")

	var resp pkgs.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code, "登录响应码不为200")

	var loginResp auth.LoginResponse
	dataBytes, _ := json.Marshal(resp.Data)
	err = json.Unmarshal(dataBytes, &loginResp)
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.AccessToken, "未能从登录响应中获取token")

	return testUser, loginResp.AccessToken
}

func TestCreateUser(t *testing.T) {
	// 为了测试CreateUser，我们需要一个拥有UserCreate权限的token
	_, token := setupUserAndLogin(t, []string{pkgs.Permissions.UserCreate.Key})

	t.Run("成功创建用户", func(t *testing.T) {
		// Arrange
		reqBody := CreateUserRequest{
			Username: "newuser_" + uuid.NewString()[:8],
			Phone:    "9876543210",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(body))
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
		assert.NotEmpty(t, resp.Data, "创建成功后应返回用户ID")

		// Cleanup
		userID := resp.Data.(string)
		_, err = testDB.Exec("DELETE FROM iacc_user WHERE id = $1", userID)
		assert.NoError(t, err)
	})

	t.Run("失败 - 缺少用户名", func(t *testing.T) {
		// Arrange
		reqBody := CreateUserRequest{
			Phone:    "9876543210",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(body))
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
		assert.Contains(t, resp.Msg, "Username", "错误消息应提示用户名字段问题")
	})
}

func TestGetUser(t *testing.T) {
	_, token := setupUserAndLogin(t, []string{pkgs.Permissions.UserView.Key})
	t.Run("成功获取用户", func(t *testing.T) {
		// Arrange
		testUser := setupTestUser(t)
		url := fmt.Sprintf("/v1/user/%s", testUser.ID)
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

		var userResp UserResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &userResp)
		require.NoError(t, err)

		assert.Equal(t, testUser.ID, userResp.ID)
		assert.Equal(t, testUser.Username, userResp.Username)
	})

	t.Run("失败 - 用户不存在", func(t *testing.T) {
		// Arrange
		nonExistentID := uuid.NewString()
		url := fmt.Sprintf("/v1/user/%s", nonExistentID)
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
		assert.Equal(t, "User not found", resp.Msg)
	})
}

func TestUpdateUser(t *testing.T) {
	_, token := setupUserAndLogin(t, []string{pkgs.Permissions.UserUpdate.Key})
	t.Run("成功更新用户", func(t *testing.T) {
		// Arrange
		testUser := setupTestUser(t)
		updatedUsername := "updated_" + testUser.Username
		reqBody := UpdateUserRequest{
			Username: &updatedUsername,
		}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/user/%s", testUser.ID)
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

		var userResp UserResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &userResp)
		require.NoError(t, err)
		assert.Equal(t, updatedUsername, userResp.Username)
	})
}

func TestListUsers(t *testing.T) {
	_, token := setupUserAndLogin(t, []string{pkgs.Permissions.UserList.Key})
	t.Run("成功列出用户", func(t *testing.T) {
		// Arrange
		_ = setupTestUser(t) // 创建至少一个用户以确保列表不为空
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var listResp ListUsersResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &listResp)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, listResp.TotalCount, int64(1))
		assert.NotEmpty(t, listResp.Users)
	})
}

func TestAssignAndRemoveRole(t *testing.T) {
	_, token := setupUserAndLogin(t, []string{pkgs.Permissions.UserAssignRole.Key})
	// Arrange
	testUser := setupTestUser(t)
	testRole := setupTestRole(t, []string{})

	t.Run("成功为用户分配角色", func(t *testing.T) {
		// Arrange
		reqBody := AssignRoleRequest{RoleID: testRole.ID}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/user/%s/role", testUser.ID)
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
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
		assert.NotEmpty(t, resp.Data, "应返回用户角色关联ID")

		// Verify in DB
		var count int
		err = testDB.Get(&count, "SELECT COUNT(*) FROM iacc_user_role WHERE user_id = $1 AND role_id = $2", testUser.ID, testRole.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "数据库中应存在关联记录")
	})

	t.Run("成功移除用户角色", func(t *testing.T) {
		// Arrange
		// 确保角色已分配
		_, err := testDB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", uuid.NewString(), testUser.ID, testRole.ID)
		require.NoError(t, err)

		url := fmt.Sprintf("/v1/user/%s/role/%s", testUser.ID, testRole.ID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		// Verify in DB
		var count int
		err = testDB.Get(&count, "SELECT COUNT(*) FROM iacc_user_role WHERE user_id = $1 AND role_id = $2", testUser.ID, testRole.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "数据库中不应再有关联记录")
	})
}

func TestGetUserPermissions(t *testing.T) {
	_, token := setupUserAndLogin(t, []string{pkgs.Permissions.UserViewPermissions.Key})
	t.Run("成功获取用户权限", func(t *testing.T) {
		// Arrange
		user := setupTestUser(t)
		role1 := setupTestRole(t, []string{"perm1", "perm2"})
		role2 := setupTestRole(t, []string{"perm2", "perm3"})

		// Assign roles to user
		_, err := testDB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3), ($4, $5, $6)",
			uuid.NewString(), user.ID, role1.ID, uuid.NewString(), user.ID, role2.ID)
		require.NoError(t, err)

		url := fmt.Sprintf("/v1/user/%s/permissions", user.ID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var perms []string
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &perms)
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"perm1", "perm2", "perm3"}, perms, "权限列表应合并去重")
	})
}
