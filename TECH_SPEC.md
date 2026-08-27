# Задачи первого этапа разработки GophProfile

В этом спринте ваша цель — создать работающий **MVP (Minimum Viable Product)**: сервис должен уметь принимать картинки, сохранять их и отдавать обратно.

## 1. Реализовать ядро сервиса и REST API

Создать HTTP-сервер на Go (**Echo/Chi**), подключить:

- PostgreSQL — для хранения метаданных;
- S3-совместимое хранилище (**MinIO/AWS S3**) — для файлов.

Необходимо реализовать эндпоинты:

- загрузки аватарок;
- получения аватарок;
- удаления аватарок;
- получения метаданных;
- получения списка аватарок пользователя.

Также необходимо интегрировать готовый фронтенд.

### 1.1. Пример API сервиса

#### Загрузка аватарки

```http
POST /api/v1/avatars
Content-Type: multipart/form-data
X-User-ID: string
```

#### Получение аватарки

```http
GET /api/v1/avatars/{avatar_id}
GET /api/v1/users/{user_id}/avatar
```

#### Удаление аватарки

```http
DELETE /api/v1/avatars/{avatar_id}
DELETE /api/v1/users/{user_id}/avatar
X-User-ID: string
```

#### Получение метаданных аватарки

```http
GET /api/v1/avatars/{avatar_id}/metadata
```

#### Список аватарок пользователя

```http
GET /api/v1/users/{user_id}/avatars
```

#### Проверка работоспособности

```http
GET /health
```

#### Веб-интерфейс

```http
GET  /web/upload
POST /web/upload
GET  /web/gallery/{user_id}
```

---

### 1.2. Архитектура и инфраструктура

Компоненты системы:

- HTTP-сервер с REST API и веб-интерфейсом;
- сервис обработки изображений;
- брокер сообщений (**RabbitMQ** или **Kafka**);
- PostgreSQL для хранения метаданных;
- S3-совместимое хранилище для файлов;
- Worker для асинхронной обработки.

Возможная структура проекта:

```text
avatars-service/
├── cmd/
│   ├── server/
│   └── worker/
├── internal/
│   ├── api/
│   ├── config/
│   ├── domain/
│   ├── handlers/
│   ├── repository/
│   ├── services/
│   └── worker/
├── pkg/
├── web/
├── migrations/
├── docker/
├── k8s/
└── tests/
```

---

### 1.3. Технические требования

Основной стек:

- Go 1.21+;
- Echo/Chi для HTTP-роутинга;
- PostgreSQL для метаданных;
- MinIO/AWS S3 для хранения файлов;
- RabbitMQ или Kafka для очередей;
- Docker;
- Docker Compose.

#### Брокер сообщений

На выбор:

**RabbitMQ**

Используйте `exchange` типа:

- `direct`;
- `topic`.

**Kafka**

Создайте топики для обработки изображений.

#### База данных

Используйте PostgreSQL:

- таблицы для хранения метаданных;
- индексы;
- миграции.

---

## 1.4. Функциональные требования

### POST `/api/v1/avatars`

Загрузка аватарки.

Требования:

- валидация формата JPEG, PNG, WebP — опционально;
- ограничение размера файла до 10 MB;
- создание миниатюр:
    - `100x100`;
    - `300x300`;
- асинхронная обработка через брокер сообщений.

Пример запроса:

```http
POST /api/v1/avatars
Content-Type: multipart/form-data
X-User-ID: string
```

Body:

```text
file: binary
```

Ограничение:

```text
max size: 10 MB
```

Успешный ответ:

```http
HTTP/1.1 201 Created
```

```json
{
  "id": "uuid",
  "user_id": "string",
  "url": "string",
  "status": "processing",
  "created_at": "2024-01-01T00:00:00Z"
}
```

Ошибка формата:

```http
HTTP/1.1 400 Bad Request
```

```json
{
  "error": "Invalid file format",
  "details": "Supported formats: jpeg, png, webp"
}
```

