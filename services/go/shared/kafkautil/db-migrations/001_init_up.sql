-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create inbox table.

create table __message_inbox
(
    id              bigserial primary key not null,
    topic           varchar               not null,
    partition       int                   not null,
    "offset"        bigint                not null,
    ts              timestamp             not null,
    high_water_mark bigint                not null,
    key             varchar               not null,
    value           text                  not null,
    event_type      varchar               not null,
    header_keys     text[]                not null,
    header_values   text[]                not null,
    status          smallint              not null,
    status_ts       timestamp             not null,
    status_by       uuid                  not null
);

create index __message_inbox_status_ix on __message_inbox (status)
    where status != 200;

-- Create outbox table.

create table __message_outbox
(
    id            bigserial primary key not null,
    topic         varchar               not null,
    created       timestamp             not null,
    key           varchar               not null,
    value         text                  not null,
    event_type    varchar               not null,
    header_keys   text[]                not null,
    header_values text[]                not null,
    status        smallint              not null,
    status_ts     timestamp             not null,
    status_by     uuid                  not null
);

create index __message_outbox_status_ix on __message_outbox (status)
    where status != 200;

-- Create ascending index for id as we will always choose the next oldest one for sending.

create index __message_outbox_id_desc_idx on __message_outbox (id asc)
    where status != 200