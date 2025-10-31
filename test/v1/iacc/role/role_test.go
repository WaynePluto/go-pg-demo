package role_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-pg-demo/internal/app"
	"go-pg-demo/internal/modules/iacc/permission"
	"go-pg-demo/internal/modules/iacc/role"
	"go-pg-demo/pkgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testDB     *sqlx.DB
	testLogger *zap.Logger
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建应用实例
	testApp, _, err := app.InitializeApp()
	if err != nil {
		// 如果应用初始化失败，直接退出
		os.Exit(1)
	}

	testDB = testApp.DB
	testLogger = testApp.Logger
	testRouter = testApp.Server

	// 运行测试
	exitCode := m.Run()

	// 退出
	os.Exit(exitCode)
}

// 在数据库中创建一个角色用于测试，并注册一个清理函数以便在测试结束后删除它
func setupTestRole(t *testing.T, name string, description *string) role.RoleEntity {
	t.Helper()

	entity := role.RoleEntity{
		Name:        name,
		Description: description,
	}

	// 直接在数据库中创建实体
	query := `INSERT INTO iacc_role (name, description) VALUES (:name, :description) RETURNING id, created_at, updated_at`
	stmt, err := testDB.PrepareNamedContext(context.Background(), query)
	assert.NoError(t, err, "准备命名查询不应该出错")
	defer stmt.Close()

	err = stmt.GetContext(context.Background(), &entity, entity)
	assert.NoError(t, err, "创建测试角色不应出错")

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM iacc_role WHERE id = :id", map[string]interface{}{"id": entity.ID})
		if err != nil {
			t.Errorf("清理测试角色失败: %v", err)
		}
	})

	return entity
}

func TestCreateRole(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		// 准备
		description := "测试角色描述"
		createReqBody := role.CreateRoleReq{
			Name:        "新的测试角色",
			Description: &description,
		}

		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.NotEmpty(t, resp.Data, "返回数据不应该为空")

		// 清理测试数据
		t.Cleanup(func() {
			if resp.Data != nil {
				if roleID, ok := resp.Data.(string); ok {
					_, err := testDB.NamedExecContext(context.Background(), "DELETE FROM iacc_role WHERE id = :id", map[string]interface{}{"id": roleID})
					if err != nil {
						t.Errorf("清理测试角色失败: %v", err)
					}
				}
			}
		})
	})

	t.Run("缺少必填字段", func(t *testing.T) {
		// 准备
		createReqBody := role.CreateRoleReq{
			// 故意不提供Name字段
		}

		bodyBytes, _ := json.Marshal(createReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200（统一错误响应格式）")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应中的Code应该是400")
		assert.Equal(t, "角色名称为必填字段", resp.Msg, "应该返回字段验证错误信息")
	})
}

func TestGetRole(t *testing.T) {
	t.Run("成功获取角色", func(t *testing.T) {
		// 准备
		description := "测试角色描述"
		entity := setupTestRole(t, "获取测试角色", &description)

		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+entity.ID, nil)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.NotNil(t, resp.Data, "返回数据不应该为空")

		roleData, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "数据应该是对象类型")
		assert.Equal(t, entity.ID, roleData["id"], "返回的角色ID应该匹配")
		assert.Equal(t, entity.Name, roleData["name"], "返回的角色名称应该匹配")
	})

	t.Run("角色不存在", func(t *testing.T) {
		// 准备
		fakeID := uuid.New().String()
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+fakeID, nil)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		// 根据handler实现，不存在的角色会返回空数据而不是404错误
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Nil(t, resp.Data, "返回数据应该为空")
	})

	t.Run("无效ID", func(t *testing.T) {
		// 准备
		invalidID := "invalid-id"
		req, _ := http.NewRequest(http.MethodGet, "/v1/role/"+invalidID, nil)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		// 根据handler实现，无效ID会返回空数据而不是404错误
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Nil(t, resp.Data, "返回数据应该为空")
	})
}

func TestUpdateRole(t *testing.T) {
	t.Run("成功更新角色", func(t *testing.T) {
		// 准备
		description := "原始描述"
		entity := setupTestRole(t, "待更新角色", &description)

		newDescription := "更新后的描述"
		updateReqBody := role.UpdateRoleReq{
			Description: &newDescription,
		}

		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/role/"+entity.ID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		// 验证数据库中的值已更新
		var updatedEntity role.RoleEntity
		query := `SELECT id, name, description, created_at, updated_at FROM iacc_role WHERE id = $1`
		err := testDB.GetContext(context.Background(), &updatedEntity, query, entity.ID)
		assert.NoError(t, err, "应该能够查询到更新后的角色")
		assert.Equal(t, newDescription, *updatedEntity.Description, "描述应该已被更新")
	})

	t.Run("更新不存在的角色", func(t *testing.T) {
		// 准备
		fakeID := uuid.New().String()
		newDescription := "更新后的描述"
		updateReqBody := role.UpdateRoleReq{
			Description: &newDescription,
		}

		bodyBytes, _ := json.Marshal(updateReqBody)
		req, _ := http.NewRequest(http.MethodPut, "/v1/role/"+fakeID, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		// 根据handler实现，更新不存在的角色不会报错，只是影响0行
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Nil(t, resp.Data, "返回数据应该为空")
	})
}

func TestDeleteRole(t *testing.T) {
	t.Run("成功删除角色", func(t *testing.T) {
		// 准备
		description := "待删除角色描述"
		entity := setupTestRole(t, "待删除角色", &description)

		req, _ := http.NewRequest(http.MethodDelete, "/v1/role/"+entity.ID, nil)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, float64(1), resp.Data, "应该影响一行数据")

		// 验证角色确实已被删除
		var count int64
		query := `SELECT COUNT(*) FROM iacc_role WHERE id = $1`
		err = testDB.GetContext(context.Background(), &count, query, entity.ID)
		assert.NoError(t, err, "应该能够执行计数查询")
		assert.Equal(t, int64(0), count, "角色应该已被删除")
	})

	t.Run("删除不存在的角色", func(t *testing.T) {
		// 准备
		fakeID := uuid.New().String()
		req, _ := http.NewRequest(http.MethodDelete, "/v1/role/"+fakeID, nil)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")
		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, float64(0), resp.Data, "应该影响零行数据")
	})
}

