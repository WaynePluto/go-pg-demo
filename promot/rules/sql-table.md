# 数据库建表规范
1. 数据表名使用单数名词；
2. 如果一个模块内有多个表，表名要加上模块名称的缩写； 
3. 非关联表（中间表）的数据表前三个字段固定为:id(PRIMARY KEY, UUIDv7)、created_at(TIMESTAMPTZ)、updated_at(TIMESTAMPTZ);
4. 关联表（中间表）只需要包含所关联的表主键，并设置外键约束，然后使用联合主键；
5. 建表语句参考 `migration/db`下面template模块的sql语句;