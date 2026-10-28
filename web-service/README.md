# Tasks API

REST API для списка задач. Данные хранятся в памяти процесса (без базы) и отдаются только в JSON.

## Запуск

Нужен Go 1.21+.

```bash
cd web-service
go run ./cmd/server
```

Сервер слушает `http://localhost:8080`.

## Контракт задачи

```json
{
  "id": 1,
  "title": "Купить молоко",
  "done": false,
  "created_at": "2026-10-28T08:00:00Z"
}
```

`id` и `created_at` задаёт сервер. `title` обязателен и не может быть пустым. При `PUT` обязательно и поле `done`.

Ошибка всегда выглядит так:

```json
{"error": "сообщение"}
```

Заголовок ответа: `Content-Type: application/json`.

## Эндпоинты

| Метод | Путь | Успех | Ошибки |
| --- | --- | --- | --- |
| GET | `/tasks` | 200, массив задач | 405 |
| POST | `/tasks` | 201, созданная задача | 400, 405 |
| GET | `/tasks/{id}` | 200, задача | 400, 404, 405 |
| PUT | `/tasks/{id}` | 200, обновлённая задача | 400, 404, 405 |
| DELETE | `/tasks/{id}` | 204, пустое тело | 400, 404, 405 |
| GET | `/health` | 200, `{"status":"ok"}` | 405 |

`PUT` заменяет `title` и `done`. `id` и `created_at` не меняются.

## Примеры

Список (пустой):

```bash
curl -i http://localhost:8080/tasks
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

[]
```

Создание:

```bash
curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Купить молоко","done":false}'
```

```http
HTTP/1.1 201 Created
Content-Type: application/json

{"id":1,"title":"Купить молоко","done":false,"created_at":"2026-10-28T08:15:00Z"}
```

Задача по id:

```bash
curl -i http://localhost:8080/tasks/1
```

Полное обновление:

```bash
curl -i -X PUT http://localhost:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Купить молоко","done":true}'
```

Удаление:

```bash
curl -i -X DELETE http://localhost:8080/tasks/1
```

```http
HTTP/1.1 204 No Content
Content-Type: application/json
```

Пустой title:

```bash
curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"  "}'
```

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{"error":"поле title обязательно"}
```

Нет такой задачи:

```bash
curl -i http://localhost:8080/tasks/999
```

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"задача не найдена"}
```

Чужой метод:

```bash
curl -i -X PATCH http://localhost:8080/tasks
```

```http
HTTP/1.1 405 Method Not Allowed
Content-Type: application/json

{"error":"метод не поддерживается"}
```

Проверка живости:

```bash
curl -i http://localhost:8080/health
```

Ручные сценарии с фактическими ответами — в [manual-tests.md](manual-tests.md).
