ENV=
ENV_FILE=.env.$(ENV)
-include $(ENV_FILE)
COMPOSE_FILE=docker-compose.$(ENV).yaml
COMPOSE=docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE)

IMAGE_NAME=ghcr.io/dani-susanto/go-account-service
IMAGE_TAG=

DB_URL=postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable
MIGRATION_NAME=
MIGRATION_STEP=
MIGRATE=docker run --rm -v $(PWD)/migrations:/migrations --network infrastructure-$(ENV)-network migrate/migrate -path /migrations -database $(DB_URL)

# ===============================
# 🛡️ GUARD
# ===============================
guard-env:
	@test -n "$(ENV)" || (echo "❌ ENV is required. Usage: make <target> ENV=local|development|production" && exit 1)
	@echo "$(ENV)" | grep -qE "^(local|development|production)$$" || (echo "❌ ENV invalid. Must be: local, development, or production" && exit 1)

guard-tag:
	@test -n "$(IMAGE_TAG)" || (echo "❌ IMAGE_TAG is required. Usage: make <target> IMAGE_TAG=v1.0.0" && exit 1)

guard-migration-name:
	@test -n "$(MIGRATION_NAME)" || (echo "❌ MIGRATION_NAME is required. Usage: make <target> MIGRATION_NAME=create_users_table" && exit 1)

guard-migration-step:
	@test -n "$(MIGRATION_STEP)" || (echo "❌ MIGRATION_STEP is required. Usage: make <target> MIGRATION_STEP=1" && exit 1)


# ===============================
# 🟢 START
# ===============================
start: guard-env
	@echo "🚀 Starting $(ENV) environment..."
	$(COMPOSE) up -d

start-build: guard-env
	@echo "🚀 Starting $(ENV) environment with build..."
	$(COMPOSE) up -d --build

start-local:
	$(MAKE) start ENV=local
	$(MAKE) log-app-local

start-local-build:
	$(MAKE) start-build ENV=local
	$(MAKE) log-app-local

start-dev:
	$(MAKE) start ENV=development
	$(MAKE) log-app-dev

start-dev-build:
	$(MAKE) start-build ENV=development
	$(MAKE) log-app-dev

start-prod:
	$(MAKE) start ENV=production
	$(MAKE) log-app-prod

start-prod-build:
	$(MAKE) start-build ENV=production
	$(MAKE) log-app-prod


# ===============================
# 🔴 STOP
# ===============================
stop: guard-env
	@echo "🛑 Stopping $(ENV) environment..."
	$(COMPOSE) down

stop-v: guard-env
	@echo "🛑 Stopping $(ENV) environment and removing volumes..."
	$(COMPOSE) down -v

stop-local:
	$(MAKE) stop ENV=local

stop-local-v:
	$(MAKE) stop-v ENV=local

stop-dev:
	$(MAKE) stop ENV=development

stop-dev-v:
	$(MAKE) stop-v ENV=development

stop-prod:
	$(MAKE) stop ENV=production

stop-prod-v:
	$(MAKE) stop-v ENV=production


# ===============================
# 🔁 RESTART
# ===============================
restart: guard-env
	@echo "🔁 Restarting $(ENV) environment..."
	$(COMPOSE) restart app worker

restart-local:
	$(MAKE) restart ENV=local
	$(MAKE) log-app-local

restart-dev:
	$(MAKE) restart ENV=development
	$(MAKE) log-app-dev

restart-prod:
	$(MAKE) restart ENV=production
	$(MAKE) log-app-prod


# ===============================
# 📋 LOG
# ===============================
log-app: guard-env
	$(COMPOSE) logs -f app

log-worker: guard-env
	$(COMPOSE) logs -f worker

log-app-local:
	$(MAKE) log-app ENV=local

log-app-dev:
	$(MAKE) log-app ENV=development

log-app-prod:
	$(MAKE) log-app ENV=production

log-worker-local:
	$(MAKE) log-worker ENV=local

log-worker-dev:
	$(MAKE) log-worker ENV=development

log-worker-prod:
	$(MAKE) log-worker ENV=production


# ===============================
# 🔨 BUILD
# ===============================
build: guard-env
	@echo "🔨 Building $(IMAGE_NAME):$(IMAGE_TAG)..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

build-local:
	$(MAKE) build ENV=local IMAGE_TAG=latest

build-dev: guard-tag
	$(MAKE) build ENV=development IMAGE_TAG=$(IMAGE_TAG)

build-prod: guard-tag
	$(MAKE) build ENV=production IMAGE_TAG=$(IMAGE_TAG)


# ===============================
# 🗄️ MIGRATION
# ===============================
migration: guard-migration-name
	docker run --rm -v $(PWD)/migrations:/migrations migrate/migrate create -ext sql -dir /migrations $(MIGRATION_NAME)

migrate: guard-env
	$(MIGRATE) up

migrate-rollback: guard-env guard-migration-step
	$(MIGRATE) down $(MIGRATION_STEP)

migrate-drop: guard-env
	$(MIGRATE) down --all

migrate-fresh: guard-env
	$(MAKE) migrate-drop ENV=$(ENV)
	$(MAKE) migrate ENV=$(ENV)
	$(MAKE) seed ENV=$(ENV)

migrate-local:
	$(MAKE) migrate ENV=local

migrate-rollback-local: guard-migration-step
	$(MAKE) migrate-rollback ENV=local MIGRATION_STEP=$(MIGRATION_STEP)

migrate-drop-local:
	$(MAKE) migrate-drop ENV=local

migrate-fresh-local:
	$(MAKE) migrate-fresh ENV=local

migrate-dev:
	$(MAKE) migrate ENV=development

migrate-rollback-dev: guard-migration-step
	$(MAKE) migrate-rollback ENV=development MIGRATION_STEP=$(MIGRATION_STEP)

migrate-drop-dev:
	$(MAKE) migrate-drop ENV=development

migrate-fresh-dev:
	$(MAKE) migrate-fresh ENV=development

migrate-prod:
	$(MAKE) migrate ENV=production

migrate-rollback-prod: guard-migration-step
	$(MAKE) migrate-rollback ENV=production MIGRATION_STEP=$(MIGRATION_STEP)

migrate-drop-prod:
	$(MAKE) migrate-drop ENV=production

migrate-fresh-prod:
	$(MAKE) migrate-fresh ENV=production


# ===============================
# 🌱 SEED
# ===============================
seed: guard-env
	$(COMPOSE) run --rm app go run cmd/seed/main.go

seed-local:
	$(MAKE) seed ENV=local

seed-dev:
	$(MAKE) seed ENV=development

seed-prod:
	$(MAKE) seed ENV=production


# ===============================
# 📝 SWAGGER
# ===============================
swag:
	$(COMPOSE) run --rm app sh -c "swag init -g cmd/http/main.go -o docs --parseDependency && swag fmt"


# ===============================
# 🧪 MOCK
# ===============================
mock:
	rm -rf ./mocks
	docker run --rm -v $(PWD):/src -w /src vektra/mockery


# ===============================
# ✅ TEST
# ===============================
test:
	go test \
		./internal/application/... \
		./internal/handler/... \
		./internal/infrastructure/cache/... \
		./internal/infrastructure/database/...

test-v:
	go test -v \
		./internal/application/... \
		./internal/handler/... \
		./internal/infrastructure/cache/... \
		./internal/infrastructure/database/...

test-cover:
	go test -cover \
		./internal/application/... \
		./internal/handler/... \
		./internal/infrastructure/cache/... \
		./internal/infrastructure/database/...
