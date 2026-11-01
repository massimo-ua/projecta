-- Down migration: 000036_add_currency_to_project_tag_budgets.down.sql

ALTER TABLE projecta_project_tag_budgets
    DROP CONSTRAINT IF EXISTS chk_tag_budget_currency;

ALTER TABLE projecta_project_tag_budgets
    DROP COLUMN IF EXISTS currency;
