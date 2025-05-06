# Red Courier

[![Go Report Card](https://goreportcard.com/badge/github.com/nathanbcrocker/red-courier)](https://goreportcard.com/report/github.com/nathanbcrocker/red-courier)
[![License](https://img.shields.io/github/license/nathanbcrocker/red-courier)](LICENSE)
[![Build Status](https://github.com/nathanbcrocker/red-courier/actions/workflows/ci.yml/badge.svg)](https://github.com/nathanbcrocker/red-courier/actions)
![Go Version](https://img.shields.io/badge/go-1.24-blue)

**Red Courier** is a lightweight Go-based service that synchronizes data from Postgres into Redis on a scheduled basis. It supports a variety of Redis data structures and incremental syncing using tracking columns. Built for e-commerce data like orders, products, or customer activity, Red Courier helps you populate Redis for caching, real-time analytics, or message processing.

## Features

* **Supports Redis data structures**:

    * `map` (HSET)
    * `list` (LPUSH)
    * `set` (SADD)
    * `sorted_set` (ZADD)
    * `stream` (XADD)
* **Incremental syncing** using a tracking column with `>` or `<` comparisons
* **Cron-style task scheduling**
* **Field-level mapping and aliasing** for flexible Redis key/value formats
* **Encapsulated Redis client** for maintainability and extensibility
* **LLM-compatible configuration guide** for easy generation of valid YAML
* Easy to containerize and deploy with Docker and GitHub Actions

## Example Use Case

Populate a Redis Stream from a Postgres `orders` table:

```yaml
tasks:
  - name: order_stream
    table: orders
    alias: order:stream
    structure: stream
    fields: [id, status, total_amount, created_at]
    column_map:
      total_amount: amount
    schedule: "@every 10s"
    tracking:
      column: created_at
      operator: ">"
      last_value_key: checkpoint:order_stream
```

## Configuration

Red Courier expects a `config.yaml` like the following:

```yaml
postgres:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  dbname: ecommerce
  sslmode: disable

redis:
  addr: localhost:6379
  password: ""
  db: 0

tasks:
  - name: recent_orders
    table: orders
    structure: map
    key: id
    value: status
    schedule: "@every 30s"
    tracking:
      column: updated_at
      operator: ">"
      last_value_key: checkpoint:recent_orders
```

## Task Options

Each task supports the following:

| Field        | Description                                                |
| ------------ | ---------------------------------------------------------- |
| `name`       | Logical name for the task                                  |
| `table`      | Postgres table to sync (schema-qualified or not)           |
| `alias`      | Optional key name override for Redis                       |
| `structure`  | One of: `map`, `list`, `set`, `sorted_set`, `stream`       |
| `key`        | Column name for Redis key (used in map/sorted\_set)        |
| `value`      | Column name for Redis value                                |
| `score`      | Column for sorted set score (only for `sorted_set`)        |
| `fields`     | List of fields to include (used for stream, list)          |
| `column_map` | Optional mapping from logical to physical Postgres columns |
| `schedule`   | Cron expression or `@every` syntax                         |
| `tracking`   | Optional object for incremental syncs (see below)          |

### Tracking Config

```yaml
tracking:
  column: updated_at           # Column used to track deltas
  operator: ">"                # Operator for comparison (">" or "<")
  last_value_key: checkpoint:orders  # Redis key to persist checkpoint value
```

## LLM Integration

Red Courier ships with a [configuration guide for LLMs](CONFIG_GUIDE.md) to help language models generate syntactically and semantically valid YAML. This is useful for:

* Prompting LLMs to produce full config files from plain language
* Generating template tasks for common e-commerce use cases
* Educating new users on how to structure and reason about task fields

To use it, paste the contents of `CONFIG_GUIDE.md` into your LLM prompt context and ask it to "generate a Red Courier task for syncing products into a stream" or similar.

## How It Works

1. Each task defines a table to pull from and a Redis structure to push to.
2. Column mapping (`column_map`) allows you to alias DB columns to logical names.
3. At the scheduled interval, the task:

    * Fetches relevant rows from Postgres
    * Transforms rows based on configuration
    * Publishes them to Redis using the configured structure

## Redis Structure Behavior

* **map**: Uses `HSET` to populate a Redis hash using `key` and `value` fields.
* **list**: Uses `LPUSH` to push values to the front of a Redis list.
* **set**: Uses `SADD` to add unique elements to a Redis set.
* **sorted\_set**: Uses `ZADD`, using `score` to order elements.
* **stream**: Uses `XADD`, with fields specified in `fields` and optionally aliased.

## Cron Syntax

Schedules follow the [robfig/cron](https://pkg.go.dev/github.com/robfig/cron) format:

* `@every 5m`: every 5 minutes
* `0 * * * *`: top of every hour

## Logging

Each task logs:

* Start and finish of execution
* Number of rows fetched and written
* Any errors during Postgres or Redis interaction

## Development

```bash
go run ./cmd/syncer
```

## License

MIT License. See [LICENSE](LICENSE) for details.

## TODO

* Support for additional filters per task
* Metrics endpoint (Prometheus-compatible)
* CLI/REST interface for runtime introspection
