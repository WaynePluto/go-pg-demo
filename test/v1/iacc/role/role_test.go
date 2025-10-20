package role_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-pg-demo/internal/app"
	"go-pg-demo/internal/modules/iacc/role"
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

func TestCreateRole(t *testing.T) {
	t.Run("成功创建角色", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleCreate.Key})
		reqBody := role.CreateRoleRequest{
			Name:        "new_role_" + uuid.NewString()[:8],
			Description: "A new role for test",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(body))
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
		roleID := resp.Data.(string)
		assert.NotEmpty(t, roleID)

		// Cleanup
		_, err = testDB.Exec("DELETE FROM iacc_role WHERE id = $1", roleID)
		assert.NoError(t, err)
	})

	t.Run("失败 - 缺少角色名", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleCreate.Key})
		reqBody := role.CreateRoleRequest{Description: "A role without name"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(body))
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
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Msg, "Name", "错误消息应提示角色名字段问题")
	})
}

func TestGetRole(t *testing.T) {
	t.Run("成功获取角色", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleView.Key})
		testRole := testUtil.SetupTestRole([]string{"perm:read"})
		url := fmt.Sprintf("/v1/role/%s", testRole.ID)
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

		var roleResp role.RoleResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &roleResp)
		require.NoError(t, err)

		assert.Equal(t, testRole.ID, roleResp.ID)
		assert.Equal(t, testRole.Name, roleResp.Name)
		assert.JSONEq(t, *testRole.Permissions, *roleResp.Permissions)
	})

	t.Run("失败 - 角色不存在", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleView.Key})
		nonExistentID := uuid.NewString()
		url := fmt.Sprintf("/v1/role/%s", nonExistentID)
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
		assert.Equal(t, "Role not found", resp.Msg)
	})
}

func TestUpdateRole(t *testing.T) {
	t.Run("成功更新角色", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleUpdate.Key})
		testRole := testUtil.SetupTestRole(nil)
		updatedName := "updated_" + testRole.Name
		updatedDesc := "Updated description"
		reqBody := role.UpdateRoleRequest{
			Name:        &updatedName,
			Description: &updatedDesc,
		}
		body, _ := json.Marshal(reqBody)
		url := fmt.Sprintf("/v1/role/%s", testRole.ID)
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

		var roleResp role.RoleResponse
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
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleDelete.Key})
		testRole := testUtil.SetupTestRole(nil)
		url := fmt.Sprintf("/v1/role/%s", testRole.ID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
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

		// Verify deletion
		var count int
		err = testDB.Get(&count, "SELECT COUNT(*) FROM iacc_role WHERE id = $1", testRole.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("失败 - 角色仍被用户使用", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleDelete.Key})

		// 1. Create a user
		u := user.User{ID: uuid.NewString(), Username: "user_with_role", Password: "pw"}
		_, err := testDB.Exec("INSERT INTO iacc_user (id, username, password) VALUES ($1, $2, $3)", u.ID, u.Username, u.Password)
		require.NoError(t, err)
		defer testDB.Exec("DELETE FROM iacc_user WHERE id = $1", u.ID)

		// 2. Create a role
		r := testUtil.SetupTestRole(nil)

		// 3. Assign role to user
		_, err = testDB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3)", uuid.NewString(), u.ID, r.ID)
		require.NoError(t, err)
		defer testDB.Exec("DELETE FROM iacc_user_role WHERE role_id = $1", r.ID)

		// 4. Attempt to delete the role
		url := fmt.Sprintf("/v1/role/%s", r.ID)
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
		assert.Equal(t, http.StatusConflict, resp.Code)
		assert.Equal(t, "该角色正在被使用，无法删除", resp.Msg)
	})
}

func TestListRoles(t *testing.T) {
	t.Run("成功列出角色", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		token := testUtil.GetAccessUserToken([]string{pkgs.Permissions.RoleList.Key})
		testUtil.SetupTestRole(nil)
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list", nil)
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

		var listResp role.ListRolesResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &listResp)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, listResp.TotalCount, int64(1))
		assert.NotEmpty(t, listResp.Roles)
	})
}
