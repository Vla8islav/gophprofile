# GophProfile

Avatar storage service: upload a profile picture, get it back in original or
thumbnail size. Metadata lives in PostgreSQL, image bytes in S3-compatible
storage (MinIO), and thumbnail generation runs asynchronously through Kafka.

Two binaries from one repo: 
* `cmd/gophprofile-server` (REST API, owns DB
migrations)
* `cmd/gophprofile-worker` (Kafka consumer: thumbnails + S3
cleanup).

## How to run

### Everything in Docker

```bash
docker compose up --build
```

Starts postgres, MinIO, single-node Kafka (KRaft), the server on
`http://localhost:8080` and the worker.

```bash
curl localhost:8080/health
# {"status":"ok","components":{"broker":"ok","database":"ok","s3":"ok"}}
```

Swagger UI: `http://localhost:8080/swagger/index.html`.
MinIO console: `http://localhost:9001` (minioadmin / minioadmin).

Quick smoke test:

```bash
TOKEN=$(curl -s -X POST localhost:8080/api/user/register \
  -H 'Content-Type: application/json' \
  -d '{"login":"me","password":"secret-pass"}' | jq -r .token)

curl -s -X POST localhost:8080/api/v1/avatars \
  -H "Authorization: Bearer $TOKEN" -F "file=@photo.jpg"
# -> {"id":"<uuid>", "url":"/api/v1/avatars/<uuid>", "status":"pending", ...}

# a couple of seconds later the worker has done its job:
curl -s localhost:8080/api/v1/avatars/<uuid>/metadata | jq .status
# "completed"
curl -s -o thumb.jpg "localhost:8080/api/v1/avatars/<uuid>?size=100x100"
```

### Binaries on the host (dev loop)

The committed compose file only exposes services inside the compose network.
To run the Go binaries locally, add a `docker-compose.override.yml` that
publishes postgres (5432), MinIO (9000) and a second Kafka listener
advertised as `localhost:19092`, then:

```bash
docker compose up -d postgres minio kafka

DATABASE_URI='postgres://gophprofile:gophprofile@localhost:5432/gophprofile?sslmode=disable' \
S3_ENDPOINT=localhost:9000 KAFKA_BROKERS=localhost:19092 \
go run ./cmd/gophprofile-server        # terminal 1

DATABASE_URI='postgres://gophprofile:gophprofile@localhost:5432/gophprofile?sslmode=disable' \
S3_ENDPOINT=localhost:9000 KAFKA_BROKERS=localhost:19092 \
go run ./cmd/gophprofile-worker        # terminal 2
```

Start the server at least once before the worker on a fresh database — the
server is the sole owner of goose migrations (the worker skips them to avoid
two processes racing `goose.Up`).

### Tests

```bash
go test ./...
```

Integration and e2e suites start real PostgreSQL, MinIO and Redpanda
(Kafka-compatible) via testcontainers, so Docker must be running. If your
Docker cannot run the Ryuk reaper sidecar, disable it once per machine:
`echo "ryuk.disabled=true" >> ~/.testcontainers.properties` — cleanup does
not depend on it.

## API

| Method | Path | Auth | |
|---|---|---|---|
| POST | `/api/user/register`, `/api/user/login` | – | returns Bearer token |
| POST | `/api/v1/avatars` | Bearer | multipart `file`; jpeg/png/webp, ≤10MB, validated by magic bytes |
| GET | `/api/v1/avatars/{id}` | – | `?size=100x100|300x300|original`; ETag/304, thumbnails fall back to original until generated |
| GET | `/api/v1/avatars/{id}/metadata` | – | status, thumbnails, timestamps |
| DELETE | `/api/v1/avatars/{id}` | Bearer | own avatars only; soft delete |
| GET | `/api/v1/users/{id}/avatar[s]` | – | latest avatar / list |
| DELETE | `/api/v1/users/{id}/avatar` | Bearer | |
| GET | `/health` | – | per-component status, 503 when degraded |

## Design notes

- **One Kafka topic** (`avatar-events`, 3 partitions), messages keyed by
  avatar id — same avatar always lands in the same partition, so an upload
  event can never be processed after that avatar's delete event. The topic is
  created explicitly with a pinned partition count (no broker auto-create).
- **At-least-once + idempotent worker**: offsets are committed only after
  handling, so every event may be delivered twice. The worker checks
  `processing_status` before working; S3 deletes are naturally idempotent.
  A poison message is logged, marked `failed` and skipped — it never blocks
  the partition.
- **Publish failures don't fail user requests**: if Kafka is down, the upload
  still returns 201 and the row stays `pending` (findable for reconciliation).
  The strict fix would be a transactional outbox — out of scope here.
- **Soft delete**: `DELETE` sets `deleted_at`; S3 objects are removed
  asynchronously by the worker from the keys carried in the event.
- Thumbnails are square (center-crop), always JPEG (pure-Go WebP encoding
  does not exist), rendered with `x/image/draw` CatmullRom.
