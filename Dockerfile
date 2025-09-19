# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/red-courier ./cmd/courier

FROM gcr.io/distroless/base-debian11:latest
WORKDIR /app
COPY --from=builder /app/red-courier ./red-courier

USER 65532:65532

ENTRYPOINT ["./red-courier"]