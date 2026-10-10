-- Down migration: 000032_asset_investments_redesign.down.sql

-- Recreate projecta_cost_categories
CREATE TABLE IF NOT EXISTS projecta_cost_categories
(
    category_id     UUID PRIMARY KEY NOT NULL,
    project_id      UUID NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    created_at      TIMESTAMP DEFAULT current_timestamp,
    updated_at      TIMESTAMP,
    CONSTRAINT projecta_cost_categories_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE
);

-- Recreate projecta_cost_types
CREATE TABLE IF NOT EXISTS projecta_cost_types
(
    type_id         UUID PRIMARY KEY NOT NULL,
    project_id      UUID NOT NULL,
    category_id     UUID NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    created_at      TIMESTAMP DEFAULT current_timestamp,
    updated_at      TIMESTAMP,
    CONSTRAINT projecta_cost_types_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE,
    CONSTRAINT projecta_cost_types_category_id_fk FOREIGN KEY (category_id) REFERENCES projecta_cost_categories(category_id) ON DELETE CASCADE
);

-- Recreate projecta_payments
CREATE TABLE IF NOT EXISTS projecta_payments
(
    payment_id          UUID PRIMARY KEY NOT NULL,
    project_id          UUID NOT NULL,
    type_id             UUID,
    amount              BIGINT NOT NULL,
    currency            CHAR(3) NOT NULL,
    description         TEXT,
    payment_date        TIMESTAMP,
    kind                VARCHAR(50) DEFAULT 'UPON_COMPLETION',
    owner_id            UUID,
    created_at          TIMESTAMP DEFAULT current_timestamp,
    updated_at          TIMESTAMP,
    CONSTRAINT projecta_payments_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE
);

-- Drop projecta_investments
DROP TRIGGER IF EXISTS update_timestamp_trigger ON projecta_investments;
DROP TABLE IF EXISTS projecta_investments;

-- Drop projecta_asset_compositions
DROP TABLE IF EXISTS projecta_asset_compositions;

-- Revert projecta_assets columns
ALTER TABLE projecta_assets
    ADD COLUMN IF NOT EXISTS type_id UUID,
    ADD COLUMN IF NOT EXISTS price BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS currency CHAR(3) DEFAULT 'UAH';

ALTER TABLE projecta_assets
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS start_date,
    DROP COLUMN IF EXISTS completed_date,
    DROP COLUMN IF EXISTS target_price,
    DROP COLUMN IF EXISTS target_currency;
