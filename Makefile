# Source directories
SRC_DIR := cmd

# Default target
.DEFAULT_GOAL := help

# =====================
# HELP
# =====================
.PHONY: help
help:
	@echo "Usage: make <command>"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# =====================
# BUILD TARGETS
# =====================
build: build-server build-client ## Build both server and client

build-server: ## Build the server binary
	go build -o $(SRC_DIR)/server/server $(SRC_DIR)/server/main.go

build-client: ## Build the client binary
	go build -o $(SRC_DIR)/client/client $(SRC_DIR)/client/main.go

# =====================
# RUN TARGETS
# =====================

run-server: ## Run the server using go run
	go run $(SRC_DIR)/server/main.go

run-client: ## Run the client using go run
	go run $(SRC_DIR)/client/main.go

# =====================
# UTILITY
# =====================
clean: ## Remove built binaries
	rm -f $(SRC_DIR)/server/server $(SRC_DIR)/client/client
