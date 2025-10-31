# 扩展新业务规范
比如要实现一个xxx_name的模块功能，参考`internal/modules/template`的实现步骤：

1. 定义数据库迁移，创建业务所需要的数据库表文件，创建文件的方式采用：终端运行命令 `migrate create -ext sql -dir migraion/db your_xxx_module_name`；

2. 创建文件后，根据用户提供的业务设计，设计数据库表（1到多张表），遵守`promot/rules/sql-table.md`规则，参考`migration/db`目录下`template`模块的sql文件，在新增的数据库迁移up和down文件中写入数据迁移相关的建表、索引，外键约束，触发器等等内容，以及对应的数据库回滚操作。

3. 从 `configs/config.dev.yaml` 等配置文件中获取数据库连接信息，然后运行数据迁移命令，例如：`migrate -path migraion/db -database "YOUR_DATABASE_CONNECTION_STRING" up`。

4. 根据用户提供的业务设计，定义api路由与控制器接口，在`internal/api/v1/intf`目录下，参考`template`模块的handler接口，写入要实现的模块内所有的handler接口，然后在`internal/api/v1/router.go`文件中，参考`template`模块，修改构造函数，扩展 RegisterXXXModuleXXXTable 方法。

5. 创建新业务目录和文件，在 internal/modules 目录下创建新的业务目录 xxx_module_name ；如果业务中涉及到多个表，则在业务目录 xxx_module_name 下面再创建各个实体表 xxx_table_name 的目录;在业务目录或者实体表目录下创建 type, handler 代码文件；如果新业务有多张表，则在业务目录下创建 services/xxx_service 文件(services目录并不是必须的，根据业务设计来判断)；所有的go代码文件再增加第一行package语句

6. 参考`internal/modules/template`的`type.go`，给新业务下某个表的`type.go`中定义该数据表涉及的数据库表模型、表相关接口的输入参数和输出参数，代码注释使用中文，且在代码行上一行。

7. 参考`internal/modules/template`的`handler.go`，给新业务下某个表`handler.go`中实现在第4步中定义的相关控制器接口。一个接口主要分为 绑定参数、验证参数、创建数据实体、数据库表操作、返回接口响应等几个步骤。在数据库操作中，表名、字段名以`migration/db`目录下的sql文件为准。最后给接口函数增加swagger文档注释，代码注释使用中文。**特别注意：使用sqlx操作数据库时遵守`promot/rules/sqlx.md`内的规则**

8. 在internal/app/wire.go文件中注册新业务涉及的所有handler构造函数，然后使用 wire.bind 绑定 api/v1/intf 内定义的接口；之后，终端运行`wire ./internal/app` 更新依赖。
