-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create operations table.

create table operations
(
    id uuid primary key not null
);

create table operation_members
(
    operation uuid not null references operations (id)
        on delete cascade on update restrict,
    "user"    uuid not null
);

-- Create users table.

create table users
(
    id uuid primary key not null
);

-- Create group table.

create table groups
(
    id          uuid primary key not null default uuid_generate_v4(),
    title       varchar          not null,
    description varchar          not null,
    operation   uuid references operations (id)
        on delete restrict on update restrict
);

-- Create members table.

create table members
(
    "group"      uuid      not null references groups (id)
        on delete cascade on update cascade,
    "user"       uuid      not null references users (id)
        on delete restrict on update restrict,
    member_since timestamp not null,
    PRIMARY KEY ("group", "user")
);