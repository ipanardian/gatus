BINARY=gatus

IMAGE ?= gatus-nobi:latest
CONTAINER ?= gatus
HOST_PORT ?= 8080
DEPLOY_ROOT ?= $(abspath $(CURDIR)/..)
CONFIG_FILE ?= $(DEPLOY_ROOT)/config.yaml
DATA_DIR ?= $(DEPLOY_ROOT)/data
PRODUCTION_ENVIRONMENT ?= production

.PHONY: install
install:
	go build -v -o $(BINARY) .

.PHONY: run
run:
	ENVIRONMENT=dev GATUS_CONFIG_PATH=./config.yaml go run main.go

.PHONY: run-binary
run-binary:
	ENVIRONMENT=dev GATUS_CONFIG_PATH=./config.yaml ./$(BINARY)

.PHONY: clean
clean:
	rm $(BINARY)

.PHONY: test
test:
	go test ./... -cover


#########################
# Production deployment #
#########################

.PHONY: production-check production-build deploy restart start stop status logs
production-check:
	@test -f "$(CURDIR)/Dockerfile" || { echo "Missing $(CURDIR)/Dockerfile"; exit 1; }
	@test -f "$(CONFIG_FILE)" || { echo "Missing $(CONFIG_FILE)"; exit 1; }

production-build: production-check
	docker build -t "$(IMAGE)" "$(CURDIR)"

deploy: production-build
	@mkdir -p "$(DATA_DIR)"
	@docker rm -f "$(CONTAINER)" >/dev/null 2>&1 || true
	docker run -d \
		--restart=unless-stopped \
		-e ENVIRONMENT="$(PRODUCTION_ENVIRONMENT)" \
		-p "$(HOST_PORT):8080" \
		-v "$(CONFIG_FILE):/config/config.yaml:ro" \
		-v "$(DATA_DIR):/data" \
		--name "$(CONTAINER)" \
		"$(IMAGE)"

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
