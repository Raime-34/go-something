# Ручная проверка

Сервер: `go run ./cmd/server` на `http://localhost:8080`.

В логе на каждый запрос пишутся метод и путь, например `GET /tasks`.

`created_at` ниже — фактическое время прогона (28 октября 2026).

## GET /tasks — пустой список

```bash
curl -i http://localhost:8080/tasks
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

[]
```

## POST /tasks — создание

```bash
curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Купить молоко","done":false}'
```

```http
HTTP/1.1 201 Created
Content-Type: application/json

{"id":1,"title":"Купить молоко","done":false,"created_at":"2026-10-28T08:28:52Z"}
```

## POST /tasks — пустой title

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

## POST /tasks — битый JSON

```bash
curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d 'not-json'
```

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{"error":"некорректный JSON"}
```

## POST /tasks — нет Content-Type

```bash
curl -i -X POST http://localhost:8080/tasks \
  -d '{"title":"без заголовка"}'
```

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{"error":"Content-Type должен быть application/json"}
```

## GET /tasks — список после двух созданий

Вторая задача создана без поля `done` (оно становится `false`):

```bash
curl -i -X POST http://localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Прочитать документацию"}'
```

```http
HTTP/1.1 201 Created
Content-Type: application/json

{"id":2,"title":"Прочитать документацию","done":false,"created_at":"2026-10-28T08:28:52Z"}
```

```bash
curl -i http://localhost:8080/tasks
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

[{"id":1,"title":"Купить молоко","done":false,"created_at":"2026-10-28T08:28:52Z"},{"id":2,"title":"Прочитать документацию","done":false,"created_at":"2026-10-28T08:28:52Z"}]
```

## GET /tasks/{id} — есть задача

```bash
curl -i http://localhost:8080/tasks/1
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"id":1,"title":"Купить молоко","done":false,"created_at":"2026-10-28T08:28:52Z"}
```

## GET /tasks/{id} — нет задачи

```bash
curl -i http://localhost:8080/tasks/999
```

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"задача не найдена"}
```

## GET /tasks/{id} — id не число

```bash
curl -i http://localhost:8080/tasks/abc
```

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{"error":"некорректный id"}
```

## PUT /tasks/{id} — полное обновление

```bash
curl -i -X PUT http://localhost:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Купить молоко","done":true}'
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"id":1,"title":"Купить молоко","done":true,"created_at":"2026-10-28T08:28:52Z"}
```

`id` и `created_at` сохранились, `done` стал `true`.

## PUT /tasks/{id} — нет задачи

```bash
curl -i -X PUT http://localhost:8080/tasks/999 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Нет такой","done":true}'
```

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"задача не найдена"}
```

## PUT /tasks/{id} — нет поля done

```bash
curl -i -X PUT http://localhost:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Только title"}'
```

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{"error":"поле done обязательно"}
```

## DELETE /tasks/{id} — удаление

```bash
curl -i -X DELETE http://localhost:8080/tasks/1
```

```http
HTTP/1.1 204 No Content
Content-Type: application/json
```

## DELETE /tasks/{id} — повторное удаление

```bash
curl -i -X DELETE http://localhost:8080/tasks/1
```

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"задача не найдена"}
```

## PATCH /tasks — неподдерживаемый метод коллекции

```bash
curl -i -X PATCH http://localhost:8080/tasks
```

```http
HTTP/1.1 405 Method Not Allowed
Allow: GET, POST
Content-Type: application/json

{"error":"метод не поддерживается"}
```

## PATCH /tasks/{id} — неподдерживаемый метод элемента

```bash
curl -i -X PATCH http://localhost:8080/tasks/2
```

```http
HTTP/1.1 405 Method Not Allowed
Allow: GET, PUT, DELETE
Content-Type: application/json

{"error":"метод не поддерживается"}
```

## GET /health

```bash
curl -i http://localhost:8080/health
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

## POST /health — чужой метод

```bash
curl -i -X POST http://localhost:8080/health
```

```http
HTTP/1.1 405 Method Not Allowed
Allow: GET
Content-Type: application/json

{"error":"метод не поддерживается"}
```

## GET /tasks — после удаления первой задачи

```bash
curl -i http://localhost:8080/tasks
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

[{"id":2,"title":"Прочитать документацию","done":false,"created_at":"2026-10-28T08:28:52Z"}]
```
