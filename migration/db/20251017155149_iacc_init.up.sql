-- 创建用户表
CREATE TABLE IF NOT EXISTS "iacc_user" (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username VARCHAR(20) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password VARCHAR(255),
    profile JSONB
);

-- 创建角色表
CREATE TABLE IF NOT EXISTS "iacc_role" (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS "iacc_permission" (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    metadata JSONB NOT NULL
);

-- 创建 user_role 表
CREATE TABLE IF NOT EXISTS "iacc_user_role" (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

-- 创建 role_permission 表
CREATE TABLE IF NOT EXISTS "iacc_role_permission" (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);


-- 添加外键约束
ALTER TABLE "iacc_user_role" 
    ADD CONSTRAINT fk_user_role_user 
    FOREIGN KEY (user_id) REFERENCES "iacc_user"(id) ON DELETE CASCADE;

ALTER TABLE "iacc_user_role" 
    ADD CONSTRAINT fk_user_role_role 
    FOREIGN KEY (role_id) REFERENCES "iacc_role"(id) ON DELETE CASCADE;

ALTER TABLE "iacc_role_permission"
    ADD CONSTRAINT fk_role_permission_role
    FOREIGN KEY (role_id) REFERENCES "iacc_role"(id) ON DELETE CASCADE;

ALTER TABLE "iacc_role_permission"
    ADD CONSTRAINT fk_role_permission_permission
    FOREIGN KEY (permission_id) REFERENCES "iacc_permission"(id) ON DELETE CASCADE;


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

-- permission 表触发器
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trigger_update_updated_at_iacc_permission'
          AND tgrelid = 'iacc_permission'::regclass
    ) THEN
        CREATE TRIGGER trigger_update_updated_at_iacc_permission
            BEFORE UPDATE ON "iacc_permission"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;


