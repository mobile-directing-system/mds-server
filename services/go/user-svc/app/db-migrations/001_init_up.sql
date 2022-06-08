-- Activate UUID support.
create extension if not exists "uuid-ossp";

-- Create user table.

create table users
(
    id         uuid primary key not null default uuid_generate_v4(),
    username   varchar          not null,
    first_name varchar          not null,
    last_name  varchar          not null,
    is_admin   boolean          not null,
    pass       varchar          not null
);

comment on column users.username is 'The username for logging in.';
comment on column users.first_name is 'The first name of the user.';
comment on column users.last_name is 'The last name of the user.';
comment on column users.pass is 'The hashed password of the user.';
