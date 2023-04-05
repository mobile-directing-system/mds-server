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
    id          uuid primary key not null,
    title       varchar          not null,
    description text             not null,
    start_ts    timestamp        not null,
    end_ts      timestamp,
    is_archived bool             not null
);

create table operation_members
(
    operation uuid not null references operations (id)
        on update cascade on delete cascade,
    "user"    uuid not null -- No ref, because of async events.
);

create index operation_members_operation_ix on operation_members (operation);
create index operation_members_user_ix on operation_members ("user");

-- Create intel table.

create table intel
(
    id         uuid primary key not null,
    created_at timestamp        not null,
    created_by uuid             not null,
    operation  uuid             not null,
    importance int              not null,
    is_valid   bool             not null
);

create index intel_operation_ix on intel (operation);
create index intel_created_at_ix on intel (created_at desc);

-- Create deliveries table.

create table active_intel_deliveries
(
    id    uuid primary key not null,
    intel uuid             not null,
    "to"  uuid             not null,
    note  varchar
);

comment on column active_intel_deliveries."to" is 'The referenced address book entry.';

-- Create delivery attempts table.

create table active_intel_delivery_attempts
(
    id       uuid primary key not null,
    delivery uuid             not null
);

create index active_intel_delivery_attempts_delivery_ix on active_intel_delivery_attempts (delivery);

-- Create table for auto-delivery.

create table auto_intel_delivery_address_book_entries
(
    entry uuid primary key not null
);

comment on table auto_intel_delivery_address_book_entries is 'Address book entries for which auto-intel-delivery is enabled and that are therefore ignored.';
