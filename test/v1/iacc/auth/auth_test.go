package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go-pg-demo/internal/app"
	"go-pg-demo/internal/modules/iacc/auth"
	"go-pg-demo/internal/utils"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
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

func TestLogin(t *testing.T) {
	t.Run("成功登录", func(t *testing.T) {
		// Arrange
		testUtil.T = t
		testUser := testUtil.SetupTestUser()
		loginReq := auth.LoginRequest{
			Username: testUser.Username,
			Password: testUser.Password,
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code, "响应状态码应为200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err, "解析响应JSON失败")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Msg)

		var loginResp auth.LoginResponse
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
		testUtil.T = t
		testUser := testUtil.SetupTestUser()
		loginReq := auth.LoginRequest{
			Username: testUser.Username,
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

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
		testUtil.T = t
		testUser := testUtil.SetupTestUser()
		loginReq := auth.LoginRequest{Username: testUser.Username, Password: testUser.Password}
		loginBody, _ := json.Marshal(loginReq)
		loginReqHttp, _ := http.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginBody))
		loginReqHttp.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()
		testRouter.ServeHTTP(loginW, loginReqHttp)
		require.Equal(t, http.StatusOK, loginW.Code)

		var loginResp pkgs.Response
		err := json.Unmarshal(loginW.Body.Bytes(), &loginResp)
		require.NoError(t, err)
		var loginData auth.LoginResponse
		dataBytes, _ := json.Marshal(loginResp.Data)
		err = json.Unmarshal(dataBytes, &loginData)
		require.NoError(t, err)
		require.NotEmpty(t, loginData.RefreshToken)
		// 登录后要延时100ms, 防止生成一样的token
		time.Sleep(1000 * time.Millisecond)

		// 2. Use the refresh token to get a new access token
		reqBody := auth.RefreshTokenRequest{RefreshToken: loginData.RefreshToken}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh-token", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var refreshResp auth.RefreshTokenResponse
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
		testUtil.T = t
		accessToken := testUtil.GetAccessUserToken([]string{})
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/profile", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code)

		var profileResp auth.UserInfoResponse
		dataBytes, _ := json.Marshal(resp.Data)
		err = json.Unmarshal(dataBytes, &profileResp)
		require.NoError(t, err)
		assert.NotEmpty(t, profileResp.ID)
		assert.NotEmpty(t, profileResp.Username)
	})

	t.Run("失败 - 无效token", func(t *testing.T) {
		// Arrange
		req, _ := http.NewRequest(http.MethodGet, "/v1/auth/profile", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.string")

		// Act
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
