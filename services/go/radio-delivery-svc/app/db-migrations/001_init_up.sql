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

-- Create operation members table.

create table operation_members
(
    operation uuid not null,
    "user"    uuid not null -- No ref, because of async events.
);

create index operation_members_operation_ix on operation_members (operation);
create index operation_members_user_ix on operation_members ("user");

-- Create channels table.

create table radio_channels
(
    id      uuid primary key not null,
    entry   uuid             not null,
    "label" varchar          not null,
    timeout bigint           not null,
    "info"  text             not null
);

comment on column radio_channels.entry is 'The id of the address book entry.';

create index radio_channels_entry_ix on radio_channels (entry);

-- Create delivery attempts table.

create table accepted_intel_delivery_attempts
(
    id                uuid primary key not null,
    intel             uuid             not null,
    intel_operation   uuid             not null,
    intel_importance  int              not null,
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

comment on column accepted_intel_delivery_attempts.intel_importance is 'The importance from the referenced intel.';
comment on column accepted_intel_delivery_attempts.assigned_to is 'The id of the user, the intel is assigned to (for this delivery).';
comment on column accepted_intel_delivery_attempts.delivery is 'The id of the referenced delivery.';
comment on column accepted_intel_delivery_attempts.channel is 'The id of the channel to use for delivery.';
comment on column accepted_intel_delivery_attempts.accepted_at is 'The timestamp when the attempt was accepted by the radio delivery service.';

create index accepted_intel_delivery_attempts_active_ix on accepted_intel_delivery_attempts (is_active)
    where is_active = true or is_active is true;

create index accepted_intel_delivery_attempts_assigned_to_ix on accepted_intel_delivery_attempts (assigned_to);

-- Create radio deliveries table.

create table radio_deliveries
(
    attempt      uuid primary key not null references accepted_intel_delivery_attempts
        on update restrict on delete restrict,
    picked_up_by uuid,
    picked_up_at timestamp,
    success      boolean,
    success_ts   timestamp        not null,
    note         varchar          not null
);

comment on column radio_deliveries.picked_up_by is 'The id of the user that picked up this delivery.';
comment on column radio_deliveries.picked_up_at is 'The timestamp when the radio-delivery was picked up.';
comment on column radio_deliveries.success is 'Indicates the success status of the radio-delivery. If NULL, the delivery is still active and awaiting results or waiting to be picked up. If true, transmission over radio was successful. Otherwise, transmission failed or was canceled due to the attempt being canceled.';
comment on column radio_deliveries.success_ts is 'The timestamp of the most recent update of success.';
comment on column radio_deliveries.note is 'Information regarding the delivery and/or success-state.';

create index radio_deliveries_active_ix on radio_deliveries (attempt)
    where success is null;
create index radio_deliveries_open_for_pickup_ix on radio_deliveries (attempt)
    where success is null and picked_up_by is null;