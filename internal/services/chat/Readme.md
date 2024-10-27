# Chat Service

## Overview

The **Chat Service** is a pivotal microservice within the Go-Chat application, responsible for managing real-time chat functionalities. It facilitates seamless WebSocket connections for users to join chat rooms, send and receive messages, and retrieve chat histories. Built with Go and leveraging WebSockets, the service ensures low-latency, high-throughput communication, making it ideal for interactive chat applications. Containerized using Docker, it promotes easy deployment and scalability within a microservices architecture.

## Table of Contents

- [Chat Service](#chat-service)
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
    - [WebSocket Endpoint](#websocket-endpoint)
      - [Connect to Chat Room](#connect-to-chat-room)
      - [Example Usage](#example-usage)
  - [Testing](#testing)
  - [Migrations](#migrations)
  - [TODOs](#todos)

## Features

- **Real-Time Communication**: Enables real-time messaging within chat rooms using WebSockets.
- **Room Management**: Allows users to connect to and disconnect from chat rooms.
- **Message Handling**: Supports sending and receiving messages with proper event handling.
- **Chat History**: Provides mechanisms to retrieve historical messages within a chat room.
- **Authentication Integration**: Integrates with the Auth Service to authenticate users before allowing access to chat functionalities.
- **Logging**: Implements comprehensive logging using Uber's Zap library for monitoring and debugging.
- **Graceful Shutdown**: Ensures that all active connections are properly closed during service shutdown.
- **Docker Support**: Containerized for consistent deployments and scalability.

## Architecture

The Chat Service adheres to a clean architecture paradigm, ensuring a clear separation of concerns and maintainability:

- **Controllers**: Handle HTTP requests, manage WebSocket connections, and delegate tasks to use cases.
- **Use Cases**: Encapsulate business logic for connecting users, disconnecting users, sending messages, and retrieving message histories.
- **Entities**: Define core business objects such as `Room`, `Message`, and `Event`.
- **Clients**: Communicate with external services like Auth Service for user authentication.
- **Storage**: Manages data persistence using PostgreSQL via GORM.
- **Middleware**: Implements authentication and authorization mechanisms for incoming connections.
- **Configuration**: Manages service configurations for different environments.
- **Docker**: Facilitates containerization for consistent and scalable deployments.

## Technologies Used

- **Go (Golang)**: Primary programming language for building the service.
- **WebSockets**: Enables real-time, bidirectional communication between clients and the server.
- **Gorilla Mux**: Powerful URL router and dispatcher for handling HTTP routes.
- **GORM**: ORM library for interacting with PostgreSQL databases.
- **Zap**: High-performance logging library for structured logging.
- **Docker**: Containerization platform for deploying the service.
- **PostgreSQL**: Relational database for storing chat data.
- **pkg/errors**: Enhanced error handling capabilities.
- **UUID**: Universally unique identifiers for entities like users and rooms.

## Prerequisites

- [Go 1.23](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Make](https://www.gnu.org/software/make/)
- [PostgreSQL](https://www.postgresql.org/download/) (for local development)
- [Protobuf Compiler](https://grpc.io/docs/protoc-installation/) (`protoc`)
- [gRPC Plugins for Go](https://grpc.io/docs/languages/go/quickstart/#installing-grpc)

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/HexArch/go-chat.git
   cd go-chat/internal/services/chat
   ```

2. **Install Dependencies**

   Ensure you are in the `chat` service directory and run:

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
     http:
       host: "0.0.0.0"
       port: 8082
       read_timeout: 10s
       write_timeout: 10s
       idle_timeout: 60s

   database:
     url: "postgres://your_db_user:your_db_password@localhost:5432/chat_db?sslmode=disable"
     max_open_conns: 25
     max_idle_conns: 25
     conn_max_lifetime: 5m

   auth_service:
     address: "localhost:9090"
     service_token: "your_service_token"

   website_service:
     address: "localhost:9091"
     service_token: "your_service_token"

   graceful_shutdown:
     timeout: 30s
   ```

## Building the Service

### Local Build

Use the provided Makefile for building the service locally.

1. **Build the Chat Service**

   ```bash
   make build
   ```

   This command compiles the `chat-service` binary and places it in the `bin/` directory.

### Docker Build

The service can be containerized using Docker for consistent deployments.

1. **Build Docker Image for Local Development**

   ```bash
   make docker-build
   ```

   This builds the Docker image tagged as `chat-service:local`.

2. **Build Docker Image for Production**

   ```bash
   make docker-build-prod
   ```

   This builds the Docker image tagged as `chat-service:prod`.

## Running the Service

### Local Run

Run the service directly on your machine using the Go command.

1. **Start the Chat Service**

   ```bash
   make run
   ```

   This command runs the `chat-service` with the production configuration.

### Docker Run

Run the service inside a Docker container.

1. **Run Docker Container for Local Environment**

   ```bash
   make docker-run
   ```

   This starts the `chat-service` container, exposing port `8082` (HTTP).

2. **Run Docker Container for Production**

   ```bash
   make docker-run-prod
   ```

   This starts the production version of the `chat-service` container.

## API Documentation

The Chat Service primarily communicates through WebSocket connections. Below is an overview of the available endpoints and their functionalities.

### WebSocket Endpoint

#### Connect to Chat Room

- **Endpoint**: `ws://<host>:8082/ws/chat/{roomID}?token=<access_token>`
- **Description**: Establishes a WebSocket connection to a specific chat room. Users must provide a valid JWT access token for authentication.

- **URL Parameters**:
  - `roomID` (string): The UUID of the chat room to connect to.
  - `token` (string, query parameter): JWT access token obtained from the Auth Service.

- **Request Flow**:
  1. **Authentication**: The server validates the provided JWT token by communicating with the Auth Service.
  2. **Connection Establishment**: Upon successful authentication, a WebSocket connection is established.
  3. **Event Handling**:
     - **User Connected**: Notifies all participants in the room about the new connection.
     - **Message Sending**: Users can send messages which are broadcasted to all room participants.
     - **Chat History**: Users can request historical messages within the room.

- **Message Types**:
  - **Message**:
    ```json
    {
      "type": "message",
      "content": "Hello, everyone!"
    }
    ```
  - **Get History**:
    ```json
    {
      "type": "get_history",
      "limit": 50,
      "offset": 0
    }
    ```

- **Response Messages**:
  - **New Message**:
    ```json
    {
      "type": "new_message",
      "data": {
        "id": "message-uuid",
        "room_id": "room-uuid",
        "user_id": "user-uuid",
        "content": "Hello, everyone!",
        "timestamp": "2024-11-01T00:00:00Z"
      }
    }
    ```
  - **Message History**:
    ```json
    {
      "type": "message_history",
      "data": [
        {
          "id": "message-uuid-1",
          "room_id": "room-uuid",
          "user_id": "user-uuid-1",
          "content": "Hello!",
          "timestamp": "2024-11-01T00:00:00Z"
        },
        {
          "id": "message-uuid-2",
          "room_id": "room-uuid",
          "user_id": "user-uuid-2",
          "content": "Hi there!",
          "timestamp": "2024-11-01T00:01:00Z"
        }
      ]
    }
    ```
  - **Error Event**:
    ```json
    {
      "type": "error",
      "data": {
        "error": "Failed to send message"
      }
    }
    ```
  - **User Connected**:
    ```json
    {
      "type": "user_connected",
      "data": {
        "user_id": "user-uuid",
        "timestamp": "2024-11-01T00:02:00Z"
      }
    }
    ```
  - **User Disconnected**:
    ```json
    {
      "type": "user_disconnected",
      "data": {
        "user_id": "user-uuid",
        "timestamp": "2024-11-01T00:05:00Z"
      }
    }
    ```

#### Example Usage

1. **Establishing a Connection**

   ```javascript
   const socket = new WebSocket('ws://localhost:8082/ws/chat/room-uuid?token=your_jwt_token');

   socket.onopen = () => {
     console.log('Connected to chat room');
   };

   socket.onmessage = (event) => {
     const message = JSON.parse(event.data);
     console.log('Received:', message);
   };

   socket.onclose = () => {
     console.log('Disconnected from chat room');
   };
   ```

2. **Sending a Message**

   ```javascript
   const message = {
     type: "message",
     content: "Hello, everyone!"
   };

   socket.send(JSON.stringify(message));
   ```

3. **Requesting Chat History**

   ```javascript
   const historyRequest = {
     type: "get_history",
     limit: 50,
     offset: 0
   };

   socket.send(JSON.stringify(historyRequest));
   ```

## Testing

To ensure the Chat Service operates correctly, follow these testing procedures:

1. **Run Unit Tests**

   Execute unit tests to verify individual components and use cases:

   ```bash
   go test ./...
   ```

2. **Run Integration Tests**

   Ensure the service is running locally or within Docker, then execute integration tests to validate interactions between components and with external services.

3. **WebSocket Testing**

   Use WebSocket clients or tools like [WebSocket King](https://websocketking.com/) or [Postman](https://www.postman.com/) to interact with the WebSocket endpoint.

   **Example: Connect and Send a Message**

   ```javascript
   const socket = new WebSocket('ws://localhost:8082/ws/chat/room-uuid?token=your_jwt_token');

   socket.onopen = () => {
     console.log('Connected');
     socket.send(JSON.stringify({
       type: "message",
       content: "Hello, World!"
     }));
   };

   socket.onmessage = (event) => {
     console.log('Message:', event.data);
   };
   ```

## Migrations

Database migrations are essential for maintaining the integrity and structure of the database. The Chat Service manages migrations to ensure the database schema aligns with the application's requirements.

1. **Run Migrations Locally**

   Execute the migration command to apply any pending migrations:

   ```bash
   make migrate
   ```

2. **Run Migrations in Docker**

   Apply migrations within a Docker container to ensure consistency across environments:

   ```bash
   make docker-migrate
   ```

   Ensure that the `config.prod.yaml` is correctly configured with your database credentials and accessible from within the Docker container.

## TODOs
- Add tests. 
- Configure CI/CD.