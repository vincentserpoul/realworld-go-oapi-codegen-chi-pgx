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

- [migrate](https://github.com/golang-migrate/migrate)
- [docker compose](https://docs.docker.com/compose/)
- [go](https://go.dev)
- [golangci-lint](https://golangci-lint.run/)

## Run

```bash
    make infra-local-up
```

```bash
    make db-migration-local-up
```

```bash
    go run cmd/api/main.go
```

# Contributing

Install [pre-commit](https://pre-commit.com/)

Setup pre-commit

```bash
pre-commit install -t commit-msg -t pre-commit -t pre-push
```
