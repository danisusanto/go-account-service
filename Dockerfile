FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /out/app ./cmd/http
RUN go build -o /out/worker ./cmd/worker

FROM alpine:latest

COPY --from=builder /out/app /usr/local/bin/app
COPY --from=builder /out/worker /usr/local/bin/worker