func TestQueryRoleList(t *testing.T) {
	t.Run("成功查询角色列表", func(t *testing.T) {
		// 准备测试数据
		description1 := "测试角色1描述"
		description2 := "测试角色2描述"
		setupTestRole(t, "列表测试角色1", &description1)
		setupTestRole(t, "列表测试角色2", &description2)

		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?page=1&pageSize=10", nil)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.NotNil(t, resp.Data, "返回数据不应该为空")

		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "数据应该是对象类型")
		assert.Contains(t, data, "list", "数据应该包含list字段")
		assert.Contains(t, data, "total", "数据应该包含total字段")
	})

	t.Run("带名称筛选查询角色列表", func(t *testing.T) {
		// 准备测试数据
		description := "筛选测试描述"
		setupTestRole(t, "筛选测试角色ABC", &description)
		setupTestRole(t, "另一个测试角色", &description)

		req, _ := http.NewRequest(http.MethodGet, "/v1/role/list?name=ABC", nil)

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		var resp pkgs.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")

		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "数据应该是对象类型")
		list := data["list"].([]interface{})
		assert.Equal(t, 1, len(list), "应该只返回一条匹配的记录")
	})
}

func TestAssignPermission(t *testing.T) {
	t.Run("成功分配权限给角色", func(t *testing.T) {
		// 准备
		// 先创建角色
		description := "分配权限测试角色"
		entity := setupTestRole(t, "权限分配测试角色", &description)

		// 创建一些测试权限
		perm1 := setupTestPermission(t, "权限1")
		perm2 := setupTestPermission(t, "权限2")

		assignReqBody := role.AssignPermissionsReq{
			PermissionIDs: []string{perm1.ID, perm2.ID},
		}

		bodyBytes, _ := json.Marshal(assignReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role/"+entity.ID+"/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200")

		// 验证权限已分配
		var count int64
		query := `SELECT COUNT(*) FROM iacc_role_permission WHERE role_id = $1`
		err := testDB.GetContext(context.Background(), &count, query, entity.ID)
		assert.NoError(t, err, "应该能够执行计数查询")
		assert.Equal(t, int64(2), count, "应该有两条权限关联记录")
	})

	t.Run("清空角色权限", func(t *testing.T) {
		// 准备
		// 先创建角色
		description := "清空权限测试角色"
		entity := setupTestRole(t, "清空权限测试角色", &description)

		// 创建测试权限并关联到角色
		perm := setupTestPermission(t, "权限A")
		
		// 手动插入关联记录
		_, err := testDB.ExecContext(context.Background(), 
			"INSERT INTO iacc_role_permission (role_id, permission_id) VALUES ($1, $2)", 
			entity.ID, perm.ID)
		assert.NoError(t, err, "应该能成功插入权限关联")

		// 发送空权限列表请求应该失败，因为验证规则要求至少一个权限
		assignReqBody := role.AssignPermissionsReq{
			PermissionIDs: []string{}, // 空列表
		}

		bodyBytes, _ := json.Marshal(assignReqBody)
		req, _ := http.NewRequest(http.MethodPost, "/v1/role/"+entity.ID+"/permission", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		// 创建 TestUtil 实例
		testUtil := &pkgs.TestUtil{Engine: testRouter, DB: testDB, T: t}
		// 获取token
		token := testUtil.GetAccessUserToken([]string{})
		req.Header.Set("Authorization", "Bearer "+token)

		// 执行
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code, "状态码应该是200（统一错误响应格式）")
		var resp pkgs.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应该能正确解析为Response结构体")
		assert.Equal(t, http.StatusBadRequest, resp.Code, "响应中的Code应该是400")
	})
}

// 在数据库中创建一个权限用于测试，并注册一个清理函数以便在测试结束后删除它
func setupTestPermission(t *testing.T, name string) permission.PermissionEntity {
	t.Helper()

	metadata := permission.Metadata{
		Method: stringPtr("get"),
		Path:   stringPtr("/test/path"),
	}
	entity := permission.PermissionEntity{
		Name:     name,
		Type:     "api",
		Metadata: metadata,
	}

	// 直接在数据库中创建实体
	query := `INSERT INTO iacc_permission (name, type, metadata) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := testDB.QueryRowContext(context.Background(), query, entity.Name, entity.Type, entity.Metadata).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	assert.NoError(t, err, "创建测试权限不应出错")

	// 使用 t.Cleanup 注册清理函数，确保测试结束后数据被删除
	t.Cleanup(func() {
		_, err := testDB.ExecContext(context.Background(), "DELETE FROM iacc_permission WHERE id = $1", entity.ID)
		if err != nil {
			t.Errorf("清理测试权限失败: %v", err)
		}
	})

	return entity
}

func stringPtr(s string) *string {
	return &s
}
