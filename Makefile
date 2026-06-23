# ====== GLOBAL ======
IMAGE_NAME=ghcr.io/dani-susanto/go-account-service

ENV ?= local
ENV_FILE=.env.$(ENV)
COMPOSE_FILE=docker-compose.$(ENV).yaml

include $(ENV_FILE)
export

DB_URL=postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable

# docker compose wrapper
COMPOSE=docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE)

# ====== BUILD ======
build:
	docker build -t ${IMAGE_NAME} .

build-prod:
	docker build -t ${IMAGE_NAME}:main.latest .

# ====== START ======
start:
	$(COMPOSE) up -d --build
	$(MAKE) log

start-local:
	$(MAKE) start ENV=local

start-dev:
	$(MAKE) start ENV=development

start-prod:
	$(MAKE) start ENV=production

# ====== STOP ======
stop:
	$(COMPOSE) down

stop-v:
	$(COMPOSE) down -v

stop-local:
	$(MAKE) stop ENV=local

stop-dev:
	$(MAKE) stop ENV=development

stop-prod:
	$(MAKE) stop ENV=production

# ====== RESTART ======
restart:
	$(MAKE) stop ENV=$(ENV)
	$(MAKE) start ENV=$(ENV)

# ====== LOG ======
log:
	$(COMPOSE) logs -f app

log-worker:
	$(COMPOSE) logs -f worker

# ====== RELOAD ======
reload:
	$(COMPOSE) restart worker
	$(COMPOSE) restart app
	$(MAKE) log

reload-worker:
	$(COMPOSE) restart worker
	$(MAKE) log-worker


# ====== GO ======
get:
	$(COMPOSE) run --rm app go get $(filter-out $@,$(MAKECMDGOALS))

tidy:
	$(COMPOSE) run --rm app go mod tidy

run:
	$(COMPOSE) run --rm app go run $(filter-out $@,$(MAKECMDGOALS))

# ====== MIGRATION ======
migration:
	$(COMPOSE) run --rm migrate create -ext sql -dir /migrations $(filter-out $@,$(MAKECMDGOALS))

migrate:
	$(COMPOSE) run --rm migrate -path /migrations -database "$(DB_URL)" up

migrate-rollback:
	$(COMPOSE) run --rm migrate -path /migrations -database "$(DB_URL)" down $(filter-out $@,$(MAKECMDGOALS))

migrate-drop:
	$(COMPOSE) run --rm migrate -path /migrations -database "$(DB_URL)" down --all

migrate-fresh:
	$(MAKE) migrate-drop ENV=$(ENV)
	$(MAKE) migrate ENV=$(ENV)
	$(MAKE) seed ENV=$(ENV)

# ====== SEED ======
seed:
	$(COMPOSE) run --rm app go run cmd/seed/main.go

# ====== SWAGGER ======
swag:
	$(COMPOSE) run --rm app sh -c "swag init -g cmd/http/main.go -o docs --parseDependency && swag fmt"

# ====== MOCK ======
mock:
	rm -rf ./mocks
	$(COMPOSE) run --rm mockery

# ====== TEST ======
# ====== TEST ======
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


# ====== IGNORE ======
%:
	@:
