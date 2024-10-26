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

# Устанавливаем рабочую директорию для сервиса frontend
WORKDIR /app/internal/services/frontend

# Создаем директории для шаблонов
RUN mkdir -p /templates/layout /templates/pages

# Собираем приложение
RUN go build -o /frontend-service ./cmd/frontend/main.go

# Финальный этап
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /frontend-service .
COPY --from=builder /app/internal/services/frontend/configs ./configs
COPY --from=builder /app/internal/services/frontend/internal/controllers/http/templates ./templates

# Открываем порт для HTTP
EXPOSE 8084

# Устанавливаем точку входа
CMD ["./frontend-service", "-config", "./configs/config.prod.yaml"]
