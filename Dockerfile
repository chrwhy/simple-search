# Build stage: compile with CGO + FTS5
FROM golang:1.23-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends gcc musl-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build --tags fts5 -o /bin/simple-search .

# Runtime stage: minimal image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /bin/simple-search .
COPY lib/ lib/
COPY dict/ dict/
COPY data/ data/
COPY .env.example .env

# DB file lives in a volume so it persists across restarts.
VOLUME ["/app/db"]

ENV SS_DB_PATH=/app/db/example.db

ENTRYPOINT ["./simple-search"]
