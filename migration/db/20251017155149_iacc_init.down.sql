-- 删除触发器
DROP TRIGGER IF EXISTS trigger_update_updated_at_iacc_user ON "iacc_user";
DROP TRIGGER IF EXISTS trigger_update_updated_at_iacc_role ON "iacc_role";
DROP TRIGGER IF EXISTS trigger_update_updated_at_iacc_permission ON "iacc_permission";

-- 删除初始admin用户
DELETE FROM "iacc_user" WHERE username = 'admin';

-- 删除表
DROP TABLE IF EXISTS "iacc_user_role";
DROP TABLE IF EXISTS "iacc_role_permission";
DROP TABLE IF EXISTS "iacc_permission";
DROP TABLE IF EXISTS "iacc_role";
DROP TABLE IF EXISTS "iacc_user";