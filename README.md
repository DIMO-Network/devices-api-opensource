# devices-api
Api for managing devices on the DIMO platform.

## Developing locally

**TL;DR**
```bash
cp settings.sample.yaml settings.yaml
mkdir ./resources/data
docker compose up -d
go run ./cmd/devices-api migrate
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

3. Migrate DB to latest: `$ go run ./cmd/devices-api migrate`

4. Run application
`$ go run ./cmd/devices-api`

5. Seed data from SmartCar:
`$ go run ./cmd/devices-api seed-smartcar`

### Linting

`brew install golangci-lint`

`golangci-lint run -E prealloc -E revive -E goimports -E deadcode -E errcheck -E gosimple -E govet -E ineffassign -E staticcheck -E structcheck -E typecheck -E unused -E varcheck --timeout=5m`

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
*Make sure you're running the docker image (ie. docker compose up)

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

If we have code base migrations in the migrations folder, we must import `_ "github.com/DIMO-INC/devices-api/migrations"` in the runner so that
it can find the migrations, otherwise get error.

## Mocks

To regenerate a mock, you can use go gen since the files that are mocked have a `//go:generate mockgen ...` at the top. For example:
`nhtsa_api_service.go`

## Helm Requirements

* cf-credentials
  ```sh
    aws secretsmanager create-secret --name infra/cf-credentials/email --description "Cloudflare email" --secret-string "xxx@xxx.xxx"
    aws secretsmanager create-secret --name infra/cf-credentials/token --description "Cloudflare token" --secret-string "XXXXXX"
    ----------------
     kubectl create secret generic cf-credentials --from-literal=email='XXX@XXX.XXX' --from-literal=token='XXX' -n infra
  ```
  
## API

Endpoints as curl commands:
```bash
curl http://localhost:3000/v1/device-definitions/all -w '\n%{time_starttransfer}\n' -v
curl http://localhost:3000/v1/device-definitions/vin/:vin
curl 'http://localhost:3000/v1/device-definitions?make=TESLA&model=MODEL%20Y&year=2021'
curl http://localhost:3000/v1/device-definitions/:id
curl http://localhost:3000/v1/device-definitions/:id/integrations
curl http://localhost:3000/v1/user/devices/me
  -H "Authorization: Bearer {token}"
curl -X POST http://localhost:3000/v1/user/devices
   -H 'Content-Type: application/json'
   -H "Authorization: Bearer {token}"
   -d '{"device_definition_id":"{existing device def id}"}'
```

To prettify json, pipe to json_pp: `| json_pp`

Some test VINs:
5YJYGDEE5MF085533
5YJ3E1EA6MF873863

Higher level env hosts:
https://devices-api.dev.dimo.zone
https://devices-api.dimo.zone

### Generating swagger / openapi spec

```bash
swag init --generalInfo ./cmd/devices-api/main.go --parseInternal true --generatedTime true
```

[declarative_comments_format](https://swaggo.github.io/swaggo.io/declarative_comments_format/)

