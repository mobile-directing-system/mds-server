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
    "user"     uuid      not null,
    permission varchar   not null,
    granted_on timestamp not null
);

comment on column permissions."user" is 'The id of the user the permission was granted to.';
comment on column permissions.permission is 'The identifier of the permissions that was granted.';
comment on column permissions.granted_on is 'The timestamp when the permission was granted.';