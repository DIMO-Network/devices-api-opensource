-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE devices_api.device_definitions
(
    uuid         uuid DEFAULT public.uuid_generate_v4(),
    make         varchar(100),
    model        varchar(100),
    year         varchar(4),
    vin_first_10 varchar(10),
    other_data   jsonb
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE devices_api.device_definitions
-- +goose StatementEnd
