# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Production Stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata postgresql-client

COPY --from=builder /app/server /app/server

EXPOSE 8000

CMD ["/app/server"]
