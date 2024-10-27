# Auth Service

## Overview

The **Auth Service** is a robust authentication and user management microservice designed for the Go-Chat application. It provides a comprehensive suite of APIs for user registration, login, token management, and user CRUD (Create, Read, Update, Delete) operations. Built with Go and gRPC, it ensures high performance, scalability, and seamless integration with other services within the monorepo architecture.

## Table of Contents

- [Auth Service](#auth-service)
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
    - [Authentication Endpoints](#authentication-endpoints)
      - [Register User](#register-user)
      - [Login](#login)
      - [Refresh Token](#refresh-token)
      - [Logout](#logout)
      - [Validate Token](#validate-token)
    - [User Management Endpoints](#user-management-endpoints)
      - [Get User](#get-user)
      - [Get Users](#get-users)
      - [Update User](#update-user)
      - [Delete User](#delete-user)
  - [Migrations](#migrations)
  - [TODOs](#todos)

## Features

- **User Registration**: Create new user accounts with email, password, and additional profile information.
- **Authentication**: Secure login with JWT-based access and refresh tokens.
- **Token Management**: Refresh and validate tokens to maintain secure sessions.
- **User Management**: Retrieve, update, and delete user profiles with pagination support.
- **Logging**: Comprehensive logging using Uber's Zap library for monitoring and debugging.
- **gRPC and REST Gateway**: Exposes APIs via gRPC with an HTTP/REST gateway for flexible client integration.
- **Docker Support**: Containerized for easy deployment and scalability.

## Architecture

The Auth Service follows a clean architecture pattern, separating concerns into different layers:

- **Protobuf Definitions**: Define the gRPC service and message structures.
- **Controllers**: Handle incoming gRPC requests and delegate to use cases.
- **Use Cases**: Encapsulate business logic for various operations like login, registration, etc.
- **Entities**: Define core business objects.
- **Interceptors**: Provide middleware for authentication and logging.
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

## Prerequisites

- [Go 1.23](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Make](https://www.gnu.org/software/make/)
- [Protobuf Compiler](https://grpc.io/docs/protoc-installation/) (`protoc`)
- [gRPC Plugins for Go](https://grpc.io/docs/languages/go/quickstart/#installing-grpc)

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/HexArch/go-chat.git
   cd go-chat/internal/services/auth
   ```

2. **Install Dependencies**

   Ensure you are in the `auth` service directory and run:

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
       port: 9090
     http:
       host: "0.0.0.0"
       port: 8080

   database:
     host: "localhost"
     port: 5432
     user: "your_db_user"
     password: "your_db_password"
     dbname: "auth_db"

   jwt:
     secret: "your_jwt_secret_key"
     access_token_expiration: 15m
     refresh_token_expiration: 7d

   service:
     token: "your_service_token"
   ```

## Building the Service

### Local Build

Use the provided Makefile for building the service locally.

1. **Build the Auth Service**

   ```bash
   make build
   ```

   This command compiles the `auth-service` binary and places it in the `bin/` directory.

### Docker Build

The service can be containerized using Docker for consistent deployments.

1. **Build Docker Image for Local Development**

   ```bash
   make docker-build
   ```

   This builds the Docker image tagged as `auth-service:local`.

2. **Build Docker Image for Production**

   ```bash
   make docker-build-prod
   ```

   This builds the Docker image tagged as `auth-service:prod`.

## Running the Service

### Local Run

Run the service directly on your machine using the Go command.

1. **Start the Auth Service**

   ```bash
   make run
   ```

   This command runs the `auth-service` with the production configuration.

### Docker Run

Run the service inside a Docker container.

1. **Run Docker Container for Local Environment**

   ```bash
   make docker-run
   ```

   This starts the `auth-service` container, exposing ports `8080` (HTTP) and `9090` (gRPC).

2. **Run Docker Container for Production**

   ```bash
   make docker-run-prod
   ```

   This starts the production version of the `auth-service` container.

## API Documentation

The Auth Service exposes both gRPC and RESTful HTTP APIs. Below are the available endpoints and their descriptions.

### Authentication Endpoints

#### Register User

- **gRPC Method**: `RegisterUser`
- **HTTP Endpoint**: `POST /api/v1/auth/register`
- **Request Body**:

  ```json
  {
    "email": "user@example.com",
    "password": "securepassword",
    "username": "user123",
    "phone": "1234567890",
    "age": 30,
    "bio": "Hello, I am a new user."
  }
  ```

- **Response**: `Empty`

#### Login

- **gRPC Method**: `Login`
- **HTTP Endpoint**: `POST /api/v1/auth/login`
- **Request Body**:

  ```json
  {
    "email": "user@example.com",
    "password": "securepassword"
  }
  ```

- **Response**:

  ```json
  {
    "access_token": "jwt_access_token",
    "refresh_token": "jwt_refresh_token",
    "access_token_expires_at": "2024-11-01T00:00:00Z",
    "refresh_token_expires_at": "2024-11-08T00:00:00Z"
  }
  ```

#### Refresh Token

- **gRPC Method**: `RefreshToken`
- **HTTP Endpoint**: `POST /api/v1/auth/refresh`
- **Request Body**:

  ```json
  {
    "refresh_token": "existing_refresh_token"
  }
  ```

- **Response**:

  ```json
  {
    "access_token": "new_jwt_access_token",
    "refresh_token": "new_jwt_refresh_token",
    "access_token_expires_at": "2024-11-01T00:00:00Z",
    "refresh_token_expires_at": "2024-11-08T00:00:00Z"
  }
  ```

#### Logout

- **gRPC Method**: `Logout`
- **HTTP Endpoint**: `POST /api/v1/auth/logout`
- **Headers**: `Authorization: Bearer <access_token>`
- **Request Body**: `Empty`
- **Response**: `Empty`

#### Validate Token

- **gRPC Method**: `ValidateToken`
- **HTTP Endpoint**: `POST /api/v1/auth/validate`
- **Request Body**:

  ```json
  {
    "token": "jwt_token_to_validate"
  }
  ```

- **Response**:

  ```json
  {
    "user": {
      "id": "user-id",
      "email": "user@example.com",
      "username": "user123",
      "phone": "1234567890",
      "age": 30,
      "bio": "Hello, I am a new user.",
      "permissions": ["READ", "WRITE"],
      "created_at": "2024-10-20T00:00:00Z",
      "updated_at": "2024-10-25T00:00:00Z"
    },
    "permissions": ["READ", "WRITE"]
  }
  ```

### User Management Endpoints

#### Get User

- **gRPC Method**: `GetUser`
- **HTTP Endpoint**: `GET /api/v1/users/{user_id}`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response**:

  ```json
  {
    "id": "user-id",
    "email": "user@example.com",
    "username": "user123",
    "phone": "1234567890",
    "age": 30,
    "bio": "Hello, I am a new user.",
    "permissions": ["READ", "WRITE"],
    "created_at": "2024-10-20T00:00:00Z",
    "updated_at": "2024-10-25T00:00:00Z"
  }
  ```

#### Get Users

- **gRPC Method**: `GetUsers`
- **HTTP Endpoint**: `GET /api/v1/users?limit=10&offset=0`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response**:

  ```json
  {
    "users": [
      {
        "id": "user-id-1",
        "email": "user1@example.com",
        "username": "user1",
        "phone": "1234567890",
        "age": 25,
        "bio": "Bio of user1",
        "permissions": ["READ"],
        "created_at": "2024-10-20T00:00:00Z",
        "updated_at": "2024-10-25T00:00:00Z"
      },
      {
        "id": "user-id-2",
        "email": "user2@example.com",
        "username": "user2",
        "phone": "0987654321",
        "age": 28,
        "bio": "Bio of user2",
        "permissions": ["READ", "WRITE"],
        "created_at": "2024-10-21T00:00:00Z",
        "updated_at": "2024-10-26T00:00:00Z"
      }
    ],
    "total": 2
  }
  ```

#### Update User

- **gRPC Method**: `UpdateUser`
- **HTTP Endpoint**: `PUT /api/v1/users/{user_id}`
- **Headers**: `Authorization: Bearer <access_token>`
- **Request Body**:

  ```json
  {
    "email": "newemail@example.com",
    "password": "newsecurepassword",
    "username": "newusername",
    "phone": "1122334455",
    "age": 31,
    "bio": "Updated bio.",
    "permissions": ["READ", "WRITE", "ADMIN"]
  }
  ```

- **Response**: `Empty`

#### Delete User

- **gRPC Method**: `DeleteUser`
- **HTTP Endpoint**: `DELETE /api/v1/users/{user_id}`
- **Headers**: `Authorization: Bearer <access_token>`
- **Response**: `Empty`

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