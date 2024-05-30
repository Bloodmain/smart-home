# syntax=docker/dockerfile:1
FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./ ./...

FROM alpine:edge as runner

WORKDIR /app

ENV HTTP_HOST="0.0.0.0"
ENV HTTP_PORT=8080
ENV DATABASE_URL="postgres://postgres:postgres@db:5432/db?sslmode=disable"

COPY --from=builder /app/server .

CMD ["./server"]
