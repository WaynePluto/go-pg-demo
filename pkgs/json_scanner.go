package pkgs

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// GenericJSONScan 通用的 JSON 扫描方法，适用于任何实现了 JSON 反序列化的类型
func GenericJSONScan[T any](target *T, value any) error {
	if value == nil {
		*target = *new(T) // 创建 T 类型的零值
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("无法将类型 %T 转换为 []byte 或 string", value)
	}

	// 如果字节数组为空，返回零值
	if len(bytes) == 0 {
		*target = *new(T)
		return nil
	}

	err := json.Unmarshal(bytes, target)
	if err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return nil
}

// GenericJSONValue 通用的 JSON 值方法，适用于任何实现了 JSON 序列化的类型
func GenericJSONValue[T any](source T) (driver.Value, error) {
	return json.Marshal(source)
}
