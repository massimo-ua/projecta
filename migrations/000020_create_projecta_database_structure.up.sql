CREATE TABLE IF NOT EXISTS projecta_projects
(
    project_id      UUID        PRIMARY KEY NOT NULL,
    name            TEXT        NOT NULL,
    description     TEXT,
    started_at      TIMESTAMP,
    ended_at        TIMESTAMP,
    owner_id        UUID        NOT NULL,
    created_at      TIMESTAMP   DEFAULT current_timestamp,
    updated_at      TIMESTAMP,
    CONSTRAINT projecta_projects_owner_id_fk FOREIGN KEY (owner_id) REFERENCES people(person_id) ON DELETE CASCADE
);

CREATE TRIGGER update_timestamp_trigger
    BEFORE UPDATE
    ON projecta_projects
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_trigger_function('updated_at');

CREATE TABLE IF NOT EXISTS projecta_cost_categories
(
    category_id     UUID        PRIMARY KEY NOT NULL,
    project_id      UUID        NOT NULL,
    name            TEXT        NOT NULL,
    description     TEXT,
    created_at      TIMESTAMP   DEFAULT current_timestamp,
    updated_at      TIMESTAMP,
    CONSTRAINT projecta_cost_categories_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE
);

CREATE TRIGGER update_timestamp_trigger
    BEFORE UPDATE
    ON projecta_cost_categories
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_trigger_function('updated_at');

CREATE TABLE IF NOT EXISTS projecta_cost_types
(
    type_id         UUID        PRIMARY KEY NOT NULL,
    project_id      UUID        NOT NULL,
    name            TEXT        NOT NULL,
    description     TEXT,
    created_at      TIMESTAMP   DEFAULT current_timestamp,
    updated_at      TIMESTAMP,
    CONSTRAINT projecta_cost_types_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE
);

CREATE TRIGGER update_timestamp_trigger
    BEFORE UPDATE
    ON projecta_cost_types
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_trigger_function('updated_at');

CREATE TABLE IF NOT EXISTS projecta_expenses
(
    expense_id          UUID        PRIMARY KEY NOT NULL,
    project_id          UUID        NOT NULL,
    category_id         UUID        NOT NULL,
    type_id             UUID        NOT NULL,
    amount              BIGINT      NOT NULL,
    currency            CHAR(3)     NOT NULL,
    description         TEXT,
    expense_date        TIMESTAMP,
    created_at          TIMESTAMP   DEFAULT current_timestamp,
    updated_at          TIMESTAMP,
    CONSTRAINT projecta_expenses_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE,
    CONSTRAINT projecta_expenses_category_id_fk FOREIGN KEY (category_id) REFERENCES projecta_cost_categories(category_id) ON DELETE CASCADE,
    CONSTRAINT projecta_expenses_type_id_fk FOREIGN KEY (type_id) REFERENCES projecta_cost_types(type_id) ON DELETE CASCADE
);

CREATE TRIGGER update_timestamp_trigger
    BEFORE UPDATE
    ON projecta_expenses
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp_trigger_function('updated_at');

CREATE TABLE IF NOT EXISTS projecta_currencies
(
    code            CHAR(3)     PRIMARY KEY NOT NULL,
    project_id      UUID        NOT NULL,
    exchange_rate   NUMERIC     NOT NULL,
    CONSTRAINT projecta_currencies_project_id_fk FOREIGN KEY (project_id) REFERENCES projecta_projects(project_id) ON DELETE CASCADE
);
