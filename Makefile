#!/usr/bin/make -f

# Project paths
ROOT_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
BACKEND_PATH := $(ROOT_PATH)/backend
FRONTEND_PATH := $(ROOT_PATH)/frontend
DOCKERFILES_PATH := $(ROOT_PATH)/dockerfiles

# Version and build info
VERSION := $(shell cat $(ROOT_PATH)/VERSION 2>/dev/null || echo "v1.0.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)

# Container registry
IMAGE_REGISTRY ?= 77kymo

# Go build environment
GO_PROXY ?= https://goproxy.cn/
GOARCH := $(shell go env GOARCH)
GOOS := linux
CGO_ENABLED ?= 0
GO_BUILD_ENV ?= GOPROXY=${GO_PROXY} GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED}
GO_VERSION := $(shell go version | awk '{print $$3}')
Code_Mode := $(CodeMode)

# Build flags module github.com/kymo-mcp/mcpcan
LDFLAGS := -X 'github.com/kymo-mcp/mcpcan/pkg/version.Version=${VERSION}' \
		-X 'github.com/kymo-mcp/mcpcan/pkg/version.BuildTime=${BUILD_TIME}' \
		-X 'github.com/kymo-mcp/mcpcan/pkg/version.Commit=${COMMIT}' \
		-X 'github.com/kymo-mcp/mcpcan/pkg/version.GoVersion=${GO_VERSION}' \
		-X 'github.com/kymo-mcp/mcpcan/pkg/version.CodeMode=${Code_Mode}'

# Backend build targets
define go_build_service
	@echo "---------- Start Go build $(1) ----------"
	@echo "cd $(BACKEND_PATH) && $(GO_BUILD_ENV) go build -ldflags \"$(LDFLAGS)\" -o $(BACKEND_PATH)/bin/$(1) $(BACKEND_PATH)/cmd/$(1)/main.go"
	@cd $(BACKEND_PATH) && $(GO_BUILD_ENV) go build -ldflags "$(LDFLAGS)" -o $(BACKEND_PATH)/bin/$(1) $(BACKEND_PATH)/cmd/$(1)/main.go
	@echo "---------- End Go build $(1) ----------"
endef

# Docker build targets
define build_docker_image
	@echo "---------- Start Docker build $(1) ----------"
	@echo "cd $(ROOT_PATH) && docker build -t $(IMAGE_REGISTRY)/$(2):$(VERSION) -f $(DOCKERFILES_PATH)/Dockerfile.$(1) ."
	@cd $(ROOT_PATH) && docker build -t $(IMAGE_REGISTRY)/$(2):$(VERSION) -f $(DOCKERFILES_PATH)/Dockerfile.$(1) .
	@echo "---------- End Docker build $(1) ----------"
endef

# Docker push targets
define push_docker_image
	@echo "---------- Start Docker push $(1) ----------"
	@echo "docker push $(IMAGE_REGISTRY)/$(1):$(VERSION)"
	@docker push $(IMAGE_REGISTRY)/$(1):$(VERSION)
	@echo "---------- End Docker push $(1) ----------"
endef

# Default target
.PHONY: all
all: help

.PHONY: print
print:
	@echo "---------- Project Configuration ----------"
	@echo "ROOT_PATH: $(ROOT_PATH)"
	@echo "BACKEND_PATH: $(BACKEND_PATH)"
	@echo "FRONTEND_PATH: $(FRONTEND_PATH)"
	@echo "DOCKERFILES_PATH: $(DOCKERFILES_PATH)"
	@echo "VERSION: $(VERSION)"
	@echo "COMMIT: $(COMMIT)"
	@echo "BUILD_TIME: $(BUILD_TIME)"
	@echo "GO_VERSION: $(GO_VERSION)"
	@echo "IMAGE_REGISTRY: $(IMAGE_REGISTRY)"
	@echo "GO_BUILD_ENV: $(GO_BUILD_ENV)"
	@echo "-------------------------------------------"

