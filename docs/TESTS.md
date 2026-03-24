# Тесты и примеры API
В этом файле представлено подробное описание тестов и примеров API

## Содержание
- [Общее](#общее)
- [Примеры API](#примеры-api)
  - [Проверка сервиса](#проверка-сервиса)
  - [Получение тестового JWT для admin](#получение-тестового-jwt-для-admin)
  - [Получение тестового JWT для user](#получение-тестового-jwt-для-user)
  - [Создание переговорки](#создание-переговорки)
  - [Создание расписания](#создание-расписания)
  - [Получение списка слотов](#получение-списка-слотов)
  - [Создание брони](#создание-брони)
  - [Получение своих броней](#получение-своих-броней)
  - [Отмена брони](#отмена-брони)
  - [Повторная отмена брони](#повторная-отмена-брони)
  - [Проверка запрета: user не может создать переговорку](#проверка-запрета-user-не-может-создать-переговорку)
  - [Проверка запрета: admin не может создать бронь](#проверка-запрета-admin-не-может-создать-бронь)
  - [Проверка конфликта: повторное создание расписания](#проверка-конфликта-повторное-создание-расписания)

## Общее
Запуск всех тестов:

```bash
go test ./...
```

Запуск тестов с покрытием:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func coverage.out
```

Было достигнуто покрытие тестами выше 70%.

## Примеры API
Ниже приведены реальные примеры ручного прогона основных сценариев.

---

### Проверка сервиса

#### Запрос
```bash
curl.exe -i http://localhost:8080/_info
```

#### Ответ
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

---

### Получение тестового JWT для admin

#### Запрос
```bash
curl.exe -i -X POST "http://localhost:8080/dummyLogin" ^
  -H "Content-Type: application/json" ^
  --data-binary "@admin.json"
```

`admin.json`:
```json
{"role":"admin"}
```

#### Ответ
```json
{
  "token": "<ADMIN_JWT>"
}
```

---

### Получение тестового JWT для user

#### Запрос
```bash
curl.exe -i -X POST "http://localhost:8080/dummyLogin" ^
  -H "Content-Type: application/json" ^
  --data-binary "@user.json"
```

`user.json`:
```json
{"role":"user"}
```

#### Ответ
```json
{
  "token": "<USER_JWT>"
}
```

---

### Создание переговорки

#### Запрос
```bash
curl.exe -s -X POST "http://localhost:8080/rooms/create" ^
  -H "Authorization: Bearer <ADMIN_JWT>" ^
  -H "Content-Type: application/json" ^
  --data-binary "@room.json"
```

`room.json`:
```json
{
  "name": "Alpha",
  "description": "Test room",
  "capacity": 6
}
```

#### Пример ответа
```json
{
  "room": {
    "id": "dc4069b8-062c-42f2-aa65-d4bf5f35701c",
    "name": "Alpha",
    "description": "Test room",
    "capacity": 6,
    "createdAt": "2026-03-21T17:46:52.091682Z"
  }
}
```

---

### Создание расписания

#### Запрос
```bash
curl.exe -s -X POST "http://localhost:8080/rooms/<ROOM_ID>/schedule/create" ^
  -H "Authorization: Bearer <ADMIN_JWT>" ^
  -H "Content-Type: application/json" ^
  --data-binary "@schedule.json"
```

`schedule.json`:
```json
{
  "roomId": "<ROOM_ID>",
  "daysOfWeek": [1, 2, 3, 4, 5],
  "startTime": "09:00",
  "endTime": "12:00"
}
```

#### Пример ответа
```json
{
  "schedule": {
    "id": "b01af25a-c837-45f0-9451-65b342009123",
    "roomId": "dc4069b8-062c-42f2-aa65-d4bf5f35701c",
    "daysOfWeek": [1, 2, 3, 4, 5],
    "startTime": "09:00",
    "endTime": "12:00",
    "createdAt": "2026-03-21T17:47:00.030729Z"
  }
}
```

---

### Получение списка слотов

#### Запрос
```bash
curl.exe -s "http://localhost:8080/rooms/<ROOM_ID>/slots/list?date=2026-03-23" ^
  -H "Authorization: Bearer <USER_JWT>"
```

#### Пример ответа
```json
{
  "slots": [
    {
      "id": "2e6f3fba-d317-5a1c-9629-d27e250c6f79",
      "roomId": "dc4069b8-062c-42f2-aa65-d4bf5f35701c",
      "start": "2026-03-23T09:00:00Z",
      "end": "2026-03-23T09:30:00Z"
    },
    {
      "id": "840a54d4-f081-5e67-99a8-834eb2a6d40b",
      "roomId": "dc4069b8-062c-42f2-aa65-d4bf5f35701c",
      "start": "2026-03-23T09:30:00Z",
      "end": "2026-03-23T10:00:00Z"
    }
  ]
}
```

---

### Создание брони

#### Запрос
```bash
curl.exe -s -X POST "http://localhost:8080/bookings/create" ^
  -H "Authorization: Bearer <USER_JWT>" ^
  -H "Content-Type: application/json" ^
  --data-binary "@booking.json"
```

`booking.json`:
```json
{
  "slotId": "<SLOT_ID>",
  "createConferenceLink": true
}
```

#### Пример ответа
```json
{
  "booking": {
    "id": "c60147fc-beb4-451c-91c4-a340b1f02bed",
    "slotId": "2e6f3fba-d317-5a1c-9629-d27e250c6f79",
    "userId": "22222222-2222-2222-2222-222222222222",
    "status": "active",
    "conferenceLink": "https://meet.mock.local/c60147fc-beb4-451c-91c4-a340b1f02bed",
    "createdAt": "2026-03-21T17:47:11.355276Z"
  }
}
```

---

### Получение своих броней

#### Запрос
```bash
curl.exe -i "http://localhost:8080/bookings/my" ^
  -H "Authorization: Bearer <USER_JWT>"
```

#### Пример ответа
```json
{
  "bookings": [
    {
      "id": "c60147fc-beb4-451c-91c4-a340b1f02bed",
      "slotId": "2e6f3fba-d317-5a1c-9629-d27e250c6f79",
      "userId": "22222222-2222-2222-2222-222222222222",
      "status": "active",
      "conferenceLink": "https://meet.mock.local/c60147fc-beb4-451c-91c4-a340b1f02bed",
      "createdAt": "2026-03-21T17:47:11.355276Z"
    }
  ]
}
```

---

### Отмена брони

#### Запрос
```bash
curl.exe -i -X POST "http://localhost:8080/bookings/<BOOKING_ID>/cancel" ^
  -H "Authorization: Bearer <USER_JWT>"
```

#### Пример ответа
```json
{
  "booking": {
    "id": "c60147fc-beb4-451c-91c4-a340b1f02bed",
    "slotId": "2e6f3fba-d317-5a1c-9629-d27e250c6f79",
    "userId": "22222222-2222-2222-2222-222222222222",
    "status": "cancelled",
    "conferenceLink": "https://meet.mock.local/c60147fc-beb4-451c-91c4-a340b1f02bed",
    "createdAt": "2026-03-21T17:47:11.355276Z"
  }
}
```

---

### Повторная отмена брони

#### Запрос
```bash
curl.exe -i -X POST "http://localhost:8080/bookings/<BOOKING_ID>/cancel" ^
  -H "Authorization: Bearer <USER_JWT>"
```

#### Пример ответа
```json
{
  "booking": {
    "id": "c60147fc-beb4-451c-91c4-a340b1f02bed",
    "slotId": "2e6f3fba-d317-5a1c-9629-d27e250c6f79",
    "userId": "22222222-2222-2222-2222-222222222222",
    "status": "cancelled",
    "conferenceLink": "https://meet.mock.local/c60147fc-beb4-451c-91c4-a340b1f02bed",
    "createdAt": "2026-03-21T17:47:11.355276Z"
  }
}
```

Повторная отмена возвращает `200 OK`, что соответствует требованию об идемпотентности.

---

### Проверка запрета: user не может создать переговорку

#### Запрос
```bash
curl.exe -i -X POST "http://localhost:8080/rooms/create" ^
  -H "Authorization: Bearer <USER_JWT>" ^
  -H "Content-Type: application/json" ^
  --data-binary "@room.json"
```

#### Ответ
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{"error":{"code":"FORBIDDEN","message":"forbidden"}}
```

---

### Проверка запрета: admin не может создать бронь

#### Запрос
```bash
curl.exe -i -X POST "http://localhost:8080/bookings/create" ^
  -H "Authorization: Bearer <ADMIN_JWT>" ^
  -H "Content-Type: application/json" ^
  --data-binary "@booking.json"
```

#### Ответ
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{"error":{"code":"FORBIDDEN","message":"forbidden"}}
```

---

### Проверка конфликта: повторное создание расписания

#### Запрос
```bash
curl.exe -i -X POST "http://localhost:8080/rooms/<ROOM_ID>/schedule/create" ^
  -H "Authorization: Bearer <ADMIN_JWT>" ^
  -H "Content-Type: application/json" ^
  --data-binary "@schedule.json"
```

#### Ответ
```http
HTTP/1.1 409 Conflict
Content-Type: application/json

{"error":{"code":"SCHEDULE_EXISTS","message":"schedule for this room already exists and cannot be changed"}}
```
