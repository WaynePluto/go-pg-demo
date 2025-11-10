package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"go-pg-demo/internal/app"
	"go-pg-demo/internal/modules/iacc/auth"
	"go-pg-demo/pkgs"
)

var (
	testDB     *sqlx.DB
	testLogger *zap.Logger
	testRouter *gin.Engine
)

// TestMain 初始化一次应用，复用数据库和路由
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	a, _, err := app.InitializeApp()
	if err != nil {
		os.Exit(1)
	}
	testDB = a.DB
	testLogger = a.Logger
	testRouter = a.Server
	code := m.Run()
	os.Exit(code)
}

// --- 登录相关测试 ---
func TestAuthLogin(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// Arrange: 创建测试用户
		util := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		u := util.SetupTestUser()
		reqBody := auth.LoginReq{Username: u.Username, Password: u.Password}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, 200, resp.Code, "响应业务码应该是200")
		dataMap, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "响应data应该是对象")
		assert.NotEmpty(t, dataMap["access_token"], "应该返回access_token")
		assert.NotEmpty(t, dataMap["refresh_token"], "应该返回refresh_token")
		assert.Greater(t, int(dataMap["expires_in"].(float64)), 0, "expires_in应大于0")
	})

	t.Run("缺少用户名", func(t *testing.T) {
		reqBody := auth.LoginReq{Password: "123456"}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusBadRequest, resp.Code, "应返回400业务码")
		assert.Contains(t, resp.Msg, "用户名", "错误信息应包含'用户名'")
	})

	t.Run("密码错误", func(t *testing.T) {
		// 创建正确用户但用错误密码
		util := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		u := util.SetupTestUser()
		reqBody := auth.LoginReq{Username: u.Username, Password: "wrong"}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Code, "应返回401业务码")
		assert.Equal(t, "用户名或密码错误", resp.Msg, "错误信息应为'用户名或密码错误'")
	})

	t.Run("用户不存在", func(t *testing.T) {
		reqBody := auth.LoginReq{Username: "not_exist_user", Password: "xxx"}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "用户名或密码错误", resp.Msg)
	})
}

// --- 刷新令牌测试 ---
func TestAuthRefreshToken(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		util := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		u := util.SetupTestUser()
		// 先登录获取 refresh_token
		loginBody := auth.LoginReq{Username: u.Username, Password: u.Password}
		loginBodyBytes, _ := json.Marshal(loginBody)
		loginReq, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBodyBytes))
		loginReq.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		testRouter.ServeHTTP(w1, loginReq)
		var loginResp pkgs.Response
		_ = json.Unmarshal(w1.Body.Bytes(), &loginResp)
		loginData := loginResp.Data.(map[string]interface{})
		refreshToken := loginData["refresh_token"].(string)

		// 使用 refresh_token 请求刷新
		refreshReq := auth.RefreshTokenReq{RefreshToken: refreshToken}
		bodyBytes, _ := json.Marshal(refreshReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh-token", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 200, resp.Code)
		dataMap := resp.Data.(map[string]interface{})
		assert.NotEmpty(t, dataMap["access_token"], "应返回新的access_token")
		assert.NotEmpty(t, dataMap["refresh_token"], "应返回新的refresh_token")
	})

	t.Run("无效token - 签名错误", func(t *testing.T) {
		// 伪造 refresh token 使用不同 secret
		fakeClaims := jwt.MapClaims{"user_id": "some-user", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, fakeClaims)
		badToken, _ := token.SignedString([]byte("wrong-secret"))
		refreshReq := auth.RefreshTokenReq{RefreshToken: badToken}
		bodyBytes, _ := json.Marshal(refreshReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh-token", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "刷新令牌无效", resp.Msg)
	})

	t.Run("无效token - 缺少user_id", func(t *testing.T) {
		fakeClaims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, fakeClaims)
		// 使用正确 secret 但缺少 user_id
		badToken, _ := token.SignedString([]byte("my-secret-key"))
		refreshReq := auth.RefreshTokenReq{RefreshToken: badToken}
		bodyBytes, _ := json.Marshal(refreshReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh-token", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "刷新令牌无效", resp.Msg)
	})
}

// --- 用户详情 ---
func TestAuthUserDetail(t *testing.T) {
	t.Run("成功 - 基本信息", func(t *testing.T) {
		util := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		u := util.SetupTestUser()
		token := util.GetAccessTokenByUser(u)
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/user-detail", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 200, resp.Code)
		dataMap := resp.Data.(map[string]interface{})
		assert.Equal(t, u.ID, dataMap["id"], "返回的用户ID应与token对应")
		assert.Equal(t, u.Username, dataMap["username"], "返回的用户名应与token对应")
		// roles/permissions 可能为空数组或未定义，做安全转换
		rolesRaw, ok := dataMap["roles"]
		if ok && rolesRaw != nil {
			if rolesSlice, ok2 := rolesRaw.([]interface{}); ok2 {
				assert.Empty(t, rolesSlice, "当前用户未分配角色应返回空数组")
			} else {
				assert.Fail(t, "roles 字段类型不正确")
			}
		} else {
			assert.True(t, ok, "roles 字段应该存在")
		}
		permsRaw, ok := dataMap["permissions"]
		if ok && permsRaw != nil {
			if permsSlice, ok2 := permsRaw.([]interface{}); ok2 {
				assert.Empty(t, permsSlice, "当前用户无权限应返回空数组")
			} else {
				assert.Fail(t, "permissions 字段类型不正确")
			}
		} else {
			assert.True(t, ok, "permissions 字段应该存在")
		}
	})

	t.Run("未授权 - 缺少Authorization", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/user-detail", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "请求头缺少 Authorization 字段", resp.Msg)
	})

	t.Run("未授权 - 无效Bearer前缀", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/user-detail", nil)
		req.Header.Set("Authorization", "Token xxxxx")
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "Authorization 字段必须以 'Bearer ' 开头", resp.Msg)
	})

	t.Run("未授权 - token签名错误", func(t *testing.T) {
		// 伪造错误签名 token
		claims := jwt.MapClaims{"user_id": "abc", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		bad, _ := tk.SignedString([]byte("other-secret"))
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/user-detail", nil)
		req.Header.Set("Authorization", "Bearer "+bad)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "无效的令牌", resp.Msg)
	})

	t.Run("用户不存在", func(t *testing.T) {
		// 构造合法签名但数据库中无此用户
		claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000000", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
		tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// secret 与配置一致
		good, _ := tk.SignedString([]byte("my-secret-key"))
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/user-detail", nil)
		req.Header.Set("Authorization", "Bearer "+good)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		var resp pkgs.Response
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		// 在 handler 中如果查询不到用户返回404
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Equal(t, "用户不存在", resp.Msg)
	})
}
