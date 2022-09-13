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

-- Create channels table.

create table notification_channels
(
    id      uuid primary key not null,
    entry   uuid             not null,
    label   varchar          not null,
    timeout bigint           not null
);

comment on column notification_channels.entry is 'The id of the address book entry.';

-- Create delivery attempts table.

create table accepted_intel_delivery_attempts
(
    id                uuid primary key not null,
    assigned_to       uuid             not null,
    assigned_to_label varchar          not null,
    assigned_to_user  uuid,
    delivery          uuid             not null,
    channel           uuid             not null,
    created_at        timestamp        not null,
    is_active         bool             not null,
    status_ts         timestamp        not null,
    note              varchar,
    accepted_at       timestamp        not null
);

comment on column accepted_intel_delivery_attempts.assigned_to is 'The id of the user, the intel is assigned to (for this delivery).';
comment on column accepted_intel_delivery_attempts.delivery is 'The id of the referenced delivery.';
comment on column accepted_intel_delivery_attempts.channel is 'The id of the channel to use for delivery.';

create index accepted_intel_delivery_attempts_active_ix on accepted_intel_delivery_attempts (is_active)
    where is_active = true;

create index accepted_intel_delivery_attempts_assigned_to_ix on accepted_intel_delivery_attempts (assigned_to);


create table intel_notification_history
(
    attempt uuid      not null references accepted_intel_delivery_attempts (id),
    ts      timestamp not null
);

comment on table intel_notification_history is 'History log for performed notifications for attempts. Used in order to identify attempts, that still require notification.';

-- Create intel table.

create table intel_to_deliver
(
    attempt    uuid      not null primary key references accepted_intel_delivery_attempts (id)
        on update cascade on delete cascade,
    id         uuid      not null,
    created_at timestamp not null,
    created_by uuid      not null references users (id)
        on delete restrict on update restrict,
    operation  uuid      not null,
    "type"     varchar   not null,
    "content"  jsonb     not null,
    importance int       not null
);
