CREATE TABLE IF NOT EXISTS person_roles (
    person_id UUID NOT NULL REFERENCES people(person_id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (person_id, role)
);

CREATE INDEX IF NOT EXISTS person_roles_person_id_idx ON person_roles (person_id);

-- Populate existing users with Administrator and User roles
INSERT INTO person_roles (person_id, role)
SELECT person_id, 'Administrator' FROM people
ON CONFLICT DO NOTHING;

INSERT INTO person_roles (person_id, role)
SELECT person_id, 'User' FROM people
ON CONFLICT DO NOTHING;
