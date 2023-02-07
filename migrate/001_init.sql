-- Write your migrate up statements here

create table if not exists users
(
    id    uuid default gen_random_uuid() not null
        constraint users_pk unique primary key,
    login varchar                        not null unique,
    password varchar                        not null
);

create table if not exists vaults
(
    id      uuid default gen_random_uuid() not null
        constraint vaults_pk unique primary key,
    user_id uuid
        constraint vaults_users_fk references users (id),
    vault bytea,
    version integer default 0 not null,
    is_deleted bool default false not null,
    s3 varchar
);

---- create above / drop below ----

drop table vaults;

drop table users;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
