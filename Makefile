.PHONY: all clean help deploy test test-race test-leak bench bench-compare bench-swagger-gen lint sec-scan vuln-scan upgrade release release-tag changelog-gen changelog-commit proto-gen proto-lint

help: ## show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

PROJECT_NAME?=realworld
API_NAME?=$(PROJECT_NAME)-api
MIGRATE_NAME?=$(PROJECT_NAME)-migrate
ENV_LIST?=local localdev dev production

SHELL = /bin/bash

########
# test #
########

test: ## launch all tests
	go test ./... -cover -race -leak

test-race: ## launch all tests with race detection
	go test ./... -cover -race

test-leak: ## launch all tests with leak detection
	go test ./... -leak

test-coverage-report: ## test with coverage report
	go test -v  ./... -cover -race -covermode=atomic -coverprofile=./coverage.out
	go tool cover -html=coverage.out

test-coveralls:
	go test -v ./... -race -leak -failfast -covermode=atomic -coverprofile=./coverage.out
	goveralls -covermode=atomic -coverprofile=./coverage.out -repotoken=$(COVERALLS_TOKEN)

#############
# benchmark #
#############

bench: ## launch benchs
	go test ./... -bench=. -benchmem | tee ./bench.txt

bench-compare: ## compare benchs results
	benchstat ./bench.txt

############
# upgrades #
############

upgrade: ## upgrade dependencies (beware, it can break everything)
	go mod tidy && \
	go get -t -u ./... && \
	go mod tidy


########
# lint #
########

lint: ## lints the entire codebase
	@golangci-lint run ./... --config=./.golangci.toml


#######
# sec #
#######

sec-scan: trivy-scan vuln-scan ## scan for security and vulnerability issues

trivy-scan: ## scan for sec issues with trivy (trivy binary needed)
	trivy fs --exit-code 1 --no-progress --severity CRITICAL ./

vuln-scan: ## scan for vulnerability issues with govulncheck (govulncheck binary needed)
	govulncheck ./...


######
# db #
######

APP_NAME_UND=$(shell echo "$(API_NAME)" | tr '-' '_')

db-pg-init: ## create users and passwords in postgres for your app
	@( \
	printf "Enter pass for db: \n"; read -rs DB_PASSWORD &&\
	printf "Enter port(5436...): \n"; read -r DB_PORT &&\
	sed \
	-e "s/DB_PASSWORD/$$DB_PASSWORD/g" \
	-e "s/APP_NAME_UND/$(APP_NAME_UND)/g" \
	./database/init/init.sql | \
	PGPASSWORD=$$DB_PASSWORD psql -h localhost -p $$DB_PORT -U postgres -f - \
	)

db-migration-local-up: ## migration up, using https://github.com/golang-migrate/migrate
	@( \
	printf "Enter database URL (for ex: postgres://postgres:localworld@localhost:5432/realworld?sslmode=disable): \n"; read -r DATABASE_URL &&\
	migrate -database $${DATABASE_URL} -path database/migrations up; \
	)

db-migration-local-down: ## migration down
	@( \
	printf "Enter database URL (for ex: postgres://postgres:localworld@localhost:5432/realworld?sslmode=disable): \n"; read -r DATABASE_URL &&\
	migrate -database $${DATABASE_URL} -path database/migrations down; \
	)

#########
# infra #
#########

infra-local-up: ## launch local infra
	docker-compose -f ./infra/local/docker-compose.yaml up -d

infra-local-down: ## remoave local infra
	docker-compose -f ./infra/local/docker-compose.yaml down

###########
#   GCI   #
###########

gci-format: ## format repo through gci linter
	gci write ./ --skip-generated -s standard -s default -s "Prefix(realworld)"
