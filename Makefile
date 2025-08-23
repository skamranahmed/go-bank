GOOSE_MIGRATOR_BINARY_NAME = goose-migrator
GOOSE_MIGRATOR_DIR = migrator

.SILENT:

.PHONY: run
run:
	go run main.go

.PHONY: build-migrator
build-migrator:
	cd $(GOOSE_MIGRATOR_DIR) && go build -o $(GOOSE_MIGRATOR_BINARY_NAME)

.PHONY: migrate-create
migrate-create: build-migrator
	cd $(GOOSE_MIGRATOR_DIR) && ./$(GOOSE_MIGRATOR_BINARY_NAME) create $(name) go

.PHONY: migrate-up
migrate-up: build-migrator
	cd $(GOOSE_MIGRATOR_DIR) && ./$(GOOSE_MIGRATOR_BINARY_NAME) up

.PHONY: migrate-down
migrate-down: build-migrator
	cd $(GOOSE_MIGRATOR_DIR) && ./$(GOOSE_MIGRATOR_BINARY_NAME) down

