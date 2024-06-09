create type credential_provider as enum ('GOOGLE', 'FACEBOOK', 'LOCAL');
create table public.credentials
(
    provider        credential_provider not null,
    person_id       uuid                not null
        references public.people
            on delete cascade,
    identity        text                not null,
    registration_id text                not null,
    primary key (provider, person_id)
);

create index credentials_person_id_idx
    on public.credentials (person_id);

