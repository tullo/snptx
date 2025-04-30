SHELL = /bin/bash -o pipefail
export PROJECT = stackwise-starter-kit
export REGISTRY_HOSTNAME = docker.io
export REGISTRY_ACCOUNT = tullo
export VERSION = 0.1.0
export DOCKER_BUILDKIT = 1
export COMPOSE_DOCKER_CLI_BUILD = 1
export DATABASE_URL ?= postgresql://root@localhost:26257/snptx?sslmode=disable

.DEFAULT_GOAL := run

browse:
	sensible-browser --new-tab https://snptx.127.0.0.1.nip.io:4200/ </dev/null >/dev/null 2>&1 & disown

all: docker-build-image go-test-coverage-profile go-tool-cover-text

run: compose-db-up

down: compose-down

docker-build-image:
	@go vet ./cmd/... ./internal/...
	@docker build \
		-f Dockerfile \
		-t $(REGISTRY_HOSTNAME)/$(REGISTRY_ACCOUNT)/snptx-amd64:$(VERSION) \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date --utc +"%Y-%m-%dT%H:%M:%S%Z"` \
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

compose-up: CMD=/cockroach/cockroach node status --insecure
compose-up: docker-build-image
	@docker-compose up -d --remove-orphans
	@docker-compose exec db sh -c 'until ${CMD}; do { printf '.'; sleep 1; }; done'
	@docker-compose logs
	@docker-compose exec snptx /app/admin migrate
	@docker-compose exec snptx /app/admin seed

compose-down:
	@docker-compose down -v

compose-db-up: CMD=/cockroach/cockroach node status --insecure
compose-db-up:
	@docker-compose up --detach --remove-orphans db
	@echo Waiting for the database to accept connections ...
	@docker-compose exec db sh -c 'until ${CMD}; do { printf '.'; sleep 1; }; done'

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
