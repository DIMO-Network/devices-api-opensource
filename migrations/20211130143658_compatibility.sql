-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

CREATE TYPE integration_type AS ENUM (
    'Hardware',
    'API'
    );
CREATE TYPE integration_style AS ENUM (
    'Addon',
    'OEM'
    );

CREATE TABLE integrations
(
    uuid       uuid PRIMARY KEY,
    type       integration_type  not null,
    style      integration_style not null,
    vendors    varchar(50)       not null,

    created_at timestamptz       not null default current_timestamp,
    updated_at timestamptz       not null default current_timestamp
);

CREATE TABLE device_integrations
(
    device_definition_uuid uuid        not null,
    integration_uuid       uuid        not null,

    created_at             timestamptz not null default current_timestamp,
    updated_at             timestamptz not null default current_timestamp,

    PRIMARY KEY (device_definition_uuid, integration_uuid),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_uuid) REFERENCES device_definitions (uuid),
    CONSTRAINT fk_integration FOREIGN KEY (integration_uuid) REFERENCES integrations (uuid)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE integrations;
DROP TABLE device_integrations
DROP TYPE integration_style;
DROP TYPE integration_type;
-- +goose StatementEnd
