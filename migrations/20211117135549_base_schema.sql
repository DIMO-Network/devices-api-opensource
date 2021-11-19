-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- uuid gen
REVOKE CREATE ON schema public FROM public; -- public schema isolation
CREATE SCHEMA devices_api;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE devices_api.migrations;
DROP SCHEMA devices_api CASCADE;
GRANT CREATE, USAGE ON schema public TO public;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
