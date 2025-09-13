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


# Default test format
TEST_FORMAT=testdox

# Override with verbose=true
# eg:
# - make test verbose=true
# - make test-pkgs pkgs=./tests/healthz verbose=true
# - make test-one pkg=./tests/healthz name=Test_CheckHealth verbose=true
ifeq ($(verbose),true)
	TEST_FORMAT=standard-verbose
endif

# Run all tests in all packages
.PHONY: test
test:
	go clean -testcache
    # Force sequential test package execution to prevent extensive resource consumption
    # Without `-p 1`, multiple test packages (authentication, healthz) would run concurrently,
    # each spinning up separate test containers and consuming excessive resources 	 	
	gotestsum --format $(TEST_FORMAT) --format-icons hivis --format-hide-empty-pkg -- -p 1 -v ./...

# Run all tests in a specific package or multiple packages
# eg: make test-pkgs pkgs=./tests/healthz
# eg: make test-pkgs pkgs="./tests/authentication ./tests/healthz"
.PHONY: test-pkgs
test-pkgs:
	test -n "$(pkgs)" || (echo "Missing argument: pkg. Example: make test-pkg pkg=./tests/healthz" && exit 1)
	go clean -testcache
    # Force sequential test package execution to prevent extensive resource consumption
    # Without `-p 1`, multiple test packages (authentication, healthz) would run concurrently,
    # each spinning up separate test containers and consuming excessive resources
	gotestsum --format $(TEST_FORMAT) --format-icons hivis --format-hide-empty-pkg -- -p 1 $(pkgs)

# Run a specific test in a specific package
# eg: make test-one pkg=./tests/healthz name=Test_CheckHealth
.PHONY: test-one
test-one:
	test -n "$(pkg)" || (echo "Missing argument: pkg. Example: make test-one pkg=./tests/healthz name=Test_CheckHealth" && exit 1)
	test -n "$(name)" || (echo "Missing argument: name. Example: make test-one pkg=./tests/healthz name=Test_CheckHealth" && exit 1)
	go clean -testcache
	gotestsum --format $(TEST_FORMAT) --format-icons hivis --format-hide-empty-pkg $(pkg) -run $(name)