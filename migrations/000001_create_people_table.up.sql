CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION update_timestamp_trigger_function()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW := json_populate_record(NEW, json_build_object(TG_ARGV[0], now()));
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

create table public.people
(
    person_id    uuid      default uuid_generate_v4() not null
        primary key,
    first_name   varchar                              not null,
    last_name    varchar                              not null,
    display_name varchar,
    created_at   timestamp default CURRENT_TIMESTAMP,
    updated_at   timestamp,
    deleted_at   timestamp
);

create trigger update_timestamp_trigger
    before update
    on public.people
    for each row
    execute procedure public.update_timestamp_trigger_function('updated_at');


