CREATE TABLE IF NOT EXISTS projecta_assets
(
    asset_id            UUID        PRIMARY KEY NOT NULL,
    name                TEXT        NOT NULL,
    description         TEXT,
    project_id          UUID        NOT NULL,
    type_id             UUID        NOT NULL,
    price               BIGINT      NOT NULL,
    currency            CHAR(3)     NOT NULL,
    acquired_at         TIMESTAMP   NOT NULL DEFAULT current_timestamp,
    created_at          TIMESTAMP   NOT NULL DEFAULT current_timestamp,
    updated_at          TIMESTAMP,
    CONSTRAINT projecta_assets_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE,
    CONSTRAINT projecta_assets_type_id_fk FOREIGN KEY (type_id) REFERENCES projecta_cost_types(type_id) ON DELETE CASCADE
);

CREATE TRIGGER update_timestamp_trigger
    BEFORE UPDATE
    ON projecta_assets
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_trigger_function('updated_at');
