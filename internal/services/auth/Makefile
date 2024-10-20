.PHONY: build run migrate docker-build docker-run docker-migrate

# Local Go build
build:
	go build -o bin/auth-service ./cmd/auth/main.go

# Local run with local config
run:
	go run ./cmd/auth/main.go -config ./configs/config.prod.yaml

# Local migration
migrate:
	go run ./cmd/migrate/main.go -config ./configs/config.prod.yaml

# Docker build for local development
docker-build:
    docker build -t auth-service:local -f Dockerfile internal/services/auth

# Docker build for production
docker-build-prod:
	docker build -t auth-service:prod --build-arg ENV=prod -f internal/services/auth/Dockerfile .

# Docker run for local environment
docker-run:
	docker run --rm -p 8080:8080 -p 9090:9090 --name auth-service auth-service:local

# Docker run for production
docker-run-prod:
	docker run --rm -p 8080:8080 -p 9090:9090 --name auth-service auth-service:prod

# Docker run migration
docker-migrate:
	docker run --rm auth-service ./migrate -config ./configs/config.prod.yaml
