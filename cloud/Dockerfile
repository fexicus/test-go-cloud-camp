# Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV CGO_ENABLED=0
RUN go build -o pickpoint ./cmd/pickpoint/

FROM alpine:latest
WORKDIR /app

# Копируем сам бинарь
COPY --from=builder /app/pickpoint ./pickpoint
RUN chmod +x ./pickpoint

# Копируем config внутрь образа
COPY config/local.yaml /config/local.yaml
RUN chmod 644 /config/local.yaml

EXPOSE 8080
ENTRYPOINT ["./pickpoint"]
