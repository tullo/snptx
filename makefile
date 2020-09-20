SHELL = /bin/bash -o pipefail
export PROJECT = stackwise-starter-kit
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 0.1.0
export DOCKER_BUILDKIT = 1
export COMPOSE_DOCKER_CLI_BUILD = 1
export SESSION_SECRET := `openssl rand -base64 32`

.DEFAULT_GOAL := run

all: docker-build-image test-cover-profile test-cover-text staticcheck

run: compose-db-up go-seed go-run

config:
	@go run ./cmd/snptx --help

go-run:
	@go vet ./cmd/... ./internal/...
	@go run ./cmd/snptx --db-disable-tls=1 --web-debug-mode=true \
		--web-session-secret=${SESSION_SECRET} \
		--aragon-memory=$$(( 64 * 1024 )) --aragon-iterations=1 --aragon-parallelism=1

go-migrate:
	@go run ./cmd/snptx-admin/main.go --db-disable-tls=1 migrate

go-seed: go-migrate
	@go run ./cmd/snptx-admin/main.go --db-disable-tls=1 seed

docker-build-image: staticcheck
	@go vet ./cmd/... ./internal/...
	@docker build \
		-f Dockerfile \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/snptx-amd64:$(VERSION) \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

docker-tag-image:
	@docker image tag \
		$(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/snptx-amd64:$(VERSION) \
		eu.gcr.io/$(PROJECT)/snptx-amd64:`git rev-parse HEAD`

docker-push-image:
	set -e ; \
	docker image push eu.gcr.io/$(PROJECT)/snptx-amd64:`git rev-parse HEAD`
	@echo '==>' listing tags for image: [eu.gcr.io/$(PROJECT)/snptx-amd64]:
	@gcloud container images list-tags eu.gcr.io/$(PROJECT)/snptx-amd64

gcloud: docker-build-image docker-tag-image docker-push-image

compose-up: docker-build-image
	@docker-compose up -d --remove-orphans
	@docker-compose logs -f

compose-db-up:
	@docker-compose up --detach --remove-orphans db
	@echo Waiting for the database to accept connections ...
	@docker-compose exec db sh -c 'until $$(nc -z localhost 5432); do { printf '.'; sleep 1; }; done'

compose-migrate:
	@docker-compose exec snptx /app/admin migrate

compose-seed: compose-migrate
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

docker-stop-all:
	@docker container stop $$(docker container ls -q --filter name=db)

docker-remove-all:
	@docker container rm $$(docker container ls -aq --filter "name=db")

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

staticcheck:
	$(shell go env GOPATH)/bin/staticcheck -go 1.15 -tests ./...

.PHONY: install
install:
	set -e ; \
	git clone git@github.com:dominikh/go-tools.git /tmp/go-tools ; \
	cd /tmp/go-tools ; \
	git checkout "2020.1.5" ; \
	go install -v ./cmd/staticcheck
	$(shell go env GOPATH)/bin/staticcheck -debug.version
