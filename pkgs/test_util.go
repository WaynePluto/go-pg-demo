// 给测试文件提供的工具函数
package pkgs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type role struct {
	ID          string
	Name        string
	Description *string
}

type permission struct {
	ID       string
	Name     string
	Type     string
	Metadata struct {
		Path   *string `json:"path"`
		Method *string `json:"method"`
		Code   *string `json:"code"`
	} `json:"metadata"`
}

type user struct {
	ID       string
	Username string
	Password string
	Phone    string
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

// SetupTestPermission 创建一个权限（methodPath: "GET /v1/template/list"）
func (testUtil *TestUtil) SetupTestPermission(methodPath string) permission {
	testUtil.T.Helper()
	var method, path string
	if methodPath != "" {
		fmt.Sscanf(methodPath, "%s %s", &method, &path)
	}
	p := permission{
		Name: uuid.NewString()[:8],
		Type: "api",
	}
	if path != "" {
		p.Metadata.Path = &path
	}
	if method != "" {
		p.Metadata.Method = &method
	}

	metaBytes, _ := json.Marshal(p.Metadata)
	query := `INSERT INTO iacc_permission (name, type, metadata) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	var createdAt, updatedAt time.Time
	err := testUtil.DB.QueryRow(query, p.Name, p.Type, string(metaBytes)).Scan(&p.ID, &createdAt, &updatedAt)
	require.NoError(testUtil.T, err, "创建测试权限失败")

	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec(`DELETE FROM iacc_permission WHERE id = $1`, p.ID)
		assert.NoError(testUtil.T, err, "清理测试权限失败")
	})
	return p
}

// SetupTestRole 创建一个测试角色
func (testUtil *TestUtil) SetupTestRole() role {
	testUtil.T.Helper()
	r := role{Name: "role_" + uuid.NewString()[:8]}
	query := `INSERT INTO iacc_role (name, description) VALUES ($1, $2) RETURNING id`
	err := testUtil.DB.QueryRow(query, r.Name, r.Description).Scan(&r.ID)
	require.NoError(testUtil.T, err, "创建测试角色失败")
	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec(`DELETE FROM iacc_role WHERE id = $1`, r.ID)
		assert.NoError(testUtil.T, err, "清理测试角色失败")
	})
	return r
}

// 创建一个用于测试的用户
func (testUtil *TestUtil) SetupTestUser() user {
	testUtil.T.Helper()
	password := "strongpassword"
	username := "testuser_" + uuid.NewString()[:8]
	// 使用UUID的一部分生成唯一的手机号
	phone := fmt.Sprintf("138%s", uuid.NewString()[:8])
	u := user{
		Username: username,
		Password: password,
		Phone:    phone,
	}

	query := `INSERT INTO "iacc_user" (username, password, phone) VALUES ($1, $2, $3) RETURNING id`
	err := testUtil.DB.QueryRow(query, u.Username, u.Password, u.Phone).Scan(&u.ID)
	require.NoError(testUtil.T, err, "创建测试用户失败")

	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec(`DELETE FROM "iacc_user" WHERE id = $1`, u.ID)
		assert.NoError(testUtil.T, err, "清理测试用户失败")
	})

	return u
}

// 给测试用户赋予角色
func (testUtil *TestUtil) AssignRoleToUser(userID, roleID string) {
	testUtil.T.Helper()
	_, err := testUtil.DB.Exec(`INSERT INTO iacc_user_role (user_id, role_id) VALUES ($1, $2)`, userID, roleID)
	require.NoError(testUtil.T, err, "分配角色给用户失败")
	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec(`DELETE FROM iacc_user_role WHERE user_id = $1 AND role_id = $2`, userID, roleID)
		assert.NoError(testUtil.T, err, "清理用户角色关联失败")
	})
}

// 分配权限到角色
func (testUtil *TestUtil) AssignPermissionToRole(roleID, permissionID string) {
	testUtil.T.Helper()
	_, err := testUtil.DB.Exec(`INSERT INTO iacc_role_permission (role_id, permission_id) VALUES ($1, $2)`, roleID, permissionID)
	require.NoError(testUtil.T, err, "分配权限到角色失败")
	testUtil.T.Cleanup(func() {
		_, err := testUtil.DB.Exec(`DELETE FROM iacc_role_permission WHERE role_id = $1 AND permission_id = $2`, roleID, permissionID)
		assert.NoError(testUtil.T, err, "清理角色权限关联失败")
	})
}

// 创建用户并分配权限，返回用户与token
func (testUtil *TestUtil) SetupUserWithPermissions(methodPaths []string) (user, string) {
	u := testUtil.SetupTestUser()
	r := testUtil.SetupTestRole()
	testUtil.AssignRoleToUser(u.ID, r.ID)
	for _, mp := range methodPaths {
		perm := testUtil.SetupTestPermission(mp)
		testUtil.AssignPermissionToRole(r.ID, perm.ID)
	}
	token := testUtil.GetAccessTokenByUser(u)
	return u, token
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

// 获取访问令牌, 传permission为空数组，就是无权限访问令牌, 单个permission就是 method+空格+path 形式的权限
func (testUtil *TestUtil) GetAccessUserToken(permissions []string) string {
	if len(permissions) == 0 {
		return testUtil.GetAccessTokenByUser(testUtil.SetupTestUser())
	}
	_, token := testUtil.SetupUserWithPermissions(permissions)
	return token
}

// 创建一个没有任何权限的用户 token
func (testUtil *TestUtil) GetNoPermissionUserToken() string {
	return testUtil.GetAccessTokenByUser(testUtil.SetupTestUser())
}