# Protocol buffer generation
.PHONY: proto-buf
proto-buf:
	@echo "---- Cleaning existing generated files ----"
	@rm -rf $(shell find $(BACKEND_PATH)/api -type f -name '*.go')
	@rm -rf $(shell find $(BACKEND_PATH)/api -type f -name '*.json')
	@echo "---- Generating protobuf files ----"
	@cd $(BACKEND_PATH)/api && buf --debug generate 
	@find $(BACKEND_PATH)/api -name "*.pb.go" -exec protoc-go-inject-tag -input={} \; || echo "No .pb.go files found for tag injection"
	@echo "---- Merging swagger files ----"
	@rm -f $(BACKEND_PATH)/api/merged.swagger.json
	@if [ -n "$$(find $(BACKEND_PATH)/api -name '*.json' -type f)" ]; then \
		swagger mixin $$(find $(BACKEND_PATH)/api -name '*.json' -type f) -o $(BACKEND_PATH)/api/merged.swagger.json; \
		echo "Swagger files merged successfully"; \
		ls -la $(BACKEND_PATH)/api/merged.swagger.json; \
	else \
		echo "No swagger JSON files found to merge"; \
		touch $(BACKEND_PATH)/api/merged.swagger.json; \
	fi

.PHONY: init
init:
	@echo "---------- Initializing Git Hooks ----------"
	@if [ -d ".git" ]; then \
		git config core.hooksPath .githooks; \
		echo "Git hooks path configured successfully."; \
	else \
		echo "Not a git repository. Skipping hook setup."; \
	fi
	@echo "-------------------------------------------"

.PHONY: pnpm-build
pnpm-build:
	@echo "---------- Start build frontend ----------"
	@echo "cd $(FRONTEND_PATH) && pnpm i && pnpm build"
	@cd $(FRONTEND_PATH) && rm -rf node_modules && CI=true pnpm i && pnpm build
	@echo "---------- End build frontend ----------"

.PHONY: go-mod-tidy
go-mod-tidy:
	@echo "---------- Start go mod tidy ----------"
	@echo "cd $(BACKEND_PATH) && go mod tidy"
	@cd $(BACKEND_PATH) && go mod tidy
	@echo "---------- End go mod tidy ----------"

.PHONY: export-go-build
export-go-build:
	@echo "---------- Extract go build artifacts ----------"
	@# 1. Build intermediate stage image (no final image generated)
	docker build --target builder -t temp-builder -f $(DOCKERFILES_PATH)/Dockerfile.export $(ROOT_PATH)
	@# 2. Create temporary container (not started, only for file extraction)
	docker create --name temp-container temp-builder
	@# 3. Copy files from temporary container to local
	mkdir -p $(ROOT_PATH)/backend/bin
	docker cp temp-container:/app/backend/bin/. $(ROOT_PATH)/backend/bin/
	rm -rf $(ROOT_PATH)/frontend/dist
	docker cp temp-container:/app/frontend/dist $(ROOT_PATH)/frontend/dist
	@# 4. Clean up temporary container
	docker rm -f temp-container
	docker rmi temp-builder
	@echo "---------- Extraction complete, artifacts located at $(ROOT_PATH)/backend/bin and $(ROOT_PATH)/frontend/dist ----------"

.PHONY: go-build-init
go-build-init:
	$(call go_build_service,init)

.PHONY: go-build-market
go-build-market: print
	$(call go_build_service,market)

.PHONY: go-build-authz
go-build-authz:
	$(call go_build_service,authz)

.PHONY: go-build-gateway
go-build-gateway:
	$(call go_build_service,gateway)

.PHONY: go-build-all 
go-build-all: proto-buf go-mod-tidy go-build-init go-build-market go-build-authz go-build-gateway

# Frontend build targets

# All build targets
.PHONY: build-all
build-all: go-build-all pnpm-build

