ALTER TABLE projecta_expenses
    ADD COLUMN compensatory_id uuid REFERENCES projecta_expenses(expense_id);

-- create compensatory_id index
CREATE INDEX projecta_expenses_compensatory_id_idx ON projecta_expenses(compensatory_id);
