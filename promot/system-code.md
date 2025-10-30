# 核心第三方库
- Web 框架：Gin
- 配置管理：Viper
- 数据库：Postgresql 18, 配合 sqlx 进行 SQL 操作
- 日志记录：Zap
- 身份认证：JWT
- API 文档：Swagger
- 测试：Go 标准测试框架 + testify

# 规则索引
`promot/rules`目录下存放有 api路径、项目文件目录、中间件、模块、工具、powershell命令、sql建表、sqlx子集、测试等功能的规则明细，如果有需要就查看一下。

# **重要注意事项**
- **不要**按你的想法来，如果有涉及`promot/rules`目录下的规则的，请你严格遵守对应的规则，而不是你觉得好的就不遵守规则。

# 需求
实现`promot/iacc/2.code-step.md`里面第一阶段的第3步。