# Builder image specific build and push with latest tag
.PHONY: docker-build-builder
docker-build-builder:
	@echo "---------- Start Docker build builder ----------"
	@echo "cd $(ROOT_PATH) && docker build -t $(IMAGE_REGISTRY)/builder:v2 -f $(DOCKERFILES_PATH)/Dockerfile.builder ."
	@cd $(ROOT_PATH) && docker build -t $(IMAGE_REGISTRY)/builder:v2 -f $(DOCKERFILES_PATH)/Dockerfile.builder .
	@echo "---------- End Docker build builder ----------"

.PHONY: docker-push-builder
docker-push-builder:
	@echo "---------- Start Docker push builder ----------"
	@echo "docker push $(IMAGE_REGISTRY)/builder:v2"
	@docker push $(IMAGE_REGISTRY)/builder:v2
	@echo "docker tag $(IMAGE_REGISTRY)/builder:v2 $(IMAGE_REGISTRY)/builder:latest"
	@docker tag $(IMAGE_REGISTRY)/builder:v2 $(IMAGE_REGISTRY)/builder:latest
	@echo "docker push $(IMAGE_REGISTRY)/builder:latest"
	@docker push $(IMAGE_REGISTRY)/builder:latest
	@echo "---------- End Docker push builder ----------"

.PHONY: docker-build-push-builder
docker-build-push-builder: docker-build-builder docker-push-builder

.PHONY: docker-build-init
docker-build-init:
	$(call build_docker_image,init,mcp-init)

.PHONY: docker-build-market
docker-build-market:
	$(call build_docker_image,market,mcp-market)

.PHONY: docker-build-openapi-to-mcp
docker-build-openapi-to-mcp:
	$(call build_docker_image,openapi-to-mcp,openapi-to-mcp)

.PHONY: docker-build-authz
docker-build-authz:
	$(call build_docker_image,authz,mcp-authz)

.PHONY: docker-build-gateway
docker-build-gateway:
	$(call build_docker_image,gateway,mcp-gateway)

.PHONY: docker-build-frontend
docker-build-frontend:
	$(call build_docker_image,frontend,mcp-web)

.PHONY: docker-build-backend
docker-build-backend:
	$(call build_docker_image,backend,mcp-backend)

.PHONY: docker-build-all
docker-build-all: export-go-build docker-build-init docker-build-market docker-build-openapi-to-mcp docker-build-authz docker-build-gateway docker-build-frontend

.PHONY: docker-push-init
docker-push-init:
	$(call push_docker_image,mcp-init)

.PHONY: docker-push-market
docker-push-market:
	$(call push_docker_image,mcp-market)

.PHONY: docker-push-openapi-to-mcp
docker-push-openapi-to-mcp:
	$(call push_docker_image,openapi-to-mcp)

.PHONY: docker-push-authz
docker-push-authz:
	$(call push_docker_image,mcp-authz)

.PHONY: docker-push-gateway
docker-push-gateway:
	$(call push_docker_image,mcp-gateway)

.PHONY: docker-push-frontend
docker-push-frontend:
	$(call push_docker_image,mcp-web)

.PHONY: docker-push-backend
docker-push-backend:
	$(call push_docker_image,mcp-backend)

.PHONY: docker-push-all
docker-push-all: docker-push-init docker-push-market docker-push-openapi-to-mcp docker-push-authz docker-push-gateway docker-push-frontend

# Docker build and push targets (using existing build and push steps)
.PHONY: docker-build-push-init
docker-build-push-init: docker-build-init docker-push-init

.PHONY: docker-build-push-market
docker-build-push-market: docker-build-market docker-push-market

.PHONY: docker-build-push-openapi-to-mcp
docker-build-push-openapi-to-mcp: docker-build-openapi-to-mcp docker-push-openapi-to-mcp

.PHONY: docker-build-push-authz
docker-build-push-authz: docker-build-authz docker-push-authz

.PHONY: docker-build-push-gateway
docker-build-push-gateway: docker-build-gateway docker-push-gateway

.PHONY: docker-build-push-frontend
docker-build-push-frontend: docker-build-frontend docker-push-frontend

