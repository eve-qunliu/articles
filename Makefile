GOLANG_DEPS_DIR=vendor
ifdef DOTENV
	DOTENV_TARGET=dotenv
else
	DOTENV_TARGET=.env
endif

DB_STRING=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):5432/$(POSTGRES_DB)?sslmode=disable

# all is the default Make target. it installs the dependencies, tests, and builds the application
all: deps test build
.PHONY: all

deps: $(DOTENV_TARGET)
	docker-compose run --rm golang make _depsGo
.PHONY: deps

start: $(DOTENV_TARGET)
	docker-compose run -p 8080:8080 --rm golang make _start
.PHONY: start

test: $(DOTENV_TARGET)
	docker-compose run --rm golang make _testUnit
.PHONY: test

# migrateUp applies all migrations scripts up to the latest version
migrateUp: $(DOTENV_TARGET)
	docker-compose run --rm golang make _dbUp
.PHONY: migrateUp

# migrateDown reverts the last migration script that was applied
migrateDown: $(DOTENV_TARGET)
	docker-compose run --rm golang make _dbDown
.PHONY: migrateDown

# .env creates .env based on .env.template if .env does not exist
.env:
	cp .env.example .env

# dotenv creates/overwrites .env with $(DOTENV)
dotenv:
	cp $(DOTENV) .env
.PHONY: dotenv

# _depsGo installs go dependencies for the project
_depsGo:
	glide install
.PHONY: _depsGo

_testUnit:
	go test -cover -v ./...
.PHONY: _testUnit

_start: _waitForDB
	go run main.go
.PHONY: _start

_dbUp: _waitForDB
	migrate -path ./migrations -database $(DB_STRING) up
.PHONY: _dbUp

_dbDown: _waitForDB
	migrate -path ./migrations -database $(DB_STRING) down 1
.PHONY: _dbDown

_waitForDB:
	dockerize -wait tcp://postgres:5432 -timeout 60s
.PHONY: _waitForDB
