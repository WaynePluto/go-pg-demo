package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-pg-demo/internal/app"
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
	testDB     *sqlx.DB
	testLogger *zap.Logger
	testRouter *gin.Engine
	testUtil   *utils.TestUtil
)

// TestMain 设置测试环境
func TestMain(m *testing.M) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建应用实例
	testApp, _, err := app.InitializeApp()
	if err != nil {
		testLogger.Fatal("创建应用实例失败", zap.Error(err))
	}

	testDB = testApp.DB
	testLogger = testApp.Logger
	testRouter = testApp.Server

	testUtil = &utils.TestUtil{
		DB:     testDB,
		Engine: testRouter,
	}

	// 运行测试
	exitCode := m.Run()

	// 退出
	os.Exit(exitCode)
}

func TestCreateUser(t *testing.T) {
	t.Run("成功创建用户", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserCreate.Key})
		reqBody := user.CreateUserRequest{
			Username: "newuser_" + uuid.NewString()[:8],
			Phone:    "13900139000",
			Password: "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code, "响应状态码应为200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err, "解析响应JSON失败")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Msg)
		assert.NotEmpty(t, resp.Data, "返回数据不应为空")

		// Cleanup
		userID := resp.Data.(string)
		_, err = testDB.Exec("DELETE FROM iacc_user WHERE id = $1", userID)
		assert.NoError(t, err)
	})

	t.Run("失败 - 缺少必填字段", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserCreate.Key})
		reqBody := user.CreateUserRequest{
			Username: "",
			Phone:    "13900139000",
			Password: "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code, "业务码应为400")
		assert.Contains(t, resp.Msg, "Username", "错误消息应提示用户名字段问题")
	})

	t.Run("失败 - 无权限", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{}) // 无权限
		reqBody := user.CreateUserRequest{
			Username: "newuser_" + uuid.NewString()[:8],
			Phone:    "13900139000",
			Password: "newpassword123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.Code, "业务码应为403 Forbidden")
	})
}

func TestGetUser(t *testing.T) {
	t.Run("成功获取用户", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserView.Key})
		testUser := testUtil.SetupTestUser()
		url := fmt.Sprintf("/v1/user/%s", testUser.ID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var userResp user.UserResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &userResp)
		require.NoError(t, err)

		assert.Equal(t, testUser.ID, userResp.ID)
		assert.Equal(t, testUser.Username, userResp.Username)
	})

	t.Run("失败 - 用户不存在", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserView.Key})
		nonExistentID := uuid.NewString()
		url := fmt.Sprintf("/v1/user/%s", nonExistentID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, "User not found", resp.Msg)
	})

	t.Run("失败 - 无效用户ID", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserView.Key})
		invalidID := "invalid-id"
		url := fmt.Sprintf("/v1/user/%s", invalidID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "Invalid user ID", resp.Msg)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("成功更新用户", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserUpdate.Key})
		testUser := testUtil.SetupTestUser()
		updatedUsername := "updated_" + testUser.Username
		updatedPhone := "13900139001"
		reqBody := user.UpdateUserRequest{
			Username: &updatedUsername,
			Phone:    &updatedPhone,
		}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/user/%s", testUser.ID)
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var userResp user.UserResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &userResp)
		require.NoError(t, err)
		assert.Equal(t, updatedUsername, userResp.Username)
		assert.Equal(t, updatedPhone, userResp.Phone)
	})

	t.Run("失败 - 用户不存在", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserUpdate.Key})
		nonExistentID := uuid.NewString()
		updatedUsername := "updated_user"
		updatedPhone := "13900139001"
		reqBody := user.UpdateUserRequest{
			Username: &updatedUsername,
			Phone:    &updatedPhone,
		}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/user/%s", nonExistentID)
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, "User not found", resp.Msg)
	})
}

func TestListUsers(t *testing.T) {
	t.Run("成功列出用户", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserList.Key})
		testUtil.SetupTestUser()
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/list", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var listResp user.ListUsersResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &listResp)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, listResp.TotalCount, int64(1))
		assert.NotEmpty(t, listResp.Users)
	})

	t.Run("成功根据用户名筛选", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserList.Key})
		testUser := testUtil.SetupTestUser()
		url := fmt.Sprintf("/v1/user/list?username=%s", testUser.Username)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var listResp user.ListUsersResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &listResp)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, listResp.TotalCount, int64(1))
		assert.NotEmpty(t, listResp.Users)
	})
}

