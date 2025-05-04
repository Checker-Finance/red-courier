# Red Courier

[![Go Report Card](https://goreportcard.com/badge/github.com/nathanbcrocker/red-courier)](https://goreportcard.com/report/github.com/nathanbcrocker/red-courier)
[![License](https://img.shields.io/github/license/nathanbcrocker/red-courier)](LICENSE)
[![Build Status](https://github.com/nathanbcrocker/red-courier/actions/workflows/ci.yml/badge.svg)](https://github.com/nathanbcrocker/red-courier/actions)
![Go Version](https://img.shields.io/badge/go-1.21-blue)

**Red Courier** is a Go-based utility that automatically synchronizes data from a PostgreSQL database to Redis on a configurable schedule. It supports a variety of Redis data structures and allows you to alias both tasks and database columns for clean, consistent Redis key organization.

## Features

- Schedule-based syncing from Postgres to Redis
- Support for Redis structures:
    - Hash maps
    - Lists
    - Sets
    - Sorted Sets
    - Streams
- Per-task cron scheduling
- Alias support for:
    - Task names
    - Redis keys
    - Database column names
- Flexible YAML configuration
- Minimal dependencies and fast execution

## Installation

```bash
git clone https://github.com/yourorg/red-courier.git
cd red-courier
go build -o red-courier ./cmd/syncer
```

## Configuration

Create a `config.yaml` file in your project root:

```yaml
postgres:
  host: localhost
  port: 5432
  user: myuser
  password: mypass
  dbname: ecommerce
  sslmode: disable

redis:
  addr: localhost:6379
  password: ""
  db: 0

tasks:
  - name: sync_customer_map
    table: customers
    alias: customer_map
    structure: map
    key: customer_id
    value: display_name
    schedule: "@every 5m"

  - name: sync_order_stream
    table: orders
    alias: order_stream
    structure: stream
    key_prefix: order_stream
    fields: [order_id, amount, created_at]
    column_map:
      order_id: id
      amount: total
      created_at: timestamp
    schedule: "0 * * * *"
```

## How It Works

1. Each task defines a table to pull from and a Redis structure to push to.
2. Column mapping (`column_map`) allows you to alias DB columns to logical names.
3. At the scheduled interval, the task:
    - Fetches relevant rows from Postgres
    - Transforms rows based on configuration
    - Publishes them to Redis using the configured structure

## Running the App

```bash
./red-courier
```

The application will start a scheduler for each task and begin syncing at the defined interval.

## Redis Structure Behavior

- **map**: Uses `HSET` to populate a Redis hash using `key` and `value` fields.
- **list**: Uses `LPUSH` to push values to the front of a Redis list.
- **set**: Uses `SADD` to add unique elements to a Redis set.
- **sorted_set**: Uses `ZADD`, using `score` to order elements.
- **stream**: Uses `XADD`, with fields specified in `fields` and optionally aliased.

## Cron Syntax

Schedules follow the [robfig/cron](https://pkg.go.dev/github.com/robfig/cron) format:

- `@every 5m`: every 5 minutes
- `0 * * * *`: top of every hour

## Logging

Each task logs:
- Start and finish of execution
- Number of rows fetched and written
- Any errors during Postgres or Redis interaction

## Development

```bash
go run ./cmd/syncer
```

## License

MIT License. See [LICENSE](LICENSE) for details.

## TODO

- Support for incremental updates (e.g., last_updated)
- Metrics endpoint (Prometheus-compatible)
- Optional filtering conditions per task
- CLI/REST interface for runtime introspection
