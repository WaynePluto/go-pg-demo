# 项目结构规范
go-pg-demo
├── api             # API 版本
│   └── v1
│       ├── intf    # 目录下的文件用来定义handler接口
│       └── router.go
├── cmd                 # 应用程序入口
│   └── server
│       └── main.go
├── configs             # 配置文件
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   └── config.yaml
├── internal            # 内部代码
│   ├── app             # 应用组装层
│   │   ├── app.go
│   │   ├── wire.go
│   │   └── wire_gen.go
│   ├── middlewares     # 中间件
│   │   ├── auth.go
│   │   ├── permission.go
│   │   ├── logger.go
│   │   ├── provider.go
│   │   └── recovery.go
│   └── modules             # 业务模块
│       └── template        # 业务参考示例模板
│           ├── handler.go
│           └── type.go
├── pkgs        
│   ├── provider.go        
│   ├── config.go
│   ├── database.go
│   ├── logger.go
│   ├── migrate.go
│   ├── test_util.go
│   └── response.go
└── readme.md

