-- Create user table.

create table users
(
    id       uuid primary key not null,
    username varchar          not null,
    is_admin bool             not null,
    pass     varchar          not null
);

comment on column users.pass is 'The hashed password of the user.';

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