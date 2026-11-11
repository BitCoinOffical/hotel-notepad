# Hotel Notepad

Записная книжка отеля. Тестовое задание Backend + Frontend.

## Запуск

1. Клонировать репозиторий:
   ```bash
   git clone https://github.com/BitCoinOffical/hotel-notepad.git
   cd hotel-notepad
   ```

2. Настроить `.env` для обоих backend-сервисов.

3. Запустить тесты:
   ```bash
   make test
   ```

4. Запустить проект:
   ```bash
   docker compose up --build -d
   ```

Открыть frontend: http://localhost:5173/

Backend:
- Auth Service: http://localhost:8080/
- Guestbook Service: http://localhost:8081/

## API

**Auth Service — `/api/v1/auth`**

| Method | Endpoint |
|---|---|
| POST | `/register` |
| POST | `/login` |
| GET | `/logout` |

**Guestbook Service — `/api/v1/guestbook`**

| Method | Endpoint |
|---|---|
| POST | `/entries` |
| GET | `/entries` |
| GET | `/entries/:guest_id` |
| PATCH | `/entries/:guest_id` |
| DELETE | `/entries/:guest_id` |
| GET | `/entries/search` |

Авторизация осуществляется через cookie `session_id`.

## JSON-запросы

**Регистрация — POST `/api/v1/auth/register`**

```json
{
  "username": "testuser",
  "password": "password123",
  "password_retry": "password123"
}
```

**Авторизация — POST `/api/v1/auth/login`**

```json
{
  "username": "testuser",
  "password": "password123"
}
```

**Создание записи — POST `/api/v1/guestbook/entries`**

```json
{
  "first_name": "Ivan",
  "mid_name": "Ivanovich",
  "last_name": "Petrov",
  "note": "Guest note",
  "phone": 79123456789,
  "arrival_date": "2026-12-01T12:00:00Z",
  "departure_date": "2026-12-10T12:00:00Z"
}
```

`note` — необязательное поле.

**Обновление записи — PATCH `/api/v1/guestbook/entries/:guest_id`**

Все поля необязательны:

```json
{
  "first_name": "Alex",
  "note": "Updated note"
}
```

Для дат в текущем DTO используется поле `current_date` для `arrival_date`.

## Стек технологий

**Backend:** Go, Gin, PostgreSQL, Redis, pgx, Squirrel, Goose, CI

**Frontend:** TypeScript, React, Vite, CSS.

**Тестирование:** Testify, GoMock, Testcontainers.

**Инфраструктура:** Docker, Docker Compose.