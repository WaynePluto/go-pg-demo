# sqlx API 子集规范

## 目标：在项目中使用一套小而统一的 sqlx API 子集，保证可读性、支持命名参数、事务和可变 IN 列表。

## 可以使用的 API：

- GetContext(ctx, dest, query, args...)
  - 用途：一行查询，结果映射到结构体或基础类型。
  - 注意：无结果返回 sql.ErrNoRows，应由业务层映射到 404 或自定义错误。

- SelectContext(ctx, destSlice, query, args...)
  - 用途：多行查询并映射到切片。

- ExecContext(ctx, query, args...)
  - 用途：执行不返回行的 DML（DELETE/UPDATE、INSERT 不带 RETURNING）。

- NamedExecContext(ctx, query, argStructOrMap)
  - 用途：使用命名参数，参数顺序不敏感、可读性高。
  - 注意：NamedExecContext 返回 sql.Result；若需 INSERT ... RETURNING，请使用 PrepareNamedContext + GetContext。

- PrepareNamedContext(ctx, ...) + NamedStmt.GetContext(ctx, dest, argStructOrMap)
  - 用途：执行带 RETURNING 的写操作或在循环中复用语句，提高性能并安全获取返回列。

- BindNamed(query, argStructOrMap) + db.Rebind(query)
  - 用途：将命名参数转换为位置占位符并生成参数切片（用于构造动态 SQL 或结合 sqlx.In）。
  - 示例：
    ```go
    q, args, err := db.BindNamed("SELECT * FROM t WHERE id=:id", map[string]interface{}{"id": id})
    q = db.Rebind(q)
    _ = db.GetContext(ctx, &dest, q, args...)
    ```

- sqlx.In(query, args...) + db.Rebind(query)
  - 用途：处理可变长度的 IN 列表（例如批量删除）。先用 sqlx.In 生成占位符，然后 Rebind 以匹配驱动。

## 事务模式（通用模板）：

```go
tx, err := db.BeginTxx(ctx, nil)
if err != nil { return err }
defer func() {
    if p := recover(); p != nil {
        tx.Rollback()
        panic(p)
    }
    if err != nil { // err 为命名返回值或外层捕获的错误变量
        tx.Rollback()
    } else {
        err = tx.Commit()
    }
}()

// 使用 tx.NamedExecContext / tx.PrepareNamedContext 等
```

## 动态 UPDATE（字段可选更新）：

- 使用指针或 omitempty 的 struct 字段来区分“未传入”与“零值”。
- 或构造 map[string]interface{}，只包含需要更新的列；手动构建安全的 SET 列表（列名需白名单校验），然后用 NamedExecContext 或 BindNamed + ExecContext 执行。

示例（map + 安全列名）：

```go
cols := []string{}
for k := range data { // data 来自业务层且已经过列名白名单校验
    cols = append(cols, fmt.Sprintf("%s = :%s", k, k))
}
query := fmt.Sprintf("UPDATE table SET %s WHERE id=:id", strings.Join(cols, ", "))
_, err := db.NamedExecContext(ctx, query, data)
```

## 全局注意事项：

1. 优先使用带 Context 的 API 支持请求取消与超时。
2. 优先使用命名参数提高可读性并降低占位符错误。
3. 对大量批量写入场景，优先考虑 PostgreSQL COPY 或专用批量接口。
4. 对于数据库中有 DEFAULT 的列，Go 侧参数应省略该列（或使用指针/omitempty），并用 RETURNING 验证生成值。

## 小结：规范旨在高可读、低出错率的 sqlx 使用模式。对复杂场景（动态 SQL、批量、性能优化）写清楚约定与安全检查点。