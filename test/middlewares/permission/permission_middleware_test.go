package permission_middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"go-pg-demo/internal/app"
	"go-pg-demo/pkgs"
)

// 复用应用实例
var (
	testDB     *sqlx.DB
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	a, _, err := app.InitializeApp()
	if err != nil {
		os.Exit(1)
	}
	testDB = a.DB
	testRouter = a.Server
	code := m.Run()
	os.Exit(code)
}

// 辅助函数：解析标准响应
func parseResponse(t *testing.T, w *httptest.ResponseRecorder) pkgs.Response {
	t.Helper()
	var resp pkgs.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "解析响应体不应出错")
	return resp
}

// 场景1：未授权 - 访问受保护接口缺少token
func TestPermissionMiddleware_Unauthorized_NoToken(t *testing.T) {
	// Arrange: 直接请求需要鉴权的接口 /v1/permission/list
	req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list", nil)
	w := httptest.NewRecorder()
	// Act
	testRouter.ServeHTTP(w, req)
	// Assert
	assert.Equal(t, http.StatusOK, w.Code, "HTTP状态码统一为200")
	resp := parseResponse(t, w)
	assert.Equal(t, http.StatusUnauthorized, resp.Code, "业务码应为401 未授权")
	assert.Contains(t, resp.Msg, "请求头缺少 Authorization", "错误信息应提示缺少Authorization")
}

// 场景2：无权限 - 有token但没有接口权限
func TestPermissionMiddleware_Forbidden_NoPermission(t *testing.T) {
	// Arrange:
	// 1. 先插入一条 GET /v1/permission/list 的权限记录（表示此接口需要鉴权）
	// 2. 创建一个无任何权限的用户 token；因为用户没有分配该接口权限，应返回403业务码
	tu := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	_ = tu.SetupTestPermission("GET /v1/permission/list")
	token := tu.GetNoPermissionUserToken()
	req, _ := http.NewRequest(http.MethodGet, "/v1/permission/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	// Act
	testRouter.ServeHTTP(w, req)
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, http.StatusForbidden, resp.Code, "业务码应为403 无权限")
	assert.Equal(t, "无接口访问权限", resp.Msg, "错误消息应为 无接口访问权限")
}

// 场景3：有权限 - 访问 GET /v1/template/list (此接口在权限中间件里被白名单直接放行, 这里测试有权限时也能正常访问)
func TestPermissionMiddleware_Success_WithPermission_TemplateList(t *testing.T) {
	// Arrange: 分配 GET /v1/template/list 权限（虽然白名单放行，但验证权限数据链路）
	tu := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	_, token := tu.SetupUserWithPermissions([]string{"GET /v1/template/list"})
	req, _ := http.NewRequest(http.MethodGet, "/v1/template/list", nil)
	req.Header.Set("Authorization", "Bearer "+token) // 白名单情况下可不加，但这里加以保证用户上下文
	w := httptest.NewRecorder()
	// Act
	testRouter.ServeHTTP(w, req)
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, http.StatusOK, resp.Code, "业务码应为200 成功")
	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok, "返回数据应为对象")
	assert.Contains(t, data, "list", "返回数据应包含list字段")
}

// 场景4：路径通配 :id 匹配 - 创建 GET /v1/role/:id 权限后访问具体ID
func TestPermissionMiddleware_Success_PathParamMatch(t *testing.T) {
	tu := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	_, token := tu.SetupUserWithPermissions([]string{"GET /v1/role/:id"})
	// 访问一个具体ID（使用有效的UUID格式，但确保不存在）
	req, _ := http.NewRequest(http.MethodGet, "/v1/role/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	resp := parseResponse(t, w)
	// 如果该ID不存在，handler 返回404（业务码），说明中间件已放行
	assert.Equal(t, http.StatusNotFound, resp.Code, "权限应放行到处理器，未找到资源应返回404")
	assert.Equal(t, "角色不存在", resp.Msg, "错误信息应为 角色不存在")
}

// 场景5：Method不匹配 - 仅创建GET权限，用POST访问应403
func TestPermissionMiddleware_Forbidden_MethodMismatch(t *testing.T) {
	// 场景：同一路径存在 GET 与 POST 两条权限记录，用户仅拥有 GET；对 POST 发起请求应403
	tu := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
	// 用户拥有 GET /v1/permission/list
	_, token := tu.SetupUserWithPermissions([]string{"GET /v1/permission/list"})
	// 仅创建（不分配）POST 权限记录，使得中间件识别该接口需要鉴权
	_ = tu.SetupTestPermission("POST /v1/permission/list")
	req, _ := http.NewRequest(http.MethodPost, "/v1/permission/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	resp := parseResponse(t, w)
	assert.Equal(t, http.StatusForbidden, resp.Code, "无POST权限应返回403")
	assert.Equal(t, "无接口访问权限", resp.Msg)
}

// 场景6：白名单接口 - /v1/template/list 无token也放行
func TestPermissionMiddleware_Whitelist_TemplateList_NoToken(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/v1/template/list", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	resp := parseResponse(t, w)
	assert.Equal(t, http.StatusOK, resp.Code, "白名单接口应直接成功")
	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, data, "list")
}
