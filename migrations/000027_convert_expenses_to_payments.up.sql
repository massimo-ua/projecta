ALTER TABLE projecta_expenses
    RENAME TO projecta_payments;

ALTER TABLE projecta_payments
    RENAME COLUMN expense_id TO payment_id;

ALTER TABLE projecta_payments
    RENAME COLUMN expense_date TO payment_date;

alter table projecta_assets add column owner_id uuid;
alter table projecta_assets add constraint projecta_assets_owner_id_fk foreign key (owner_id) references people(person_id) on delete cascade;
-- copy all entries from projecta_payments where compensatory_id is not null to projecta_assets
INSERT INTO projecta_assets (asset_id, name, description, project_id, type_id, price, currency, acquired_at, owner_id, created_at, updated_at)
SELECT payment_id, description, description, project_id, type_id, amount, currency, payment_date, owner_id, created_at, updated_at
FROM projecta_payments
WHERE compensatory_id IS NOT NULL;
-- remove all entries from projecta_payments where compensatory_id is not null or amount is less than 0
DELETE FROM projecta_payments
WHERE compensatory_id IS NOT NULL OR amount < 0;
-- remove compensatory_id column from projecta_payments
ALTER TABLE projecta_payments
    DROP COLUMN compensatory_id;
