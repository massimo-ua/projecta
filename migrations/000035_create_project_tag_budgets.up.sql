-- Up migration: 000035_create_project_tag_budgets.up.sql

CREATE TABLE IF NOT EXISTS projecta_project_tag_budgets
(
    project_id          UUID NOT NULL,
    tag                 VARCHAR(100) NOT NULL,
    frequency           VARCHAR(20) NOT NULL DEFAULT 'MONTHLY', -- 'ONCE', 'DAILY', 'WEEKLY', 'MONTHLY', 'ANNUALLY'
    target_month        SMALLINT NOT NULL DEFAULT 1,            -- 1..12 (1 = January .. 12 = December, calendar month used for ONCE / ANNUALLY)
    resource_type       VARCHAR(20) NOT NULL DEFAULT 'MONEY',   -- 'MONEY', 'TIME', 'GOODS'
    amount              BIGINT NOT NULL,                        -- Normalized valuation per occurrence in minor units (cents) of project main_currency
    time_hours          NUMERIC(8, 2),
    time_hourly_rate    BIGINT,                                 -- Minor units (cents)
    goods_quantity      NUMERIC(10, 2),
    goods_unit          VARCHAR(50),
    goods_item_name     TEXT,
    description         TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMP NOT NULL DEFAULT current_timestamp,

    PRIMARY KEY (project_id, tag),
    CONSTRAINT chk_tag_budget_frequency CHECK (frequency IN ('ONCE', 'DAILY', 'WEEKLY', 'MONTHLY', 'ANNUALLY')),
    CONSTRAINT chk_tag_budget_target_month CHECK (target_month >= 1 AND target_month <= 12),
    CONSTRAINT chk_tag_budget_resource_type CHECK (resource_type IN ('MONEY', 'TIME', 'GOODS')),
    CONSTRAINT chk_tag_budget_amount CHECK (amount > 0),
    CONSTRAINT fk_tag_budgets_project
        FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_project_tag_budgets_project ON projecta_project_tag_budgets(project_id);
