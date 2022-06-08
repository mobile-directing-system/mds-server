-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create user table.

create table users
(
    username varchar not null,
    is_admin bool    not null,
    pass     varchar not null
);

comment on column users.pass is 'The hashed password of the user.';

-- Create permissions table.

create table permissions
(
    username   uuid      not null,
    permission varchar   not null,
    granted_on timestamp not null default now()
);

comment on column permissions.username is 'The username of the user the permission was granted to.';
comment on column permissions.permission is 'The identifier of the permissions that was granted.';
comment on column permissions.granted_on is 'The timestamp when the permission was granted.';