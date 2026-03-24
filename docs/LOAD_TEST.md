# Нагрузочное тестирование
В этоме файле представлено подробное описание реализации нагрузочного тестирования

## Содержание
- [Цель](#цель)
- [Реализация](#реализация)
- [Пример запуска](#пример-запуска)
- [Результаты](#результаты)
- [Вывод](#вывод)

## Цель
Проверить поведение самого нагруженного эндпоинта — получения доступных слотов по переговорке и дате.

Согласно ТЗ, именно этот сценарий является основным по нагрузке, а ориентир по времени ответа составляет 200 мс.

## Реализация
В репозитории добавлен сценарий `scripts/loadtest.js` для k6, ориентированный на самый важный эндпоинт `/rooms/{roomId}/slots/list`.

Для проверки производительности сервиса был выполнен нагрузочный тест endpoint:

- `GET /rooms/{roomId}/slots/list?date=YYYY-MM-DD`

Использовался `k6` со следующим сценарием:
- executor: `constant-arrival-rate`
- rate: `100 req/s`
- duration: `30s`
- preAllocatedVUs: `20`
- maxVUs: `100`

## Пример запуска:

```bash
k6 run scripts/loadtest.js -e BASE_URL=http://localhost:8080 -e TOKEN=<jwt> -e ROOM_ID=<room-id>
```

Либо из Docker Network (я проводил так, из за того, что проверял на Windows Powershell):

```bash
$TOKEN = "<user-jwt>"
$ROOM_ID = "<room-id>"
$DATE = "2026-03-24"

docker run --rm `
  --network room-booking-service_default `
  -v "${PWD}/scripts:/scripts" `
  -e BASE_URL=http://app:8080 `
  -e TOKEN=$TOKEN `
  -e ROOM_ID=$ROOM_ID `
  -e DATE=$DATE `
  grafana/k6 run /scripts/loadtest.js
```

## Результаты 
- всего запросов: 3001
- успешных запросов: 100%
- `http_req_failed`: 0.00%
- `p(95)` по `http_req_duration`: 4.22 ms
- среднее время ответа: 3.19 ms

### Вывод:
Сервис успешно справился с нагрузочным тестированием
