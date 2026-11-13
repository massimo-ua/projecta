DROP INDEX IF EXISTS idx_projecta_assets_project_schedule;

ALTER TABLE projecta_assets
    DROP COLUMN IF EXISTS target_completion_date;
