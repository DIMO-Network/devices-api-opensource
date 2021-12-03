-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE devices_api.device_definitions
(
    uuid         uuid PRIMARY KEY,
    vin_first_10 varchar(10) not null, -- aka squishVin
    make         varchar(100) not null,
    model        varchar(100) not null,
    year         smallint    not null,
    sub_model    varchar(100),
    metadata     jsonb,
    created_at   timestamptz not null default current_timestamp,
    updated_at   timestamptz not null default current_timestamp
);
CREATE UNIQUE INDEX vin_idx ON devices_api.device_definitions (vin_first_10);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE devices_api.device_definitions
-- +goose StatementEnd
