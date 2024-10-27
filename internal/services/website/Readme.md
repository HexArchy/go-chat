# Website Service

## Overview

The **Website Service** is a core microservice within the Go-Chat application responsible for managing chat rooms. It provides a suite of APIs to create, retrieve, update, delete, and search chat rooms. Built with Go and gRPC, the service ensures high performance, scalability, and seamless integration with other services such as the Auth Service. Containerized using Docker, it facilitates easy deployment and maintenance within a microservices architecture.

## Table of Contents

- [Website Service](#website-service)
  - [Overview](#overview)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Architecture](#architecture)
  - [Technologies Used](#technologies-used)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
  - [Building the Service](#building-the-service)
    - [Local Build](#local-build)
    - [Docker Build](#docker-build)
  - [Running the Service](#running-the-service)
    - [Local Run](#local-run)
    - [Docker Run](#docker-run)
  - [API Documentation](#api-documentation)
    - [Room Management Endpoints](#room-management-endpoints)
      - [Create Room](#create-room)
      - [Get Room](#get-room)
      - [Get Owner Rooms](#get-owner-rooms)
      - [Search Rooms](#search-rooms)
      - [Delete Room](#delete-room)
      - [Get All Rooms](#get-all-rooms)
  - [Testing](#testing)
  - [Migrations](#migrations)
  - [TODOs](#todos)

## Features

- **Room Creation**: Create new chat rooms with unique names and assign ownership.
- **Room Retrieval**: Fetch details of individual rooms or lists of rooms with pagination.
- **Room Deletion**: Remove existing rooms, ensuring only authorized users can perform deletions.
- **Room Search**: Search for rooms by name with support for pagination.
- **Ownership Management**: Manage room ownership to control access and modifications.
- **Logging**: Comprehensive logging using Uber's Zap library for monitoring and debugging.
- **gRPC and REST Gateway**: Exposes APIs via gRPC with an HTTP/REST gateway for flexible client integration.
- **Docker Support**: Containerized for easy deployment and scalability.

## Architecture

The Website Service follows a clean architecture pattern, separating concerns into different layers:

- **Protobuf Definitions**: Define the gRPC service and message structures.
- **Controllers**: Handle incoming gRPC requests and delegate to use cases.
- **Use Cases**: Encapsulate business logic for various operations like creating rooms, searching rooms, etc.
- **Entities**: Define core business objects such as `Room`.
- **Middleware**: Provide authentication and authorization mechanisms.
- **Configuration**: Manage service configurations for different environments.
- **Docker**: Facilitate containerization for consistent deployments.

## Technologies Used

- **Go (Golang)**: Primary programming language.
- **gRPC**: High-performance RPC framework for API communication.
- **Protobuf**: Interface definition language for gRPC services.
- **Docker**: Containerization platform for deploying the service.
- **Zap**: High-performance logging library.
- **gRPC-Gateway**: Translates RESTful HTTP API into gRPC.
- **CORS**: Handles Cross-Origin Resource Sharing for HTTP endpoints.
- **GORM**: ORM library for database interactions.
- **PostgreSQL**: Relational database for storing room data.

## Prerequisites

- [Go 1.23](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Make](https://www.gnu.org/software/make/)
- [Protobuf Compiler](https://grpc.io/docs/protoc-installation/) (`protoc`)
- [gRPC Plugins for Go](https://grpc.io/docs/languages/go/quickstart/#installing-grpc)
- [PostgreSQL](https://www.postgresql.org/download/) (for local development)

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/HexArch/go-chat.git
   cd go-chat/internal/services/website
   ```

2. **Install Dependencies**

   Ensure you are in the `website` service directory and run:

   ```bash
   go mod download
   ```

## Configuration

Configuration files are located in the `configs` directory. The service uses YAML for configuration management.

1. **Create Configuration File**

   Duplicate the provided example configuration and customize it as needed:

   ```bash
   cp configs/config.example.yaml configs/config.prod.yaml
   ```

2. **Edit Configuration**

   Open `configs/config.prod.yaml` and update the settings according to your environment:

   ```yaml
   server:
     grpc:
       host: "0.0.0.0"
       port: 9091
     http:
       host: "0.0.0.0"
       port: 8081
       read_timeout: 15s
       write_timeout: 15s

   database:
     url: "postgres://your_db_user:your_db_password@localhost:5432/website_db?sslmode=disable"
     max_open_conns: 25
     max_idle_conns: 25
     conn_max_lifetime: 5m

   auth_service:
     address: "localhost:9090"
     jwt_secret: "your_jwt_secret_key"

   graceful_shutdown:
     timeout: 30s
   ```

## Building the Service

### Local Build

Use the provided Makefile for building the service locally.

1. **Build the Website Service**

   ```bash
   make build
   ```

   This command compiles the `website-service` binary and places it in the `bin/` directory.

### Docker Build

The service can be containerized using Docker for consistent deployments.

1. **Build Docker Image for Local Development**

   ```bash
   make docker-build
   ```

   This builds the Docker image tagged as `website-service:local`.

2. **Build Docker Image for Production**

   ```bash
   make docker-build-prod
   ```

   This builds the Docker image tagged as `website-service:prod`.

## Running the Service

### Local Run

Run the service directly on your machine using the Go command.

1. **Start the Website Service**

   ```bash
   make run
   ```

   This command runs the `website-service` with the production configuration.

### Docker Run

Run the service inside a Docker container.

1. **Run Docker Container for Local Environment**

   ```bash
   make docker-run
   ```

   This starts the `website-service` container, exposing ports `8081` (HTTP) and `9091` (gRPC).

2. **Run Docker Container for Production**

   ```bash
   make docker-run-prod
   ```

   This starts the production version of the `website-service` container.

## API Documentation

The Website Service exposes both gRPC and RESTful HTTP APIs. Below are the available endpoints and their descriptions.

### Room Management Endpoints

#### Create Room

- **gRPC Method**: `CreateRoom`
- **HTTP Endpoint**: `POST /api/v1/rooms`
- **Request Body**:

  ```json
  {
    "name": "General Chat",
    "owner_id": "owner-uuid"
  }
  ```

- **Response**:

  ```json
  {
    "room": {
      "id": "room-uuid",
      "name": "General Chat",
      "owner_id": "owner-uuid",
      "created_at": "2024-11-01T00:00:00Z",
      "updated_at": "2024-11-01T00:00:00Z"
    }
  }
  ```

#### Get Room

- **gRPC Method**: `GetRoom`
- **HTTP Endpoint**: `GET /api/v1/rooms/{room_id}`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response**:

  ```json
  {
    "id": "room-uuid",
    "name": "General Chat",
    "owner_id": "owner-uuid",
    "created_at": "2024-11-01T00:00:00Z",
    "updated_at": "2024-11-01T00:00:00Z"
  }
  ```

#### Get Owner Rooms

- **gRPC Method**: `GetOwnerRooms`
- **HTTP Endpoint**: `GET /api/v1/users/{owner_id}/rooms`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response**:

  ```json
  {
    "rooms": [
      {
        "id": "room-uuid-1",
        "name": "General Chat",
        "owner_id": "owner-uuid",
        "created_at": "2024-11-01T00:00:00Z",
        "updated_at": "2024-11-01T00:00:00Z"
      },
      {
        "id": "room-uuid-2",
        "name": "Random",
        "owner_id": "owner-uuid",
        "created_at": "2024-11-02T00:00:00Z",
        "updated_at": "2024-11-02T00:00:00Z"
      }
    ]
  }
  ```

#### Search Rooms

- **gRPC Method**: `SearchRooms`
- **HTTP Endpoint**: `GET /api/v1/rooms/search?name=chat&limit=10&offset=0`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response**:

  ```json
  {
    "rooms": [
      {
        "id": "room-uuid-1",
        "name": "Chat Room 1",
        "owner_id": "owner-uuid",
        "created_at": "2024-11-01T00:00:00Z",
        "updated_at": "2024-11-01T00:00:00Z"
      },
      {
        "id": "room-uuid-2",
        "name": "Chat Room 2",
        "owner_id": "owner-uuid",
        "created_at": "2024-11-02T00:00:00Z",
        "updated_at": "2024-11-02T00:00:00Z"
      }
    ]
  }
  ```

#### Delete Room

- **gRPC Method**: `DeleteRoom`
- **HTTP Endpoint**: `DELETE /api/v1/rooms/{room_id}`
- **Headers**: `Authorization: Bearer <access_token>`
- **Request Body**:

  ```json
  {
    "room_id": "room-uuid",
    "owner_id": "owner-uuid"
  }
  ```

- **Response**: `Empty`

#### Get All Rooms

- **gRPC Method**: `GetAllRooms`
- **HTTP Endpoint**: `GET /api/v1/rooms?limit=10&offset=0`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response**:

  ```json
  {
    "rooms": [
      {
        "id": "room-uuid-1",
        "name": "General Chat",
        "owner_id": "owner-uuid",
        "created_at": "2024-11-01T00:00:00Z",
        "updated_at": "2024-11-01T00:00:00Z"
      },
      {
        "id": "room-uuid-2",
        "name": "Random",
        "owner_id": "owner-uuid",
        "created_at": "2024-11-02T00:00:00Z",
        "updated_at": "2024-11-02T00:00:00Z"
      }
    ]
  }
  ```

## Testing

To ensure the Website Service functions correctly, follow these steps:

1. **Run Unit Tests**

   ```bash
   go test ./...
   ```

2. **Run Integration Tests**

   Ensure the service is running locally or within Docker, then execute integration tests as per your testing framework.

3. **API Testing**

   Use tools like [Postman](https://www.postman.com/) or [cURL](https://curl.se/) to interact with the APIs.

   **Example: Create Room**

   ```bash
   curl -X POST http://localhost:8081/api/v1/rooms \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <access_token>" \
     -d '{
       "name": "General Chat",
       "owner_id": "owner-uuid"
     }'
   ```

## Migrations

Database migrations are managed via the `migrate` command.

1. **Run Migrations Locally**

   ```bash
   make migrate
   ```

2. **Run Migrations in Docker**

   ```bash
   make docker-migrate
   ```

   Ensure the `config.prod.yaml` is correctly configured with your database credentials.

## TODOs
- Add tests. 
- Configure CI/CD.