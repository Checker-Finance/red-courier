
# CONFIG_GUIDE.md

This guide provides detailed instructions and examples for configuring Red Courier, a tool that syncs data from Postgres to Redis using scheduled tasks.

This document is optimized to help both humans and large language models (LLMs) generate and validate correct `config.yaml` files.

---

## 🔧 Top-Level Structure

A valid `config.yaml` consists of three sections:

```yaml
postgres:
  host: ...
  port: ...
  user: ...
  password: ...
  dbname: ...
  sslmode: ...

redis:
  addr: ...
  password: ...
  db: ...

tasks:
  - name: ...
    ...
```

---

## 🛠 postgres

Defines how to connect to your PostgreSQL database:

| Key       | Type   | Required | Example          |
|-----------|--------|----------|------------------|
| `host`    | string | ✅        | `"localhost"`     |
| `port`    | int    | ✅        | `5432`            |
| `user`    | string | ✅        | `"postgres"`      |
| `password`| string | ✅        | `"secret"`        |
| `dbname`  | string | ✅        | `"ecommerce"`     |
| `sslmode` | string | ❌        | `"disable"`       |

---

## 🔌 redis

Defines how to connect to Redis:

| Key       | Type   | Required | Example         |
|-----------|--------|----------|-----------------|
| `addr`    | string | ✅        | `"localhost:6379"` |
| `password`| string | ❌        | `""`              |
| `db`      | int    | ✅        | `0`               |

---

## 📦 tasks

A list of one or more task configurations. Each task defines how to query a Postgres table and write the results to Redis.

### Common Fields

| Key         | Type     | Required | Description |
|--------------|----------|----------|-------------|
| `name`       | string   | ✅        | Logical name for this sync task |
| `table`      | string   | ✅        | Postgres table or schema-qualified table (`schema.table`) |
| `alias`      | string   | ❌        | Override the Redis key prefix |
| `structure`  | string   | ✅        | One of: `map`, `list`, `set`, `sorted_set`, `stream` |
| `key`        | string   | ✅ for `map` and `sorted_set` | Postgres column to use as Redis key or member |
| `value`      | string   | ✅ for `map` | Postgres column to use as Redis value |
| `score`      | string   | ✅ for `sorted_set` | Column to use as Redis score |
| `fields`     | list     | ✅ for `stream`, `list`, `set` | List of fields to extract and write |
| `column_map` | object   | ❌        | Map of logical field name → DB column name |
| `schedule`   | string   | ✅        | Cron expression or `@every 10s` style syntax |
| `tracking`   | object   | ❌        | See below for delta sync support |

---

## ⏱ `tracking`

Optional block for incremental sync. If present, Red Courier will:
- Only fetch rows where the tracking column is newer/older than the last checkpoint.
- Store that checkpoint in Redis.

### Fields

| Key             | Type   | Required | Description |
|------------------|--------|----------|-------------|
| `column`         | string | ✅        | DB column used for delta tracking |
| `operator`       | string | ✅        | Either `">"` or `"<"` |
| `last_value_key` | string | ✅        | Redis key to persist the last checkpoint |

---

## ✅ Examples

### Example 1: Streaming New Orders

```yaml
tasks:
  - name: stream_orders
    table: public.orders
    alias: order_stream
    structure: stream
    fields: [id, status, amount, created_at]
    column_map:
      amount: total_amount
    schedule: "@every 15s"
    tracking:
      column: created_at
      operator: ">"
      last_value_key: checkpoint:order_stream
```

### Example 2: Caching Clients in a Redis Hash

```yaml
tasks:
  - name: sync_clients
    table: client_management.t_client
    alias: clients
    structure: map
    key: s_id_client
    value: s_name
    schedule: "@every 5m"
```

---

## 🤖 Prompt Pattern for LLMs

> You are configuring a Red Courier task.  
> Output valid YAML where:
> - You pull data from a table called `transactions`
> - You use a stream structure
> - You want to include fields: `id`, `user_id`, `amount`, `created_at`
> - You only want new rows based on `created_at`

---

## 🔍 Validation Notes

- Every task must declare a `structure`.
- If `structure: map` or `structure: sorted_set`, then `key` is required.
- `score` is only required for `sorted_set`.
- If `tracking` is used, `last_value_key` must be unique per task.

---

## 🔗 Links

- [Main README](./README.md)
- [Red Courier GitHub](https://github.com/Checker-Finance/red-courier)

---

This guide is designed to help humans and LLMs generate well-structured Red Courier configs that are production-ready.
