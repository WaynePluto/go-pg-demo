# 常用命令规则
## 运行项目内所有测试
```sh
go test ./... | Where-Object { $_ -notmatch 'no test files' }
```
## 数据库迁移
```sh
# 创建迁移文件，new_module_name是新的模块名字
migrate create -ext sql -dir migration/db new_module_name

# 使用 cli 同步，其中postgre的连接地址通过config文件获取
migrate -path migration/db -database "postgres://postgres:0000@localhost:5432/db_demo?sslmode=disable" up
```