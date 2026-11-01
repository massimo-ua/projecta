-- Up migration: 000036_add_currency_to_project_tag_budgets.up.sql

ALTER TABLE projecta_project_tag_budgets
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'UAH';

-- Backfill existing tag budgets from parent project's main_currency
UPDATE projecta_project_tag_budgets tb
SET currency = COALESCE(p.main_currency, 'UAH')
FROM projecta_projects p
WHERE tb.project_id = p.project_id;

ALTER TABLE projecta_project_tag_budgets
    DROP CONSTRAINT IF EXISTS chk_tag_budget_currency;

ALTER TABLE projecta_project_tag_budgets
    ADD CONSTRAINT chk_tag_budget_currency
    CHECK (currency IN ('UAH', 'USD', 'EUR', 'PLN'));
