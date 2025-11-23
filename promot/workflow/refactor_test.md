# 重构测试代码工作流

## 项目背景

### 核心第三方库
- Web 框架：Gin
- 配置管理：Viper
- 数据库：Postgresql 18, 配合 sqlx 进行 SQL 操作
- 日志记录：Zap
- 身份认证：JWT
- API 文档：Swagger
- 测试：Go 标准测试框架 + testify

### 规则索引
`promot/rules`目录下存放有 api路径、项目文件目录、中间件、模块、工具、powershell命令、sql建表、sqlx子集、测试等功能的规则明细，如果有需要就查看一下。

### **重要注意事项**
- **不要**按你的想法来，如果有涉及`promot/rules`目录下的规则的，请你严格遵守对应的规则。
- 如果你发现有规则定义出现了前后矛盾，请你停下任务，把矛盾的地方告诉我。

## 任务需求

请你帮我重构一下`test`目录下的测试代码。
要求（以`test/v1/template`代码为例）：
- 测试代码不应该依赖具体的实现代码，比如template表的测试代码中，不应该出现`go-pg-demo/internal/modules/template`依赖项，涉及到接口的入参和出参，使用`map[string]any`作为JSON的数据类型。
- 不要使用interface{}，使用any。
- 目前的测试代码都能运行通过，请你重构后，也运行一下重构的测试代码，保证重构正常。