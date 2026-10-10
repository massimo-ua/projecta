-- Down migration: 000033_add_tags_to_assets.down.sql

DROP INDEX IF EXISTS idx_projecta_assets_tags;

ALTER TABLE projecta_assets
    DROP COLUMN IF EXISTS tags;
