# Go 后端项目搭建指南

## 一、 核心第三方依赖

| 分类 | 依赖包 | 作用 |
| :--- | :--- | :--- |
| **Web 框架** | `github.com/gin-gonic/gin` | 提供路由、中间件等功能。 |
| **数据库驱动** | `github.com/lib/pq` | PostgreSQL 的官方驱动。 |
| **SQL 辅助** | `github.com/jmoiron/sqlx` | 使用原生 SQL 时，将结果扫描到结构体，并支持命名参数。 |
| **数据库迁移** | `github.com/golang-migrate/migrate/v4` | 数据库迁移核心库 |
| **依赖注入** | `github.com/google/wire/cmd/wire@latest` | 自动依赖注入 |
| **配置管理** | `github.com/spf13/viper` | 支持多种格式（JSON, YAML, TOML）和环境变量。 |
| **日志** | `go.uber.org/zap` | 高性能、结构化的日志库 |
| **JWT 认证** | `github.com/golang-jwt/jwt/v5` | 处理 JWT 的生成和解析。 |
| **数据验证** | `github.com/go-playground/validator/v10` | 结构体和字段验证器，与 Gin 无缝集成。 |
| **测试辅助** | `github.com/stretchr/testify` | 提供丰富的断言和模拟功能，让单元测试代码更简洁、易读。 |

## 二、 项目目录结构

```
my-project/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go         # main 函数，初始化并启动服务
├── configs/                # 配置文件
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── internal/               # 内部私有代码，外部无法导入
│   ├── api/                    # API 版本与路由定义 (对外入口)
│   │   └── v1/
│   │       └── router.go       # v1 版本的路由注册
│   ├── app/                        # 应用组装层 
│   │   ├── app.go                  # 定义 App 结构体，持有 Server, DB 等
│   │   ├── wire.go                 # 核心依赖注入定义文件
│   │   └── wire_gen.go              # 由 wire 自动生成
│   ├── middlewares/         # 全局中间件
│   │   ├── provider.go     # 注册中间件函数的构造函数
│   │   ├── auth.go         # JWT 认证中间件
│   │   ├── logger.go       # 日志中间件
│   │   └── recovery.go     # 异常恢复中间件
│   ├── modules/            # 业务模块 (垂直分块)
│   │   ├── template/                   # 示例模块
│   │   │   ├── handler_test.go         # HTTP 处理器测试代码
│   │   │   ├── handler.go              # HTTP 处理器，校验输入，包装输出
│   │   │   ├── routes.go               # 注册路由组
│   │   │   └── type.go                 # 类型定义，比如entity、dto
│   │   ├── user/...            # 用户模块 (参考示例)
│   │   ├── permission/...      # 权限模块 (参考示例)
│   │   └── role/...            # 角色模块 (参考示例)
│   └── pkgs/                # 内部共享包
│       ├── provider.go       # wire.NewSet包装配置集合(NewConf, NewDB, NewLogger)
│       ├── config.go         # 配置加载逻辑
│       ├── database.go       # 数据库连接与初始化
│       ├── migrate.go        # 数据库迁移，在database.go连接后执行
│       ├── migrations        # 数据库迁移集合
│       ├── logger.go         # zap 日志
│       └── response.go       # 统一响应格式，不需要作为提供者
├── scripts/                # 构建、部署等脚本
├── go.mod
└── go.sum

```

## 三、 目录作用详解
*   `api/`: 负责路由，集成各个module内的routes。
*   `cmd/server/main.go`: **应用的启动器**。唯一职责是组装所有组件（配置、数据库、路由等）并启动 HTTP 服务器。保持其尽可能简洁。
*   `internal/module/`: **项目的核心**。每个子目录（如 `user`, `product`）代表一个独立的业务模块。
*   `internal/middleware`: **横切关注点**。如身份认证、日志记录、请求恢复等，可以被应用到多个路由上。
*   `internal/pkg/`: 存放内部可复用的组件，如数据库连接池、数据库迁移、配置加载器、日志工具等。它不是业务逻辑，而是基础设施。
*   `internal/app.go`: 导入所有模块和基础包,在 wire.Build 中注册所有 Provider,定义最终要生成的“根”对象（通常是 App 结构体）。
*   `configs/`: 存放配置文件，如 `config.yaml`。
  
## 四、 核心编码原则
1.  **依赖注入**: 使用`wire.go`生成的初始化函数。
2.  **显式错误处理**: 永远不要忽略错误，使用 `if err != nil` 来处理每一个可能出错的操作。
   
## 五、 典型请求调用链路 (使用 Wire 依赖注入)
假设一个 `POST /api/v1/users` 的请求，其调用链路在 `wire` 的介入下，变得更加清晰和自动化：
1.  **启动与依赖注入**:
    *   **开发者定义**: 开发者在各个模块（如 `internal/modules/user`）和共享包（如 `internal/pkgs`）中编写 **Provider 函数**。这些函数通常以 `New` 开头，负责创建并返回一个组件实例。
    *   **开发者定义**: 开发者在 `internal/app.go` 定义实例结构体，定义要包含哪些依赖；专门的 `internal/app/wire.go` 文件中，定义一个 **Injector 函数签名**，例如 `func InitializeApp() (*App, error)`。这个函数声明了最终我们想要得到的根对象（在这里是包含了所有依赖的 `App` 结构体）。
    *   **Wire 生成**: 开发者运行 `wire` 命令（通常通过 `go generate` 触发）。`wire` 工具会分析所有 Provider 函数以及 Injector 函数签名，构建一个依赖图，并自动生成一个名为 `wire_gen.go` 的文件。这个文件里包含了 `InitializeApp` 函数的完整实现，其内部代码就是按正确顺序调用所有 Provider 函数来创建和组装依赖。
    *   **应用启动**: `cmd/server/main.go` 变得极其简洁。它不再关心任何组件的创建细节，只需直接调用 `wire` 生成的 `InitializeApp()` 函数，即可获得一个完全初始化、所有依赖都已注入的 `App` 实例。然后，它调用 `app.Run()` 或 `app.Server.ListenAndServe()` 来启动服务。
    **简而言之，`main.go` 从一个复杂的“组装工厂”变成了一个简单的“启动按钮”。**
2.  **路由**: 请求到达服务器，Gin 的路由器根据 `api/router.go` 中定义的规则，找到 `POST /api/v1/users` 对应的处理函数是 `handlers.CreateUser`。这里的 `handler` 实例已经在 `wire` 的生成代码中被创建并注入到了路由中。
3.  **中间件**: 在请求到达 `handler` 之前，会先经过配置好的中间件（如 `LoggerMiddleware` 记录请求，`AuthMiddleware` 验证 JWT）。这些中间件本身也可以作为依赖，由 `wire` 进行管理和注入。
4.  **Handler**: `handlers.CreateUser` 函数被调用。它：
    a.绑定请求参数
    b.用请求参数创建实体
    c.数据库操作实体
    d.返回数据json

## 六、 测试
集成测试：“给 API 一个输入，返回一个符合预期的输出”。
