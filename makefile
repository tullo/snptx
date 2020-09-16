SHELL = /bin/bash -o pipefail
export PROJECT = tullo-starter-kit
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 0.1.0
export DOCKER_BUILDKIT = 1
export COMPOSE_DOCKER_CLI_BUILD = 1

.DEFAULT_GOAL := run

all: snptx test-cover-profile test-cover-text check

run: up-db go-seed go-run

go-run:
	@go run ./cmd/snptx --db-disable-tls=1

go-migrate:
	@go run ./cmd/snptx-admin/main.go --db-disable-tls=1 migrate

go-seed: go-migrate
	@go run ./cmd/snptx-admin/main.go --db-disable-tls=1 seed

snptx:
	@docker build \
		-f Dockerfile \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/snptx-amd64:$(VERSION) \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	@docker image tag \
		$(REGISTRY_ACCOUNT)/snptx-amd64:$(VERSION) \
		gcr.io/$(PROJECT)/snptx-amd64:$(VERSION)

up:
	@docker-compose up --remove-orphans

up-db:
	@docker-compose up --detach --remove-orphans db
	@sleep 2

migrate:
	@docker-compose exec snptx /app/admin migrate

seed: migrate
	@docker-compose exec snptx /app/admin seed

down:
	@docker-compose down

test:
	@go test -count=1 -failfast -test.timeout=30s ./...

test-cover-profile:
	@go test -test.timeout=30s -coverprofile=/tmp/profile.out ./...

test-cover-text:
	@go tool cover -func=/tmp/profile.out

test-cover-html:
	@go tool cover -html=/tmp/profile.out

stop-all:
	@docker container stop $$(docker container ls -q --filter name=web_db)

remove-all:
	@docker container rm $$(docker container ls -aq --filter "name=web_db")

tidy:
	@go mod tidy
	@go mod vendor

deps-reset:
	@git checkout -- go.mod
	@go mod tidy
	@go mod vendor

deps-upgrade:
	@go get -d -t -u -v ./...
#   -d flag ...download the source code needed to build ...
#   -t flag ...consider modules needed to build tests ...
#   -u flag ...use newer minor or patch releases when available 

deps-cleancache:
	@go clean -modcache

check:
	$(shell go env GOPATH)/bin/staticcheck -go 1.15 -tests ./...

.PHONY: clone
clone:
	@git clone git@github.com:dominikh/go-tools.git /tmp/go-tools \
		&& cd /tmp/go-tools \
		&& git checkout "2020.1.5" \

.PHONY: install
install:
	@cd /tmp/go-tools && go install -v ./cmd/staticcheck
	$(shell go env GOPATH)/bin/staticcheck -debug.version
