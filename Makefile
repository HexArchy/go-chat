.PHONY: build run migrate docker-build docker-run docker-migrate up down swagger

# Build the auth service binary
build:
	go build -o bin/auth-service ./internal/services/auth/cmd/auth/main.go

# Run the auth service locally
run:
	go run ./internal/services/auth/cmd/auth/main.go -config ./internal/services/auth/configs/config.prod.yaml

# Run migrations locally
migrate:
	go run ./internal/services/auth/cmd/migrate/main.go -config ./internal/services/auth/configs/config.prod.yaml

# Docker build for the auth service
docker-build:
	docker build -t auth-service:local -f internal/services/auth/Dockerfile .

# Run the auth service in Docker
docker-run:
	docker run --rm -p 8080:8080 -p 9090:9090 --name auth-service auth-service:local

# Run migrations in Docker
docker-migrate:
	docker run --rm auth-service:local ./migrate -config ./configs/config.prod.yaml

# Start all services with Docker Compose
up:
	docker-compose up --build

# Stop all services
down:
	docker-compose down

# Start Swagger UI
swagger:
	docker-compose up swagger-ui
