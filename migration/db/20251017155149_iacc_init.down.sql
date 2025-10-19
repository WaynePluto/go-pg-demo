-- 删除触发器
DROP TRIGGER IF EXISTS trigger_update_updated_at_iacc_user ON "iacc_user";
DROP TRIGGER IF EXISTS trigger_update_updated_at_iacc_role ON "iacc_role";
DROP TRIGGER IF EXISTS trigger_update_updated_at_iacc_user_role ON "iacc_user_role";

-- 删除外键约束
ALTER TABLE "iacc_user_role" 
    DROP CONSTRAINT IF EXISTS fk_user_role_user;
    
ALTER TABLE "iacc_user_role" 
    DROP CONSTRAINT IF EXISTS fk_user_role_role;

-- 删除表
DROP TABLE IF EXISTS "iacc_user_role";
DROP TABLE IF EXISTS "iacc_role";
DROP TABLE IF EXISTS "iacc_user";

-- 删除触发器函数（仅当没有其他表使用时）
DROP FUNCTION IF EXISTS update_updated_at_column();