.PHONY: all clean \
		help \
		test test-race test-leak test-coverage-report test-coveralls \
		bench bench-compare \
		upgrade \
		lint \
		sec-scan sec-trivy-scan sec-vuln-scan \
		build-docker-api build-docker-generic \
		db-pg-init db-migration-up db-migration-down \
		infra-local-up infra-local-down \
		gci-format

help: ## show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

PROJECT_NAME?=realworld
API_NAME?=$(PROJECT_NAME)-api


SHELL = /bin/bash

########
# test #
########

test: ## launch all tests
	go test ./... -race -failfast -cover

test-coverage-report: ## test with coverage report
	go test ./internal/... -race -failfast -coverpkg=./... -covermode=atomic -coverprofile=./coverage.out
	go tool cover -html=coverage.out

test-coveralls:
	go test ./... -race -failfast -coverpkg=./internal/... -covermode=atomic -coverprofile=./coverage.out
	go tool goveralls -covermode=atomic -coverprofile=./coverage.out -repotoken=$(COVERALLS_TOKEN)


test-clean-cache: ## clean test cache
	go clean -testcache


#############
# benchmark #
#############

bench: ## launch benchs
	go test ./... -bench=. -benchmem | tee ./bench.txt

bench-compare: ## compare benchs results
	go tool benchstat ./bench.txt

############
# upgrades #
############

upgrade: ## upgrade dependencies (beware, it can break everything)
	go mod tidy && \
	go get -t -u ./... && \
	go mod tidy


upgrade-tools: ## upgrade all tools listed in go.mod
	@echo "Upgrading tools..."
	@tools=$$(sed -n '/^tool (/,/)/p' go.mod | grep -E '^\s*github.com'); \
	  for tool in $$tools; do \
	    echo "Upgrading $$tool"; \
	    go get -tool $$tool; \
	  done

########
# lint #
########

lint: ## lints the entire codebase
	@go tool golangci-lint run ./... --config=./.golangci.toml

lint-clean-cache: ## clean the linter cache
	@go tool golangci-lint cache clean

#######
# sec #
#######

sec-scan: sec-trivy-scan sec-vuln-scan ## scan for security and vulnerability issues

sec-trivy-scan: ## scan for sec issues with trivy (trivy binary needed)
	trivy fs --exit-code 1 --no-progress --severity CRITICAL ./

sec-vuln-scan: ## scan for vulnerability issues with govulncheck (govulncheck binary needed)
	go tool govulncheck ./...


#########
# build #
#########

build-docker-api: TAG_NAME=$(API_NAME) ## docker build for api
build-docker-api: BINARY_NAME="api"
build-docker-api: build-docker-generic

build-docker-generic:
	if [[ -n "${PLATFORM}" ]]; then \
		PLATFORM_FLAG="--platform ${PLATFORM}"; \
	fi; \
	docker buildx build \
		-f Dockerfile \
		-t $(TAG_NAME) \
		$$PLATFORM_FLAG \
		--build-arg BINARY_NAME=$(BINARY_NAME) \
		--build-arg BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg LAST_MAIN_COMMIT_HASH=$(shell git rev-parse --short HEAD) \
		--progress=plain \
		--load \
		./

######
# db #
######

db-migration-up: ## migration up, using https://github.com/golang-migrate/migrate
	@( \
	printf "Enter database URL (for ex: postgres://postgres:localworld@localhost:5432/realworld?sslmode=disable): \n"; read -r DATABASE_URL &&\
	migrate -database $${DATABASE_URL} -path database/migrations up; \
	)

db-migration-down: ## migration down
	@( \
	printf "Enter database URL (for ex: postgres://postgres:localworld@localhost:5432/realworld?sslmode=disable): \n"; read -r DATABASE_URL &&\
	migrate -database $${DATABASE_URL} -path database/migrations down; \
	)

#########
# infra #
#########

infra-local-up: ## launch local infra
	docker compose -f ./infra/local/docker-compose.yaml up -d

infra-local-down: ## remoave local infra
	docker compose -f ./infra/local/docker-compose.yaml down

###########
#   GCI   #
###########

gci-format: ## format repo through gci linter
	go tool gci ./ --skip-generated -s standard -s default -s "Prefix(realworld)"
