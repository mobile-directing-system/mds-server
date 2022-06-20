-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create users table.

create table users
(
    id uuid primary key not null
);

-- Create permissions table.

create table permissions
(
    "user"  uuid    not null
        constraint permissions_users_fk
            references users
            on update restrict
            on delete restrict,
    name    varchar not null,
    options jsonb
);

comment on column permissions."user" is 'The id of the user the permission is granted to.';
comment on column permissions.name is 'The identifier of the permissions that was granted.';
comment on column permissions.options is 'Additional options for the permission.';

create index idx_permissions_user_permission on permissions ("user", name);