package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-pg-demo/internal/modules/iacc/user"
	"go-pg-demo/pkgs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestProfileValueScan 测试 Profile 类型的 Value 和 Scan 方法
func TestProfileValueScan(t *testing.T) {
	t.Run("Value方法 - 正常序列化", func(t *testing.T) {
		email := "test@example.com"
		profile := user.Profile{
			Email: &email,
		}

		value, err := profile.Value()
		assert.NoError(t, err, "Value方法不应出错")

		// 验证返回的值是有效的JSON
		var result map[string]any
		err = json.Unmarshal(value.([]byte), &result)
		assert.NoError(t, err, "返回的值应该是有效的JSON")
		assert.Equal(t, email, result["email"], "JSON中的email应该匹配")
	})

	t.Run("Value方法 - 空Profile", func(t *testing.T) {
		profile := user.Profile{}

		value, err := profile.Value()
		assert.NoError(t, err, "Value方法不应出错")

		// 验证返回的值是有效的JSON
		var result map[string]any
		err = json.Unmarshal(value.([]byte), &result)
		assert.NoError(t, err, "返回的值应该是有效的JSON")
		assert.Empty(t, result, "空Profile应该序列化为空对象")
	})

	t.Run("Scan方法 - 从[]byte扫描", func(t *testing.T) {
		email := "test@example.com"
		jsonData := []byte(`{"email":"test@example.com"}`)

		var profile user.Profile
		err := profile.Scan(jsonData)
		assert.NoError(t, err, "Scan方法不应出错")
		assert.NotNil(t, profile.Email, "Email字段不应为nil")
		assert.Equal(t, email, *profile.Email, "Email值应该匹配")
	})

	t.Run("Scan方法 - 从string扫描", func(t *testing.T) {
		email := "test@example.com"
		jsonData := `{"email":"test@example.com"}`

		var profile user.Profile
		err := profile.Scan(jsonData)
		assert.NoError(t, err, "Scan方法不应出错")
		assert.NotNil(t, profile.Email, "Email字段不应为nil")
		assert.Equal(t, email, *profile.Email, "Email值应该匹配")
	})

	t.Run("Scan方法 - nil值", func(t *testing.T) {
		var profile user.Profile
		err := profile.Scan(nil)
		assert.NoError(t, err, "Scan方法处理nil值不应出错")
		assert.Equal(t, user.Profile{}, profile, "扫描nil值应返回空Profile")
	})

	t.Run("Scan方法 - 空字节数组", func(t *testing.T) {
		var profile user.Profile
		err := profile.Scan([]byte{})
		assert.NoError(t, err, "Scan方法处理空字节数组不应出错")
		assert.Equal(t, user.Profile{}, profile, "扫描空字节数组应返回空Profile")
	})

	t.Run("Scan方法 - 无效JSON", func(t *testing.T) {
		var profile user.Profile
		err := profile.Scan([]byte(`{invalid json`))
		assert.Error(t, err, "Scan方法处理无效JSON应该出错")
		assert.Contains(t, err.Error(), "解析 JSON 失败", "错误信息应该包含JSON解析失败")
	})

	t.Run("Scan方法 - 不支持的类型", func(t *testing.T) {
		var profile user.Profile
		err := profile.Scan(123)
		assert.Error(t, err, "Scan方法处理不支持的类型应该出错")
		assert.Contains(t, err.Error(), "无法将类型", "错误信息应该包含类型转换失败")
	})
}

// TestProfileJSONSerialization 测试 Profile 的 JSON 序列化和反序列化
func TestProfileJSONSerialization(t *testing.T) {
	t.Run("JSON序列化 - 有数据", func(t *testing.T) {
		email := "test@example.com"
		profile := user.Profile{
			Email: &email,
		}

		jsonBytes, err := json.Marshal(profile)
		assert.NoError(t, err, "JSON序列化不应出错")

		var result map[string]any
		err = json.Unmarshal(jsonBytes, &result)
		assert.NoError(t, err, "反序列化不应出错")
		assert.Equal(t, email, result["email"], "JSON中的email应该匹配")
	})

	t.Run("JSON序列化 - 空数据", func(t *testing.T) {
		profile := user.Profile{}

		jsonBytes, err := json.Marshal(profile)
		assert.NoError(t, err, "JSON序列化不应出错")

		var result map[string]any
		err = json.Unmarshal(jsonBytes, &result)
		assert.NoError(t, err, "反序列化不应出错")
		assert.Empty(t, result, "空Profile应该序列化为空对象")
	})

	t.Run("JSON反序列化 - 有数据", func(t *testing.T) {
		jsonData := `{"email":"test@example.com"}`
		var profile user.Profile
		err := json.Unmarshal([]byte(jsonData), &profile)
		assert.NoError(t, err, "JSON反序列化不应出错")
		assert.NotNil(t, profile.Email, "Email字段不应为nil")
		assert.Equal(t, "test@example.com", *profile.Email, "Email值应该匹配")
	})

	t.Run("JSON反序列化 - 空对象", func(t *testing.T) {
		jsonData := `{}`
		var profile user.Profile
		err := json.Unmarshal([]byte(jsonData), &profile)
		assert.NoError(t, err, "JSON反序列化不应出错")
		assert.Nil(t, profile.Email, "Email字段应为nil")
	})
}

