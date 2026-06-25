FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o mssql-exporter ./

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/mssql-exporter .
EXPOSE 9187
ENTRYPOINT ["./mssql-exporter"]
