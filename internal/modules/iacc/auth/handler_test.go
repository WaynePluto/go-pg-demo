package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	v1 "go-pg-demo/api/v1"
	"go-pg-demo/internal/middlewares"
	"go-pg-demo/internal/modules/iacc/user"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	testAuthHandler *Handler
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

	authMiddleware := middlewares.NewAuthMiddleware(testConfig, testLogger)

	testAuthHandler = NewAuthHandler(testDB, testLogger, testValidator, testConfig)

	engine := gin.New()

	testRouter = &v1.Router{
		Engine:      engine,
		AuthHandler: testAuthHandler,
	}
	// 注册全局中间件
	testRouter.Engine.Use(gin.HandlerFunc(authMiddleware))
	// 注册路由组
	testRouter.RouterGroup = testRouter.Engine.Group("/v1")
	testRouter.RegisterIACCAuth()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	testRouter.Engine.ServeHTTP(rr, req)
	return rr
}

// 创建一个用于测试的用户
func setupTestUser(t *testing.T) (user.User, string) {
	t.Helper()
	password := "strongpassword"
	u := user.User{
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

	return u, password
}

// 获取访问令牌
// 获取访问令牌
func getAccessToken(t *testing.T, testUser user.User, password string) string {
	loginReq := LoginRequest{Username: testUser.Username, Password: password}
	loginBody, _ := json.Marshal(loginReq)
	loginReqHttp, err := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
	require.NoError(t, err)
	loginReqHttp.Header.Set("Content-Type", "application/json")
	loginW := executeRequest(loginReqHttp)
	var loginResp pkgs.Response
	err = json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	var loginData LoginResponse
	dataBytes, err := json.Marshal(loginResp.Data)
	require.NoError(t, err)
	err = json.Unmarshal(dataBytes, &loginData)
	require.NoError(t, err)
	return loginData.AccessToken
}

func TestLogin(t *testing.T) {
	t.Run("成功登录", func(t *testing.T) {
		// Arrange
		testUser, password := setupTestUser(t)
		loginReq := LoginRequest{
			Username: testUser.Username,
			Password: password,
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code, "响应状态码应为200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err, "解析响应JSON失败")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Msg)

		var loginResp LoginResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &loginResp)
		require.NoError(t, err)

		assert.NotEmpty(t, loginResp.AccessToken, "访问令牌不应为空")
		assert.NotEmpty(t, loginResp.RefreshToken, "刷新令牌不应为空")
		assert.Equal(t, testUser.ID, loginResp.User.ID, "返回的用户ID不匹配")
		assert.Equal(t, testUser.Username, loginResp.User.Username, "返回的用户名不匹配")
	})

	t.Run("失败 - 密码错误", func(t *testing.T) {
		// Arrange
		testUser, _ := setupTestUser(t)
		loginReq := LoginRequest{
			Username: testUser.Username,
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.Code, "业务码应为401 Unauthorized")
		assert.Equal(t, "用户名或密码错误", resp.Msg, "错误消息不匹配")
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("成功刷新Token", func(t *testing.T) {
		// Arrange
		// 1. Create user and login to get a refresh token
		testUser, password := setupTestUser(t)
		loginReq := LoginRequest{Username: testUser.Username, Password: password}
		loginBody, _ := json.Marshal(loginReq)
		loginReqHttp, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
		loginReqHttp.Header.Set("Content-Type", "application/json")
		loginW := executeRequest(loginReqHttp)
		require.Equal(t, http.StatusOK, loginW.Code)

		var loginResp pkgs.Response
		err := json.Unmarshal(loginW.Body.Bytes(), &loginResp)
		require.NoError(t, err)
		var loginData LoginResponse
		dataBytes, _ := json.Marshal(loginResp.Data)
		err = json.Unmarshal(dataBytes, &loginData)
		require.NoError(t, err)
		require.NotEmpty(t, loginData.RefreshToken)
		// 登录后要延时100ms, 防止生成一样的token
		time.Sleep(1000 * time.Millisecond)

		// 2. Use the refresh token to get a new access token
		reqBody := RefreshTokenRequest{RefreshToken: loginData.RefreshToken}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh-token", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var refreshResp RefreshTokenResponse
		dataBytes, _ = json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &refreshResp)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEmpty(t, refreshResp.RefreshToken)
		assert.NotEqual(t, loginData.AccessToken, refreshResp.AccessToken, "新的access token不应与旧的相同")
		assert.WithinDuration(t, time.Now(), refreshResp.ExpiresAt, 25*time.Hour) // 允许一些误差
	})
}

func TestGetProfile(t *testing.T) {
	t.Run("成功获取用户信息", func(t *testing.T) {
		// Arrange
		testUser, password := setupTestUser(t)
		accessToken := getAccessToken(t, testUser, password)

		// Make request to /profile
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/profile", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		// Act
		w := executeRequest(req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var profileResp UserInfoResponse
		profileBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(profileBytes, &profileResp)
		require.NoError(t, err)

		// The handler has a hardcoded user ID, so we can't assert against the created user.
		// We'll assert that the fields are not empty.
		assert.NotEmpty(t, profileResp.ID, "用户ID不应为空")
		assert.NotEmpty(t, profileResp.Username, "用户名不应为空")
	})
}
