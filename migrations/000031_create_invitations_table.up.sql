CREATE TABLE IF NOT EXISTS invitations (
    invitation_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    code_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_by UUID NOT NULL REFERENCES people(person_id) ON DELETE CASCADE,
    person_id UUID NULL REFERENCES people(person_id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS invitations_email_idx ON invitations (email);
CREATE INDEX IF NOT EXISTS invitations_created_by_idx ON invitations (created_by);
CREATE INDEX IF NOT EXISTS invitations_code_hash_idx ON invitations (code_hash);
CREATE INDEX IF NOT EXISTS invitations_person_id_idx ON invitations (person_id);
