-- Create user table.

create table users
(
    id        uuid primary key not null,
    username  varchar          not null,
    is_admin  bool             not null,
    pass      varchar          not null,
    is_active bool             not null
);

comment on column users.pass is 'The hashed password of the user.';

-- Create tokens table.

create table session_tokens
(
    "user"     uuid      not null
        constraint session_tokens_users_id_fk
            references users
            on update cascade on delete restrict,
    token      varchar   not null,
    created_ts timestamp not null
);

create index session_tokens_token_ix on session_tokens (token);

-- Create permissions table.

create table permissions
(
    "user"  uuid    not null,
    name    varchar not null,
    options jsonb
);

comment on column permissions."user" is 'The id of the user the permission was granted to.';
comment on column permissions.name is 'The identifier of the permissions that was granted.';
comment on column permissions.options is 'Additional options for the permission.';

create index permissions_user_ix on permissions ("user");