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
    "user"    uuid not null -- No ref because of async events.
);

-- Create entries table.

create table address_book_entries
(
    id          uuid primary key not null default uuid_generate_v4(),
    label       varchar          not null,
    description varchar          not null,
    operation   uuid, -- No ref because of async events.
    "user"      uuid  -- No ref because of async events.
);

-- Create intel table.

create table intel
(
    id          uuid primary key not null default uuid_generate_v4(),
    created_at  timestamp        not null,
    created_by  uuid             not null references users (id)
        on delete restrict on update cascade,
    operation   uuid             not null references operations (id)
        on delete restrict on update cascade,
    "type"      varchar          not null,
    content     jsonb            not null,
    search_text varchar,
    importance  int              not null,
    is_valid    bool             not null
);

-- Create assignments table.

create table intel_assignments
(
    id    uuid primary key not null default uuid_generate_v4(),
    intel uuid             not null references intel (id)
        on delete restrict on update restrict,
    "to"  uuid             not null -- No ref for possible deletion.
)