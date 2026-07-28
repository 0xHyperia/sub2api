.PHONY: bootstrap doctor dev dev-infra-up dev-infra-down deploy-validate integration-config
.PHONY: build build-backend build-frontend test test-backend test-frontend test-frontend-critical

MISE := mise exec --
DEV_COMPOSE := docker compose --env-file deploy/.env.dev -f deploy/compose.dev.yaml

bootstrap:
	command -v mise >/dev/null
	mise install
	$(MISE) pnpm --dir frontend install --frozen-lockfile
	@test -f deploy/.env.dev || cp deploy/.env.dev.example deploy/.env.dev

doctor:
	@command -v docker >/dev/null && docker compose version >/dev/null
	@command -v mise >/dev/null
	@$(MISE) go version
	@$(MISE) pnpm --version

dev-infra-up:
	@test -f deploy/.env.dev || cp deploy/.env.dev.example deploy/.env.dev
	$(DEV_COMPOSE) up -d --wait

dev-infra-down:
	$(DEV_COMPOSE) down --remove-orphans

dev:
	@test -f deploy/.env.dev || { echo "run make dev-infra-up first"; exit 1; }
	@set -a; . deploy/.env.dev; set +a; \
	  trap 'kill 0' EXIT; \
	  (cd backend && AUTO_SETUP=true SERVER_HOST=0.0.0.0 SERVER_PORT=8080 SERVER_MODE=debug \
	    DATABASE_HOST=127.0.0.1 DATABASE_PORT="$${SUB2API_POSTGRES_PORT:-5432}" \
	    DATABASE_USER=sub2api DATABASE_PASSWORD="$$SUB2API_POSTGRES_PASSWORD" DATABASE_DBNAME=sub2api DATABASE_SSLMODE=disable \
	    REDIS_HOST=127.0.0.1 REDIS_PORT="$${SUB2API_REDIS_PORT:-6379}" REDIS_PASSWORD="$$SUB2API_REDIS_PASSWORD" \
	    REDIS_DB=0 REDIS_ENABLE_TLS=false ADMIN_EMAIL="$$SUB2API_ADMIN_EMAIL" ADMIN_PASSWORD="$$SUB2API_ADMIN_PASSWORD" \
	    JWT_SECRET="$$SUB2API_JWT_SECRET" TOTP_ENCRYPTION_KEY="$$SUB2API_TOTP_KEY" \
	    $(MISE) go run ./cmd/server) & \
	  ($(MISE) pnpm --dir frontend dev --host 127.0.0.1 --port 5173) & wait

deploy-validate:
	deploy/production/validate.sh

integration-config:
	docker compose --env-file deploy/.env.dev.example -f deploy/compose.integration.yaml config --quiet

FRONTEND_CRITICAL_VITEST := \
	src/views/auth/__tests__/LinuxDoCallbackView.spec.ts \
	src/views/auth/__tests__/WechatCallbackView.spec.ts \
	src/views/user/__tests__/PaymentView.spec.ts \
	src/views/user/__tests__/PaymentResultView.spec.ts \
	src/components/user/profile/__tests__/ProfileInfoCard.spec.ts \
	src/views/admin/__tests__/SettingsView.spec.ts

# 一键编译前后端
build: build-backend build-frontend

# 编译后端（复用 backend/Makefile）
build-backend:
	@$(MAKE) -C backend build

# 编译前端（需要已安装依赖）
build-frontend:
	@pnpm --dir frontend run build

# 运行测试（后端 + 前端）
test: test-backend test-frontend

test-backend:
	@$(MAKE) -C backend test

test-frontend:
	@pnpm --dir frontend run lint:check
	@pnpm --dir frontend run typecheck
	@$(MAKE) test-frontend-critical

test-frontend-critical:
	@pnpm --dir frontend exec vitest run $(FRONTEND_CRITICAL_VITEST)
