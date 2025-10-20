// 给测试文件提供的工具函数
package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"go-pg-demo/pkgs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type role struct {
	ID          string
	Name        string
	Description string
	Permissions *string
}

type user struct {
	ID       string
	Username string
	Password string
}

type loginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// 创建一个用于测试的角色
func SetupTestRole(t *testing.T, testDB *sqlx.DB, permissions []string) role {
	t.Helper()
	var permsStr *string
	if permissions != nil {
		permsJSON, err := json.Marshal(permissions)
		require.NoError(t, err)
		s := string(permsJSON)
		permsStr = &s
	}

	r := role{
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

// 创建一个用于测试的用户
func SetupTestUser(t *testing.T, testDB *sqlx.DB) user {
	t.Helper()
	password := "strongpassword"
	u := user{
		Username: "testuser_" + uuid.NewString()[:8],
		Password: password,
	}

	query := `INSERT INTO iacc_user (username, password) VALUES ($1, $2) RETURNING id`
	err := testDB.QueryRow(query, u.Username, u.Password).Scan(&u.ID)
	require.NoError(t, err, "创建测试用户失败")

	t.Cleanup(func() {
		_, err := testDB.Exec("DELETE FROM iacc_user WHERE id = $1", u.ID)
		assert.NoError(t, err, "清理测试用户失败")
	})

	return u
}

// 给测试用户赋予角色
func AssignRoleToUser(t *testing.T, testDB *sqlx.DB, userID, roleID string) {
	t.Helper()
	assocID := uuid.NewString()
	_, err := testDB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3)", assocID, userID, roleID)
	require.NoError(t, err, "给用户赋予角色失败")

	t.Cleanup(func() {
		_, err := testDB.Exec("DELETE FROM iacc_user_role WHERE id = $1", assocID)
		assert.NoError(t, err, "清理用户角色关联失败")
	})
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// 获取访问令牌
func GetAccessToken(t *testing.T, engine *gin.Engine, testUser user) string {
	loginReq := loginRequest{Username: testUser.Username, Password: testUser.Password}
	loginBody, _ := json.Marshal(loginReq)
	loginReqHttp, err := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	require.NoError(t, err)
	loginReqHttp.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, loginReqHttp)
	var loginResp pkgs.Response
	err = json.Unmarshal(rr.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	var loginData LoginResponse
	dataBytes, err := json.Marshal(loginResp.Data)
	require.NoError(t, err)
	err = json.Unmarshal(dataBytes, &loginData)
	require.NoError(t, err)
	return loginData.AccessToken
}

func SetupAccessUserToken(t *testing.T, engine *gin.Engine, testDB *sqlx.DB, permissions []string) string {
	roleId := SetupTestRole(t, testDB, []string{pkgs.Permissions.RoleCreate.Key}).ID
	testUser := SetupTestUser(t, testDB)
	AssignRoleToUser(t, testDB, testUser.ID, roleId)
	return GetAccessToken(t, engine, testUser)
}
