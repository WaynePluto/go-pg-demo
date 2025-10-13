package pkgs

import "github.com/google/wire"

// ProviderSet 包含了基础组件的提供者集合
var ProviderSet = wire.NewSet(
	NewConfig,
	NewConnection,
	NewLogger,
)