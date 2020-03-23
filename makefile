SHELL := /bin/bash

export PROJECT = tullo-starter-kit
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 0.1.0
export DOCKER_BUILDKIT = 1

all: snptx test-cover-profile test-cover

migrate:
	go run ./cmd/web-admin/main.go --db-disable-tls=1 migrate

snptx:
	docker build \
		-f Dockerfile \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/snptx-amd64:$(VERSION) \
		--build-arg PACKAGE_NAME=snptx \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker image tag \
		$(REGISTRY_ACCOUNT)/snptx-amd64:$(VERSION) \
		gcr.io/$(PROJECT)/snptx-amd64:$(VERSION)

up:
	docker-compose up --remove-orphans

down:
	docker-compose down

test:
	go test -mod=vendor ./... -count=1

test-cover-profile:
	go test -mod=vendor -coverprofile=/tmp/profile.out ./...

test-cover:
	go tool cover -func=/tmp/profile.out

stop-all:
	docker container stop $$(docker container ls -q --filter name=web_db)

remove-all:
	docker container rm $$(docker container ls -aq --filter "name=web_db")

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -t -d -v ./...

deps-cleancache:
	go clean -modcache
