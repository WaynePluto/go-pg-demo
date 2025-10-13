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
├── api/                    # API 版本与路由定义 (对外入口)
│   └── v1/
│       └── router.go       # v1 版本的路由注册
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go         # main 函数，初始化并启动服务
├── internal/               # 内部私有代码，外部无法导入
│   ├── app/                        # 应用组装层 
│   │   ├── app.go                  # 定义 App 结构体，持有 Server, DB 等
│   │   ├── wire.go                 # 核心依赖注入定义文件
│   │   └── wiregen.go              # 由 wire 自动生成
│   ├── modules/            # 业务模块 (垂直分块)
│   │   ├── template/                   # 示例模块
│   │   │   ├── provider.go             # 模块内 Provider 集合，提供routes, handler, service, repo等具体实现的构造集合(wire.NewSet)
│   │   │   ├── routes.go               # 注册路由组，函数本身就是一个提供者
│   │   │   ├── handler.go              # HTTP 处理器，校验输入，包装输出
│   │   │   ├── handler_test.go         # handler 单元测试
│   │   │   ├── service.go              # 业务逻辑接口定义, ITemplateService
│   │   │   ├── service_test.go         # 单元测试
│   │   │   ├── service_impl.go         # 业务逻辑具体实现  
│   │   │   ├── repository.go           # 数据持久化接口定义, ITemplateRepo
│   │   │   ├── repository_test.go 
│   │   │   ├── repository_impl_pg.go   # 数据持久化基于pg的具体实现
│   │   │   └── model.go                # 领域模型，比如entity、dto
│   │   ├── user/...            # 用户模块 (参考示例)
│   │   ├── permission/...      # 权限模块 (参考示例)
│   │   └── role/...            # 角色模块 (参考示例)
│   ├── middlewares/         # 全局中间件
│   │   ├── provider.go     # wire.NewSet包装中间件集合
│   │   ├── auth.go         # JWT 认证中间件
│   │   ├── logger.go       # 日志中间件
│   │   └── recovery.go     # 异常恢复中间件
│   └── pkgs/                # 内部共享包
│       ├── provider.go       # wire.NewSet包装配置集合(NewConf, NewDB, NewLogger)
│       ├── config.go         # 配置加载逻辑
│       ├── database.go       # 数据库连接与初始化
│       ├── migrate.go        # 数据库迁移，在database.go连接后执行
│       ├── migrations        # 数据库迁移sql文件目录
│       ├── logger.go         # zap 日志
│       └── response.go       # 统一响应格式，不需要作为提供者
├── configs/                # 配置文件
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── tests/                         # 测试代码的根目录
│   └── integration/               # 集成测试
│       ├── user_test.go           # 测试 user 模块的 API 端点
│       ├── template_test.go       # 测试 template 模块的 API 端点
│       └── xxx_test.go            # 其他...
├── scripts/                # 构建、部署等脚本
├── go.mod
└── go.sum

