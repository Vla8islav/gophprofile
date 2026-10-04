# GophProfile

Avatar storage service: upload a profile picture, get it back in original or
thumbnail size. Metadata lives in PostgreSQL, image bytes in S3-compatible
storage (MinIO), and thumbnail generation runs asynchronously through Kafka.

Three binaries from one repo:
* `cmd/gophprofile-server` - REST API
* `cmd/gophprofile-worker` - Kafka consumer: thumbnails + S3 cleanup
* `cmd/gophprofile-migrate` - goose migration runner (run as a k8s Job)

[monitoring & alerts](docs/monitoring.md) · Swagger UI at `/swagger/index.html`.

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

Frontend: `http://localhost:8080/web/`
Swagger UI: `http://localhost:8080/swagger/index.html`.
MinIO console: `http://localhost:9001` (minioadmin / minioadmin).

Quick smoke test:

```bash
TOKEN=$(curl -s -X POST localhost:8080/api/user/register \
  -H 'Content-Type: application/json' \
  -d '{"login":"me","password":"secret-pass"}' | jq -r .token)

curl -s -X POST localhost:8080/api/v1/avatars \
  -H "Authorization: Bearer $TOKEN" -F "file=@thumb.jpg"

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
MIGRATIONS_FOLDER=./migrations \
go run ./cmd/gophprofile-server        # terminal 1

DATABASE_URI='postgres://gophprofile:gophprofile@localhost:5432/gophprofile?sslmode=disable' \
S3_ENDPOINT=localhost:9000 KAFKA_BROKERS=localhost:19092 \
go run ./cmd/gophprofile-worker        # terminal 2
```

Note: **nothing applies migrations unless told to.** Set
`MIGRATIONS_FOLDER=./migrations` (or `-m`) on the server, or run
`cmd/gophprofile-migrate`, for a fresh database. In compose and k8s this is
already wired.

### Kubernetes - from scratch

Developed against single-node k3s v1.36, but any k8s should work.

**1. Cluster.** On the target Linux host:

```bash
curl -sfL https://get.k3s.io | sh -
```

k3s bundles traefik (ingress) and metrics-server (HPA)

**2. Local tooling & kubeconfig.** You need `docker`, `kubectl`, `helm`, and
ssh access to the node. 

```bash
ssh <user>@<node-ip> sudo cat /etc/rancher/k3s/k3s.yaml > ~/.kube/gophprofile.yaml
sed -i '' 's/127.0.0.1/<node-ip>/' ~/.kube/gophprofile.yaml
export KUBECONFIG=~/.kube/gophprofile.yaml
```

**3. Site-specific values** 

* `k8s-deploy.sh` - the ssh target for the image import;
* `deploy/k8s/30-ingress.yaml` and `41-grafana-ingress.yaml` - hostnames
  embed the node IP (`*.<node-ip>.nip.io`);

**4. Monitoring stack**: installs the ServiceMonitor
CRDs that the app manifests depend on.

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace -f deploy/monitoring/kps-values.yaml
```

**5. Deploy**

```bash
./k8s-deploy.sh
```

builds the image from `HEAD`, imports it into k3s over ssh, then applies
`deploy/k8s/` in order: namespace - postgres/MinIO/Kafka/Jaeger - migration
Job (**gates** the app rollout via `kubectl wait`, its built-in DB retry
absorbs postgres's first-boot init) - server + worker - ingress,
ServiceMonitors, NetworkPolicies.

**6. Verify**: `kubectl get pods -n gophprofile` all Running/Completed, then
the smoke test below; the *GophProfile HTTP* dashboard appears in Grafana
about a minute after the first deploy.

URLs (nip.io resolves `<anything>.<ip>.nip.io` to the embedded IP(192.168.88.37 is just a sample) - no DNS
setup; substitute your node IP):

| | |
|---|---|
| App | `http://gophprofile.192.168.88.37.nip.io/web` |
| Jaeger | `http://jaeger.192.168.88.37.nip.io` |
| Grafana | `http://grafana.192.168.88.37.nip.io` (admin / see `deploy/monitoring/kps-values.yaml`) |
