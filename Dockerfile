# ---------- Build stage ----------
FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/build/betteryd ./cmd/betteryd

# ---------- Runtime stage ----------
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates curl jq && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/build/betteryd /usr/local/bin/betteryd

COPY ./shell/init.sh /usr/local/bin/init.sh
COPY ./shell/runNode.sh /usr/local/bin/runNode.sh
COPY ./shell/runFirstNode.sh /usr/local/bin/runFirstNode.sh

RUN chmod +x /usr/local/bin/init.sh
RUN chmod +x /usr/local/bin/runNode.sh
RUN chmod +x /usr/local/bin/runFirstNode.sh

EXPOSE 26656 26657 1317 9090