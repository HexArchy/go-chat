.PHONY: docker-migrate up down swagger gen gen-website gen-auth

# Start all services with Docker Compose
up:
	docker-compose up --build

# Stop all services
down:
	docker-compose down

# Start Swagger UI
swagger:
	docker-compose up swagger-ui

# Generate
gen: gen-website gen-auth gen-chat

gen-website:
	protoc -I . \
		-I $GOPATH/src \
		-I $GOPATH/src/github.com/googleapis/api-common-protos \
		--go_out=./internal/api/generated \
		--go-grpc_out=./internal/api/generated \
		--grpc-gateway_out=./internal/api/generated \
		--openapiv2_out=./internal/api/generated \
		internal/api/proto/website/website.proto
gen-auth:
	protoc -I . \
		-I $GOPATH/src \
		-I $GOPATH/src/github.com/googleapis/api-common-protos \
		--go_out=./internal/api/generated \
		--go-grpc_out=./internal/api/generated \
		--grpc-gateway_out=./internal/api/generated \
		--openapiv2_out=./internal/api/generated \
		internal/api/proto/auth/auth.proto
