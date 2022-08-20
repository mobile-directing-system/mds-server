-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create actions table.

create table __search_actions
(
    id              bigserial primary key not null,
    created         timestamp             not null,
    index           varchar               not null,
    document_id     varchar,
    action_type     varchar               not null,
    options         jsonb                 not null,
    remaining_tries int                   not null,
    retry_cooldown  bigint                not null,
    processing_ts   timestamp             not null,
    err_message     varchar
);

comment on column __search_actions.id is 'The id of the entry. Also used for sorting.';
comment on column __search_actions.created is 'The timestamp for the action. Only for human-readability.';
comment on column __search_actions.index is 'The index name in search.';
comment on column __search_actions.document_id is 'The id of the document. If set, this allows concurrent processing.';
comment on column __search_actions.action_type is 'The type of action to perform.';
comment on column __search_actions.options is 'Additional options for the action.';
comment on column __search_actions.remaining_tries is 'Remaining tries to perform. Used for indicating open actions.';
comment on column __search_actions.retry_cooldown is 'Cooldown in nanoseconds when processing fails.';
comment on column __search_actions.processing_ts
    is 'Timestamp of the last operation performed on this action. On creation, must be the same as creation_ts.';
comment on column __search_actions.err_message is 'Optional error message in case of error. Otherwise, NULL.';

create index __search_actions_status_ix on __search_actions (document_id)
    where remaining_tries > 0;

-- Create ascending index for id as we will always choose the next oldest one for sending.

create index __search_actions_id_desc_idx on __search_actions (id asc)
    where remaining_tries > 0;