package template_test

import (
	"testing"
)

// createSortTestTemplate 创建用于排序测试的模板
// 保持向后兼容性，内部调用 createTestTemplate
func createSortTestTemplate(t *testing.T, name string, num int) string {
	t.Helper()
	entity := createTestTemplate(t, name, &num)
	return entity["id"].(string)
}

// TestQueryListTemplates_Sort 已移至 template_query_test.go 文件中