// TestProfileInDatabase 测试 Profile 在数据库中的存储和检索
func TestProfileInDatabase(t *testing.T) {
	t.Run("数据库存储和检索 - 有Profile数据", func(t *testing.T) {
		// 准备测试数据
		username := "profileuser_" + uuid.NewString()[:5] // 确保用户名不超过20个字符
		phone := "138" + uuid.NewString()[:7]
		email := "dbtest@example.com"

		profileData := user.Profile{
			Email: &email,
		}

		// 直接在数据库中创建用户
		var userID string
		query := `INSERT INTO "iacc_user" (username, phone, password, profile) VALUES ($1, $2, $3, $4) RETURNING id`
		err := testDB.QueryRowContext(context.Background(), query, username, phone, "password123", profileData).Scan(&userID)
		assert.NoError(t, err, "创建带Profile的用户不应出错")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = $1`, userID)
			assert.NoError(t, err, "清理创建的用户不应出错")
		})

		// 从数据库检索用户
		var retrievedUser struct {
			ID      string       `db:"id"`
			Profile user.Profile `db:"profile"`
		}
		retrieveQuery := `SELECT id, profile FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &retrievedUser, retrieveQuery, userID)
		assert.NoError(t, err, "检索用户不应出错")
		assert.Equal(t, userID, retrievedUser.ID, "用户ID应该匹配")
		assert.NotNil(t, retrievedUser.Profile.Email, "Profile的Email字段不应为nil")
		assert.Equal(t, email, *retrievedUser.Profile.Email, "Profile的Email值应该匹配")
	})

	t.Run("数据库存储和检索 - NULL Profile", func(t *testing.T) {
		// 准备测试数据
		username := "nullprofile_" + uuid.NewString()[:5] // 确保用户名不超过20个字符
		phone := "138" + uuid.NewString()[:8]             // 确保手机号至少11个字符

		// 直接在数据库中创建用户，设置Profile为空JSON对象
		emptyProfile := user.Profile{}
		var userID string
		query := `INSERT INTO "iacc_user" (username, phone, password, profile) VALUES ($1, $2, $3, $4) RETURNING id`
		err := testDB.QueryRowContext(context.Background(), query, username, phone, "password123", emptyProfile).Scan(&userID)
		assert.NoError(t, err, "创建用户不应出错")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = $1`, userID)
			assert.NoError(t, err, "清理创建的用户不应出错")
		})

		// 从数据库检索用户
		var retrievedUser struct {
			ID      string       `db:"id"`
			Profile user.Profile `db:"profile"`
		}
		retrieveQuery := `SELECT id, profile FROM "iacc_user" WHERE id = $1`
		err = testDB.GetContext(context.Background(), &retrievedUser, retrieveQuery, userID)
		assert.NoError(t, err, "检索用户不应出错")
		assert.Equal(t, userID, retrievedUser.ID, "用户ID应该匹配")
		assert.Equal(t, user.Profile{}, retrievedUser.Profile, "空Profile应该被正确处理为空结构体")
	})
}

// TestProfileInAPI 测试 Profile 在 API 请求和响应中的处理
func TestProfileInAPI(t *testing.T) {
	t.Run("创建用户带Profile - API测试", func(t *testing.T) {
		// 准备请求数据
		username := "apiuser_" + uuid.NewString()[:5] // 确保用户名不超过20个字符
		phone := "138" + uuid.NewString()[:8]
		email := "apitest@example.com"

		createReqBody := map[string]any{
			"username": username,
			"phone":    phone,
			"password": "password123",
			"profile": map[string]any{
				"email": email,
			},
		}
		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/user", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行请求
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var createResp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, 200, createResp.Code, "响应码应该是 200")

		createdID, ok := createResp.Data.(string)
		assert.True(t, ok, "响应数据应该是字符串类型的 ID")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = $1`, createdID)
			assert.NoError(t, err, "清理创建的用户不应出错")
		})
	})

	t.Run("更新用户Profile - API测试", func(t *testing.T) {
		// 创建测试用户
		entity := setupTestUser(t)

		// 准备更新数据
		newEmail := "updated@example.com"
		updateReqBody := map[string]any{
			"profile": map[string]any{
				"email": newEmail,
			},
		}
		bodyBytes, _ := json.Marshal(updateReqBody)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		token := testUtil.GetAccessUserToken([]string{})

		// 执行更新请求
		req, _ := http.NewRequest(http.MethodPut, "/v1/user/"+entity["id"].(string), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")
	})

	t.Run("获取用户详情验证Profile - API测试", func(t *testing.T) {
		// 创建一个带有Profile数据的用户
		username := "getprofile_" + uuid.NewString()[:5] // 确保用户名不超过20个字符
		phone := "138" + uuid.NewString()[:8]            // 确保手机号至少11个字符
		email := "getprofile@example.com"

		// 使用Profile结构体而不是直接JSON
		profile := user.Profile{
			Email: &email,
		}

		var userID string
		query := `INSERT INTO "iacc_user" (username, phone, password, profile) VALUES ($1, $2, $3, $4) RETURNING id`
		err := testDB.QueryRowContext(context.Background(), query, username, phone, "password123", profile).Scan(&userID)
		assert.NoError(t, err, "创建带Profile的用户不应出错")

		// 清理
		t.Cleanup(func() {
			_, err := testDB.ExecContext(context.Background(), `DELETE FROM "iacc_user" WHERE id = $1`, userID)
			assert.NoError(t, err, "清理创建的用户不应出错")
		})

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		token := testUtil.GetAccessUserToken([]string{})

		// 执行获取请求
		req, _ := http.NewRequest(http.MethodGet, "/v1/user/"+userID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是 200")
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "解析响应体不应出错")
		assert.Equal(t, http.StatusOK, resp.Code, "响应码应该是 200")

		// 验证Profile字段
		data, ok := resp.Data.(map[string]any)
		assert.True(t, ok, "响应数据应该是一个 map")

		profileData, ok := data["profile"].(map[string]any)
		assert.True(t, ok, "Profile字段应该是一个map")
		assert.Equal(t, email, profileData["email"], "Profile中的email应该与创建时一致")
	})
}

