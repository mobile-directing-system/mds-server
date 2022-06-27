-- Activate UUID support.
create extension if not exists "uuid-ossp";

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