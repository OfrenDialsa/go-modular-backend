.PHONY: help dev-up dev-down dev-build dev-restart dev-logs migrate-up migrate-down migrate-status build run clean mocks clean-mocks test monitor-up monitor-down monitor-logs monitor-restart

DC = docker compose -f docker-compose.yml
DC_MONITORING = docker compose -f docker/monitoring/docker-compose.yml
DC_MINIO = docker compose -f docker/minio/docker-compose.yml
DB_CONTAINER = postgres-starter
DB_USER = postgres
DB_NAME = go-gin-starter
API_CONTAINER=api

MOCKERY := ~/go/bin/mockery
REPO_DIR := internal/repository
SERVICE_DIR := internal/service
STORAGE_DIR := internal/infra/storage
MAILER_DIR := internal/infra/mailer
MOCK_DIR := tests/mocks
MOCK_PKG := mocks

COVERAGE_FILE = coverage.out
HTML_FILE = coverage.html

FEATURE_NAME=$(name)

up:
	$(DC) up -d

build:
	$(DC) build

up-build:
	$(DC) up --build

down:
	$(DC) down

restart:
	$(DC) restart

logs:
	$(DC) logs -f

logs-api:
	$(DC) logs -f api

ps:
	$(DC) ps

minio-up:
	$(DC_MINIO) up -d

minio-down:
	$(DC_MINIO) down

minio-logs:
	$(DC_MINIO) logs -f

minio-restart:
	$(DC_MINIO) restart

monitor-up:
	$(DC_MONITORING) up -d
	
monitor-down:
	$(DC_MONITORING) down

monitor-logs:
	$(DC_MONITORING) logs -f

monitor-restart:
	$(DC_MONITORING) restart

dev-logs:
	@echo "Menampilkan logs dari API..."
	$(DC) logs -f

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "=X= Error: 'name' is required. Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	@go run cmd/migrate/main.go -action=create -name=$(name)

migrate-up:
	$(DC) exec $(API_CONTAINER) go run cmd/migrate/main.go -action=up

migrate-down:
	$(DC) exec $(API_CONTAINER) go run cmd/migrate/main.go -action=down

migrate-status:
	$(DC) exec $(API_CONTAINER) ls -l migrations

db-shell:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

clean:
	go clean

MOCK_FILTER = 2>&1 | grep -E "Generating mock|Unable to find|could not find|error" | awk '\
    /Generating mock/ { \
        split($$0, a, "interface="); \
        split(a[2], b, " "); \
        print "  -> Mocking interface:", b[1] \
    } \
    /Unable to find|could not find|error/ { \
        print "  \033[0;31m[!] Error:\033[0m Interface not found" \
    }'

clean-mocks:
	@echo "[X] Cleaning old mocks in $(MOCK_DIR)..."
	@rm -rf $(MOCK_DIR)

mocks:
	@echo "[>>] Generating repository mocks..."
	@mkdir -p $(MOCK_DIR)
	@$(MOCKERY) --all --dir $(REPO_DIR) --output $(MOCK_DIR) --outpkg $(MOCK_PKG) --case snake --disable-version-string $(MOCK_FILTER)

	@echo "[>>] Generating specific mocks..."
	@$(MOCKERY) --name TxStarter --dir $(SERVICE_DIR) --output $(MOCK_DIR) --outpkg $(MOCK_PKG) --case snake --disable-version-string $(MOCK_FILTER)
	@$(MOCKERY) --name StorageService --dir $(STORAGE_DIR) --output $(MOCK_DIR) --outpkg $(MOCK_PKG) --case snake --disable-version-string $(MOCK_FILTER)
	
	@echo "[>>] Generating external mocks..."
	@$(MOCKERY) --name Tx --srcpkg github.com/jackc/pgx/v5 --output $(MOCK_DIR) --outpkg $(MOCK_PKG) --case snake --disable-version-string $(MOCK_FILTER)
	
	@echo "[>>] Generating mailer mock..."
	@$(MOCKERY) --name MailerService --dir $(MAILER_DIR) --output $(MOCK_DIR) --outpkg $(MOCK_PKG) --case snake --disable-version-string $(MOCK_FILTER)
	
	@echo "[v] Done generating mocks"

test:
	@echo "Running tests and generating coverage profile..."
	go test -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic -coverpkg=./internal/service/... ./tests/...
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

cover-html: test
	@echo "Generating HTML coverage report..."
	go tool cover -html=$(COVERAGE_FILE) -o $(HTML_FILE)
	@echo "Coverage report generated: $(HTML_FILE)"

cover-open: cover-html
	@open $(HTML_FILE) || xdg-open $(HTML_FILE) || start $(HTML_FILE)

clean-cover:
	rm -f $(COVERAGE_FILE) $(HTML_FILE)

swag-init:
	swag init -g main.go --output docs