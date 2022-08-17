# devices-api
Api for managing devices on the DIMO platform.

## Table of contents

- [Developing locally](#developing-locally)
  - [Kafka test producer](#kafka-test-producer)
  - [Linting](#linting)
  - [Database ORM](#database-orm)
- [Migrations](#migrations)
  - [Managing migrations from k8s](#managing-migrations-from-k8s)
- [Mocks](#mocks)
- [Helm requirements](#helm-requirements)
- [API](#api)
  - [Generating Swagger / OpenAPI spec](#generating-swagger--openapi-spec)


## Developing locally

**TL;DR**
```bash
cp settings.sample.yaml settings.yaml
docker compose up -d
go run ./cmd/devices-api migrate
go run ./cmd/devices-api
```

1. Create a settings file by copying the sample
   ```sh
   cp settings.sample.yaml settings.yaml
   ```
   Adjust these as necessary—the sample file should have what you need for local development. (Make sure you do this step each time you run `git pull` in case there have been any changes to the sample settings file.)

2. Start the services
   ```sh
   docker compose up -d
   ```
   This will start a bunch of services. Briefly:

   - Postgres, used to store the basic data models, on port 5432.
   - [Redis](https://redis.io), used by the [taskq library](https://taskq.uptrace.dev) to enqueue interactions with the AutoPi API, on port 6379.
   - [ElasticSearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html), only used by the sub-command `search-sync-dds`, on port 9200. Kibana provides a UI for this on port 5601.
   - [LocalStack](https://localstack.cloud), for testing our use of AWS S3 to store user documents and NFTs, takes up ports 4566–4583.
   - [IPFS](https://ipfs.tech), which we hope to use to store device definitions, takes up ports 4001, 8080, 8081, and 5001.
   - [Kafka](https://kafka.apache.org) is used to receive vehicle and task status updates, and emit events. It lives on port 9092, and the supporting Zookeeper service lives on port 2181.

   If you get a port conflict, you can find the existing process using the port with, e.g., `lsof -i :5432`. Most of these containers have attached volumes, so their data will persist across restarts. To check container status, run `docker ps`.

3. You can log into the database now with
   ```sh
   psql -h localhost -p 5432 -U dimo
   ```
   using password `dimo`, or use your favorite UI like [DataGrip](https://www.jetbrains.com/datagrip/). To do anything useful, you'll have to apply the database migrations from the `migrations` folder: 
   ```sh
   go run ./cmd/devices-api migrate
   ```

5. You are now ready to run the application:
   ```sh
   go run ./cmd/devices-api
   ```

It may be helpful to seed the database with test data:

6. Scrape device definitions from Edmunds:
   ```sh
   go run ./cmd/devices-api edmunds-vehicles-sync --mergemmy
   ```
7. Sync Smartcar integration compatibility:
   ```sh
   go run ./cmd/devices-api smartcar-sync
   ```
8. Scrape vehicle images from edmunds:
   ```sh
   go run ./cmd/devices-api edmunds-images [--overwrite]
   ```

Finally, if you want to test document uploads:

9. Execute the following command to point the AWS CLI at LocalStack:
   ```sh
   aws --endpoint-url=http://localhost:4566 s3 mb s3://documents
   ```

### Kafka test producer

This tool can be useful to test the consumer when running locally.
`$ go run ./cmd/test-producer <integrationID> <userDeviceID>`

Above integration and vehicle ID's aka userDeviceID should exist in your local DB. 

### Linting

`brew install golangci-lint`

`golangci-lint run`

This should use the settings from `.golangci.yml`, which you can override.

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
sqlboiler psql --no-tests --wipe
```
*Make sure you're running the docker image (ie. docker compose up)

## Migrations

To install goose in GO:
```bash
$ go get github.com/pressly/goose/v3/cmd/goose@v3.5.3
export GOOSE_DRIVER=postgres
```

To install goose CLI:
```bash
$ go install github.com/pressly/goose/v3/cmd/goose
export GOOSE_DRIVER=postgres
```

Add a migrations:
`$ goose -dir migrations create <migration_name> sql`

Migrate DB to latest:
`$ go run ./cmd/devices-api migrate`

Clear DB to start over:
```bash
docker ps
docker stop <container_id>
rm -R ./resources/data/ && mkdir ./resources/data/ 
docker compose up -d
```

If we have code base migrations in the migrations folder, we must import `_ "github.com/DIMO-Network/devices-api/migrations"` in the runner so that
it can find the migrations, otherwise get error.

### Managing migrations from k8s
```bash
kc get pods -n dev
kc exec devices-api-dev-65f8f47ff5-94dp4 -n dev -it -- /bin/sh
./devices-api migrate down # brings the last migration down
```

## Mocks

To regenerate a mock, you can use go gen since the files that are mocked have a `//go:generate mockgen ...` at the top. For example:
`nhtsa_api_service.go`

## Helm requirements

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

Note that swagger must be served from fiber-swagger library v2.31.1 +, since they fixed an issue in previous version. 

To check what cli version you have installed: `swag --version`. As of this writing v1.8.1 is working for us. 
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/devices-api/main.go --parseDependency --parseInternal --generatedTime true 
# optionally add `--parseDepth 2` if have issues
```

[declarative_comments_format](https://swaggo.github.io/swaggo.io/declarative_comments_format/)

