# 数据库建表规范
1. 数据表名使用单数名词；
2. 对于关系模型，请你根据领域设计模型决定：1.选择关联表；2.选择json/jsonb字段内嵌到主表中；
3. 如果一个领域模型内有多个表，表名要加上领域名称的缩写； 
4. 数据表前三个字段固定为:id(UUIDv7)、created_at(TIMESTAMPTZ)、updated_at(TIMESTAMPTZ)
