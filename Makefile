.PHONY: docker-migrate docker-migrate-website docker-migrate-chat up down swagger gen gen-website gen-auth gen-chat migrate-all migrate-services build build-all test lint clean help

# Default target
all: up

# Start all services with Docker Compose
up:
	docker-compose up --build

# Stop all services
down:
	docker-compose down

# Start Swagger UI
swagger:
	docker-compose up swagger-ui

# Generate all proto files
gen: gen-website gen-auth gen-chat

# Generate proto files for the Website service
gen-website:
	protoc -I . \
		-I $(GOPATH)/src \
		-I $(GOPATH)/src/github.com/googleapis/api-common-protos \
		--go_out=./internal/api/generated \
		--go-grpc_out=./internal/api/generated \
		--grpc-gateway_out=./internal/api/generated \
		--openapiv2_out=./internal/api/generated \
		internal/api/proto/website/website.proto

# Generate proto files for the Auth service
gen-auth:
	protoc -I . \
		-I $(GOPATH)/src \
		-I $(GOPATH)/src/github.com/googleapis/api-common-protos \
		--go_out=./internal/api/generated \
		--go-grpc_out=./internal/api/generated \
		--grpc-gateway_out=./internal/api/generated \
		--openapiv2_out=./internal/api/generated \
		internal/api/proto/auth/auth.proto

# Generate proto files for the Chat service
gen-chat:
	protoc -I . \
		-I $(GOPATH)/src \
		-I $(GOPATH)/src/github.com/googleapis/api-common-protos \
		--go_out=./internal/api/generated \
		--go-grpc_out=./internal/api/generated \
		--grpc-gateway_out=./internal/api/generated \
		--openapiv2_out=./internal/api/generated \
		internal/api/proto/chat/chat.proto

# Run all migration services
migrate-all: migrate migrate-website migrate-chat

# Run Auth service migrations
migrate:
	docker-compose run --rm migrate

# Run Website service migrations
migrate-website:
	docker-compose run --rm website-migrate

# Run Chat service migrations
migrate-chat:
	docker-compose run --rm chat-migrate

# Build all services without starting them
build-all:
	docker-compose build

# Build a specific service
build:
	docker-compose build $(service)

# Run tests for the project
test:
	go test ./... -cover

# Run linting (предполагается использование golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	go clean
	rm -rf bin/
	rm -rf internal/api/generated/

# Display help
help:
	@echo "Available targets:"
	@echo "  all                      : Default target. Builds and starts all services."
	@echo "  up                       : Start all services with Docker Compose."
	@echo "  down                     : Stop all services."
	@echo "  swagger                  : Start Swagger UI."
	@echo "  gen                      : Generate all proto files."
	@echo "  gen-website              : Generate proto files for the Website service."
	@echo "  gen-auth                 : Generate proto files for the Auth service."
	@echo "  gen-chat                 : Generate proto files for the Chat service."
	@echo "  migrate-all              : Run all migration services."
	@echo "  migrate                  : Run Auth service migrations."
	@echo "  migrate-website          : Run Website service migrations."
	@echo "  migrate-chat             : Run Chat service migrations."
	@echo "  build-all                : Build all services without starting them."
	@echo "  build [service]          : Build a specific service. Example: make build service=auth-service"
	@echo "  test                     : Run tests for the project."
	@echo "  lint                     : Run linting."
	@echo "  clean                    : Clean build artifacts."
	@echo "  help                     : Display this help message."
