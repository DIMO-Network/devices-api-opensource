-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
alter table autopi_units alter column unit_id type char(36) using unit_id::char(36);
alter table autopi_units alter column device_id type varchar(50) using device_id::varchar(50);
ALTER TABLE user_device_api_integrations alter column unit_id type char(36) using unit_id::char(36);
ALTER TABLE autopi_jobs alter column unit_id type char(36) using unit_id::char(36);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
alter table autopi_units alter column unit_id type char(50) using unit_id::char(50);
alter table autopi_units alter column device_id type char(50) using device_id::char(50);
alter table user_device_api_integrations alter column unit_id type char(50) using unit_id::char(50);
alter table autopi_jobs alter column unit_id type char(50) using unit_id::char(50);
-- +goose StatementEnd
