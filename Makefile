APP_NAME=autonotif-scheduler
OUTPUT_DIR=deployment/tmp
DOCKER_NETWORK=autonotif-network
CONFIG_PATH?=./config/config.yaml

.PHONY: tidy
tidy:
	env GO111MODULE=on go mod tidy

.PHONY: test
test:
	go test -race ./...

.PHONY: migrate
migrate:
	docker run --rm -v $(shell pwd)/db/migrations:/migrations --network $(DOCKER_NETWORK) migrate/migrate -path=/migrations/ -database $(url) up

.PHONY: rollback
rollback:
	docker run --rm -v $(shell pwd)/db/migrations:/migrations --network $(DOCKER_NETWORK) migrate/migrate -path=/migrations/ -database $(url) down 1

.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)
	docker image prune --force --filter='label=$(APP_NAME)'

.PHONY: compile
compile:
	mkdir -p $(OUTPUT_DIR)
	env GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(OUTPUT_DIR)/$(APP_NAME) cmd/scheduler/main.go

.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME) --label $(APP_NAME) -f deployment/Dockerfile .

.PHONY: docker-run
docker-run:
	docker run -d --rm \
		-p 10.104.0.4:8080:8080 \
		-v $(shell pwd)/config:/app/config \
		-e CONFIG_PATH=$(CONFIG_PATH) \
		--net $(DOCKER_NETWORK) \
		--name $(APP_NAME) \
		$(APP_NAME):latest

.PHONE: run
run: clean compile docker-network docker-postgres docker-build docker-run

.PHONY: go-run
go-run:
	CONFIG_PATH=$(CONFIG_PATH) go run cmd/scheduler/main.go

.PHONE: docker-network
docker-network:
	docker network create $(DOCKER_NETWORK) || true

.PHONE: docker-postgres
docker-postgres:
	docker-compose -f db/docker-compose.yml up -d