Слишком большой файл:

```http
HTTP/1.1 413 Payload Too Large
```

```json
{
  "error": "File too large",
  "max_size": 10485760
}
```

---

### GET `/api/v1/avatars/{avatar_id}`

Получение аватарки.

Опционально поддерживаются query-параметры:

```http
GET /api/v1/avatars/{avatar_id}?size=300x300&format=webp
```

Параметры:

| Параметр | Обязательный | Возможные значения |
|---|---|---|
| `size` | Нет | `100x100`, `300x300`, `original` |
| `format` | Нет | `jpeg`, `png`, `webp` |

Успешный ответ:

```http
HTTP/1.1 200 OK
Content-Type: image/jpeg
Cache-Control: max-age=86400
ETag: "hash"
```

Body:

```text
Binary image data
```

`Content-Type` должен соответствовать фактическому формату изображения:

```text
image/jpeg
image/png
image/webp
```

Если аватарка не найдена:

```http
HTTP/1.1 404 Not Found
```

```json
{
  "error": "Avatar not found"
}
```

---

### GET `/api/v1/avatars/{avatar_id}/metadata`

Получение метаданных аватарки.

Ответ:

```http
HTTP/1.1 200 OK
```

```json
{
  "id": "uuid",
  "user_id": "string",
  "file_name": "avatar.jpg",
  "mime_type": "image/jpeg",
  "size": 1024000,
  "dimensions": {
    "width": 1920,
    "height": 1080
  },
  "thumbnails": [
    {
      "size": "100x100",
      "url": "..."
    },
    {
      "size": "300x300",
      "url": "..."
    }
  ],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### DELETE `/api/v1/avatars/{avatar_id}`

Удаление аватарки.

Требования:

- мягкое удаление записи в БД;
- асинхронное удаление файлов из S3;
- пользователь может удалять только собственные аватарки.

Пример запроса:

```http
DELETE /api/v1/avatars/{avatar_id}
X-User-ID: string
```

Успешный ответ:

```http
HTTP/1.1 204 No Content
```

Если пользователь пытается удалить чужую аватарку:

```http
HTTP/1.1 403 Forbidden
```

```json
{
  "error": "Forbidden",
  "details": "You can only delete your own avatars"
}
```

---

### GET `/health`

Healthcheck сервиса.

Необходимо проверять:

- HTTP-сервис;
- PostgreSQL;
- S3;
- брокер сообщений.

Ответ должен содержать JSON со статусами компонентов.

Пример:

```json
{
  "status": "ok",
  "components": {
    "database": "ok",
    "s3": "ok",
    "broker": "ok"
  }
}
```

---

## Веб-интерфейс

Необходимо реализовать или подключить:

- форму загрузки;
- превью изображения;
- галерею аватарок пользователя.

Опционально:

- drag & drop;
- прогресс загрузки.

### Фронтенд уже готов

Необязательно писать фронтенд с нуля. Можно взять готовое SPA-приложение из репозитория и изменить его на свой вкус.

---

## 1.5. Модель данных

Пример PostgreSQL-схемы:

```sql
CREATE TABLE avatars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    thumbnail_s3_keys JSONB,
    upload_status VARCHAR(50) DEFAULT 'uploading',
    processing_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_avatars_user_id
    ON avatars(user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_avatars_status
    ON avatars(upload_status, processing_status);
```

---

# 2. Настроить асинхронную обработку изображений

Необходимо внедрить брокер сообщений:

- RabbitMQ;
- или Kafka.

Также необходимо реализовать отдельный **Worker-сервис** для:

- создания миниатюр;
- удаления файлов;
- асинхронной обработки изображений.

Worker должен учитывать идемпотентность операций.

## 2.1. События для брокера

### Событие загрузки

```go
type AvatarUploadEvent struct {
    AvatarID string `json:"avatar_id"`
    UserID   string `json:"user_id"`
    S3Key    string `json:"s3_key"`
}
```

### Событие обработки

```go
type AvatarProcessEvent struct {
    AvatarID   string         `json:"avatar_id"`
    Operations []ProcessingOp `json:"operations"`
}
```

### Событие удаления

```go
type AvatarDeleteEvent struct {
    AvatarID string   `json:"avatar_id"`
    S3Keys   []string `json:"s3_keys"`
}
```

---

## 2.2. Идемпотентность

Необходимо:

- использовать уникальные идентификаторы сообщений;
- проверять статус обработки перед выполнением операции;
- реализовать retry;
- использовать экспоненциальный backoff.

---

## 2.3. Примеры

### Отправка события после загрузки

```go
func (s *AvatarService) PublishUploadEvent(
    avatarID,
    userID,
    s3Key string,
) error {
    event := AvatarUploadEvent{
        AvatarID: avatarID,
        UserID:   userID,
        S3Key:    s3Key,
    }

    // RabbitMQ
    return s.publisher.Publish(
        "avatars.exchange",
        "avatar.uploaded",
        event,
    )

    // Kafka
    return s.producer.Send(&sarama.ProducerMessage{
        Topic: "avatar-events",
        Key:   sarama.StringEncoder(avatarID),
        Value: sarama.JSONEncoder(event),
    })
}
```

> В реальном коде, конечно, должен использоваться только один выбранный брокер.

### Обработка события в Worker

```go
func (w *Worker) HandleUploadEvent(event AvatarUploadEvent) error {
    // Получаем метаданные из БД.
    avatar, err := w.repo.GetAvatar(event.AvatarID)
    if err != nil {
        return err
    }

    // Загружаем оригинал из S3.
    image, err := w.s3.Download(event.S3Key)
    if err != nil {
        return err
    }

    // Создаём миниатюры.
    thumbnails := []struct {
        size string
        data []byte
    }{
        {"100x100", w.resizer.Resize(image, 100, 100)},
        {"300x300", w.resizer.Resize(image, 300, 300)},
    }

    // Сохраняем миниатюры в S3.
    for _, thumb := range thumbnails {
        key := fmt.Sprintf(
            "thumbnails/%s/%s.jpg",
            event.AvatarID,
            thumb.size,
        )

        if err := w.s3.Upload(key, thumb.data); err != nil {
            return err
        }
    }

    // Обновляем статус в БД.
    return w.repo.UpdateProcessingStatus(
        event.AvatarID,
        "completed",
    )
}
```

---

# 3. Обеспечить качество и контейнеризацию сервиса

Необходимо:

- написать unit-тесты;
- добиться покрытия тестами более 50%;
- упаковать приложение в Docker-образы;
- настроить запуск окружения через Docker Compose.

## 3.1. Тестирование

Unit-тестами необходимо покрыть:

- HTTP-обработчики;
- сервисный слой;
- репозитории;
- утилиты обработки изображений.

Рекомендуемые инструменты:

- `go test` — unit-тесты;
- `testify/suite` — интеграционные тесты;
- `testcontainers-go` — тестовое окружение;
- `golangci-lint` — статический анализ.

Пример запуска:

```bash
go test ./...
```

С покрытием:

```bash
go test -cover ./...
```

---

## 3.2. Docker

Пример multi-stage `Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/worker .
COPY --from=builder /app/web ./web/
```

---

## Docker Compose для разработки

Через Docker Compose должны запускаться:

- Server;
- Worker;
- PostgreSQL;
- RabbitMQ или Kafka;
- MinIO.

Пример состава:

```text
docker-compose.yml
├── server
├── worker
├── postgres
├── rabbitmq / kafka
└── minio
```

---

# 4. Бонусная задача: обеспечить безопасность

Если основные задачи выполнены раньше срока, можно добавить дополнительные механизмы безопасности и покрыть их тестами.

Необходимо или рекомендуется реализовать:

- валидацию MIME-типов;
- проверку magic bytes;
- ограничение размера файлов;
- rate limiting для API;
- настройку CORS;
- валидацию `X-User-ID`.