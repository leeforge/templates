.PHONY: help dev generate swagger migrate clean setup-db setup-redis smoke-full \
	check-config test build build-arm save tag push \
	remote-pull remote-clean local-clean push-compose-file \
	remote-deploy remote-status remote-logs

.DEFAULT_GOAL := help

# Colors
GREEN  = \033[0;32m
YELLOW = \033[0;33m
RED    = \033[0;31m
NC     = \033[0m

APP_NAME ?= leeforge-examples
GO_ENV ?= GOWORK=off GOCACHE=/tmp/go-build-cache
GO_GENERATE_ENV ?= GOFLAGS=-mod=mod

# ============================================================
# Development
# ============================================================

help:
	@echo "$(GREEN)Leeforge Examples Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  $(GREEN)make dev$(NC)          - Run development server (air)"
	@echo "  $(GREEN)make generate$(NC)     - Run code generation (Ent)"
	@echo "  $(GREEN)make swagger$(NC)      - Generate Swagger docs"
	@echo "  $(GREEN)make clean$(NC)        - Clean build artifacts"
	@echo "  $(GREEN)make setup-db$(NC)     - Setup PostgreSQL database"
	@echo "  $(GREEN)make setup-redis$(NC)  - Setup Redis cache"
	@echo "  $(GREEN)make smoke-full$(NC)   - Run realignment verification"
	@echo ""
	@echo "$(YELLOW)Deployment (ENV_MODE=$(GREEN)$(ENV_MODE)$(YELLOW)):$(NC)"
	@echo "  $(GREEN)make check-config$(NC)       - Validate deployment config"
	@echo "  $(GREEN)make build$(NC)              - Build amd64 image via buildx"
	@echo "  $(GREEN)make build-arm$(NC)          - Build image (native arch)"
	@echo "  $(GREEN)make tag$(NC)                - Tag image for registry"
	@echo "  $(GREEN)make push$(NC)               - Push image to registry"
	@echo "  $(GREEN)make remote-deploy$(NC)      - Full remote deployment"
	@echo "  $(GREEN)make remote-status$(NC)      - Check remote compose status"
	@echo "  $(GREEN)make remote-logs$(NC)        - Tail remote logs"

dev:
	@echo "$(GREEN)Starting development server...$(NC)"
	@air -c .air.toml

generate:
	@echo "$(GREEN)Generating Ent code...$(NC)"
	@cd ent && go generate ./...

swagger:
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	@swag init -g cmd/server/main.go -o docs --parseDependency --parseDepth 1

migrate:
	@echo "$(YELLOW)No standalone migration command in examples; use verify script instead.$(NC)"

clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	@rm -f $(APP_NAME)
	@rm -rf tmp
	@$(GO_ENV) go clean -cache

setup-db:
	@echo "$(GREEN)Setting up PostgreSQL database...$(NC)"
	@docker-compose -f ./docker/postgres-18.yaml up -d
	@echo "$(GREEN)PostgreSQL is running on port 15436$(NC)"

setup-redis:
	@echo "$(GREEN)Setting up Redis cache...$(NC)"
	@docker-compose -f ./docker/redis.yaml up -d
	@echo "$(GREEN)Redis is running on port 16379$(NC)"

smoke-full:
	@bash scripts/verify-preview-core-realignment.sh

# ============================================================
# Deployment
# ============================================================

VERSION ?= latest
ENV_MODE ?= test
USE_SUDO ?= true
SUDO_CMD = $(if $(filter 1 true yes on,$(USE_SUDO)),sudo,)
MONOREPO_ROOT ?= .

DEPLOY_COMMON_FILE ?= .deploy.env.common
DEPLOY_ENV_FILE ?= .deploy.env.$(ENV_MODE)
-include $(DEPLOY_COMMON_FILE)
-include $(DEPLOY_ENV_FILE)

LOCAL_COMPOSE_FILE ?= docker/docker-compose.$(ENV_MODE).yaml
ifeq ($(ENV_MODE),local)
LOCAL_COMPOSE_FILE = docker/docker-compose.local.yaml
endif

FULL_REGISTRY_IMAGE = $(REGISTRY_HOST)/$(APP_NAME):$(VERSION)

check-config:
	@test -f $(DEPLOY_COMMON_FILE) || (printf "$(RED)Missing $(DEPLOY_COMMON_FILE)$(NC)\n" && exit 1)
	@test -f $(DEPLOY_ENV_FILE) || (printf "$(RED)Missing $(DEPLOY_ENV_FILE)$(NC)\n" && exit 1)
	@test -n "$(REGISTRY_HOST)" || (printf "$(RED)Missing REGISTRY_HOST$(NC)\n" && exit 1)
	@test -n "$(REMOTE_USER)" || (printf "$(RED)Missing REMOTE_USER$(NC)\n" && exit 1)
	@test -n "$(REMOTE_HOST)" || (printf "$(RED)Missing REMOTE_HOST$(NC)\n" && exit 1)
	@test -n "$(REMOTE_PORT)" || (printf "$(RED)Missing REMOTE_PORT$(NC)\n" && exit 1)
	@test -n "$(REMOTE_COMPOSE_PATH)" || (printf "$(RED)Missing REMOTE_COMPOSE_PATH$(NC)\n" && exit 1)
	@test -n "$(LOCAL_COMPOSE_FILE)" || (printf "$(RED)Missing LOCAL_COMPOSE_FILE$(NC)\n" && exit 1)
	@case "$(REMOTE_PORT)" in ""|*[!0-9]*) printf "$(RED)REMOTE_PORT must be numeric$(NC)\n"; exit 1;; esac
	@if [ "$(REMOTE_PORT)" -lt 1 ] || [ "$(REMOTE_PORT)" -gt 65535 ]; then printf "$(RED)REMOTE_PORT out of range$(NC)\n"; exit 1; fi

test: ## Run local compose smoke
	docker compose -f docker/docker-compose.local.yaml up --build

build-arm: ## Build image from monorepo root
	@printf "$(YELLOW)Building image from monorepo root...$(NC)\n"
	docker build \
		-t $(APP_NAME):$(VERSION) \
		-f docker/Dockerfile \
		$(MONOREPO_ROOT)

build: ## Build amd64 image via buildx
	docker buildx build --platform linux/amd64 \
		-t $(APP_NAME):$(VERSION) \
		-f docker/Dockerfile \
		$(MONOREPO_ROOT)

save: build ## Save image tarball
	docker save $(APP_NAME):$(VERSION) -o ./$(APP_NAME)-$(VERSION).tar
	@printf "$(GREEN)Image saved to ./$(APP_NAME)-$(VERSION).tar$(NC)\n"

tag: check-config build ## Tag image
	docker tag $(APP_NAME):$(VERSION) $(FULL_REGISTRY_IMAGE)

push: check-config tag ## Push image
	docker push $(FULL_REGISTRY_IMAGE)

remote-pull: check-config push ## Pull image on remote host
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "$(SUDO_CMD) docker pull $(FULL_REGISTRY_IMAGE)"

remote-clean: check-config ## Cleanup dangling images on remote host
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "$(SUDO_CMD) docker image prune -f"

local-clean: ## Cleanup local images
	docker rmi $(APP_NAME):$(VERSION) || true
	docker rmi $(FULL_REGISTRY_IMAGE) || true

push-compose-file: check-config push ## Upload compose file to remote host
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p $(REMOTE_COMPOSE_PATH)"
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "chmod 750 $(REMOTE_COMPOSE_PATH) || true"
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "rm -f $(REMOTE_COMPOSE_PATH)/$(APP_NAME).yaml"
	scp -P $(REMOTE_PORT) $(LOCAL_COMPOSE_FILE) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_COMPOSE_PATH)/$(APP_NAME).yaml

remote-deploy: check-config push local-clean push-compose-file ## Deploy on remote host
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_COMPOSE_PATH) && $(SUDO_CMD) docker compose -f $(APP_NAME).yaml down"
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_COMPOSE_PATH) && $(SUDO_CMD) docker compose -f $(APP_NAME).yaml pull"
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_COMPOSE_PATH) && $(SUDO_CMD) docker compose -f $(APP_NAME).yaml up -d"

remote-status: check-config ## Check remote compose status
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_COMPOSE_PATH) && $(SUDO_CMD) docker compose -f $(APP_NAME).yaml ps"

remote-logs: check-config ## Tail recent logs on remote host
	ssh -p $(REMOTE_PORT) $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_COMPOSE_PATH) && $(SUDO_CMD) docker compose -f $(APP_NAME).yaml logs --tail=200"
