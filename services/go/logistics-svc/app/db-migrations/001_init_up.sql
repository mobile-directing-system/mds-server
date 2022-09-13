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

create index operation_members_user_ix on operation_members ("user");

-- Create entries table.

create table address_book_entries
(
    id          uuid primary key not null default uuid_generate_v4(),
    label       varchar          not null,
    description varchar          not null,
    operation   uuid references operations (id)
        on delete restrict on update restrict,
    "user"      uuid references users (id)
        on delete restrict on update restrict
);

-- Create groups table.

create table groups
(
    id          uuid primary key not null,
    title       varchar          not null,
    description varchar          not null,
    operation   uuid
);

create table group_members
(
    "group" uuid not null references groups (id)
        on delete cascade on update cascade,
    "user"  uuid not null
);

create index group_members_user_ix on group_members ("user");

-- Create channels table.

create table channels
(
    id             uuid primary key not null default uuid_generate_v4(),
    entry          uuid             not null references address_book_entries (id)
        on delete restrict on update restrict,
    label          varchar          not null,
    "type"         varchar          not null,
    priority       int              not null,
    min_importance numeric          not null,
    timeout        bigint           not null
);

-- Create direct table.

create table direct_channels
(
    channel uuid primary key not null references channels (id)
        on delete restrict on update restrict,
    "info"  text             not null
);

-- Create email addresses table.

create table email_channels
(
    channel uuid primary key not null references channels (id)
        on delete restrict on update restrict,
    email   varchar          not null
);

-- Create phone numbers table.

create table phone_call_channels
(
    channel uuid primary key not null references channels (id)
        on delete restrict on update restrict,
    phone   varchar          not null
);

-- Create radio table.

create table radio_channels
(
    channel uuid primary key not null references channels (id)
        on delete restrict on update restrict,
    "info"  text             not null
);

-- Create in-app-notification table.

create table in_app_notification_channels
(
    channel uuid primary key not null references channels (id)
        on delete restrict on update restrict
);

-- Create forward-to-user table.

create table forward_to_user_channel_entries
(
    channel         uuid not null references channels (id)
        on delete restrict on update restrict,
    forward_to_user uuid not null references users (id)
        on delete restrict on update restrict
);

-- Create forward-to-group table.

create table forward_to_group_channel_entries
(
    channel          uuid not null references channels (id)
        on delete restrict on update restrict,
    forward_to_group uuid not null references groups (id)
        on delete restrict on update restrict
);

-- Create intel table.

create table intel
(
    id          uuid primary key not null,
    created_at  timestamp        not null,
    created_by  uuid             not null references users (id)
        on delete restrict on update restrict,
    operation   uuid             not null references operations (id)
        on delete restrict on update restrict,
    "type"      varchar          not null,
    "content"   jsonb            not null,
    importance  int              not null,
    search_text varchar,
    is_valid    bool             not null
);

-- Create assignments table.

create table intel_assignments
(
    id    uuid primary key not null default uuid_generate_v4(),
    intel uuid             not null references intel (id)
        on delete restrict on update restrict,
    "to"  uuid             not null -- No ref for possible deletion.
);

create index intel_assignments_to_ix on intel_assignments ("to");

-- Create deliveries table.

create table intel_deliveries
(
    id           uuid primary key not null default uuid_generate_v4(),
    "assignment" uuid             not null references intel_assignments (id)
        on delete restrict on update restrict,
    is_active    bool             not null,
    success      bool             not null,
    note         varchar
);

create index intel_deliveries_active_ix on intel_deliveries (is_active)
    where is_active = true;

create index intel_deliveries_failed_ix on intel_deliveries (is_active, success)
    where is_active = false and success = false;

-- Create delivery attempts table.

create table intel_delivery_attempts
(
    id         uuid primary key not null default uuid_generate_v4(),
    delivery   uuid             not null references intel_deliveries (id)
        on delete restrict on update restrict,
    channel    uuid             not null references channels (id)
        on delete restrict on update restrict,
    created_at timestamp        not null,
    is_active  bool             not null,
    status     varchar          not null,
    status_ts  timestamp        not null,
    note       varchar
);

create index intel_delivery_attempts_active_ix on intel_delivery_attempts (is_active)
    where is_active = true;