# devices-api
Api for managing devices on the DIMO platform.

## Developing locally

**TL;DR**
```bash
cp settings.sample.yaml settings.yaml
mkdir ./resources/data
docker compose up -d
go run ./cmd/db_migrate
go run ./cmd/devices-api
```

1. Copy settings: `$ cp settings.sample.yaml settings.yaml`
Adjust secrets or settings as necessary. The sample file should have what you need with correct defaults for local dev.

2. Make sure a data folder exists under: `$ mkdir ./resources/data`
Start Database: `$ docker compose up -d`
This will start the db on port 5432, if you have conflicting port issue can check with: `$ lsof -i :5432`. 
Data will be persisted across sessions b/c we have the volume set. 
To check container status: `$ docker ps`
You can connect to db eg: `psql -h localhost -p 5432 -U dimo` or with your favorite db IDE

3. Migrate DB to latest: `$ go run ./cmd/db_migrate`

4. Run application
`$ go run ./cmd/devices-api`

### Database ORM

This is using [sqlboiler](https://github.com/volatiletech/sqlboiler). The ORM models are code generated. If the db changes,
you must update the models.

Make sure you have sqlboiler installed:
```bash
go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
```

To generate the models:
```bash
sqlboiler psql --no-tests
```

## Migrations

To install goose CLI:
```bash
$ go install github.com/pressly/goose/v3/cmd/goose
export GOOSE_DRIVER=postgres
```

Add a migrations:
`$ goose -dir migrations create <migration_name> sql`

Migrate DB to latest:
`$ go run ./cmd/db_migrate`

Clear DB to start over:
```bash
docker ps
docker stop <container_id>
rm -R ./resources/data/ && mkdir ./resources/data/ 
docker compose up -d
```

When running migrations from CD, we will want to set the following env vars:
- SERVICE_ACCOUNT_PASSWORD: This will be the password the `service` account will use to connect to, which is the account the application should connect with in HL envs.
- 

## Helm Requirements

* cf-credentials
  ```sh
    aws secretsmanager create-secret --name infra/cf-credentials/email --description "Cloudflare email" --secret-string "xxx@xxx.xxx"
    aws secretsmanager create-secret --name infra/cf-credentials/token --description "Cloudflare token" --secret-string "XXXXXX"
    ----------------
     kubectl create secret generic cf-credentials --from-literal=email='XXX@XXX.XXX' --from-literal=token='XXX' -n infra
  ```