BINARY=gatus

TAG ?=
IMAGE ?= gatus-nobi:$(TAG)
CONTAINER ?= gatus
HOST_PORT ?= 8080
CONFIG_FILE ?= ./config.yaml
ENV_FILE ?= ./.env
COMPOSE_FILE ?= ./docker-compose.yml

.PHONY: help
help:
	@printf '%s\n' \
		'make install          Build the Gatus binary' \
		'make run              Run Gatus in development mode' \
		'make run-binary       Run the built binary in development mode' \
		'make test             Run all Go tests with coverage' \
		'make clean            Remove the built binary' \
		'make production-build Build the production Docker image' \
		'make build-deploy     Build and deploy the production image' \
		'make deploy           Deploy the existing production image' \
		'make restart          Restart the production container' \
		'make start            Start the production container' \
		'make stop             Stop the production container' \
		'make status           Show the production container status' \
		'make logs             Follow the production container logs' \
		'make docker-build     Build the default Gatus Docker image' \
		'make docker-run       Run the default Gatus Docker image' \
		'make docker-build-and-run Build and run the default Docker image' \
		'make frontend-install Install frontend dependencies' \
		'make frontend-build   Build the frontend' \
		'make frontend-dev     Start the frontend development server'

.PHONY: install
install:
	go build -v -o $(BINARY) .

.PHONY: run
run:
	ENVIRONMENT=dev GATUS_CONFIG_PATH=$(CONFIG_FILE) go run main.go

.PHONY: run-binary
run-binary:
	ENVIRONMENT=dev GATUS_CONFIG_PATH=$(CONFIG_FILE) ./$(BINARY)

.PHONY: clean
clean:
	rm $(BINARY)

.PHONY: test
test:
	go test ./... -cover


#########################
# Production deployment #
#########################

.PHONY: tag-check production-build-check deploy-check production-build build-deploy deploy restart start stop status logs
tag-check:
	@test -n "$(TAG)" || { echo "Missing TAG; use TAG=<image-tag>"; exit 1; }

production-build-check: tag-check
	@test -f "$(CURDIR)/Dockerfile" || { echo "Missing $(CURDIR)/Dockerfile"; exit 1; }

deploy-check: tag-check
	@test -f "$(COMPOSE_FILE)" || { echo "Missing $(COMPOSE_FILE)"; exit 1; }
	@test -f "$(CONFIG_FILE)" || { echo "Missing $(CONFIG_FILE)"; exit 1; }
	@test -f "$(ENV_FILE)" || { echo "Missing $(ENV_FILE); copy .env.example to .env"; exit 1; }

production-build: production-build-check
	docker build -t "$(IMAGE)" "$(CURDIR)"

build-deploy: production-build
	$(MAKE) deploy

deploy: deploy-check
	TAG="$(TAG)" \
	GATUS_PORT="$(HOST_PORT)" \
	GATUS_CONFIG_FILE="$(CONFIG_FILE)" \
	docker compose --env-file "$(ENV_FILE)" -f "$(COMPOSE_FILE)" up -d --no-build gatus

restart:
	docker restart "$(CONTAINER)"

start:
	docker start "$(CONTAINER)"

stop:
	docker stop "$(CONTAINER)"

status:
	docker ps -a --filter "name=^/$(CONTAINER)$$"

logs:
	docker logs --tail=100 -f "$(CONTAINER)"


##########
# Docker #
##########

.PHONY: docker-build
docker-build:
	docker build -t twinproduction/gatus:latest .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 --name gatus twinproduction/gatus:latest

.PHONY: docker-build-and-run
docker-build-and-run: docker-build docker-run


#############
# Front end #
#############

.PHONY: frontend-install
frontend-install:
	npm --prefix web/app install

.PHONY: frontend-build
frontend-build:
	npm --prefix web/app run build

.PHONY: frontend-dev
frontend-dev:
	npm --prefix web/app run serve
