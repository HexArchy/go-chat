# Build Stage
FROM golang:1.23-alpine AS builder

# Устанавливаем переменные окружения для производства
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

# Копируем go.mod и go.sum из корня монорепозитория
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь монорепозиторий в контейнер
COPY . .

# Устанавливаем рабочую директорию для сервиса auth
WORKDIR /app/internal/services/auth

# Собираем приложение
RUN go build -o /auth-service ./cmd/auth/main.go

# Собираем команду migrate
RUN go build -o /migrate ./cmd/migrate/main.go

# Финальный этап
FROM alpine:latest

WORKDIR /root/

# Копируем бинарники из builder
COPY --from=builder /auth-service .
COPY --from=builder /migrate .
COPY --from=builder /app/internal/services/auth/configs ./configs

# Открываем порты для HTTP и gRPC
EXPOSE 8080
EXPOSE 9090
EXPOSE 9100

# Устанавливаем точку входа
CMD ["./auth-service", "-config", "./configs/config.prod.yaml"]