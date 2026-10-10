-- Up migration: 000032_asset_investments_redesign.up.sql

-- 1. Update projecta_assets table with status, dates, and target benchmark pricing
ALTER TABLE projecta_assets
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    ADD COLUMN IF NOT EXISTS start_date TIMESTAMP NOT NULL DEFAULT current_timestamp,
    ADD COLUMN IF NOT EXISTS completed_date TIMESTAMP,
    ADD COLUMN IF NOT EXISTS target_price BIGINT,
    ADD COLUMN IF NOT EXISTS target_currency CHAR(3);

-- Copy existing price to target_price if target_price is null
UPDATE projecta_assets 
SET target_price = price, target_currency = currency 
WHERE target_price IS NULL AND price > 0;

-- Drop legacy cost type dependency on assets
ALTER TABLE projecta_assets DROP CONSTRAINT IF EXISTS projecta_assets_type_id_fk;
ALTER TABLE projecta_assets DROP COLUMN IF EXISTS type_id;

-- Make price and currency nullable or drop price/currency in favor of dynamic aggregation
-- We drop price/currency from projecta_assets because asset cost is dynamically aggregated
ALTER TABLE projecta_assets DROP COLUMN IF EXISTS price;
ALTER TABLE projecta_assets DROP COLUMN IF EXISTS currency;

-- 2. Create projecta_asset_compositions (Many-to-Many parent-child links)
CREATE TABLE IF NOT EXISTS projecta_asset_compositions
(
    parent_asset_id     UUID NOT NULL,
    child_asset_id      UUID NOT NULL,
    share_percentage    NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    created_at          TIMESTAMP NOT NULL DEFAULT current_timestamp,
    
    PRIMARY KEY (parent_asset_id, child_asset_id),
    CONSTRAINT chk_no_self_link CHECK (parent_asset_id <> child_asset_id),
    CONSTRAINT chk_valid_share CHECK (share_percentage > 0 AND share_percentage <= 100),
    CONSTRAINT projecta_comp_parent_fk 
        FOREIGN KEY (parent_asset_id) REFERENCES projecta_assets(asset_id) ON DELETE CASCADE,
    CONSTRAINT projecta_comp_child_fk 
        FOREIGN KEY (child_asset_id) REFERENCES projecta_assets(asset_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_asset_comp_parent ON projecta_asset_compositions(parent_asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_comp_child ON projecta_asset_compositions(child_asset_id);

-- 3. Create projecta_investments (replaces projecta_payments)
CREATE TABLE IF NOT EXISTS projecta_investments
(
    investment_id       UUID PRIMARY KEY NOT NULL,
    project_id          UUID NOT NULL,
    asset_id            UUID NOT NULL,
    contributor_id      UUID NOT NULL,
    resource_type       VARCHAR(20) NOT NULL DEFAULT 'MONEY', -- 'MONEY', 'TIME', 'GOODS'
    
    -- Monetary Amount (mandatory normalized valuation)
    amount              BIGINT NOT NULL,
    currency            CHAR(3) NOT NULL,
    
    -- Non-monetary resource metadata
    time_hours          NUMERIC(8, 2),
    time_hourly_rate    BIGINT,
    goods_quantity      NUMERIC(10, 2),
    goods_unit          VARCHAR(50),
    goods_item_name     TEXT,
    
    -- Common attributes
    description         TEXT NOT NULL,
    date                TIMESTAMP NOT NULL DEFAULT current_timestamp,
    tags                TEXT[] NOT NULL DEFAULT '{}',
    
    created_at          TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at          TIMESTAMP,
    
    CONSTRAINT projecta_investments_project_id_fk 
        FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE,
    CONSTRAINT projecta_investments_asset_id_fk 
        FOREIGN KEY (asset_id) REFERENCES projecta_assets(asset_id) ON DELETE CASCADE,
    CONSTRAINT projecta_investments_contributor_id_fk 
        FOREIGN KEY (contributor_id) REFERENCES people(person_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_projecta_investments_asset ON projecta_investments(asset_id);
CREATE INDEX IF NOT EXISTS idx_projecta_investments_project ON projecta_investments(project_id);
CREATE INDEX IF NOT EXISTS idx_projecta_investments_tags ON projecta_investments USING GIN (tags);

CREATE TRIGGER update_timestamp_trigger
    BEFORE UPDATE
    ON projecta_investments
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_trigger_function('updated_at');

-- 4. Data Migration from projecta_payments to projecta_investments
-- Step A: For projects with payments where no asset exists, create a default "General Operations" asset
INSERT INTO projecta_assets (asset_id, name, description, project_id, owner_id, status, start_date, created_at)
SELECT 
    gen_random_uuid(),
    'General Operations',
    'Auto-generated asset for existing payments and overhead',
    p.project_id,
    p.owner_id,
    'ACTIVE',
    COALESCE(p.started_at, current_timestamp),
    current_timestamp
FROM projecta_projects p
WHERE EXISTS (
    SELECT 1 FROM projecta_payments pay WHERE pay.project_id = p.project_id
)
AND NOT EXISTS (
    SELECT 1 FROM projecta_assets a WHERE a.project_id = p.project_id
);

-- Step B: Migrate payments into investments, mapping legacy category/type into tags
INSERT INTO projecta_investments (
    investment_id, project_id, asset_id, contributor_id, resource_type,
    amount, currency, description, date, tags, created_at, updated_at
)
SELECT 
    pay.payment_id,
    pay.project_id,
    (SELECT a.asset_id FROM projecta_assets a WHERE a.project_id = pay.project_id ORDER BY a.created_at ASC LIMIT 1),
    pay.owner_id,
    'MONEY',
    pay.amount,
    pay.currency,
    pay.description,
    pay.payment_date,
    ARRAY_REMOVE(ARRAY[
        LOWER(REPLACE(t.name, ' ', '-')), 
        LOWER(REPLACE(c.name, ' ', '-')),
        LOWER(REPLACE(pay.kind, '_', '-'))
    ], NULL),
    pay.created_at,
    pay.updated_at
FROM projecta_payments pay
LEFT JOIN projecta_cost_types t ON pay.type_id = t.type_id
LEFT JOIN projecta_cost_categories c ON t.category_id = c.category_id;

-- 5. Drop legacy payments, cost types, and cost categories
DROP TABLE IF EXISTS projecta_payments;
DROP TABLE IF EXISTS projecta_cost_types;
DROP TABLE IF EXISTS projecta_cost_categories;
