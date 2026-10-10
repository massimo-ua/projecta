-- Up migration: 000033_add_tags_to_assets.up.sql

ALTER TABLE projecta_assets
    ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_projecta_assets_tags ON projecta_assets USING GIN (tags);
