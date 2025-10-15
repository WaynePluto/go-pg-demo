# go-pg-demo

一个基于 Go 语言构建的后端服务模板，旨在提供一个结构清晰、可扩展性强的 Web 服务基础框架。通过集成主流库和遵循最佳实践，帮助开发者快速搭建生产级应用。

## 功能特性

- 基于 Gin 框架的 HTTP 路由与中间件支持
- 使用 PostgreSQL 作为持久化存储，配合 sqlx 进行原生 SQL 操作
- JWT 认证与权限控制（通过中间件实现）
- 配置文件加载（YAML 格式 + 环境变量支持）
- 数据库自动迁移（migrate/v4）
- 统一响应格式封装
- 依赖注入（Google Wire）
- Swagger API 文档自动生成

## 技术栈

- Web 框架: github.com/gin-gonic/gin v1.11.0
- 数据库驱动: github.com/lib/pq v1.10.9
- SQL 辅助: github.com/jmoiron/sqlx v1.4.0
- 数据库迁移: github.com/golang-migrate/migrate/v4 v4.19.0
- 依赖注入: github.com/google/wire v0.7.0
- 配置管理: github.com/spf13/viper v1.21.0
- 日志库: go.uber.org/zap v1.27.0
- JWT 认证: github.com/golang-jwt/jwt/v5 v5.3.0
- Swagger 文档: github.com/swaggo/gin-swagger v1.6.0
- 数据验证: github.com/go-playground/validator/v10
- 测试辅助: github.com/stretchr/testify

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 12+

### 安装依赖

```bash
go mod download
```

### 配置数据库

复制配置文件:

```bash
cp configs/config.dev.yaml.example configs/config.dev.yaml
```

修改 `configs/config.dev.yaml` 中的数据库配置:

```yaml
database:
  host: localhost
  port: 5432
  user: your_user
  password: your_password
  dbname: your_database
  sslmode: disable
```

### 运行服务

```bash
cd cmd/server
go run main.go
```

### 访问 Swagger 文档

服务启动后，可以通过以下地址访问 Swagger API 文档：

```
http://localhost:8080/swagger/index.html
```

## 项目结构

```
.
├── cmd                 # 应用程序入口
│   └── server
│       └── main.go
├── configs             # 配置文件
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   └── config.yaml
├── internal            # 内部代码
│   ├── api             # API 版本
│   │   └── v1
│   │       └── router.go
│   ├── app             # 应用组装层
│   │   ├── app.go
│   │   ├── wire.go
│   │   └── wire_gen.go
│   ├── middlewares     # 中间件
│   │   ├── auth.go
│   │   ├── logger.go
│   │   ├── provider.go
│   │   └── recovery.go
│   ├── modules         # 业务模块
│   │   ├── permission
│   │   │   └── permission.go
│   │   ├── role
│   │   │   └── role.go
│   │   ├── template
│   │   │   ├── handler.go
│   │   │   ├── handler_test.go
│   │   │   ├── routes.go
│   │   │   └── type.go
│   │   └── user
│   │       └── user.go
│   └── pkgs            # 内部共享包
│       ├── migrations
│       │   ├── 20251013153018_template.down.sql
│       │   └── 20251013153018_template.up.sql
│       ├── config.go
│       ├── database.go
│       ├── logger.go
│       ├── migrate.go
│       ├── provider.go
│       └── response.go
└── readme.md
```

## API 文档

项目集成了 Swagger 来自动生成 API 文档。服务运行后，访问 `http://localhost:8080/swagger/index.html` 即可查看完整的 API 文档。

## 依赖注入

项目使用 Google Wire 进行依赖注入，简化了组件之间的依赖关系管理。

生成依赖注入代码：

```bash
cd internal/app
go generate
```

## 数据库迁移

项目使用 golang-migrate 工具进行数据库迁移。在服务启动时会自动运行迁移脚本。

## 测试

运行测试：

```bash
go test ./...
```