.PHONY: help up down test lint migrate build

help:
@echo "Conduit-TH — available commands:"
@echo ""
@echo "  make up             start all local services"
@echo "  make down           stop all local services"
@echo "  make test           run all tests"
@echo "  make test svc=auth  run tests for specific service"
@echo "  make lint           run linter"
@echo "  make migrate        run database migrations up"
@echo "  make migrate-down   rollback 1 migration"
@echo "  make migrate-reset  rollback all + migrate up"
@echo "  make build          build all binaries"

up:
docker compose -f infra/docker/docker-compose.yml up -d
@echo "✓ services started"
@echo "  PostgreSQL : localhost:5432"
@echo "  Redis      : localhost:6379"
@echo "  MinIO      : localhost:9000 (console: 9001)"
@echo "  Adminer    : localhost:8080"

down:
docker compose -f infra/docker/docker-compose.yml down

migrate:
migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-reset:
migrate -path migrations -database "$(DATABASE_URL)" down -all
migrate -path migrations -database "$(DATABASE_URL)" up

test:
ifdef svc
@echo "── testing $(svc) ──"
cd services/$(svc) && go test -v -race -coverprofile=coverage.out ./...
else
@for svc in auth job user document media material notify report config line-webhook ai-worker gateway; do \
echo "── testing $$svc ──"; \
cd services/$$svc && go test -race ./... && cd ../..; \
done
endif

lint:
cd shared && golangci-lint run ./...
@for svc in auth job user document media material notify report config line-webhook ai-worker gateway; do \
echo "── linting $$svc ──"; \
cd services/$$svc && golangci-lint run ./... && cd ../..; \
done

build:
@mkdir -p bin
@for svc in auth job user document media material notify report config line-webhook ai-worker gateway; do \
echo "── building $$svc ──"; \
cd services/$$svc && go build -o ../../bin/$$svc . && cd ../..; \
done
@echo "✓ all binaries in ./bin/"
