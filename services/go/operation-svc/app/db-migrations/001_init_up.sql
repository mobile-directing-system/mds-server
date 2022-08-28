-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create users table.

create table users
(
    id         uuid primary key not null,
    username   varchar          not null,
    first_name varchar          not null,
    last_name  varchar          not null,
    is_active  bool             not null
);

-- Create operations table.

create table operations
(
    id          uuid primary key not null default uuid_generate_v4(),
    title       varchar          not null,
    description text             not null,
    start_ts    timestamp        not null,
    end_ts      timestamp,
    is_archived bool             not null
);

create table operation_members
(
    operation uuid not null references operations (id)
        on delete restrict on update restrict,
    "user"    uuid not null references users (id)
        on delete restrict on update restrict
);

create index idx_operation_members_operation
    on operation_members (operation);