SHELL = /bin/bash -o pipefail
export PROJECT = stackwise-starter-kit
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 0.1.0
export DOCKER_BUILDKIT = 1
export COMPOSE_DOCKER_CLI_BUILD = 1
export SESSION_SECRET = $(shell openssl rand -base64 32)

.DEFAULT_GOAL := run

browse:
	sensible-browser --new-tab https://snptx.127.0.0.1.nip.io:4200/ </dev/null >/dev/null 2>&1 & disown

all: docker-build-image go-test-coverage-profile go-tool-cover-text staticcheck

run: compose-db-up go-seed go-run

go-deps-upgrade:
	@go get -d -t -u -v ./...
#   -d flag ...download the source code needed to build ...
#   -t flag ...consider modules needed to build tests ...
#   -u flag ...use newer minor or patch releases when available 

go-deps-reset:
	@git checkout -- go.mod
	@go mod tidy
	@go mod vendor

go-config:
	@go run ./cmd/snptx --help

go-mod-tidy:
	@go mod tidy
	@go mod vendor

go-mod-list-final:
	@echo '==>' Final versions that will be used in a build for all direct and indirect dependencies
	@go list -mod=readonly -m all
#	-m flag causes list to list modules instead of packages.

go-mod-list-updates:
	@go list -mod=readonly -json -m -u all
#	-u flag adds information about available upgrades.

go-mod-why:
	@go mod why -m golang.org/x/sys

go-run:
	@go vet ./cmd/... ./internal/...
	@echo '==>' Activating debug mode to get detailed errors and stack traces in the http response.
	@go run ./cmd/snptx --db-disable-tls=1 --web-debug-mode=true \
		--web-session-secret=${SESSION_SECRET} \
		--aragon-memory=$$(( 64 * 1024 )) --aragon-iterations=1 --aragon-parallelism=1

go-migrate:
	@go run ./cmd/snptx-admin/main.go --db-disable-tls=1 migrate

go-seed: go-migrate
	@go run ./cmd/snptx-admin/main.go --db-disable-tls=1 seed

go-test: staticcheck
	@go test -count=1 -failfast -test.timeout=30s ./...

go-test-coverage-summary:
	@go test -cover ./...

go-test-coverage-profile:
	@go test -test.timeout=30s -covermode=count -coverprofile=/tmp/profile.out ./...

go-tool-cover-text:
	@go tool cover -func=/tmp/profile.out

go-tool-cover-html:
	@go tool cover -html=/tmp/profile.out

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

compose-config:
	@docker-compose config

compose-logs:
	@docker-compose logs -f

compose-up: docker-build-image
	@docker-compose up -d --remove-orphans
	@docker-compose exec db sh -c 'until $$(nc -z localhost 5432); do { printf '.'; sleep 1; }; done'
	@docker-compose logs
	@docker-compose exec snptx /app/admin migrate
	@docker-compose exec snptx /app/admin seed

compose-down:
	@docker-compose down

compose-db-up:
	@docker-compose up --detach --remove-orphans db
	@echo Waiting for the database to accept connections ...
	@docker-compose exec db sh -c 'until $$(nc -z localhost 5432); do { printf '.'; sleep 1; }; done'

compose-migrate:
	@docker-compose exec snptx /app/admin migrate

compose-seed: compose-migrate
	@docker-compose exec snptx /app/admin seed

compose-psql: compose-db-up
	@docker-compose exec db psql -U postgres

docker-stop-all:
	@docker container stop $$(docker container ls -aq --filter name=db --filter name=snptx)

docker-remove-all: docker-stop-all
	@docker container rm $$(docker container ls -aq --filter name=db --filter name=snptx)

deps-cleancache:
	@go clean -modcache

staticcheck:
	$(shell go env GOPATH)/bin/staticcheck -go 1.15 -tests ./...

staticcheck-install:
	set -e ; \
	git clone git@github.com:dominikh/go-tools.git /tmp/go-tools ; \
	cd /tmp/go-tools ; \
	git checkout 2020.2.1 ; \
	go get ./...; \
	go install ./... ; \
	rm -fr /tmp/go-tools
	$(shell go env GOPATH)/bin/staticcheck -debug.version

mkcert-install:
	@echo make sure libnss3-tools is installed \"apt install libnss3-tools\"
	set -e ; \
	git clone https://github.com/FiloSottile/mkcert /tmp/mkcert ; \
	cd /tmp/mkcert ; \
	go install -v -ldflags "-X main.Version=$$(git describe --tags)" ; \
	rm -fr /tmp/mkcert
	$$(go env GOPATH)/bin/mkcert

mkcert-install-rootCA:
	$$(go env GOPATH)/bin/mkcert -install

mkcert-generate-certs:
	@mkdir -p tls/localhost
	$$(go env GOPATH)/bin/mkcert -cert-file ./tls/localhost/cert.pem -key-file ./tls/localhost/key.pem \
		snptx.127.0.0.1.nip.io snptx.test snptx 0.0.0.0 localhost 127.0.0.1 ::1
