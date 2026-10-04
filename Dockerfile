FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder
WORKDIR /app
RUN uname -m

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /build/ ./cmd/...

RUN mkdir -p /out/var/log/gophprofile

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=builder /build/gophprofile-server /usr/local/bin/gophprofile-server
COPY --from=builder /build/gophprofile-worker /usr/local/bin/gophprofile-worker
COPY --from=builder /build/gophprofile-migrate /usr/local/bin/gophprofile-migrate
COPY --from=builder --chmod=755 /app/migrations /app/migrations
COPY --from=builder --chown=nonroot:nonroot /out/var/log/gophprofile /var/log/gophprofile

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gophprofile-server"]
