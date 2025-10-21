// 给测试文件提供的工具函数
package pkgs

import (
	"bytes"
	"context"
	"encoding/json"
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

type TestUtil struct {
	Engine *gin.Engine
	DB     *sqlx.DB
	T      *testing.T
}

// 创建一个用于测试的角色
func (testUtil *TestUtil) SetupTestRole(permissions []string) role {
	testUtil.T.Helper()
	var permsStr *string
	if permissions != nil {
		permsJSON, err := json.Marshal(permissions)
		require.NoError(testUtil.T, err)
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
	err := testUtil.DB.QueryRowContext(context.Background(), query, r.Name, r.Description, r.Permissions).Scan(&r.ID)
	require.NoError(testUtil.T, err, "创建测试角色失败")

	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec("DELETE FROM iacc_role WHERE id = $1", r.ID)
		assert.NoError(testUtil.T, err, "清理测试角色失败")
	})

	return r
}

// 创建一个用于测试的用户
func (testUtil *TestUtil) SetupTestUser() user {
	testUtil.T.Helper()
	password := "strongpassword"
	u := user{
		Username: "testuser_" + uuid.NewString()[:8],
		Password: password,
	}

	query := `INSERT INTO iacc_user (username, password) VALUES ($1, $2) RETURNING id`
	err := testUtil.DB.QueryRow(query, u.Username, u.Password).Scan(&u.ID)
	require.NoError(testUtil.T, err, "创建测试用户失败")

	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec("DELETE FROM iacc_user WHERE id = $1", u.ID)
		assert.NoError(testUtil.T, err, "清理测试用户失败")
	})

	return u
}

// 给测试用户赋予角色
func (testUtil *TestUtil) AssignRoleToUser(userID, roleID string) {
	testUtil.T.Helper()
	assocID := uuid.NewString()
	_, err := testUtil.DB.Exec("INSERT INTO iacc_user_role (id, user_id, role_id) VALUES ($1, $2, $3)", assocID, userID, roleID)
	require.NoError(testUtil.T, err, "给用户赋予角色失败")

	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec("DELETE FROM iacc_user_role WHERE id = $1", assocID)
		assert.NoError(testUtil.T, err, "清理用户角色关联失败")
	})
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (testUtil *TestUtil) GetAccessTokenByUser(testUser user) string {
	loginReq := loginRequest{Username: testUser.Username, Password: testUser.Password}
	loginBody, _ := json.Marshal(loginReq)
	loginReqHttp, err := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	require.NoError(testUtil.T, err)
	loginReqHttp.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	testUtil.Engine.ServeHTTP(rr, loginReqHttp)
	var loginResp Response
	err = json.Unmarshal(rr.Body.Bytes(), &loginResp)
	require.NoError(testUtil.T, err)
	var loginData LoginResponse
	dataBytes, err := json.Marshal(loginResp.Data)
	require.NoError(testUtil.T, err)
	err = json.Unmarshal(dataBytes, &loginData)
	require.NoError(testUtil.T, err)
	return loginData.AccessToken
}

// 获取访问令牌, 传permission为空数组，就是无权限访问令牌
func (testUtil *TestUtil) GetAccessUserToken(permissions []string) string {
	roleId := testUtil.SetupTestRole(permissions).ID
	testUser := testUtil.SetupTestUser()
	testUtil.AssignRoleToUser(testUser.ID, roleId)
	return testUtil.GetAccessTokenByUser(testUser)
}
