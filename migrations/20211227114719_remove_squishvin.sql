-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
ALTER TABLE device_definitions DROP COLUMN vin_first_10;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
ALTER TABLE device_definitions ADD COLUMN vin_first_10 varchar(10);
-- +goose StatementEnd
