# LogSwarm Auth Service

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/PostgreSQL-15-336791?style=for-the-badge&logo=postgresql" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/License-Apache--2.0-yellow?style=for-the-badge" alt="License">
</p>

## Описание

**LogSwarm** — микросервис аутентификации и авторизации пользователей, построенный на принципах чистой архитектуры (Clean Architecture). Сервис обеспечивает полный цикл работы с пользователями: регистрация, аутентификация, управление токенами доступа.

### Ключевые возможности

- **Регистрация пользователей** — создание учётных записей с валидацией данных
- **Аутентификация** — вход по email и паролю с использованием bcrypt
- **JWT-токены** — короткоживущие токены доступа (default: 15 минут)
- **Refresh-токены** — длительные токены для обновления доступа (default: 7 дней)
- **Swagger-документация** — интерактивная документация API

## Архитектура

Проект построен на принципах **Clean Architecture** с чётким разделением слоёв:

```
├── cmd/server/          # Точка входа, запуск приложения
├── internal/
│   ├── domain/         # Бизнес-сущности (User, RefreshToken)
│   ├── usecase/        # Бизнес-логика (AuthService, UserService)
│   ├── repository/    # Работа с данными (UserRepository, TokenRepository)
│   ├── handlers/      # HTTP-обработчики (Auth, User)
│   ├── middleware/    # Промежуточное ПО (Auth middleware)
│   └── config/         # Конфигурация приложения
├── generated/db/       # Сгенерированный код (sqlc)
├── migrations/          # Миграции базы данных
└── docs/               # Swagger-документация
```

### Используемые технологии

| Компонент | Технология | Версия |
|-----------|------------|--------|
| Язык | Go | 1.26+ |
| База данных | PostgreSQL | 15+ |
| HTTP-роутер | gorilla/mux | v1.8.1 |
| JWT | golang-jwt/jwt | v5 |
| Хеширование | bcrypt | (via golang.org/x/crypto) |
| СУБД-клиент | jackc/pgx | v5.9.2 |
| Конфигурация | spf13/viper | v1.21.0 |
| Документация | swaggo/swag | v1.16.6 |
| Контейнеризация | Docker + docker-compose | — |

## API Reference

### Аутентификация

#### Регистрация пользователя

```http
POST /api/auth/register
Content-Type: application/json

{
  "name": "ivanov",
  "email": "ivanov@example.com",
  "password": "securePassword123"
}
```

**Успешный ответ (201 Created)**

```json
{
  "id": 1,
  "name": "ivanov",
  "email": "ivanov@example.com",
  "createdAt": "2026-04-30T12:00:00Z",
  "isActive": false
}
```

#### Вход в систему

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "ivanov@example.com",
  "password": "securePassword123"
}
```

**Успешный ответ (200 OK)**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4e5f6...",
  "user": {
    "id": 1,
    "name": "ivanov",
    "email": "ivanov@example.com",
    "createdAt": "2026-04-30T12:00:00Z",
    "isActive": false
  }
}
```

#### Обновление токена

```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "a1b2c3d4e5f6..."
}
```

**Успешный ответ (200 OK)**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "f6e5d4c3b2a1..."
}
```

### Защищённые endpoints

Для доступа к защищённым endpoints необходимо передавать JWT-токен в заголовке:

```http
Authorization: Bearer <access_token>
```

## Быстрый старт

### Требования

- Go 1.26+
- Docker и Docker Compose
- PostgreSQL 15+ (или Docker)

### Запуск

```bash
# Запуск PostgreSQL
docker-compose up -d postgres

# Запуск сервера
task build | go build -o ./bin/server cmd/server/main.go
tasl run | go run cmd/server/main.go 
```

### Конфигурация

Параметры конфигурации могут быть переданы через переменные окружения или файл `config.env`:

| Параметр | Описание | По умолчанию |
|----------|-----------|--------------|
| `SERVER_HOST` | Хост сервера | `0.0.0.0` |
| `SERVER_PORT` | Порт сервера | `8080` |
| `DATABASE_HOST` | Хост PostgreSQL | `localhost` |
| `DATABASE_PORT` | Порт PostgreSQL | `5432` |
| `DATABASE_USER` | Пользователь БД | `postgres` |
| `DATABASE_PASSWORD` | Пароль БД | `secret` |
| `DATABASE_NAME` | Имя БД | `auth` |
| `JWT_SECRET` | Секретный ключ для JWT | `default-secret-key-change-in-production` |
| `JWT_ACCESS_EXPIRY` | Время жизни access-токена | `15m` |
| `JWT_REFRESH_EXPIRY` | Время жизни refresh-токена | `168h` |


## Разработка

### Генерация кода

```bash
# Генерация SQL-кода (sqlc)
task generate

# Генерация Swagger-документации
task swagger
```

### Миграции

```bash
# Применение миграций
go run cmd/server/main.go --migrate

# Создание новой миграции
touch migrations/002_<name>.sql
```

## Безопасность

### Рекомендации для production

1. **JWT_SECRET** — использовать криптографически стойкий секретный ключ (минимум 256 бит)
2. **HTTPS** — обязательно использовать TLS для production-развёртывания
3. **Rate limiting** — реализовать ограничение частоты запросов
4. **Логирование** — настроить централизованный сбор логов
5. **Мониторинг** — интегр��ровать метрики (Prometheus, Grafana)

### Алерты безопасности

- ❌ Не использовать значения по умолчанию для секретов
- ❌ Не хранить пароли в открытом виде (используется bcrypt)
- ❌ Не использовать HTTP в production

## Лицензия

Copyright © 2026 LogSwarm. Licensed under [Apache License 2.0](LICENSE).

---

<p align="center">
  <sub>Проект разработан с использованием принципов Clean Architecture</sub>
</p>