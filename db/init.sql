CREATE SCHEMA IF NOT EXISTS org;

-- =========================================================
-- 1. Org Service 表
-- =========================================================
CREATE TABLE org.organizations
(
    id          BIGSERIAL PRIMARY KEY,
    parent_id   BIGINT REFERENCES org.organizations (id) ON DELETE CASCADE,
    name        VARCHAR(120) NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    disabled_at TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    
    -- 添加约束确保逻辑删除的数据不能被禁用
    CONSTRAINT chk_deleted_not_disabled CHECK (
        (deleted_at IS NULL) OR (disabled_at IS NULL)
    ),
    
    -- 确保时间戳的逻辑性
    CONSTRAINT chk_timestamps CHECK (
        created_at <= updated_at AND
        (disabled_at IS NULL OR disabled_at >= created_at) AND
        (deleted_at IS NULL OR deleted_at >= created_at)
    )
);

-- 创建触发器函数自动更新updated_at字段
CREATE OR REPLACE FUNCTION org.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建触发器
CREATE TRIGGER trigger_update_organizations_updated_at
    BEFORE UPDATE ON org.organizations
    FOR EACH ROW
    EXECUTE FUNCTION org.update_updated_at_column();

-- 索引优化
CREATE INDEX idx_org_parent ON org.organizations (parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_org_active ON org.organizations (id) WHERE deleted_at IS NULL AND disabled_at IS NULL;
CREATE INDEX idx_org_name ON org.organizations (name) WHERE deleted_at IS NULL;
CREATE INDEX idx_org_created_at ON org.organizations (created_at);
CREATE INDEX idx_org_updated_at ON org.organizations (updated_at);

-- 复合索引用于常见查询场景
CREATE INDEX idx_org_parent_active ON org.organizations (parent_id, id) 
    WHERE deleted_at IS NULL AND disabled_at IS NULL;

-- 用于软删除查询的索引
CREATE INDEX idx_org_deleted_at ON org.organizations (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_org_disabled_at ON org.organizations (disabled_at) WHERE disabled_at IS NOT NULL;

-- 添加注释
COMMENT ON TABLE org.organizations IS '组织机构表，支持树形结构';
COMMENT ON COLUMN org.organizations.id IS '主键ID';
COMMENT ON COLUMN org.organizations.parent_id IS '父级组织ID，支持树形结构';
COMMENT ON COLUMN org.organizations.name IS '组织名称，不能为空或纯空格';
COMMENT ON COLUMN org.organizations.created_at IS '创建时间';
COMMENT ON COLUMN org.organizations.updated_at IS '更新时间，通过触发器自动维护';
COMMENT ON COLUMN org.organizations.disabled_at IS '禁用时间，NULL表示未禁用';
COMMENT ON COLUMN org.organizations.deleted_at IS '软删除时间，NULL表示未删除';