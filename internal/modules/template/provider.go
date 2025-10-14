package template

import (
	"github.com/google/wire"
)

// ProviderSet 模板模块的Provider集合
var ProviderSet = wire.NewSet(
	NewTemplateHandler,
	RegisterRoutesV1,
)