.PHONY: docker-build-push-backend
docker-build-push-backend: docker-build-push-init docker-build-push-market docker-build-push-openapi-to-mcp docker-build-push-authz docker-build-push-gateway

.PHONY: docker-build-push-all
docker-build-push-all: docker-build-all docker-push-all

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BACKEND_PATH)/bin/*
	@rm -rf $(FRONTEND_PATH)/dist
	@rm -rf $(FRONTEND_PATH)/node_modules/.cache

# Test targets
.PHONY: test-backend
test-backend:
	@echo "Running backend tests..."
	@cd $(BACKEND_PATH) && go test ./...

.PHONY: test-frontend
test-frontend:
	@echo "Running frontend tests..."
	@cd $(FRONTEND_PATH) && pnpm test

.PHONY: test-all
test-all: test-backend test-frontend

# Lint targets
.PHONY: lint-backend
lint-backend:
	@echo "Linting backend code..."
	@cd $(BACKEND_PATH) && golangci-lint run

.PHONY: lint-frontend
lint-frontend:
	@echo "Linting frontend code..."
	@cd $(FRONTEND_PATH) && pnpm lint

.PHONY: lint-all
lint-all: lint-backend lint-frontend

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  go-build-init         - Build init service binary [Go]"
	@echo "  go-build-market       - Build market service binary [Go]"
	@echo "  go-build-authz        - Build authz service binary [Go]"
	@echo "  go-build-gateway      - Build gateway service binary [Go]"
	@echo "  go-build-all          - Build all backend services [Go]"
	@echo "  pnpm-build             - Build frontend application [Node]"
	@echo "  build-all                  - Build all services and frontend [Go+Node]"
	@echo ""
	@echo "Docker targets: Builder"
	@echo "  docker-build-builder       - Build builder Docker image"
	@echo "  docker-push-builder        - Push builder Docker image"
	@echo "  docker-build-push-builder  - Build and push builder Docker image with latest tag"
	@echo ""
	@echo "Docker targets: Services"
	@echo "  docker-build-init          - Build init Docker image"
	@echo "  docker-build-market        - Build market Docker image"
	@echo "  docker-build-openapi-to-mcp- Build openapi-to-mcp Docker image"
	@echo "  docker-build-authz         - Build authz Docker image"
	@echo "  docker-build-gateway       - Build gateway Docker image"
	@echo "  docker-build-frontend      - Build frontend Docker image"
	@echo "  docker-build-all           - Build all Docker images"
	@echo "  docker-push-init           - Push init Docker image"
	@echo "  docker-push-market         - Push market Docker image"
	@echo "  docker-push-openapi-to-mcp         - Push openapi-to-mcp Docker image"
	@echo "  docker-push-authz          - Push authz Docker image"
	@echo "  docker-push-gateway        - Push gateway Docker image"
	@echo "  docker-push-frontend       - Push frontend Docker image"
	@echo "  docker-push-all            - Push all Docker images"
	@echo "  docker-build-push-init     - Build and push init Docker image"
	@echo "  docker-build-push-market   - Build and push market Docker image"
	@echo "  docker-build-push-openapi-to-mcp   - Build and push openapi-to-mcp Docker image"
	@echo "  docker-build-push-authz    - Build and push authz Docker image"
	@echo "  docker-build-push-gateway  - Build and push gateway Docker image"
	@echo "  docker-build-push-frontend - Build and push frontend Docker image"
	@echo "  docker-build-push-backend  - Build and push all backend services"
	@echo "  docker-build-push-all      - Build and push all Docker images"
	@echo ""
	@echo "Utility targets:"
	@echo "  proto-buf                  - Generate protobuf and swagger files"
	@echo "  go mod tidy                - Tidy Go modules"
	@echo "  clean                      - Clean build artifacts"
	@echo "  test-all                   - Run all tests"
	@echo "  lint-all                   - Run all linters"
	@echo "  print                      - Print configuration"
	@echo "  help                       - Show this help message"