func TestAssignRole(t *testing.T) {
	t.Run("成功分配角色", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserAssignRole.Key})
		testUser := testUtil.SetupTestUser()

		// 创建测试角色
		var permsStr *string
		permsJSON, _ := json.Marshal([]string{"perm:read"})
		s := string(permsJSON)
		permsStr = &s

		r := struct {
			ID          string
			Name        string
			Description string
			Permissions *string
		}{
			Name:        "testrole_" + uuid.NewString()[:8],
			Description: "A role for testing",
			Permissions: permsStr,
		}

		query := `INSERT INTO iacc_role (name, description, permissions) VALUES ($1, $2, $3) RETURNING id`
		err := testDB.QueryRowContext(context.Background(), query, r.Name, r.Description, r.Permissions).Scan(&r.ID)
		require.NoError(t, err, "创建测试角色失败")

		t.Cleanup(func() {
			_, err := testDB.Exec("DELETE FROM iacc_role WHERE id = $1", r.ID)
			assert.NoError(t, err, "清理测试角色失败")
		})

		// 准备请求
		reqBody := user.AssignRoleRequest{
			RoleID: r.ID,
		}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/user/%s/role", testUser.ID)
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.NotEmpty(t, resp.Data, "返回数据不应为空")

		// 清理用户角色关联
		assocID := resp.Data.(string)
		_, err = testDB.Exec("DELETE FROM iacc_user_role WHERE id = $1", assocID)
		assert.NoError(t, err)
	})

	t.Run("失败 - 用户不存在", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserAssignRole.Key})
		nonExistentUserID := uuid.NewString()

		// 创建测试角色
		var permsStr *string
		permsJSON, _ := json.Marshal([]string{"perm:read"})
		s := string(permsJSON)
		permsStr = &s

		r := struct {
			ID          string
			Name        string
			Description string
			Permissions *string
		}{
			Name:        "testrole_" + uuid.NewString()[:8],
			Description: "A role for testing",
			Permissions: permsStr,
		}

		query := `INSERT INTO iacc_role (name, description, permissions) VALUES ($1, $2, $3) RETURNING id`
		err := testDB.QueryRowContext(context.Background(), query, r.Name, r.Description, r.Permissions).Scan(&r.ID)
		require.NoError(t, err, "创建测试角色失败")

		t.Cleanup(func() {
			_, err := testDB.Exec("DELETE FROM iacc_role WHERE id = $1", r.ID)
			assert.NoError(t, err, "清理测试角色失败")
		})

		// 准备请求
		reqBody := user.AssignRoleRequest{
			RoleID: r.ID,
		}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/user/%s/role", nonExistentUserID)
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, "User not found", resp.Msg)
	})
}

func TestRemoveRole(t *testing.T) {
	t.Run("成功移除角色", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserDelete.Key})
		testUser := testUtil.SetupTestUser()

		// 创建测试角色
		var permsStr *string
		permsJSON, _ := json.Marshal([]string{"perm:read"})
		s := string(permsJSON)
		permsStr = &s

		r := struct {
			ID          string
			Name        string
			Description string
			Permissions *string
		}{
			Name:        "testrole_" + uuid.NewString()[:8],
			Description: "A role for testing",
			Permissions: permsStr,
		}

		query := `INSERT INTO iacc_role (name, description, permissions) VALUES ($1, $2, $3) RETURNING id`
		err := testDB.QueryRowContext(context.Background(), query, r.Name, r.Description, r.Permissions).Scan(&r.ID)
		require.NoError(t, err, "创建测试角色失败")

		t.Cleanup(func() {
			_, err := testDB.Exec("DELETE FROM iacc_role WHERE id = $1", r.ID)
			assert.NoError(t, err, "清理测试角色失败")
		})

		// 分配角色给用户
		assocID := uuid.NewString()
		_, err = testDB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3)", assocID, testUser.ID, r.ID)
		require.NoError(t, err, "给用户分配角色失败")

		// 准备请求
		url := fmt.Sprintf("/v1/user/%s/role/%s", testUser.ID, r.ID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "Role removed successfully", resp.Data)
	})

	t.Run("失败 - 用户角色关联不存在", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.UserAssignRole.Key})
		testUser := testUtil.SetupTestUser()

		// 创建测试角色
		var permsStr *string
		permsJSON, _ := json.Marshal([]string{"perm:read"})
		s := string(permsJSON)
		permsStr = &s

		r := struct {
			ID          string
			Name        string
			Description string
			Permissions *string
		}{
			Name:        "testrole_" + uuid.NewString()[:8],
			Description: "A role for testing",
			Permissions: permsStr,
		}

		query := `INSERT INTO iacc_role (name, description, permissions) VALUES ($1, $2, $3) RETURNING id`
		err := testDB.QueryRowContext(context.Background(), query, r.Name, r.Description, r.Permissions).Scan(&r.ID)
		require.NoError(t, err, "创建测试角色失败")

		t.Cleanup(func() {
			_, err := testDB.Exec("DELETE FROM iacc_role WHERE id = $1", r.ID)
			assert.NoError(t, err, "清理测试角色失败")
		})

		// 确保没有分配角色给用户
		_, err = testDB.Exec("DELETE FROM iacc_user_role WHERE user_id = $1 AND role_id = $2", testUser.ID, r.ID)
		require.NoError(t, err)

		// 准备请求
		url := fmt.Sprintf("/v1/user/%s/role/%s", testUser.ID, r.ID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, "User role not found", resp.Msg)
	})
}
