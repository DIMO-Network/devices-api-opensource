-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

CREATE TABLE autopi_units
(
    unit_id         char(50) PRIMARY KEY,
    device_id char(50),
    user_id          text not null,
    nft_address        varchar(50) not null,
    created_at   timestamptz not null default current_timestamp,
    updated_at   timestamptz not null default current_timestamp
);
CREATE UNIQUE INDEX autopi_units_nft_address_idx ON autopi_units (nft_address);
CREATE UNIQUE INDEX autopi_units_device_id_idx ON autopi_units (device_id);

ALTER TABLE user_device_api_integrations
ADD COLUMN unit_id char(50);

ALTER TABLE user_device_api_integrations ADD CONSTRAINT user_device_api_integrations_autopi_units FOREIGN KEY (unit_id) REFERENCES autopi_units(unit_id);

ALTER TABLE autopi_jobs
ADD COLUMN unit_id char(50);

ALTER TABLE autopi_jobs ADD CONSTRAINT autopi_jobs_autopi_units FOREIGN KEY (unit_id) REFERENCES autopi_units(unit_id);

INSERT INTO autopi_units (unit_id, device_id, user_id, nft_address)
SELECT
  json_extract_path_text(ud.metadata::json,'autoPiUnitId') unit_id,
  ud.id,
  ud.user_id,
  '' nft_address
FROM
    user_device_api_integrations udai
    INNER JOIN integrations i ON ( udai.integration_id = i.id ) 
    INNER JOIN user_devices ud ON ( udai.user_device_id = ud.id )
WHERE
    i.vendor = 'AutoPI';

UPDATE
    user_device_api_integrations udai
SET
    unit_id = a.unit_id
FROM
    autopi_units a
WHERE
    a.device_id = udai.user_device_id;

UPDATE
    autopi_jobs aj
SET
    unit_id = a.unit_id
FROM
    autopi_units a
WHERE
    a.device_id = aj.user_device_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE autopi_jobs drop column unit_id;
ALTER TABLE user_device_api_integrations DROP COLUMN unit_id;

DROP TABLE autopi_units;

-- +goose StatementEnd
