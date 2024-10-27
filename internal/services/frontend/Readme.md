# Frontend Service

## Overview

The **Frontend Service** is the user-facing component of the Go-Chat application, responsible for rendering web pages, handling user interactions, and interfacing with backend microservices such as Auth, Website, and Chat. Built with Go and leveraging Gorilla Mux for routing and WebSockets for real-time communication, the service ensures a seamless and interactive user experience. It utilizes HTML templates (`.tmpl` files) for dynamic content rendering and manages user sessions securely. Containerized using Docker, the Frontend Service promotes easy deployment and scalability within a microservices architecture.

## Table of Contents

- [Frontend Service](#frontend-service)
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
    - [HTTP Endpoints](#http-endpoints)
      - [Home Page](#home-page)
      - [Registration](#registration)
      - [Login](#login)
      - [Logout](#logout)
      - [User Profile](#user-profile)
      - [Rooms Management](#rooms-management)
  - [WebSocket Endpoint](#websocket-endpoint)
    - [Connect to Chat Room](#connect-to-chat-room)
      - [Connection Flow](#connection-flow)
      - [Message Types](#message-types)
      - [Response Messages](#response-messages)
      - [Example Usage](#example-usage)
  - [Testing](#testing)
  - [Migrations](#migrations)
  - [TODOs](#todos)

## Features

- **User Authentication**: Handles user registration, login, and logout by interfacing with the Auth Service.
- **Session Management**: Manages user sessions securely using encrypted cookies.
- **Room Management**: Allows users to create, view, search, and delete chat rooms by communicating with the Website Service.
- **Real-Time Chat**: Facilitates real-time messaging within chat rooms using WebSockets and the Chat Service.
- **Profile Management**: Enables users to view and edit their profiles.
- **Template Rendering**: Utilizes HTML templates for dynamic content rendering.
- **Logging**: Implements structured logging using Uber's Zap library for monitoring and debugging.
- **Graceful Shutdown**: Ensures that all active connections and sessions are properly closed during service shutdown.
- **Docker Support**: Containerized for consistent deployments and scalability.

## Architecture

The Frontend Service follows a modular and clean architecture, ensuring separation of concerns and maintainability:

- **Controllers**: Handle HTTP requests, manage WebSocket connections, and delegate tasks to use cases.
- **Use Cases**: Encapsulate business logic for authentication, profile management, room management, and real-time chat.
- **Entities**: Define core business objects such as `User`, `Room`, and `Message`.
- **Clients**: Communicate with external services like Auth, Website, and Chat Services.
- **Token Manager**: Manages JWT tokens for user authentication and session management.
- **Middleware**: Implements authentication checks and session management.
- **Templates**: Store HTML templates for rendering web pages.
- **Configuration**: Manages service configurations for different environments.
- **Docker**: Facilitates containerization for consistent and scalable deployments.

## Technologies Used

- **Go (Golang)**: Primary programming language for building the service.
- **Gorilla Mux**: Powerful URL router and dispatcher for handling HTTP routes.
- **Gorilla Sessions**: Manages user sessions using secure cookies.
- **Gorilla WebSocket**: Enables real-time, bidirectional communication between clients and the server.
- **GORM**: ORM library for interacting with PostgreSQL databases.
- **Zap**: High-performance logging library for structured logging.
- **HTML Templates**: Utilizes Go's `html/template` package for dynamic content rendering.
- **Docker**: Containerization platform for deploying the service.
- **PostgreSQL**: Relational database for storing user and room data.

## Prerequisites

- [Go 1.23](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Make](https://www.gnu.org/software/make/)
- [PostgreSQL](https://www.postgresql.org/download/) (for local development)
- [Gorilla Mux](https://github.com/gorilla/mux) and other Go dependencies

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/HexArch/go-chat.git
   cd go-chat/internal/services/frontend
   ```

2. **Install Dependencies**

   Ensure you are in the `frontend` service directory and run:

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
       port: 8080
       read_timeout: 15s
       write_timeout: 15s
       templates_path: "./templates"

   database:
     url: "postgres://your_db_user:your_db_password@localhost:5432/frontend_db?sslmode=disable"
     max_open_conns: 25
     max_idle_conns: 25
     conn_max_lifetime: 5m

   auth_service:
     address: "localhost:9090"

   website_service:
     address: "localhost:9091"

   chat_service:
     address: "localhost:8082"

   session:
     secret: "your_session_secret"
     max_age: 86400  # in seconds

   graceful_shutdown:
     timeout: 30s
   ```

## Building the Service

### Local Build

Use the provided Makefile for building the service locally.

1. **Build the Frontend Service**

   ```bash
   make build
   ```

   This command compiles the `frontend-service` binary and places it in the `bin/` directory.

### Docker Build

The service can be containerized using Docker for consistent deployments.

1. **Build Docker Image for Local Development**

   ```bash
   make docker-build
   ```

   This builds the Docker image tagged as `frontend-service:local`.

2. **Build Docker Image for Production**

   ```bash
   make docker-build-prod
   ```

   This builds the Docker image tagged as `frontend-service:prod`.

## Running the Service

### Local Run

Run the service directly on your machine using the Go command.

1. **Start the Frontend Service**

   ```bash
   make run
   ```

   This command runs the `frontend-service` with the production configuration.

### Docker Run

Run the service inside a Docker container.

1. **Run Docker Container for Local Environment**

   ```bash
   make docker-run
   ```

   This starts the `frontend-service` container, exposing port `8080` (HTTP).

2. **Run Docker Container for Production**

   ```bash
   make docker-run-prod
   ```

   This starts the production version of the `frontend-service` container.

## API Documentation

The Frontend Service primarily serves HTML pages and handles user interactions. It communicates with backend microservices (Auth, Website, Chat) through HTTP clients and WebSockets. Below is an overview of the available HTTP endpoints and their functionalities.

### HTTP Endpoints

#### Home Page

- **Endpoint**: `GET /`
- **Description**: Renders the home page. If the user is authenticated, it displays user-specific information.
- **Template**: `home.tmpl`

#### Registration

- **Render Registration Page**
  - **Endpoint**: `GET /register`
  - **Description**: Displays the user registration form.
  - **Template**: `register.tmpl`

- **Process Registration**
  - **Endpoint**: `POST /register`
  - **Description**: Handles user registration by collecting form data and interfacing with the Auth Service.
  - **Form Data**:
    - `email` (string): User's email address.
    - `password` (string): User's password.
    - `username` (string): Desired username.
    - `phone` (string): User's phone number.
    - `age` (integer): User's age.
    - `bio` (string): User's biography.
  - **Redirect**: Upon successful registration, redirects to the login page.

#### Login

- **Render Login Page**
  - **Endpoint**: `GET /login`
  - **Description**: Displays the user login form.
  - **Template**: `login.tmpl`

- **Process Login**
  - **Endpoint**: `POST /login`
  - **Description**: Handles user authentication by collecting form data and interfacing with the Auth Service.
  - **Form Data**:
    - `email` (string): User's email address.
    - `password` (string): User's password.
  - **Redirect**: Upon successful login, redirects to the rooms list page.

#### Logout

- **Endpoint**: `POST /logout`
- **Description**: Logs out the user by clearing session data and redirecting to the home page.
- **Redirect**: Upon successful logout, redirects to the home page.

#### User Profile

- **View Profile**
  - **Endpoint**: `GET /profile`
  - **Description**: Displays the authenticated user's profile information.
  - **Template**: `profile.tmpl`

- **Edit Profile**
  - **Endpoint**: `GET /profile/edit`
  - **Description**: Displays the profile editing form.
  - **Template**: `profile_edit.tmpl`

  - **Endpoint**: `POST /profile/edit`
  - **Description**: Processes profile updates by collecting form data and interfacing with the Profile Use Case.
  - **Form Data**:
    - `email` (string, optional): New email address.
    - `username` (string, optional): New username.
    - `phone` (string, optional): New phone number.
    - `bio` (string, optional): New biography.
    - `password` (string, optional): New password.
  - **Redirect**: Upon successful update, redirects to the profile page.

#### Rooms Management

- **View User's Rooms**
  - **Endpoint**: `GET /rooms`
  - **Description**: Displays a list of chat rooms owned by the authenticated user.
  - **Template**: `rooms.tmpl`

- **View All Rooms**
  - **Endpoint**: `GET /rooms/all`
  - **Description**: Displays a paginated list of all available chat rooms.
  - **Template**: `all_rooms.tmpl`
  - **Query Parameters**:
    - `limit` (integer, optional): Number of rooms to display per page. Default is `100`.
    - `offset` (integer, optional): Starting point for pagination. Default is `0`.

- **Create Room**
  - **Endpoint**: `GET /rooms/create`
  - **Description**: Displays the room creation form.
  - **Template**: `room_create.tmpl`

  - **Endpoint**: `POST /rooms/create`
  - **Description**: Processes room creation by collecting form data and interfacing with the Room Use Case.
  - **Form Data**:
    - `name` (string): Name of the new chat room.
  - **Redirect**: Upon successful creation, redirects to the newly created room's view page.

- **View Room**
  - **Endpoint**: `GET /rooms/{id}`
  - **Description**: Displays the chat interface for a specific room.
  - **Template**: `room_view.tmpl`
  - **URL Parameters**:
    - `id` (string): UUID of the chat room.

- **Delete Room**
  - **Endpoint**: `POST /rooms/{id}/delete`
  - **Description**: Deletes a specific chat room by interfacing with the Room Use Case.
  - **URL Parameters**:
    - `id` (string): UUID of the chat room.
  - **Redirect**: Upon successful deletion, redirects to the rooms list page.

- **Search Rooms**
  - **Endpoint**: `GET /rooms/search`
  - **Description**: Allows users to search for chat rooms by name.
  - **Template**: `room_search.tmpl`
  - **Query Parameters**:
    - `q` (string): Search query. Must be at least 2 characters long.

## WebSocket Endpoint

The Frontend Service facilitates real-time chat functionalities by establishing WebSocket connections to the Chat Service.

### Connect to Chat Room

- **Endpoint**: `ws://<host>:8080/ws/chat/{roomID}?token=<access_token>`
- **Description**: Establishes a WebSocket connection to a specific chat room. Users must provide a valid JWT access token for authentication.
- **URL Parameters**:
  - `roomID` (string): UUID of the chat room to connect to.
  - `token` (string, query parameter): JWT access token obtained from the Auth Service.

#### Connection Flow

1. **Authentication**: The server validates the provided JWT token by communicating with the Auth Service.
2. **Connection Establishment**: Upon successful authentication, a WebSocket connection is established.
3. **Event Handling**:
   - **User Connected**: Notifies all participants in the room about the new connection.
   - **Message Sending**: Users can send messages which are broadcasted to all room participants.
   - **Chat History**: Users can request historical messages within the room.

#### Message Types

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

#### Response Messages

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
   const socket = new WebSocket('ws://localhost:8080/ws/chat/room-uuid?token=your_jwt_token');

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

To ensure the Frontend Service operates correctly, follow these testing procedures:

1. **Run Unit Tests**

   Execute unit tests to verify individual components and use cases:

   ```bash
   go test ./...
   ```

2. **Run Integration Tests**

   Ensure the service is running locally or within Docker, then execute integration tests to validate interactions between components and with external services.

3. **HTTP Testing**

   Use tools like [Postman](https://www.postman.com/) or [cURL](https://curl.se/) to interact with the HTTP endpoints.

   **Example: Register User**

   ```bash
   curl -X POST http://localhost:8080/register \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "email=user@example.com&password=securepassword&username=user123&phone=1234567890&age=30&bio=Hello, I am a new user."
   ```

4. **WebSocket Testing**

   Use WebSocket clients or tools like [WebSocket King](https://websocketking.com/) or [Postman](https://www.postman.com/) to interact with the WebSocket endpoint.

   **Example: Connect and Send a Message**

   ```javascript
   const socket = new WebSocket('ws://localhost:8080/ws/chat/room-uuid?token=your_jwt_token');

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

Database migrations are essential for maintaining the integrity and structure of the database. The Frontend Service manages migrations to ensure the database schema aligns with the application's requirements.

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