```

## 三、 目录作用详解
*   `api/`: 负责路由，集成各个module内的routes。
*   `cmd/server/main.go`: **应用的启动器**。唯一职责是组装所有组件（配置、数据库、路由等）并启动 HTTP 服务器。保持其尽可能简洁。
*   `internal/module/`: **项目的核心**。每个子目录（如 `user`, `product`）代表一个独立的业务模块。
    *   **高内聚**：一个模块的所有东西（模型、业务逻辑、数据访问、API 处理）都在这里。
    *   **低耦合**：模块之间通过 `service` 接口通信，而不是直接访问对方的 `repository` 或 `handler`。
*   `internal/middleware`: **横切关注点**。如身份认证、日志记录、请求恢复等，可以被应用到多个路由上。
*   `internal/pkg/`: 存放内部可复用的组件，如数据库连接池、数据库迁移、配置加载器、日志工具等。它不是业务逻辑，而是基础设施。
*   `internal/app.go`: 导入所有模块和基础包,在 wire.Build 中注册所有 Provider,定义最终要生成的“根”对象（通常是 App 结构体）。
*   `configs/`: 存放配置文件，如 `config.yaml`。
*   `tests/`: 存放单元测试和集成测试。
  
## 四、 核心编码原则
1.  **关注点分离**: 每一层只做自己的事，互不越界。
2.  **依赖倒置原则**: 高层模块不应依赖低层模块，两者都应依赖抽象。
    *   **实践**: `service` 层不直接依赖 `repository` 的具体实现，而是依赖 `repository` 定义的 **接口**。例如，`UserService` 依赖 `IUserRepo` 接口，而不是 `UserRepoImplPostgre` 结构体。
    详细描述示例： 在 internal/modules/user/repository.go 中定义 IUserRepo 接口。在 internal/modules/user/repository_impl_pg.go 中定义 UserRepoImplPostgre 结构体，并实现 IUserRepo 接口的所有方法。在 main.go 中，创建 UserRepoImplPostgre 实例。将这个实例（作为 IUserRepo 接口类型）注入到 UserService 的构造函数中。这样一来，UserService 只关心接口契约，而不在乎底层是 PostgreSQL 还是 mongodb。
3.  **依赖注入**: 使用`wire.go`生成的初始化函数。
4.  **显式错误处理**: 永远不要忽略错误，使用 `if err != nil` 来处理每一个可能出错的操作。
5.  **接口驱动设计**: 先定义接口，再写实现。在设计阶段就思考好模块间的契约，并为测试和未来的扩展铺平道路。
   
## 五、 典型请求调用链路 (使用 Wire 依赖注入)
假设一个 `POST /api/v1/users` 的请求，其调用链路在 `wire` 的介入下，变得更加清晰和自动化：
1.  **启动与依赖注入**:
    *   **开发者定义**: 开发者在各个模块（如 `internal/modules/user`）和共享包（如 `internal/pkgs`）中编写 **Provider 函数**。这些函数通常以 `New` 开头（例如 `NewUserRepo`, `NewUserService`, `NewDB`），负责创建并返回一个组件（如 `repository` 实例、`service` 实例、数据库连接）。
    *   **开发者定义**: 开发者在 `internal/app.go` 或专门的 `internal/wire/wire.go` 文件中，定义一个 **Injector 函数签名**，例如 `func InitializeApp() (*App, error)`。这个函数声明了最终我们想要得到的根对象（在这里是包含了所有依赖的 `App` 结构体）。
    *   **Wire 生成**: 开发者运行 `wire` 命令（通常通过 `go generate` 触发）。`wire` 工具会分析所有 Provider 函数以及 Injector 函数签名，构建一个依赖图，并自动生成一个名为 `wire_gen.go` 的文件。这个文件里包含了 `InitializeApp` 函数的完整实现，其内部代码就是按正确顺序调用所有 Provider 函数来创建和组装依赖。
    *   **应用启动**: `cmd/server/main.go` 变得极其简洁。它不再关心任何组件的创建细节，只需直接调用 `wire` 生成的 `InitializeApp()` 函数，即可获得一个完全初始化、所有依赖都已注入的 `App` 实例。然后，它调用 `app.Run()` 或 `app.Server.ListenAndServe()` 来启动服务。
    **简而言之，`main.go` 从一个复杂的“组装工厂”变成了一个简单的“启动按钮”。**
2.  **路由**: 请求到达服务器，Gin 的路由器根据 `api/routes.go` 中定义的规则，找到 `POST /api/v1/users` 对应的处理函数是 `handlers.CreateUser`。这里的 `handler` 实例已经在 `wire` 的生成代码中被创建并注入到了路由中。
3.  **中间件**: 在请求到达 `handler` 之前，会先经过配置好的中间件（如 `LoggerMiddleware` 记录请求，`AuthMiddleware` 验证 JWT）。这些中间件本身也可以作为依赖，由 `wire` 进行管理和注入。
4.  **Handler**: `handlers.CreateUser` 函数被调用。它：
    a. 使用 `c.ShouldBindJSON()` 将请求体绑定到 `models.User` 结构体。
    b. 调用其内部已经由 `wire` 注入的 `userService` 实例的 `Create(user)` 方法。
    c. 根据 `service` 返回的结果和错误，封装成统一的 `response.Response` 格式，并用 `c.JSON()` 返回。
5.  **Service**: `userService.Create(user)` 函数被调用。它：
    a. 检查用户名是否已存在（调用其内部由 `wire` 注入的 `userRepo` 实例的 `GetByUsername()` 方法）。
    b. 调用 `userRepo.Create(user)` 将用户存入数据库。
    c. 返回创建成功的用户信息或错误。
6.  **Repository**: `userRepo.Create(user)` 函数被调用。它：
    a. 使用其内部由 `wire` 注入的 `sqlx.DB` 实例，通过 `NamedExec` 方法执行 `INSERT INTO users ...` 的原生 SQL 语句。
    b. 返回数据库操作的结果或错误。
7.  **响应**: 错误或结果沿着调用链层层返回，最终由 `handler` 打包成 JSON 响应给客户端。

## 六、 测试
测试分层进行：
1. Handler 层测试 (HTTP 测试)
目的：测试 HTTP 请求/响应的正确性。例如：状态码是否正确、响应头是否设置、响应体（JSON）是否符合预期。
方法：使用 Go 标准库的 net/http/httptest 来模拟 HTTP 请求，而无需启动真实的 HTTP 服务器。
关键：在这一层，我们需要 Mock 掉 Service 层，因为我们只想测试 Handler 的逻辑，不想依赖真实的业务逻辑和数据库。
2. Service 层测试 (业务逻辑单元测试)
目的：测试核心业务逻辑的正确性。例如：创建用户时，业务规则是否被遵守。
方法：这是最核心的单元测试。需要 Mock 掉 Repository 层，因为不想依赖真实的数据库。可以使用像 testify/mock 来生成 Mock 对象。
关键：确保测试覆盖了所有业务分支，包括成功和失败的场景。
3. Repository 层测试 (数据访测试)
目的：测试 SQL 语句的正确性和与数据库的交互。
方法：需要一个真实的测试数据库。
关键：测试 Prepare 语句是否正确、能否正确插入、查询、更新数据，以及事务处理是否正确。
4. 集成测试
它不关心 UserService 内部某个复杂算法的具体实现细节（那是单元测试的事），它只关心“我给 API 一个输入，它是否给我一个符合预期的输出”。
