# 项目结构规范
go-pg-demo
├── api                  # API 版本
│   └── v1
│       ├── intf         # 目录下的文件用来定义handler接口
│       └── router.go
├── cmd                  # 应用程序入口
│   └── server
│       └── main.go
├── configs              # 配置文件
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   └── config.yaml
├── docs                 # API文档
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal             # 内部代码
│   ├── app              # 应用组装层
│   │   ├── app.go
│   │   ├── wire.go
│   │   └── wire_gen.go
│   ├── middlewares      # 中间件
│   │   ├── auth.go
│   │   ├── permission.go
│   │   ├── logger.go
│   │   ├── provider.go
│   │   └── recovery.go
│   └── modules          # 业务模块
│       ├── iacc         # IACC业务模块
│       │   ├── auth     # 认证模块
│       │   │   ├── handler.go      # HTTP处理器实现
│       │   │   ├── repository.go    # 数据访问层
│       │   │   └── type.go         # 数据类型定义
│       │   ├── permission  # 权限模块
│       │   │   ├── handler.go      # HTTP处理器实现
│       │   │   ├── repository.go    # 数据访问层
│       │   │   └── type.go         # 数据类型定义
│       │   ├── role        # 角色模块
│       │   │   ├── handler.go      # HTTP处理器实现
│       │   │   ├── repository.go    # 数据访问层
│       │   │   └── type.go         # 数据类型定义
│       │   └── user        # 用户模块
│       │       ├── handler.go      # HTTP处理器实现
│       │       ├── repository.go    # 数据访问层
│       │       └── type.go         # 数据类型定义
│       └── template      # 业务参考示例模板
│           ├── handler.go          # HTTP处理器实现
│           ├── repository.go        # 数据访问层
│           └── type.go             # 数据类型定义
├── migration            # 数据库迁移
│   ├── migrate.go
│   └── db
│       ├── 20251012153018_init.up.sql
│       ├── 20251012153018_init.down.sql
│       ├── 20251013153018_template.up.sql
│       ├── 20251013153018_template.down.sql
│       ├── 20251017155149_iacc_init.up.sql
│       └── 20251017155149_iacc_init.down.sql
├── pkgs                 # 公共包
│   ├── bind.go          # 数据绑定
│   ├── config.go        # 配置管理
│   ├── database.go      # 数据库连接
│   ├── error.go         # 错误处理
│   ├── init_admin_root.go # 初始化管理员
│   ├── logger.go        # 日志管理
│   ├── provider.go      # 依赖注入
│   ├── response.go      # 响应格式化
│   ├── scheduler.go     # 任务调度
│   ├── test_util.go     # 测试工具
│   └── validator.go     # 数据验证
├── promot               # 项目文档和规则
│   ├── rules            # 编码规范
│   ├── iacc             # IACC模块文档
│   ├── system-code.md   # 系统代码规范
│   └── workflow         # 工作流文档
├── test                 # 测试文件
│   ├── middlewares      # 中间件测试
│   │   └── permission
│   │       └── permission_middleware_test.go
│   └── v1               # API v1 测试
│       ├── iacc         # IACC模块测试
│       │   ├── auth
│       │   │   └── auth_test.go
│       │   ├── permission
│       │   │   └── permission_test.go
│       │   ├── role
│       │   │   └── role_test.go
│       │   └── user
│       │       └── user_test.go
│       └── template
│           └── template_test.go
├── .vscode              # IDE配置（可选）
├── go.mod               # Go模块定义
├── go.sum               # Go模块依赖校验
├── readme.md            # 项目说明文档
└── instruction.md       # 项目说明文档

