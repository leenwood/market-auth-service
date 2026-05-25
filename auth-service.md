# ТЗ: auth-service

Сервис аутентификации и авторизации пользователей маркетплейса.

**Стек:** Go 1.26+ · Gin · PostgreSQL 16 (pgx v5) · Redis (хранение refresh-токенов) · bcrypt · JWT RS256

---

## Функциональные требования

1. Регистрация пользователя по email + пароль
2. Вход — выдача пары access/refresh токенов
3. Обновление access-токена по refresh-токену
4. Получение профиля текущего пользователя
5. Выход (инвалидация refresh-токена)
6. Мягкое удаление аккаунта

---

## HTTP API

Base path: `/api/v1/auth`

| Метод | Путь | Тело запроса | Описание |
|---|---|---|---|
| `POST` | `/register` | `RegisterRequest` | Регистрация |
| `POST` | `/login` | `LoginRequest` | Вход |
| `POST` | `/refresh` | `RefreshRequest` | Обновление токена |
| `POST` | `/logout` | `RefreshRequest` | Выход |
| `GET` | `/me` | — (Bearer JWT) | Профиль пользователя |
| `DELETE` | `/me` | — (Bearer JWT) | Удалить аккаунт |

### Системные маршруты

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/health` | Liveness probe |
| `GET` | `/ready` | Readiness probe (ping DB + Redis) |
| `GET` | `/metrics` | Prometheus метрики |

---

## Схемы данных

### RegisterRequest
```json
{
  "email": "user@example.com",
  "password": "минимум 8 символов",
  "name": "Иван Иванов"
}
```

Валидация:
- `email` — обязательный, валидный формат, уникальный
- `password` — обязательный, минимум 8 символов
- `name` — обязательный, 2–100 символов

### LoginRequest
```json
{
  "email": "user@example.com",
  "password": "..."
}
```

### TokenResponse
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 900
}
```

`access_token` — JWT RS256, TTL 15 минут  
`refresh_token` — opaque UUID, TTL 30 дней, хранится в Redis

### RefreshRequest
```json
{
  "refresh_token": "uuid"
}
```

### UserResponse
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "Иван Иванов",
  "role": "buyer",
  "created_at": "RFC3339"
}
```

Роли: `buyer` (покупатель), `seller` (продавец), `admin`.  
При регистрации всегда присваивается `buyer`.

---

## База данных

### Таблица `users`

```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'buyer',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
```

### Refresh-токены (Redis)

Ключ: `refresh:{token_uuid}`  
Значение: `user_id`  
TTL: 30 дней

---

## JWT

- Алгоритм: RS256
- Приватный ключ хранится в переменной окружения `JWT_PRIVATE_KEY`
- Публичный ключ публикуется на `GET /.well-known/jwks.json` — gateway и другие сервисы проверяют токены через него
- Payload claims: `sub` (user_id), `email`, `role`, `iat`, `exp`

---

## Коды ответов

| Код | Ситуация |
|---|---|
| `200` | OK |
| `201` | Пользователь зарегистрирован |
| `400` | Невалидный запрос |
| `401` | Неверный email/пароль или токен истёк |
| `409` | Email уже зарегистрирован |
| `500` | Внутренняя ошибка |

Тело ошибки: `{ "error": "описание" }`

---

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `DATABASE_DSN` | **обязательно** | PostgreSQL DSN |
| `REDIS_ADDR` | `localhost:6379` | Адрес Redis |
| `REDIS_PASSWORD` | `""` | Пароль Redis |
| `REDIS_DB` | `0` | Номер БД Redis |
| `JWT_PRIVATE_KEY` | **обязательно** | RSA приватный ключ (PEM) |
| `JWT_PUBLIC_KEY` | **обязательно** | RSA публичный ключ (PEM) |
| `ACCESS_TOKEN_TTL` | `15m` | TTL access-токена |
| `REFRESH_TOKEN_TTL` | `720h` (30 дней) | TTL refresh-токена |
| `HTTP_ADDR` | `:8081` | Адрес сервера |
| `HTTP_READ_TIMEOUT` | `15s` | |
| `HTTP_WRITE_TIMEOUT` | `15s` | |
| `HTTP_IDLE_TIMEOUT` | `60s` | |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT` | `json` | `json` / `text` |
| `OTEL_ENABLED` | `false` | Включить трейсинг |
| `OTEL_EXPORTER` | `stdout` | `stdout` / `otlp` |
| `OTEL_ENDPOINT` | — | OTLP endpoint |
| `OTEL_SERVICE_NAME` | `auth-service` | |
