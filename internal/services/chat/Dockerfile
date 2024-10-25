# internal/services/chat/Dockerfile
# Build Stage
FROM golang:1.23-alpine AS builder

# Устанавливаем переменные окружения для сборки в Linux
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

# Копируем go.mod и go.sum из корня монорепозитория
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь монорепозиторий в контейнер
COPY . .

# Устанавливаем рабочую директорию для сервиса chat
WORKDIR /app/internal/services/chat

# Собираем приложение
RUN go build -o /chat-service ./cmd/chat/main.go

# Собираем команду migrate
RUN go build -o /migrate ./cmd/migrate/main.go

# Финальный этап
FROM alpine:latest

WORKDIR /root/

# Копируем бинарники из builder
COPY --from=builder /chat-service .
COPY --from=builder /migrate .
COPY --from=builder /app/internal/services/chat/configs ./configs

# Открываем порты для HTTP/WebSocket и gRPC
EXPOSE 8082
EXPOSE 9092

# Устанавливаем точку входа
CMD ["./chat-service", "-config", "./configs/config.prod.yaml"]