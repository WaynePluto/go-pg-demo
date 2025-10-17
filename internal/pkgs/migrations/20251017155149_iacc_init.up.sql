-- 创建 user 表
CREATE TABLE IF NOT EXISTS "iacc_user" (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username VARCHAR(50) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password VARCHAR(255),
    profile JSONB
);

-- 创建 role 表
CREATE TABLE IF NOT EXISTS "iacc_role" (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(50) UNIQUE,
    description TEXT,
    permissions JSONB
);

-- 创建 user_role 表
CREATE TABLE IF NOT EXISTS "iacc_user_role" (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID,
    role_id UUID,
    UNIQUE(user_id, role_id)
);

-- 添加外键约束
ALTER TABLE "iacc_user_role" 
    ADD CONSTRAINT fk_user_role_user 
    FOREIGN KEY (user_id) REFERENCES "iacc_user"(id);

ALTER TABLE "iacc_user_role" 
    ADD CONSTRAINT fk_user_role_role 
    FOREIGN KEY (role_id) REFERENCES "iacc_role"(id);

-- 创建触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为每个表创建触发器
-- user 表触发器
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trigger_update_updated_at_iacc_user'
          AND tgrelid = 'iacc_user'::regclass
    ) THEN
        CREATE TRIGGER trigger_update_updated_at_iacc_user
            BEFORE UPDATE ON "iacc_user"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- role 表触发器
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trigger_update_updated_at_iacc_role'
          AND tgrelid = 'iacc_role'::regclass
    ) THEN
        CREATE TRIGGER trigger_update_updated_at_iacc_role
            BEFORE UPDATE ON "iacc_role"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- user_role 表触发器
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trigger_update_updated_at_iacc_user_role'
          AND tgrelid = 'iacc_user_role'::regclass
    ) THEN
        CREATE TRIGGER trigger_update_updated_at_iacc_user_role
            BEFORE UPDATE ON "iacc_user_role"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;