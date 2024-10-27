# Go-Chat

[![License: CC BY-NC 4.0](https://mirrors.creativecommons.org/presskit/icons/cc.svg?ref=chooser-v1)](https://creativecommons.org/licenses/by-nc/4.0/?ref=chooser-v1) ![BY](https://mirrors.creativecommons.org/presskit/icons/by.svg?ref=chooser-v1) ![NC](https://mirrors.creativecommons.org/presskit/icons/nc.svg?ref=chooser-v1)

Go-Chat — это многофункциональное приложение для обмена сообщениями в реальном времени, разработанное на языке программирования Go. Оно состоит из нескольких микросервисов, обеспечивающих аутентификацию пользователей, управление чатами, профиль пользователей и взаимодействие через веб-интерфейс.

## Известные проблемы
- Большие тайминги ожидания:
  Действия идут слишком долго в рамках докер контейнеров, при тесте в 200 rps на сервис auth на localhost, средний timeout составил 20 мс. (rpc запросы, macbook m2 pro). 
- При загрузке чата его нужно обновить:
  Нужно посмотреть template room_view, при заходе в чат, отправляются токены, устанавливается соединение, но для дальнейшей работы нужно обновить страницу (cmd+r or ctr+r)

## Содержание

- [Go-Chat](#go-chat)
  - [Известные проблемы](#известные-проблемы)
  - [Содержание](#содержание)
  - [Особенности](#особенности)
  - [Архитектура](#архитектура)
  - [Используемые Технологии](#используемые-технологии)
  - [Предварительные Требования](#предварительные-требования)
  - [Установка](#установка)
  - [Конфигурация](#конфигурация)
  - [Сборка и Запуск](#сборка-и-запуск)
  - [Миграции](#миграции)
  - [Генерация Протофайлов](#генерация-протофайлов)
  - [Тестирование](#тестирование)
  - [Линтинг](#линтинг)
  - [Очистка Артефактов](#очистка-артефактов)
  - [Лицензия](#лицензия)

## Особенности

- **Аутентификация Пользователей**: Регистрация, вход и выход пользователей с использованием сервиса аутентификации.
- **Управление Сессиями**: Безопасное управление сессиями через зашифрованные куки.
- **Управление Чатами**: Создание, просмотр, поиск и удаление чатов с помощью сервиса веб-сайта.
- **Реальное Время Чата**: Обмен сообщениями в реальном времени через WebSockets и чат-сервис.
- **Управление Профилем**: Просмотр и редактирование профиля пользователя.
- **Рендеринг Шаблонов**: Динамическое отображение контента с использованием HTML-шаблонов.
- **Логирование**: Структурированное логирование с помощью библиотеки Zap.
- **Грейсфул Шатдаун**: Корректное завершение работы сервисов при выключении.
- **Поддержка Docker**: Контейнеризация для удобного развертывания и масштабирования.

## Архитектура

Go-Chat построен на основе модульной архитектуры с разделением ответственности между микросервисами:

- **Auth Service**: Управление аутентификацией и авторизацией пользователей.
- **Website Service**: Обработка запросов, связанных с чатами и пользователями.
- **Chat Service**: Обеспечение функциональности чата в реальном времени.
- **Frontend Service**: Веб-интерфейс для взаимодействия пользователей.
- **Vault**: Управление секретами и конфиденциальными данными.
- **PostgreSQL**: Реляционная база данных для хранения данных пользователей и чатов.
- **PgAdmin**: Интерфейс для управления базой данных PostgreSQL.
- **Nginx**: Обратный прокси-сервер для маршрутизации запросов.

## Используемые Технологии

- **Go (Golang)**: Основной язык программирования.
- **Docker & Docker Compose**: Контейнеризация и оркестрация микросервисов.
- **Gorilla Mux & WebSocket**: Маршрутизация и поддержка WebSocket-соединений.
- **GORM**: ORM для взаимодействия с PostgreSQL.
- **Zap**: Высокопроизводительное логирование.
- **Protobuf & gRPC**: Генерация API и коммуникация между сервисами.
- **Swagger UI**: Документация API.
- **Staticcheck & Golangci-lint**: Инструменты для статического анализа кода.

## Предварительные Требования

- [Go 1.23](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)
- [PostgreSQL](https://www.postgresql.org/download/) (для локальной разработки)
- [Protoc](https://grpc.io/docs/protoc-installation/) (Protocol Buffers Compiler)

## Установка

1. **Клонирование Репозитория**

    ```bash
    git clone https://github.com/HexArch/go-chat.git
    cd go-chat
    ```

2. **Установка Зависимостей**

    Убедитесь, что вы находитесь в корневой директории проекта и выполните:

    ```bash
    go mod download
    ```

## Конфигурация

Конфигурационные файлы расположены в директории `internal/services/*/configs`. Используется формат YAML для управления настройками.

1. **Создание Конфигурационных Файлов**

    Скопируйте пример конфигурации и адаптируйте её под своё окружение:

    ```bash
    cp internal/services/auth/configs/config.prod.yaml.example internal/services/auth/configs/config.prod.yaml
    cp internal/services/website/configs/config.prod.yaml.example internal/services/website/configs/config.prod.yaml
    cp internal/services/chat/configs/config.prod.yaml.example internal/services/chat/configs/config.prod.yaml
    cp internal/services/frontend/configs/config.prod.yaml.example internal/services/frontend/configs/config.prod.yaml
    ```

2. **Редактирование Конфигураций**

    Откройте каждый файл `config.prod.yaml` и обновите параметры согласно вашему окружению:

    ```yaml
    server:
      http:
        host: "0.0.0.0"
        port: 8080
        read_timeout: 15s
        write_timeout: 15s
        templates_path: "./templates"

    database:
      url: "postgres://gochatuser:gochatpass@postgres:5432/gochat?sslmode=disable"
      max_open_conns: 25
      max_idle_conns: 25
      conn_max_lifetime: 5m

    auth_service:
      address: "http://auth-service:9090"

    website_service:
      address: "http://website-service:9091"

    chat_service:
      address: "http://chat-service:9092"

    session:
      secret: "your-session-secret"
      max_age: 86400  # в секундах

    graceful_shutdown:
      timeout: 30s
    ```

## Сборка и Запуск

Используйте `Makefile` для удобного управления сборкой и запуском сервисов.

1. **Запуск всех сервисов**

    ```bash
    make up
    ```

    Эта команда соберёт и запустит все сервисы, описанные в `docker-compose.yml`.

2. **Остановка всех сервисов**

    ```bash
    make down
    ```

3. **Запуск Swagger UI**

    ```bash
    make swagger
    ```

4. **Сборка всех сервисов без их запуска**

    ```bash
    make build-all
    ```

5. **Сборка конкретного сервиса**

    Например, для сборки `auth-service`:

    ```bash
    make build service=auth-service
    ```

## Миграции

Миграции необходимы для поддержания структуры базы данных в актуальном состоянии.

1. **Запуск всех миграций**

    ```bash
    make migrate-all
    ```

2. **Запуск миграций для конкретного сервиса**

    Например, для `auth-service`:

    ```bash
    make migrate
    ```

    Для `website-service`:

    ```bash
    make migrate-website
    ```

    Для `chat-service`:

    ```bash
    make migrate-chat
    ```

## Генерация Протофайлов

Протофайлы используются для генерации API и клиентских библиотек.

1. **Генерация всех протофайлов**

    ```bash
    make gen
    ```

2. **Генерация протофайлов для конкретного сервиса**

    - Для `website-service`:

        ```bash
        make gen-website
        ```

    - Для `auth-service`:

        ```bash
        make gen-auth
        ```

    - Для `chat-service`:

        ```bash
        make gen-chat
        ```

## Тестирование

Запуск тестов для всего проекта (TODO):

```bash
make test
```

## Линтинг

Используйте `golangci-lint` для статического анализа кода и обнаружения потенциальных ошибок.

1. **Запуск линтера**

    ```bash
    make lint
    ```

2. **Установка `golangci-lint`**

    Если `golangci-lint` ещё не установлен, установите его следующим образом:

    ```bash
    brew install golangci-lint
    ```

    Или используя скрипт установки:

    ```bash
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1
    ```

## Очистка Артефактов

Удаление скомпилированных артефактов и сгенерированных файлов:

```bash
make clean
```

## Лицензия

Этот проект лицензирован под лицензией [Creative Commons Attribution-NonCommercial 4.0 International](https://creativecommons.org/licenses/by-nc/4.0/?ref=chooser-v1).

---

*Этот проект создан [Belyakov Nikita](https://github.com/HexArchy) и лицензирован под [Creative Commons Attribution-NonCommercial 4.0 International](https://creativecommons.org/licenses/by-nc/4.0/?ref=chooser-v1).*