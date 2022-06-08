-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create permissions table.

create table permissions
(
    "user"     uuid      not null,
    permission varchar   not null,
    granted_on timestamp not null default now()
);

comment on column permissions."user" is 'The id of the user the permission is granted to.';
comment on column permissions.permission is 'The identifier of the permissions that was granted.';
comment on column permissions.granted_on is 'The timestamp when the permission was granted.';

create index idx_permissions_user_permission on permissions ("user", permission);