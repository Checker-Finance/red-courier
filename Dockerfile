# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy && \
    go build -o red-courier ./cmd/courier

FROM gcr.io/distroless/base-debian11
WORKDIR /red-courier
COPY --from=builder /app/red-courier .
COPY config.yaml .

ENTRYPOINT ["./red-courier"]