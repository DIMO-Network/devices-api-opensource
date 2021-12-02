-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

ALTER TABLE device_integrations ADD COLUMN capabilities jsonb;

ALTER TABLE device_definitions AlTER COLUMN vin_first_10 DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
ALTER TABLE device_integrations DROP COLUMN capabilities;
ALTER TABLE device_definitions AlTER COLUMN vin_first_10 SET NOT NULL;
-- +goose StatementEnd
