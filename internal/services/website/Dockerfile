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

# Устанавливаем рабочую директорию для сервиса website
WORKDIR /app/internal/services/website

# Собираем приложение
RUN go build -o /website-service ./cmd/website/main.go

# Собираем команду migrate
RUN go build -o /migrate ./cmd/migrate/main.go

# Финальный этап
FROM alpine:latest

WORKDIR /root/

# Копируем бинарники из builder
COPY --from=builder /website-service .
COPY --from=builder /migrate .
COPY --from=builder /app/internal/services/website/configs ./configs

# Открываем порты для HTTP и gRPC
EXPOSE 8081
EXPOSE 9091
EXPOSE 9101

# Устанавливаем точку входа
CMD ["./website-service", "-config", "./configs/config.prod.yaml"]
