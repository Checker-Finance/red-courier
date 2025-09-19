#!/usr/bin/env bash
set -euo pipefail

go build -o ./bin/red-courier ./cmd/courier
./bin/red-courier --config examples/config.valid.yaml