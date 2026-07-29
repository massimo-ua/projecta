ALTER TABLE projecta_projects ADD COLUMN IF NOT EXISTS share_token UUID UNIQUE DEFAULT gen_random_uuid();
UPDATE projecta_projects SET share_token = gen_random_uuid() WHERE share_token IS NULL;
ALTER TABLE projecta_projects ALTER COLUMN share_token SET NOT NULL;

CREATE TABLE IF NOT EXISTS projecta_project_shares
(
    share_id    UUID        PRIMARY KEY NOT NULL,
    project_id  UUID        NOT NULL,
    person_id   UUID        NOT NULL,
    created_at  TIMESTAMP   DEFAULT current_timestamp,
    CONSTRAINT projecta_project_shares_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE,
    CONSTRAINT projecta_project_shares_person_id_fk FOREIGN KEY (person_id) REFERENCES people(person_id) ON DELETE CASCADE,
    CONSTRAINT projecta_project_shares_unique UNIQUE (project_id, person_id)
);
