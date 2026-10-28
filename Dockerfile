# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w \
    -X github.com/viniciussouzao/mssql-exporter/internal/version.Version=${VERSION} \
    -X github.com/viniciussouzao/mssql-exporter/internal/version.Commit=${COMMIT} \
    -X github.com/viniciussouzao/mssql-exporter/internal/version.Date=${DATE}" \
    -o /out/mssql-exporter ./

FROM alpine:3.24

ARG VERSION=dev
LABEL org.opencontainers.image.title="mssql-exporter" \
      org.opencontainers.image.description="Prometheus exporter for Microsoft SQL Server" \
      org.opencontainers.image.source="https://github.com/viniciussouzao/mssql-exporter" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.licenses="Apache-2.0"

RUN apk add --no-cache ca-certificates && \
    addgroup -S exporter && adduser -S -G exporter -h /app exporter

WORKDIR /app
COPY --from=builder --chown=exporter:exporter /out/mssql-exporter .

USER exporter
EXPOSE 10060

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:10060/healthz || exit 1

ENTRYPOINT ["./mssql-exporter"]
