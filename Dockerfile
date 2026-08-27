FROM golang:1.26-bookworm AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /build/gophprofile-server ./cmd/gophprofile-server && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /build/gophprofile-worker ./cmd/gophprofile-worker

RUN mkdir -p /out/var/log/gophprofile

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=builder /build/gophprofile-server /usr/local/bin/gophprofile-server
COPY --from=builder /build/gophprofile-worker /usr/local/bin/gophprofile-worker
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder --chown=nonroot:nonroot /out/var/log/gophprofile /var/log/gophprofile

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gophprofile-server"]
