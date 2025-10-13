package template

import (
	"context"
	"fmt"
	"go-pg-demo/internal/pkgs"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	config := &pkgs.Config{
		Database: pkgs.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Username:        "postgres",
			Password:        "0000",
			DBName:          "db_demo",
			SSLMode:         "disable",
			MaxIdleConns:    5,
			MaxOpenConns:    20,
			ConnMaxLifetime: time.Hour,
		},
	}

	// Build PostgreSQL connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.Username,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", connStr)

	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func teardownTestDB(t *testing.T, db *sqlx.DB) {

	db.Close()
}

func TestTemplateRepoImplPostgre_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewTemplateRepo(db)

	template := &TemplateEntity{
		Name: "Test Template",
		Num:  intPtr(10),
	}

	err := repo.Create(context.Background(), template)

	assert.NoError(t, err)

	// 验证数据是否正确插入
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM template WHERE id = $1", template.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestTemplateRepoImplPostgre_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewTemplateRepo(db)

	now := time.Now()
	expected := &TemplateEntity{
		ID:        "test-id-get",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Test Template",
		Num:       intPtr(10),
	}

	// 先插入测试数据
	_, err := db.NamedExec(`INSERT INTO template (id, created_at, updated_at, name, num) 
	                     VALUES (:id, :created_at, :updated_at, :name, :num)`, expected)
	assert.NoError(t, err)

	result, err := repo.GetByID(context.Background(), expected.ID)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Name, result.Name)
	if expected.Num != nil && result.Num != nil {
		assert.Equal(t, *expected.Num, *result.Num)
	}
}

func TestTemplateRepoImplPostgre_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewTemplateRepo(db)

	now := time.Now()
	template := &TemplateEntity{
		ID:        "test-id-update",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "Original Template",
		Num:       intPtr(10),
	}

	// 先插入测试数据
	_, err := db.NamedExec(`INSERT INTO template (id, created_at, updated_at, name, num) 
	                     VALUES (:id, :created_at, :updated_at, :name, :num)`, template)
	assert.NoError(t, err)

	// 更新数据
	template.Name = "Updated Template"
	template.Num = intPtr(20)
	template.UpdatedAt = time.Now()

	err = repo.Update(context.Background(), template)

	assert.NoError(t, err)

	// 验证数据是否正确更新
	var updated TemplateEntity
	err = db.Get(&updated, "SELECT id, name, num FROM template WHERE id = $1", template.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Template", updated.Name)
	assert.Equal(t, 20, *updated.Num)
}

func TestTemplateRepoImplPostgre_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewTemplateRepo(db)

	// 插入测试数据
	_, err := db.Exec("INSERT INTO template (id, created_at, updated_at, name) VALUES ($1, $2, $3, $4)",
		"test-id-delete", time.Now(), time.Now(), "Test Template")
	assert.NoError(t, err)

	err = repo.Delete(context.Background(), "test-id-delete")

	assert.NoError(t, err)

	// 验证数据是否被删除
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM template WHERE id = $1", "test-id-delete")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTemplateRepoImplPostgre_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewTemplateRepo(db)

	now := time.Now()
	templates := []*TemplateEntity{
		{
			ID:        "test-id-list-1",
			CreatedAt: now,
			UpdatedAt: now,
			Name:      "Test Template 1",
			Num:       intPtr(10),
		},
		{
			ID:        "test-id-list-2",
			CreatedAt: now,
			UpdatedAt: now,
			Name:      "Test Template 2",
			Num:       intPtr(20),
		},
	}

	// 插入测试数据
	for _, template := range templates {
		_, err := db.NamedExec(`INSERT INTO template (id, created_at, updated_at, name, num) 
		                     VALUES (:id, :created_at, :updated_at, :name, :num)`, template)
		assert.NoError(t, err)
	}

	result, totalCount, err := repo.List(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 2)
	assert.GreaterOrEqual(t, totalCount, int64(2))

	// 检查是否包含我们插入的数据
	found := 0
	for _, item := range result {
		if item.ID == "test-id-list-1" || item.ID == "test-id-list-2" {
			found++
		}
	}
	assert.GreaterOrEqual(t, found, 2)
}
