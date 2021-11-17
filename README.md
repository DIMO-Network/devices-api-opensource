# devices-api
Api for managing devices on the DIMO platform.

## Developing locally

1. Make sure a data folder exists under: `./resources/data`
Start Database: `$ docker compose up`
This will start the db on port 5432, if you have conflicting port issue can check with: `$ lsof -i :5432`. 
Data will be persisted across sessions b/c we have the volume set. 
To check container status: `$ docker ps`
You can connect to db eg: psql -h localhost -p 5432 -U dimo

2. Run application
`$ go run ./cmd/devices-api` 

## Migrations

To add a migrations:
`$ goose -dir migrations create <migration_name> sql`

env vars to export:
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=DBSTRING
