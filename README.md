[![checks sec, lint](https://github.com/vincentserpoul/realworld-go-oapi-codegen-chi-pgx/actions/workflows/check.yaml/badge.svg)](https://github.com/vincentserpoul/realworld-go-oapi-codegen-chi-pgx/actions/workflows/check.yaml) [![build](https://github.com/vincentserpoul/realworld-go-oapi-codegen-chi-pgx/actions/workflows/build.yaml/badge.svg)](https://github.com/vincentserpoul/realworld-go-oapi-codegen-chi-pgx/actions/workflows/build.yml) [![Coverage Status](https://coveralls.io/repos/github/vincentserpoul/realworld-go-oapi-codegen-chi-pgx/badge.svg?branch=main)](https://coveralls.io/github/vincentserpoul/realworld-go-oapi-codegen-chi-pgx?branch=main)

# ![RealWorld Example App](logo.png)

> ### codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.


### [Demo](https://demo.realworld.io/)&nbsp;&nbsp;&nbsp;&nbsp;[RealWorld](https://github.com/gothinkster/realworld)


This codebase was created to demonstrate a fully fledged fullstack application built with pgx, oapi-codegen and chi including CRUD operations, authentication, routing, pagination, and more.

For more information on how to this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.


# How it works

Forget your ORMs, this implementation is leveraging deepmap/oapi-codegen/v2, go-chi/v5 and pgx/v5.

# Getting started

## What you need

- [migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [docker compose](https://docs.docker.com/compose/)
- [go](https://go.dev)
- [golangci-lint](https://golangci-lint.run/)

## Setting up the config

Copy the content of config/api/secrets.sample.yaml to config/api/local.secrets.yaml

## Run

```bash
    make infra-local-up
```

```bash
    make db-migration-up
```

```bash
    go run cmd/api/main.go
```

## Enabling otel traces

Just go to 127.0.0.1:5080, log in with the credentials you find in infra/local/docker-compose.yaml.
Then press the tab "ingestion" and look into the Authorization header key.
Update the header in the file you just created (config/api/local.secrets.yaml) and restart!

## Running the test suite

After all the setup is done, in one terminal, run:

```bash
go run ./cmd/api/main.go
```

and in another, you run:

```bash
APIURL=http://localhost:8083 ./api/run-api-tests.sh
```

## Contributing

Make sure you install [pre-commit](https://pre-commit.com/) and set it up as following:

```bash
pre-commit install -t commit-msg -t pre-commit -t pre-push
```
