.DEFAULT_GOAL := help

.PHONY: help setup infra up down logs dev dev-web dev-api test check build migrate migrate-status migrate-down sqlc sqlc-check test-integration

help:
	@echo "setup     Generate private local credentials and install web dependencies"
	@echo "infra     Start PostgreSQL and MinIO only"
	@echo "up        Build and start the complete application in Docker"
	@echo "dev       Start infra; run make dev-web and make dev-api in two terminals"
	@echo "down      Stop containers without deleting data"
	@echo "test      Run local setup, Go, and web unit tests"
	@echo "check     Run tests, Go vet, web lint and typecheck"
	@echo "build     Build API and production website"
	@echo "migrate   Apply migrations (loads trusted .env if present)"
	@echo "migrate-status / migrate-down   Inspect / roll back ONE migration"
	@echo "sqlc / sqlc-check   Generate queries / verify no generation drift"
	@echo "test-integration   Disposable PostgreSQL; never uses development volumes"

setup:
	python3 scripts/setup_local.py
	npm --prefix apps/web ci

infra:
	docker compose up -d --build --wait postgres minio
	docker compose run --build --rm storage-init

up:
	docker compose --profile application up -d --build --wait

down:
	docker compose --profile application down

logs:
	docker compose --profile application logs -f

dev: infra
	@echo "Run make dev-web and make dev-api in separate terminals."

dev-web:
	npm --prefix apps/web run dev

dev-api:
	@set -a; . ./.env; set +a; cd apps/api && go run ./cmd/api

test:
	python3 -m unittest discover -s tests -v
	cd apps/api && go test -race -cover ./...
	npm --prefix apps/web test

check: test
	cd apps/api && go vet ./...
	npm --prefix apps/web run lint
	npm --prefix apps/web run typecheck

build:
	cd apps/api && go build -o bin/api ./cmd/api
	npm --prefix apps/web run build

migrate:
	@set -a; if [ -f .env ]; then . ./.env; fi; set +a; python3 scripts/api_tools.py up

migrate-status:
	@set -a; if [ -f .env ]; then . ./.env; fi; set +a; python3 scripts/api_tools.py status

# Destructive: reverts one migration, including its data. Back up first.
migrate-down:
	@set -a; if [ -f .env ]; then . ./.env; fi; set +a; python3 scripts/api_tools.py down

sqlc:
	python3 scripts/api_tools.py sqlc

sqlc-check:
	python3 scripts/api_tools.py sqlc-check

test-integration:
	python3 scripts/api_tools.py integration
