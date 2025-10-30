-- 创建表
CREATE TABLE IF NOT EXISTS "template" (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(50),
    num int
);

-- 创建触发器
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trigger_update_updated_at_template'
          AND tgrelid = 'template'::regclass
    ) THEN
        CREATE TRIGGER trigger_update_updated_at_template
            BEFORE UPDATE ON "template"
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;