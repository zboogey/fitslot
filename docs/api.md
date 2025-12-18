# Fitslot API

## Auth

### POST /auth/register

- Доступ: public
- Описание: Регистрация нового пользователя и выдача пары access/refresh токенов.
- Request body (JSON):
  - name: string
  - email: string
  - password: string
- Responses:
  - 201: LoginResponse (accessToken, refreshToken, user)
  - 400: { "error": "..." } — невалидный JSON/валидация
  - 409: { "error": "Email already registered" }
  - 500: { "error": "..." }

### POST /auth/login

- Доступ: public
- Описание: Авторизация по email и паролю, возвращает access/refresh токены.
- Request body (JSON):
  - email: string
  - password: string
- Responses:
  - 200: LoginResponse (accessToken, refreshToken, user)
  - 400: { "error": "..." } — невалидный JSON/валидация
  - 401: { "error": "Invalid email or password" }
  - 500: { "error": "..." }

### POST /auth/refresh

- Доступ: public
- Описание: Обновление access‑токена по refresh‑токену.
- Request body (JSON):
  - refresh_token: string
- Responses:
  - 200: { "access_token": string, "user": User }
  - 400: { "error": "refresh_token is required" }
  - 401: { "error": "invalid or expired refresh token" }
  - 404: { "error": "user not found" }
  - 500: { "error": "failed to generate access token" }

## User

### GET /me

- Доступ: user (Authorization: Bearer <access_token>)
- Описание: Получение профиля текущего пользователя.
- Request: без тела.
- Responses:
  - 200: User
  - 401: { "error": "User not authenticated" }
  - 404: { "error": "User not found" }

## Gyms

### GET /gyms

- Доступ: user/admin
- Описание: Получение списка залов.
- Request: без тела.
- Responses:
  - 200: массив Gym
  - 500: { "error": "Failed to fetch gyms" }

### GET /gyms/{gymID}/slots

- Доступ: user/admin
- Описание: Получение слотов зала с информацией о доступности.
- Path params:
  - gymID: integer
- Responses:
  - 200: массив TimeSlotWithAvailability
  - 400: { "error": "Invalid gym ID" }
  - 404: { "error": "Gym not found" }
  - 500: { "error": "Failed to fetch time slots" }

## Bookings (user)

### POST /slots/{slotID}/book

- Доступ: user
- Описание: Бронирование слота.
- Path params:
  - slotID: integer
- Responses:
  - 201 или 200: информация о созданной брони (структура Booking)
  - 400/404/409/500: ошибки (невалидные данные, слот не найден, нет мест, и т.д.)

### POST /bookings/{bookingID}/cancel

- Доступ: user
- Описание: Отмена собственной брони.
- Path params:
  - bookingID: integer
- Responses:
  - 200: обновленная бронь или подтверждение отмены
  - 400/403/404/500: ошибки

### GET /bookings

- Доступ: user
- Описание: Получение списка броней текущего пользователя.
- Responses:
  - 200: массив Booking
  - 500: ошибка сервера

## Admin

### POST /admin/gyms

- Доступ: admin
- Описание: Создание нового зала.
- Request body (JSON):
  - name: string
  - location: string
- Responses:
  - 201: Gym
  - 400: { "error": "..." }
  - 500: { "error": "Failed to create gym" }

### GET /admin/gyms

- Доступ: admin
- Описание: Получение списка залов (админский доступ).
- Responses:
  - 200: массив Gym
  - 500: { "error": "Failed to fetch gyms" }

### POST /admin/gyms/{gymID}/slots

- Доступ: admin
- Описание: Создание слота для зала.
- Path params:
  - gymID: integer
- Request body (JSON):
  - start_time: string (RFC3339)
  - end_time: string (RFC3339)
  - capacity: integer
- Responses:
  - 201: TimeSlot
  - 400: { "error": "Invalid gym ID" } или ошибки формата времени
  - 404: { "error": "Gym not found" }
  - 500: { "error": "Failed to create time slot" }

### GET /admin/gyms/{gymID}/slots

- Доступ: admin
- Описание: Получение всех слотов зала (включая прошедшие).
- Path params:
  - gymID: integer
- Responses:
  - 200: массив TimeSlotWithAvailability
  - 400: { "error": "Invalid gym ID" }
  - 404: { "error": "Gym not found" }
  - 500: { "error": "Failed to fetch time slots" }

### GET /admin/slots/{slotID}/bookings

- Доступ: admin
- Описание: Получение всех броней по конкретному слоту.
- Path params:
  - slotID: integer
- Responses:
  - 200: массив Booking
  - 400/404/500: ошибки

### GET /admin/gyms/{gymID}/bookings

- Доступ: admin
- Описание: Получение всех броней по конкретному залу.
- Path params:
  - gymID: integer
- Responses:
  - 200: массив Booking
  - 400/404/500: ошибки

### GET /admin/analytics/bookings

- Доступ: admin
- Описание: Получение агрегированной аналитики по бронированиям.
- Responses:
  - 200: объект с метриками/агрегациями
  - 500: ошибка сервера

## Health

### GET /health

- Доступ: public
- Описание: Проверка, что сервис жив.
- Responses:
  - 200: { "status": "ok" }
