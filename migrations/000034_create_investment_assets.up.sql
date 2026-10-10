-- Up migration: 000034_create_investment_assets.up.sql

CREATE TABLE IF NOT EXISTS projecta_investment_assets
(
    investment_id       UUID NOT NULL,
    asset_id            UUID NOT NULL,
    share_percentage    NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    created_at          TIMESTAMP NOT NULL DEFAULT current_timestamp,
    
    PRIMARY KEY (investment_id, asset_id),
    CONSTRAINT chk_investment_share CHECK (share_percentage > 0 AND share_percentage <= 100),
    CONSTRAINT fk_inv_assets_investment 
        FOREIGN KEY (investment_id) REFERENCES projecta_investments(investment_id) ON DELETE CASCADE,
    CONSTRAINT fk_inv_assets_asset 
        FOREIGN KEY (asset_id) REFERENCES projecta_assets(asset_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_inv_assets_investment ON projecta_investment_assets(investment_id);
CREATE INDEX IF NOT EXISTS idx_inv_assets_asset ON projecta_investment_assets(asset_id);

-- Backfill existing single-asset investments into projecta_investment_assets
INSERT INTO projecta_investment_assets (investment_id, asset_id, share_percentage)
SELECT investment_id, asset_id, 100.00
FROM projecta_investments
WHERE asset_id IS NOT NULL
ON CONFLICT (investment_id, asset_id) DO NOTHING;
