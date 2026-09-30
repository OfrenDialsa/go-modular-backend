# --- STAGE 1: BASE ---
FROM golang:1.26.1-alpine AS base
WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# --- STAGE 2: DEVELOPMENT ---
FROM base AS dev
RUN go install github.com/air-verse/air@v1.64.5
COPY . .
RUN go build -o migrate ./cmd/migrate
RUN mkdir -p /app/docker/air/tmp && chmod -R 777 /app/docker/air/tmp
CMD ["air", "-c", ".air.toml"]

# --- STAGE 3: BUILDER ---
FROM base AS builder
COPY . .
RUN go build -o main .
RUN go build -o migrate ./cmd/migrate

# --- STAGE 4: RUNNER ---
FROM alpine:edge AS runner
ENV GIN_MODE=release
WORKDIR /app

RUN mkdir -p /app/data /app/env /app/license /app/migrations
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations

RUN addgroup -S appgroup && adduser -S appuser -G appgroup && \
    chown -R appuser:appgroup /app
USER appuser

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/main"]