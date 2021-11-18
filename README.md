# devices-api
Api for managing devices on the DIMO platform.

## Developing locally

1. Make sure a data folder exists under: `mkdir ./resources/data`
Start Database: `$ docker compose up -d`
This will start the db on port 5432, if you have conflicting port issue can check with: `$ lsof -i :5432`. 
Data will be persisted across sessions b/c we have the volume set. 
To check container status: `$ docker ps`
You can connect to db eg: `psql -h localhost -p 5432 -U dimo` or with your favorite db IDE

2. Migrate DB to latest: `$ go run ./cmd/db_migrate`

3. Run application
`$ go run ./cmd/devices-api` 

## Migrations

To install goose CLI:
```bash
$ go get -u github.com/pressly/goose/v3/cmd/goose
export GOOSE_DRIVER=postgres
```

Add a migrations:
`$ goose -dir migrations create <migration_name> sql`

Migrate DB to latest:
`$ go run ./cmd/db_migrate`

Clear DB to start over:
```bash
$ docker ps
$ docker stop <container_id>
$ rm -R ./resources/data/
```

When running migrations from CD, we will want to set the following env vars:
- SERVICE_ACCOUNT_PASSWORD: This will be the password the `service` account will use to connect to, which is the account the application should connect with in HL envs.
- 
