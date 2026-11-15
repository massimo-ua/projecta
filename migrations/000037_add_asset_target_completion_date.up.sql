ALTER TABLE projecta_assets
    ADD COLUMN IF NOT EXISTS target_completion_date TIMESTAMP DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_projecta_assets_project_schedule
    ON projecta_assets (project_id, start_date, target_completion_date);