// TestProfileEdgeCases 测试 Profile 的边界情况
func TestProfileEdgeCases(t *testing.T) {
	t.Run("空字符串Email", func(t *testing.T) {
		email := ""
		profile := user.Profile{
			Email: &email,
		}

		value, err := profile.Value()
		assert.NoError(t, err, "Value方法不应出错")

		var result map[string]any
		err = json.Unmarshal(value.([]byte), &result)
		assert.NoError(t, err, "返回的值应该是有效的JSON")
		assert.Equal(t, "", result["email"], "JSON中的email应该为空字符串")
	})

	t.Run("特殊字符Email", func(t *testing.T) {
		email := "test+special@example-domain.com"
		profile := user.Profile{
			Email: &email,
		}

		value, err := profile.Value()
		assert.NoError(t, err, "Value方法不应出错")

		var result map[string]any
		err = json.Unmarshal(value.([]byte), &result)
		assert.NoError(t, err, "返回的值应该是有效的JSON")
		assert.Equal(t, email, result["email"], "JSON中的email应该匹配")
	})

	t.Run("JSON包含额外字段", func(t *testing.T) {
		// 测试当JSON包含额外字段时的处理
		jsonData := `{"email":"test@example.com","extra_field":"extra_value"}`
		var profile user.Profile
		err := profile.Scan([]byte(jsonData))
		assert.NoError(t, err, "Scan方法不应出错")
		assert.NotNil(t, profile.Email, "Email字段不应为nil")
		assert.Equal(t, "test@example.com", *profile.Email, "Email值应该匹配")
		// 额外字段应该被忽略
	})

	t.Run("JSON缺少email字段", func(t *testing.T) {
		// 测试当JSON缺少email字段时的处理
		jsonData := `{"other_field":"other_value"}`
		var profile user.Profile
		err := profile.Scan([]byte(jsonData))
		assert.NoError(t, err, "Scan方法不应出错")
		assert.Nil(t, profile.Email, "Email字段应为nil")
	})
}
