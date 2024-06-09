ALTER TABLE projecta_expenses
    ADD COLUMN owner_id UUID NOT NULL REFERENCES people(person_id) ON DELETE CASCADE;
