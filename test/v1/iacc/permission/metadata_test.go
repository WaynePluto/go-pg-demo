package permission_test

import (
	"encoding/json"
	"testing"

	"go-pg-demo/internal/modules/iacc/permission"

	"github.com/stretchr/testify/assert"
)

// TestMetadataValueScan 测试 Metadata 类型的 Value 和 Scan 方法
func TestMetadataValueScan(t *testing.T) {
	t.Run("Value方法 - 正常序列化", func(t *testing.T) {
		path := "/api/test"
		method := "GET"
		code := "API_TEST"

		metadata := permission.Metadata{
			Path:   &path,
			Method: &method,
			Code:   &code,
		}

		value, err := metadata.Value()
		assert.NoError(t, err, "Value方法不应出错")

		// 验证返回的值是有效的JSON
		var result map[string]any
		err = json.Unmarshal(value.([]byte), &result)
		assert.NoError(t, err, "返回的值应该是有效的JSON")
		assert.Equal(t, path, result["path"], "JSON中的path应该匹配")
		assert.Equal(t, method, result["method"], "JSON中的method应该匹配")
		assert.Equal(t, code, result["code"], "JSON中的code应该匹配")
	})

	t.Run("Value方法 - 空Metadata", func(t *testing.T) {
		metadata := permission.Metadata{}

		value, err := metadata.Value()
		assert.NoError(t, err, "Value方法不应出错")

		// 验证返回的值是有效的JSON
		var result map[string]any
		err = json.Unmarshal(value.([]byte), &result)
		assert.NoError(t, err, "返回的值应该是有效的JSON")
		assert.Empty(t, result, "空Metadata应该序列化为空对象")
	})

	t.Run("Scan方法 - 从[]byte扫描", func(t *testing.T) {
		path := "/api/test"
		method := "GET"
		code := "API_TEST"
		jsonData := []byte(`{"path":"/api/test","method":"GET","code":"API_TEST"}`)

		var metadata permission.Metadata
		err := metadata.Scan(jsonData)
		assert.NoError(t, err, "Scan方法不应出错")
		assert.NotNil(t, metadata.Path, "Path字段不应为nil")
		assert.NotNil(t, metadata.Method, "Method字段不应为nil")
		assert.NotNil(t, metadata.Code, "Code字段不应为nil")
		assert.Equal(t, path, *metadata.Path, "Path值应该匹配")
		assert.Equal(t, method, *metadata.Method, "Method值应该匹配")
		assert.Equal(t, code, *metadata.Code, "Code值应该匹配")
	})

	t.Run("Scan方法 - 从string扫描", func(t *testing.T) {
		path := "/api/test"
		method := "GET"
		code := "API_TEST"
		jsonData := `{"path":"/api/test","method":"GET","code":"API_TEST"}`

		var metadata permission.Metadata
		err := metadata.Scan(jsonData)
		assert.NoError(t, err, "Scan方法不应出错")
		assert.NotNil(t, metadata.Path, "Path字段不应为nil")
		assert.NotNil(t, metadata.Method, "Method字段不应为nil")
		assert.NotNil(t, metadata.Code, "Code字段不应为nil")
		assert.Equal(t, path, *metadata.Path, "Path值应该匹配")
		assert.Equal(t, method, *metadata.Method, "Method值应该匹配")
		assert.Equal(t, code, *metadata.Code, "Code值应该匹配")
	})

	t.Run("Scan方法 - nil值", func(t *testing.T) {
		var metadata permission.Metadata
		err := metadata.Scan(nil)
		assert.NoError(t, err, "Scan方法处理nil值不应出错")
		assert.Equal(t, permission.Metadata{}, metadata, "扫描nil值应返回空Metadata")
	})

	t.Run("Scan方法 - 空字节数组", func(t *testing.T) {
		var metadata permission.Metadata
		err := metadata.Scan([]byte{})
		assert.NoError(t, err, "Scan方法处理空字节数组不应出错")
		assert.Equal(t, permission.Metadata{}, metadata, "扫描空字节数组应返回空Metadata")
	})

	t.Run("Scan方法 - 无效JSON", func(t *testing.T) {
		var metadata permission.Metadata
		err := metadata.Scan([]byte(`{invalid json`))
		assert.Error(t, err, "Scan方法处理无效JSON应该出错")
		assert.Contains(t, err.Error(), "解析 JSON 失败", "错误信息应该包含JSON解析失败")
	})

	t.Run("Scan方法 - 不支持的类型", func(t *testing.T) {
		var metadata permission.Metadata
		err := metadata.Scan(123)
		assert.Error(t, err, "Scan方法处理不支持的类型应该出错")
		assert.Contains(t, err.Error(), "无法将类型", "错误信息应该包含类型转换失败")
	})

	t.Run("Scan方法 - JSON包含额外字段", func(t *testing.T) {
		// 测试当JSON包含额外字段时的处理
		jsonData := `{"path":"/api/test","method":"GET","code":"API_TEST","extra_field":"extra_value"}`
		var metadata permission.Metadata
		err := metadata.Scan([]byte(jsonData))
		assert.NoError(t, err, "Scan方法不应出错")
		assert.NotNil(t, metadata.Path, "Path字段不应为nil")
		assert.NotNil(t, metadata.Method, "Method字段不应为nil")
		assert.NotNil(t, metadata.Code, "Code字段不应为nil")
		assert.Equal(t, "/api/test", *metadata.Path, "Path值应该匹配")
		assert.Equal(t, "GET", *metadata.Method, "Method值应该匹配")
		assert.Equal(t, "API_TEST", *metadata.Code, "Code值应该匹配")
		// 额外字段应该被忽略
	})

	t.Run("Scan方法 - JSON缺少部分字段", func(t *testing.T) {
		// 测试当JSON缺少部分字段时的处理
		jsonData := `{"path":"/api/test","method":"GET"}`
		var metadata permission.Metadata
		err := metadata.Scan([]byte(jsonData))
		assert.NoError(t, err, "Scan方法不应出错")
		assert.NotNil(t, metadata.Path, "Path字段不应为nil")
		assert.NotNil(t, metadata.Method, "Method字段不应为nil")
		assert.Nil(t, metadata.Code, "Code字段应为nil")
		assert.Equal(t, "/api/test", *metadata.Path, "Path值应该匹配")
		assert.Equal(t, "GET", *metadata.Method, "Method值应该匹配")
	})

	t.Run("ValueScan往返测试", func(t *testing.T) {
		// 测试Value和Scan方法的往返一致性
		original := permission.Metadata{
			Path:   stringPtr("/api/test"),
			Method: stringPtr("GET"),
			Code:   stringPtr("API_TEST"),
		}

		// 使用Value方法序列化
		value, err := original.Value()
		assert.NoError(t, err, "Value方法不应出错")

		// 使用Scan方法反序列化
		var restored permission.Metadata
		err = restored.Scan(value)
		assert.NoError(t, err, "Scan方法不应出错")

		// 验证往返一致性
		assert.Equal(t, original.Path, restored.Path, "Path字段应该一致")
		assert.Equal(t, original.Method, restored.Method, "Method字段应该一致")
		assert.Equal(t, original.Code, restored.Code, "Code字段应该一致")
	})
}
