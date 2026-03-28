.PHONY: all build run dev test clean docker-up docker-down

# Variables
APP_NAME := storyflow
BACKEND_DIR := backend
FRONTEND_DIR := frontend

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Main targets
all: deps build

deps:
	cd $(BACKEND_DIR) && $(GOMOD) download
	cd $(FRONTEND_DIR) && npm install

build:
	cd $(BACKEND_DIR) && $(GOBUILD) -o bin/server ./cmd/server
	cd $(FRONTEND_DIR) && npm run build

run:
	cd $(BACKEND_DIR) && $(GOCMD) run ./cmd/server

dev:
	@echo "Starting backend..."
	@cd $(BACKEND_DIR) && $(GOCMD) run ./cmd/server &
	@echo "Starting frontend..."
	@cd $(FRONTEND_DIR) && npm run dev

test:
	cd $(BACKEND_DIR) && $(GOTEST) -v ./...
	cd $(FRONTEND_DIR) && npm test

clean:
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

# Database
db-migrate:
	@echo "Run database migrations..."
	# Add migration commands here

db-reset:
	@echo "Reset database..."
	# Add reset commands here

# ComfyUI
comfyui-start:
	@echo "Starting ComfyUI..."
	# Add ComfyUI start command, e.g.:
	# python $(COMFYUI_DIR)/main.py --listen

# Development
watch:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	cd $(BACKEND_DIR) && air

# Setup
setup:
	cp .env.example .env
	@echo "Please edit .env with your API keys"
	$(MAKE) deps

# Help
help:
	@echo "Available targets:"
	@echo "  make deps       - Install all dependencies"
	@echo "  make build      - Build backend and frontend"
	@echo "  make run        - Run backend server"
	@echo "  make dev        - Run both backend and frontend in development mode"
	@echo "  make test       - Run all tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make docker-up  - Start Docker containers"
	@echo "  make docker-down- Stop Docker containers"
	@echo "  make setup      - Initial project setup"