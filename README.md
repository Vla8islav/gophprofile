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

curl -s localhost:8080/api/v1/avatars/<uuid>/metadata | jq .status
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