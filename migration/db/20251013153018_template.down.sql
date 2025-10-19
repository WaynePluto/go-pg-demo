-- 删除触发器
DROP TRIGGER IF EXISTS trigger_update_updated_at_template ON "template";

-- 删除触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除表
DROP TABLE IF EXISTS "template";