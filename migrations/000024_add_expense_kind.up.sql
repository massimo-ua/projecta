CREATE TYPE expense_kind AS ENUM (
    'DOWN_PAYMENT',
    'UPON_COMPLETION',
    'CREDIT_PAYMENT'
);

ALTER TABLE projecta_expenses
    ADD COLUMN kind expense_kind NOT NULL DEFAULT 'UPON_COMPLETION';
