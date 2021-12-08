-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

CREATE TABLE user_devices
(
    id char(27) not null,
    user_id text not null,
    device_definition_id char(27)    not null,

    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,

    PRIMARY KEY (id),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id) REFERENCES device_definitions (id)
);

alter table device_definitions add column source text; -- where the information came from
alter table device_definitions add column verified boolean not null DEFAULT true; -- whether this info has been verified and should show up in our endpoints outward

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table user_devices;
-- +goose StatementEnd
